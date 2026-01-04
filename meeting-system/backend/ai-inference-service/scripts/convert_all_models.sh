#!/bin/bash
#
# 一键转换所有 AI 模型到 ONNX 格式
#
# 使用方法:
#   ./convert_all_models.sh [--simple]
#
# 选项:
#   --simple    使用简化模型（不下载大型预训练模型）
#

set -e  # 遇到错误立即退出

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="/work/models"

echo "================================================================================"
echo "🎯 AI 模型转换工具"
echo "================================================================================"
echo ""
echo "📂 输出目录: $OUTPUT_DIR"
echo "📂 脚本目录: $SCRIPT_DIR"
echo ""

# 检查参数
USE_SIMPLE=false
if [[ "$1" == "--simple" ]]; then
    USE_SIMPLE=true
    echo "⚙️ 模式: 简化模型（不下载预训练模型）"
else
    echo "⚙️ 模式: 完整模型（下载预训练模型）"
fi
echo ""

# 创建输出目录
echo "📁 创建输出目录..."
mkdir -p "$OUTPUT_DIR"
echo "✅ 输出目录已创建"
echo ""

# 检查 Python 环境
echo "🐍 检查 Python 环境..."
if ! command -v python3 &> /dev/null; then
    echo "❌ 错误: 未找到 python3"
    exit 1
fi

PYTHON_VERSION=$(python3 --version)
echo "✅ Python 版本: $PYTHON_VERSION"
echo ""

# 检查依赖
echo "📦 检查 Python 依赖..."
REQUIRED_PACKAGES=("torch" "onnx" "numpy")

for package in "${REQUIRED_PACKAGES[@]}"; do
    if python3 -c "import $package" 2>/dev/null; then
        echo "   ✅ $package"
    else
        echo "   ❌ $package (未安装)"
        echo ""
        echo "请安装缺失的依赖:"
        echo "   pip install torch onnx numpy onnxruntime"
        exit 1
    fi
done
echo ""

# 转换 ASR 模型
echo "================================================================================"
echo "1️⃣ 转换 ASR 模型 (语音识别)"
echo "================================================================================"
echo ""

if [ "$USE_SIMPLE" = true ]; then
    python3 "$SCRIPT_DIR/convert_whisper_to_onnx.py" \
        --simple \
        --output-dir "$OUTPUT_DIR"
else
    echo "⚠️ 注意: 下载 Whisper 模型可能需要几分钟..."
    echo ""
    python3 "$SCRIPT_DIR/convert_whisper_to_onnx.py" \
        --model-size base \
        --output-dir "$OUTPUT_DIR"
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ ASR 模型转换成功"
else
    echo ""
    echo "❌ ASR 模型转换失败"
    exit 1
fi
echo ""

# 转换 Emotion Detection 模型
echo "================================================================================"
echo "2️⃣ 转换 Emotion Detection 模型 (情感检测)"
echo "================================================================================"
echo ""

if [ "$USE_SIMPLE" = true ]; then
    python3 "$SCRIPT_DIR/convert_emotion_to_onnx.py" \
        --output-dir "$OUTPUT_DIR"
else
    echo "⚠️ 注意: 下载预训练模型可能需要几分钟..."
    echo ""
    python3 "$SCRIPT_DIR/convert_emotion_to_onnx.py" \
        --pretrained \
        --output-dir "$OUTPUT_DIR"
    
    # 如果预训练模型下载失败，回退到简化模型
    if [ $? -ne 0 ]; then
        echo ""
        echo "⚠️ 预训练模型下载失败，使用简化模型..."
        echo ""
        python3 "$SCRIPT_DIR/convert_emotion_to_onnx.py" \
            --output-dir "$OUTPUT_DIR"
    fi
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Emotion Detection 模型转换成功"
else
    echo ""
    echo "❌ Emotion Detection 模型转换失败"
    exit 1
fi
echo ""

# 转换 Synthesis Detection 模型
echo "================================================================================"
echo "3️⃣ 转换 Synthesis Detection 模型 (深度伪造检测)"
echo "================================================================================"
echo ""

python3 "$SCRIPT_DIR/convert_synthesis_to_onnx.py" \
    --output-dir "$OUTPUT_DIR"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Synthesis Detection 模型转换成功"
else
    echo ""
    echo "❌ Synthesis Detection 模型转换失败"
    exit 1
fi
echo ""

# 显示结果
echo "================================================================================"
echo "🎉 所有模型转换完成！"
echo "================================================================================"
echo ""
echo "📊 模型文件:"
ls -lh "$OUTPUT_DIR"/*.onnx
echo ""

echo "📈 模型大小统计:"
du -sh "$OUTPUT_DIR"
echo ""

echo "✅ 下一步:"
echo "   1. 将模型放入 Triton model repository"
echo "   2. 重启/热加载 Triton"
echo "   3. 测试真实模型推理"
echo ""
echo "================================================================================"
