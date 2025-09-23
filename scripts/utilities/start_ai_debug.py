#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
AI Service Debug Startup Script
"""

import sys
import os
import uvicorn
from pathlib import Path

# 添加AI服务目录到Python路径
ai_service_path = Path(__file__).parent / "ai-service"
sys.path.insert(0, str(ai_service_path))

def main():
    """启动AI服务"""
    print("=" * 60)
    print("AI Service Debug Startup")
    print("=" * 60)
    
    # 检查AI服务目录
    if not ai_service_path.exists():
        print(f"❌ AI服务目录不存在: {ai_service_path}")
        return 1
    
    print(f"✅ AI服务目录: {ai_service_path}")
    
    # 检查必要文件
    main_py = ai_service_path / "main.py"
    if not main_py.exists():
        print(f"❌ main.py不存在: {main_py}")
        return 1
    
    print(f"✅ main.py存在: {main_py}")
    
    # 检查依赖
    try:
        import fastapi
        import uvicorn
        import redis
        import numpy
        print("✅ 主要依赖包已安装")
    except ImportError as e:
        print(f"❌ 依赖包缺失: {e}")
        return 1
    
    # 切换到AI服务目录
    os.chdir(ai_service_path)
    print(f"✅ 切换到目录: {os.getcwd()}")
    
    # 启动服务
    print("\n启动AI服务...")
    print("服务地址: http://localhost:5000")
    print("按 Ctrl+C 停止服务")
    print("-" * 60)
    
    try:
        uvicorn.run(
            "main:app",
            host="0.0.0.0",
            port=5000,
            reload=True,
            log_level="info"
        )
    except KeyboardInterrupt:
        print("\nAI服务已停止")
    except Exception as e:
        print(f"❌ 启动失败: {e}")
        return 1
    
    return 0

if __name__ == "__main__":
    sys.exit(main()) 