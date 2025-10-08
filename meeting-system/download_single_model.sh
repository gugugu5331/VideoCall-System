#!/bin/bash

# 下载单个模型
MODEL_ID=$1
MODEL_PATH=$2
MODEL_NAME=$3

echo "========================================"
echo "下载模型: $MODEL_NAME"
echo "模型ID: $MODEL_ID"
echo "保存路径: $MODEL_PATH"
echo "========================================"

python3 << EOF
from huggingface_hub import snapshot_download
import os

try:
    os.makedirs("$MODEL_PATH", exist_ok=True)
    snapshot_download(
        repo_id="$MODEL_ID",
        local_dir="$MODEL_PATH",
        local_dir_use_symlinks=False
    )
    print("✅ $MODEL_NAME 下载完成")
    print("文件大小:")
    os.system("du -sh $MODEL_PATH")
except Exception as e:
    print(f"❌ 下载失败: {e}")
    import traceback
    traceback.print_exc()
    exit(1)
EOF

