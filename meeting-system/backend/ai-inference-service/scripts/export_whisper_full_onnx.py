#!/usr/bin/env python3
"""
完整导出 Whisper 模型为 ONNX（包括 Encoder 和 Decoder）

解决 scaled_dot_product_attention 的 is_causal 参数问题
通过 monkey-patch 修改 Whisper 的 attention 实现
"""

import torch
import torch.nn.functional as F
import onnx
import numpy as np
from pathlib import Path
import sys


def patch_whisper_for_onnx():
    """修改 Whisper 模型以兼容 ONNX 导出"""
    import whisper.model as whisper_model
    
    # 保存原始的 qkv_attention 方法
    original_qkv_attention = whisper_model.MultiHeadAttention.qkv_attention
    
    def onnx_compatible_qkv_attention(self, q, k, v, mask=None):
        """ONNX 兼容的 QKV attention"""
        n_batch, n_ctx, n_state = q.shape
        scale = (n_state // self.n_head) ** -0.25
        q = q.view(*q.shape[:2], self.n_head, -1).permute(0, 2, 1, 3) * scale
        k = k.view(*k.shape[:2], self.n_head, -1).permute(0, 2, 3, 1) * scale
        v = v.view(*v.shape[:2], self.n_head, -1).permute(0, 2, 1, 3)

        # 使用传统的 attention 计算，避免 scaled_dot_product_attention
        qk = q @ k
        if mask is not None:
            qk = qk + mask[:n_ctx, :n_ctx]
        qk = qk.float()

        w = F.softmax(qk, dim=-1).to(q.dtype)
        out = (w @ v).permute(0, 2, 1, 3).flatten(start_dim=2)
        return out, qk.detach()
    
    # 替换方法
    whisper_model.MultiHeadAttention.qkv_attention = onnx_compatible_qkv_attention
    
    print("✅ Whisper 模型已修改为 ONNX 兼容模式")
    
    return original_qkv_attention


def export_whisper_encoder():
    """导出 Whisper Encoder"""
    print("=" * 80)
    print("🎯 导出 Whisper Encoder 为 ONNX")
    print("=" * 80)
    print()
    
    try:
        import whisper
        
        model_size = "base"
        
        print(f"📥 加载 Whisper 模型: {model_size}")
        model = whisper.load_model(model_size, device="cpu")
        model = model.cpu()
        model.eval()
        
        print(f"✅ 模型加载成功")
        print()
        
        # 获取配置
        n_mels = 80
        n_audio_ctx = model.dims.n_audio_ctx
        mel_length = n_audio_ctx * 2
        
        print(f"📊 Encoder 配置:")
        print(f"   n_mels: {n_mels}")
        print(f"   n_audio_ctx: {n_audio_ctx}")
        print(f"   mel_length: {mel_length}")
        print()
        
        # 导出 Encoder
        print(f"💾 导出 Encoder 为 ONNX...")
        
        dummy_mel = torch.randn(1, n_mels, mel_length).cpu()
        encoder_output_path = Path("/work/models/whisper-encoder.onnx")
        
        with torch.no_grad():
            torch.onnx.export(
                model.encoder,
                dummy_mel,
                str(encoder_output_path),
                export_params=True,
                opset_version=14,
                do_constant_folding=True,
                input_names=['mel'],
                output_names=['encoder_output'],
                dynamic_axes={
                    'mel': {0: 'batch_size'},
                    'encoder_output': {0: 'batch_size'}
                }
            )
        
        print(f"✅ Encoder ONNX 导出成功: {encoder_output_path}")
        print(f"   文件大小: {encoder_output_path.stat().st_size / 1024 / 1024:.2f} MB")
        print()
        
        return True
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return False


def export_whisper_decoder():
    """导出 Whisper Decoder（完整版本）"""
    print("=" * 80)
    print("🎯 导出 Whisper Decoder 为 ONNX（完整版本）")
    print("=" * 80)
    print()
    
    try:
        import whisper
        
        # Patch Whisper for ONNX compatibility
        original_qkv_attention = patch_whisper_for_onnx()
        
        model_size = "base"
        
        print(f"📥 加载 Whisper 模型: {model_size}")
        model = whisper.load_model(model_size, device="cpu")
        model = model.cpu()
        model.eval()
        
        print(f"✅ 模型加载成功")
        print()
        
        # 获取配置
        n_audio_ctx = model.dims.n_audio_ctx
        n_audio_state = model.dims.n_audio_state
        n_text_ctx = model.dims.n_text_ctx
        
        print(f"📊 Decoder 配置:")
        print(f"   n_audio_ctx: {n_audio_ctx}")
        print(f"   n_audio_state: {n_audio_state}")
        print(f"   n_text_ctx: {n_text_ctx}")
        print()
        
        # 创建 Decoder wrapper（单步解码）
        class WhisperDecoderOneStep(torch.nn.Module):
            def __init__(self, decoder, n_text_ctx):
                super().__init__()
                self.decoder = decoder
                self.n_text_ctx = n_text_ctx
                
                # 创建因果掩码
                self.register_buffer(
                    "mask",
                    torch.empty(n_text_ctx, n_text_ctx).fill_(-np.inf).triu_(1)
                )
            
            def forward(self, tokens, encoder_output):
                """
                单步解码
                tokens: (batch, seq_len) - 当前已生成的 token 序列
                encoder_output: (batch, n_audio_ctx, n_audio_state) - encoder 输出
                返回: (batch, seq_len, n_vocab) - 每个位置的 logits
                """
                # 使用 decoder 的 forward 方法
                # 注意：需要传递正确的 mask
                x = self.decoder(tokens, encoder_output)
                return x
        
        decoder_wrapper = WhisperDecoderOneStep(model.decoder, n_text_ctx)
        decoder_wrapper.eval()
        
        print(f"💾 导出 Decoder 为 ONNX...")
        
        # 创建示例输入
        batch_size = 1
        seq_len = 10  # 当前序列长度
        
        # 示例 tokens: [<|startoftranscript|>, <|zh|>, <|transcribe|>, <|notimestamps|>, ...]
        dummy_tokens = torch.tensor([[50258, 50260, 50359, 50363] + [0] * (seq_len - 4)], dtype=torch.long)
        dummy_encoder_output = torch.randn(batch_size, n_audio_ctx, n_audio_state)
        
        decoder_output_path = Path("/work/models/whisper-decoder.onnx")
        
        with torch.no_grad():
            torch.onnx.export(
                decoder_wrapper,
                (dummy_tokens, dummy_encoder_output),
                str(decoder_output_path),
                export_params=True,
                opset_version=14,
                do_constant_folding=True,
                input_names=['tokens', 'encoder_output'],
                output_names=['logits'],
                dynamic_axes={
                    'tokens': {0: 'batch_size', 1: 'seq_len'},
                    'encoder_output': {0: 'batch_size'},
                    'logits': {0: 'batch_size', 1: 'seq_len'}
                }
            )
        
        print(f"✅ Decoder ONNX 导出成功: {decoder_output_path}")
        print(f"   文件大小: {decoder_output_path.stat().st_size / 1024 / 1024:.2f} MB")
        print()
        
        # 验证模型
        print(f"🔍 验证 ONNX 模型...")
        onnx_model = onnx.load(str(decoder_output_path))
        onnx.checker.check_model(onnx_model)
        print(f"✅ ONNX 模型验证通过")
        print()
        
        # 测试推理
        print(f"🧪 测试 ONNX 推理...")
        import onnxruntime as ort
        
        session = ort.InferenceSession(str(decoder_output_path))
        
        tokens_data = dummy_tokens.numpy().astype(np.int64)
        encoder_output_data = dummy_encoder_output.numpy().astype(np.float32)
        
        outputs = session.run(None, {
            'tokens': tokens_data,
            'encoder_output': encoder_output_data
        })
        
        print(f"✅ ONNX 推理成功")
        print(f"   tokens 形状: {tokens_data.shape}")
        print(f"   encoder_output 形状: {encoder_output_data.shape}")
        print(f"   logits 形状: {outputs[0].shape}")
        print()
        
        # 测试贪婪解码
        print(f"🧪 测试贪婪解码...")
        logits = outputs[0]  # (batch, seq_len, n_vocab)
        
        # 获取最后一个 token 的预测
        last_logits = logits[0, -1, :]  # (n_vocab,)
        predicted_token = np.argmax(last_logits)
        
        print(f"   最后一个位置的预测 token: {predicted_token}")
        print(f"   概率: {np.exp(last_logits[predicted_token]) / np.sum(np.exp(last_logits)):.4f}")
        print()
        
        print(f"=" * 80)
        print(f"🎉 Whisper Decoder ONNX 导出完成！")
        print(f"=" * 80)
        print()
        
        return True
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return False


def save_model_info():
    """保存模型信息和配置"""
    print("=" * 80)
    print("🎯 保存模型信息和配置")
    print("=" * 80)
    print()
    
    try:
        import whisper
        import json
        
        model_size = "base"
        model = whisper.load_model(model_size, device="cpu")
        tokenizer = whisper.tokenizer.get_tokenizer(multilingual=True)
        
        # 保存完整的词汇表
        print("💾 保存词汇表...")
        vocab = {}
        for i in range(tokenizer.encoding.n_vocab):
            try:
                token = tokenizer.decode([i])
                vocab[i] = token
            except:
                vocab[i] = f"<token_{i}>"
        
        vocab_path = Path("/work/models/whisper_vocab.json")
        with open(vocab_path, 'w', encoding='utf-8') as f:
            json.dump(vocab, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 词汇表已保存: {vocab_path}")
        print(f"   词汇表大小: {len(vocab)}")
        print()
        
        # 保存特殊 token
        print("💾 保存特殊 token...")
        special_tokens = {
            "sot": int(tokenizer.sot),
            "eot": int(tokenizer.eot),
            "sot_prev": int(tokenizer.sot_prev),
            "no_speech": int(tokenizer.no_speech),
            "no_timestamps": int(tokenizer.no_timestamps),
            "timestamp_begin": int(tokenizer.timestamp_begin),
            "language_tokens": {
                "zh": int(tokenizer.sot + 1 + 50259 - tokenizer.sot),  # <|zh|>
                "en": int(tokenizer.sot + 1 + 50258 - tokenizer.sot),  # <|en|>
            },
            "task_tokens": {
                "transcribe": int(tokenizer.transcribe),
                "translate": int(tokenizer.translate),
            }
        }
        
        special_tokens_path = Path("/work/models/whisper_special_tokens.json")
        with open(special_tokens_path, 'w', encoding='utf-8') as f:
            json.dump(special_tokens, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 特殊 token 已保存: {special_tokens_path}")
        print()
        
        # 保存模型配置
        print("💾 保存模型配置...")
        model_config = {
            "model_size": model_size,
            "n_mels": 80,
            "n_audio_ctx": model.dims.n_audio_ctx,
            "mel_length": model.dims.n_audio_ctx * 2,
            "n_audio_state": model.dims.n_audio_state,
            "n_audio_head": model.dims.n_audio_head,
            "n_audio_layer": model.dims.n_audio_layer,
            "n_vocab": tokenizer.encoding.n_vocab,
            "n_text_ctx": model.dims.n_text_ctx,
            "n_text_state": model.dims.n_text_state,
            "n_text_head": model.dims.n_text_head,
            "n_text_layer": model.dims.n_text_layer,
        }
        
        config_path = Path("/work/models/whisper_config.json")
        with open(config_path, 'w', encoding='utf-8') as f:
            json.dump(model_config, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 模型配置已保存: {config_path}")
        print()
        
        print(f"=" * 80)
        print(f"🎉 模型信息保存完成！")
        print(f"=" * 80)
        print()
        
        return True
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="完整导出 Whisper 模型为 ONNX")
    parser.add_argument(
        "--component",
        type=str,
        choices=["encoder", "decoder", "all"],
        default="all",
        help="导出哪个组件"
    )
    
    args = parser.parse_args()
    
    success = True
    
    if args.component in ["encoder", "all"]:
        if not export_whisper_encoder():
            success = False
    
    if args.component in ["decoder", "all"]:
        if not export_whisper_decoder():
            success = False
    
    if args.component == "all":
        if not save_model_info():
            success = False
    
    sys.exit(0 if success else 1)

