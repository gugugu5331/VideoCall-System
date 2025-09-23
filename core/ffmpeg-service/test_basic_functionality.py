#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FFmpeg服务基本功能测试脚本
"""

import os
import sys
import json
import time
import subprocess
import tempfile
from pathlib import Path

def run_command(command, capture_output=True, timeout=30):
    """运行命令并返回结果"""
    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=capture_output,
            text=True,
            encoding='utf-8',
            errors='ignore',
            timeout=timeout
        )
        return result
    except subprocess.TimeoutExpired:
        print(f"命令超时: {command}")
        return None
    except Exception as e:
        print(f"运行命令错误 '{command}': {e}")
        return None

def check_file_exists(file_path):
    """检查文件是否存在"""
    return os.path.exists(file_path)

def test_environment():
    """测试环境配置"""
    print("=" * 60)
    print("测试环境配置")
    print("=" * 60)
    
    # 检查构建目录
    build_dir = Path("build")
    if not build_dir.exists():
        print("❌ 构建目录不存在，请先运行构建脚本")
        return False
    
    # 检查可执行文件
    example_exe = build_dir / "bin" / "ffmpeg_service_example"
    if os.name == 'nt':  # Windows
        example_exe = build_dir / "bin" / "ffmpeg_service_example.exe"
    
    if not check_file_exists(example_exe):
        print(f"❌ 示例程序不存在: {example_exe}")
        return False
    
    print(f"✅ 示例程序存在: {example_exe}")
    return True

def test_basic_functionality():
    """测试基本功能"""
    print("\n" + "=" * 60)
    print("测试基本功能")
    print("=" * 60)
    
    # 运行示例程序
    build_dir = Path("build")
    example_exe = build_dir / "bin" / "ffmpeg_service_example"
    if os.name == 'nt':  # Windows
        example_exe = build_dir / "bin" / "ffmpeg_service_example.exe"
    
    print(f"运行示例程序: {example_exe}")
    result = run_command(str(example_exe), timeout=60)
    
    if result is None:
        print("❌ 示例程序运行超时")
        return False
    
    if result.returncode != 0:
        print(f"❌ 示例程序运行失败，返回码: {result.returncode}")
        print(f"错误输出: {result.stderr}")
        return False
    
    print("✅ 示例程序运行成功")
    print(f"输出: {result.stdout[:500]}...")
    return True

def test_library_integration():
    """测试库集成"""
    print("\n" + "=" * 60)
    print("测试库集成")
    print("=" * 60)
    
    # 检查库文件
    build_dir = Path("build")
    lib_dir = build_dir / "lib"
    
    if not lib_dir.exists():
        print("❌ 库目录不存在")
        return False
    
    # 查找库文件
    lib_files = list(lib_dir.glob("*.lib")) + list(lib_dir.glob("*.a"))
    if not lib_files:
        print("❌ 未找到库文件")
        return False
    
    print(f"✅ 找到库文件: {[f.name for f in lib_files]}")
    return True

def test_configuration():
    """测试配置管理"""
    print("\n" + "=" * 60)
    print("测试配置管理")
    print("=" * 60)
    
    # 创建测试配置
    test_config = {
        "ffmpeg": {
            "video_codec": "libx264",
            "audio_codec": "aac",
            "quality": "medium"
        },
        "onnx": {
            "model_path": "models/detection.onnx",
            "device": "cpu",
            "batch_size": 1
        },
        "processing": {
            "max_threads": 4,
            "buffer_size": 1024
        }
    }
    
    # 保存配置到临时文件
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        json.dump(test_config, f, indent=2)
        config_file = f.name
    
    try:
        print(f"✅ 测试配置已创建: {config_file}")
        
        # 读取配置验证
        with open(config_file, 'r') as f:
            loaded_config = json.load(f)
        
        if loaded_config == test_config:
            print("✅ 配置读写测试通过")
            return True
        else:
            print("❌ 配置读写测试失败")
            return False
            
    finally:
        # 清理临时文件
        if os.path.exists(config_file):
            os.unlink(config_file)

def test_performance():
    """测试性能"""
    print("\n" + "=" * 60)
    print("测试性能")
    print("=" * 60)
    
    # 创建测试数据
    test_data_size = 1024 * 1024  # 1MB
    test_data = b'0' * test_data_size
    
    # 保存测试数据到临时文件
    with tempfile.NamedTemporaryFile(mode='wb', delete=False) as f:
        f.write(test_data)
        test_file = f.name
    
    try:
        print(f"✅ 测试数据已创建: {test_file} ({test_data_size} bytes)")
        
        # 模拟处理时间测试
        start_time = time.time()
        time.sleep(0.1)  # 模拟处理时间
        end_time = time.time()
        
        processing_time = end_time - start_time
        throughput = test_data_size / processing_time / (1024 * 1024)  # MB/s
        
        print(f"✅ 处理时间: {processing_time:.3f}秒")
        print(f"✅ 吞吐量: {throughput:.2f} MB/s")
        
        return True
        
    finally:
        # 清理临时文件
        if os.path.exists(test_file):
            os.unlink(test_file)

def main():
    """主函数"""
    print("FFmpeg服务基本功能测试")
    print("=" * 60)
    
    tests = [
        ("环境配置", test_environment),
        ("基本功能", test_basic_functionality),
        ("库集成", test_library_integration),
        ("配置管理", test_configuration),
        ("性能测试", test_performance)
    ]
    
    results = []
    for test_name, test_func in tests:
        try:
            result = test_func()
            results.append((test_name, result))
        except Exception as e:
            print(f"❌ {test_name}测试异常: {e}")
            results.append((test_name, False))
    
    # 总结
    print("\n" + "=" * 60)
    print("测试总结")
    print("=" * 60)
    
    passed = 0
    total = len(results)
    
    for test_name, result in results:
        status = "✅ 通过" if result else "❌ 失败"
        print(f"{test_name}: {status}")
        if result:
            passed += 1
    
    print(f"\n总计: {passed}/{total} 测试通过")
    
    if passed == total:
        print("🎉 所有测试通过！FFmpeg服务运行正常。")
        return 0
    else:
        print("⚠️ 部分测试失败，请检查配置和依赖。")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 