#!/bin/bash
#
# 安装模型转换所需的 Python 依赖
#

set -e

echo "================================================================================"
echo "📦 安装 AI 模型转换依赖"
echo "================================================================================"
echo ""

# 检查 Python
echo "🐍 检查 Python..."
if ! command -v python3 &> /dev/null; then
    echo "❌ 错误: 未找到 python3"
    echo "请先安装 Python 3.8+"
    exit 1
fi

PYTHON_VERSION=$(python3 --version)
echo "✅ Python 版本: $PYTHON_VERSION"
echo ""

# 检查 pip
echo "📦 检查 pip..."
if ! command -v pip3 &> /dev/null; then
    echo "❌ 错误: 未找到 pip3"
    echo "请先安装 pip"
    exit 1
fi

PIP_VERSION=$(pip3 --version)
echo "✅ pip 版本: $PIP_VERSION"
echo ""

# 升级 pip
echo "⬆️ 升级 pip..."
pip3 install --upgrade pip
echo ""

# 安装基础依赖
echo "📦 安装基础依赖..."
echo ""

echo "1️⃣ 安装 PyTorch..."
pip3 install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu
echo ""

echo "2️⃣ 安装 ONNX..."
pip3 install onnx onnxruntime
echo ""

echo "3️⃣ 安装 NumPy..."
pip3 install numpy
echo ""

# 可选依赖
echo "================================================================================"
echo "📦 安装可选依赖（用于下载预训练模型）"
echo "================================================================================"
echo ""

read -p "是否安装 Whisper? (用于 ASR 模型) [y/N]: " install_whisper
if [[ "$install_whisper" =~ ^[Yy]$ ]]; then
    echo "4️⃣ 安装 OpenAI Whisper..."
    pip3 install openai-whisper
    echo ""
fi

read -p "是否安装 Transformers? (用于 Emotion Detection 模型) [y/N]: " install_transformers
if [[ "$install_transformers" =~ ^[Yy]$ ]]; then
    echo "5️⃣ 安装 Transformers..."
    pip3 install transformers
    echo ""
fi

echo "6️⃣ 安装音频处理库..."
pip3 install librosa soundfile
echo ""

# 验证安装
echo "================================================================================"
echo "✅ 验证安装"
echo "================================================================================"
echo ""

python3 << 'EOF'
import sys

packages = {
    "torch": "PyTorch",
    "onnx": "ONNX",
    "onnxruntime": "ONNX Runtime",
    "numpy": "NumPy",
}

optional_packages = {
    "whisper": "OpenAI Whisper",
    "transformers": "Transformers",
    "librosa": "Librosa",
}

print("📦 已安装的包:")
print()

for package, name in packages.items():
    try:
        __import__(package)
        print(f"   ✅ {name}")
    except ImportError:
        print(f"   ❌ {name} (未安装)")

print()
print("📦 可选包:")
print()

for package, name in optional_packages.items():
    try:
        __import__(package)
        print(f"   ✅ {name}")
    except ImportError:
        print(f"   ⚪ {name} (未安装)")

print()
EOF

echo "================================================================================"
echo "🎉 依赖安装完成！"
echo "================================================================================"
echo ""
echo "✅ 下一步:"
echo "   运行模型转换脚本:"
echo "   ./scripts/convert_all_models.sh --simple"
echo ""

