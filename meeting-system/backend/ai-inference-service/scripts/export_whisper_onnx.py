#!/usr/bin/env python3
"""
正确导出 Whisper 模型为 ONNX 格式

解决位置编码维度不匹配的问题
"""

import torch
import onnx
import numpy as np
from pathlib import Path
import sys


def export_whisper_encoder_onnx():
    """导出 Whisper Encoder 为 ONNX"""
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
        
        # 获取 Whisper 的配置
        n_mels = 80
        n_audio_ctx = model.dims.n_audio_ctx  # 1500
        # Whisper 使用 2 个 conv1d 层，每个 stride=2
        # 所以输入 mel-spectrogram 长度需要是 n_audio_ctx * 2
        mel_length = n_audio_ctx * 2  # 3000

        print(f"📊 Whisper 配置:")
        print(f"   n_mels: {n_mels}")
        print(f"   n_audio_ctx: {n_audio_ctx}")
        print(f"   mel_length: {mel_length}")
        print(f"   n_audio_state: {model.dims.n_audio_state}")
        print()

        # 导出 Encoder（使用固定长度）
        print(f"💾 导出 Encoder 为 ONNX（固定长度）...")

        # 创建固定长度的输入
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
        
        # 验证模型
        print(f"🔍 验证 ONNX 模型...")
        onnx_model = onnx.load(str(encoder_output_path))
        onnx.checker.check_model(onnx_model)
        print(f"✅ ONNX 模型验证通过")
        print()
        
        # 测试推理
        print(f"🧪 测试 ONNX 推理...")
        import onnxruntime as ort

        session = ort.InferenceSession(str(encoder_output_path))
        input_data = np.random.randn(1, n_mels, mel_length).astype(np.float32)
        outputs = session.run(None, {'mel': input_data})
        
        print(f"✅ ONNX 推理成功")
        print(f"   输入形状: {input_data.shape}")
        print(f"   输出形状: {outputs[0].shape}")
        print()
        
        # 保存 tokenizer 和配置
        print(f"💾 保存 Tokenizer 和配置...")
        
        tokenizer = whisper.tokenizer.get_tokenizer(multilingual=True)
        
        # 保存完整的词汇表
        vocab = {}
        for i in range(tokenizer.encoding.n_vocab):
            try:
                token = tokenizer.decode([i])
                vocab[i] = token
            except:
                vocab[i] = f"<token_{i}>"
        
        # 保存为 JSON
        import json
        
        vocab_path = Path("/work/models/whisper_vocab.json")
        with open(vocab_path, 'w', encoding='utf-8') as f:
            json.dump(vocab, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 词汇表已保存: {vocab_path}")
        print(f"   词汇表大小: {len(vocab)}")
        print()
        
        # 保存特殊 token
        special_tokens = {
            "sot": int(tokenizer.sot),
            "eot": int(tokenizer.eot),
            "sot_prev": int(tokenizer.sot_prev),
            "no_speech": int(tokenizer.no_speech),
            "no_timestamps": int(tokenizer.no_timestamps),
            "timestamp_begin": int(tokenizer.timestamp_begin),
            "language_tokens": {
                "zh": int(tokenizer.sot + 1 + tokenizer.encoding.encode(" Chinese")[0]),
                "en": int(tokenizer.sot + 1 + tokenizer.encoding.encode(" English")[0]),
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
        model_config = {
            "model_size": model_size,
            "n_mels": n_mels,
            "n_audio_ctx": n_audio_ctx,
            "mel_length": mel_length,
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
        print(f"🎉 Whisper Encoder ONNX 导出完成！")
        print(f"=" * 80)
        print()
        
        return True
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return False


def export_whisper_decoder_onnx():
    """导出 Whisper Decoder 为 ONNX（简化版本）"""
    print("=" * 80)
    print("🎯 创建简化的 Whisper Decoder")
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
        
        # 创建一个简化的 Decoder wrapper
        class WhisperDecoderWrapper(torch.nn.Module):
            def __init__(self, decoder):
                super().__init__()
                self.decoder = decoder
            
            def forward(self, tokens, encoder_output):
                # tokens: (batch, seq_len)
                # encoder_output: (batch, n_audio_ctx, n_audio_state)
                return self.decoder(tokens, encoder_output)
        
        decoder_wrapper = WhisperDecoderWrapper(model.decoder)
        decoder_wrapper.eval()
        
        print(f"💾 导出 Decoder 为 ONNX...")
        
        # 创建示例输入
        batch_size = 1
        seq_len = 10
        n_audio_ctx = 1500
        n_audio_state = model.dims.n_audio_state
        
        dummy_tokens = torch.tensor([[50258, 50259, 50359, 50363] + [0] * (seq_len - 4)], dtype=torch.long)
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


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="导出 Whisper 模型为 ONNX")
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
        if not export_whisper_encoder_onnx():
            success = False
    
    if args.component in ["decoder", "all"]:
        if not export_whisper_decoder_onnx():
            success = False
    
    sys.exit(0 if success else 1)

