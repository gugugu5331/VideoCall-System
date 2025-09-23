#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
系统状态检查脚本
"""

import requests
import json
import time

# 配置
BACKEND_URL = "http://localhost:8000"
AI_SERVICE_URL = "http://localhost:5000"

def check_backend_service():
    """检查后端服务"""
    print("🔍 检查后端服务...")
    try:
        response = requests.get(f"{BACKEND_URL}/health", timeout=5)
        if response.status_code == 200:
            print("✅ 后端服务正常运行")
            data = response.json()
            print(f"   状态: {data.get('status', 'unknown')}")
            print(f"   消息: {data.get('message', 'unknown')}")
            return True
        else:
            print(f"❌ 后端服务异常: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ 后端服务连接失败: {e}")
        return False

def check_ai_service():
    """检查AI服务"""
    print("\n🔍 检查AI服务...")
    try:
        response = requests.get(f"{AI_SERVICE_URL}/health", timeout=5)
        if response.status_code == 200:
            print("✅ AI服务正常运行")
            data = response.json()
            print(f"   状态: {data.get('status', 'unknown')}")
            print(f"   服务: {data.get('service', 'unknown')}")
            return True
        else:
            print(f"❌ AI服务异常: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ AI服务连接失败: {e}")
        return False

def test_backend_api():
    """测试后端API功能"""
    print("\n🔍 测试后端API功能...")
    
    # 测试根端点
    try:
        response = requests.get(f"{BACKEND_URL}/", timeout=5)
        if response.status_code == 200:
            print("✅ 后端根端点正常")
        else:
            print(f"❌ 后端根端点异常: {response.status_code}")
    except Exception as e:
        print(f"❌ 后端根端点测试失败: {e}")
    
    # 测试注册端点
    try:
        test_user = {
            "username": "test_status_user",
            "email": "test@status.com",
            "password": "password123",
            "full_name": "Test Status User"
        }
        response = requests.post(f"{BACKEND_URL}/api/v1/auth/register", json=test_user, timeout=5)
        if response.status_code in [201, 409]:  # 成功或用户已存在
            print("✅ 后端注册API正常")
        else:
            print(f"❌ 后端注册API异常: {response.status_code}")
    except Exception as e:
        print(f"❌ 后端注册API测试失败: {e}")

def test_ai_service_api():
    """测试AI服务API功能"""
    print("\n🔍 测试AI服务API功能...")
    
    # 测试根端点
    try:
        response = requests.get(f"{AI_SERVICE_URL}/", timeout=5)
        if response.status_code == 200:
            print("✅ AI服务根端点正常")
            data = response.json()
            print(f"   版本: {data.get('version', 'unknown')}")
        else:
            print(f"❌ AI服务根端点异常: {response.status_code}")
    except Exception as e:
        print(f"❌ AI服务根端点测试失败: {e}")
    
    # 测试检测端点
    try:
        test_request = {
            "detection_type": "voice",
            "audio_data": "test_data"
        }
        response = requests.post(f"{AI_SERVICE_URL}/detect", json=test_request, timeout=5)
        if response.status_code == 200:
            print("✅ AI服务检测API正常")
            data = response.json()
            print(f"   检测ID: {data.get('detection_id', 'unknown')}")
        else:
            print(f"❌ AI服务检测API异常: {response.status_code}")
    except Exception as e:
        print(f"❌ AI服务检测API测试失败: {e}")

def check_ports():
    """检查端口状态"""
    print("\n🔍 检查端口状态...")
    
    import subprocess
    import re
    
    try:
        # 检查端口8000
        result = subprocess.run(['netstat', '-an'], capture_output=True, text=True)
        if '8000' in result.stdout and 'LISTENING' in result.stdout:
            print("✅ 端口8000 (后端服务) 正在监听")
        else:
            print("❌ 端口8000 (后端服务) 未监听")
        
        # 检查端口5000
        if '5000' in result.stdout and 'LISTENING' in result.stdout:
            print("✅ 端口5000 (AI服务) 正在监听")
        else:
            print("❌ 端口5000 (AI服务) 未监听")
            
    except Exception as e:
        print(f"❌ 端口检查失败: {e}")

def main():
    print("🚀 开始系统状态检查")
    print("=" * 50)
    
    # 检查端口
    check_ports()
    
    # 检查服务
    backend_ok = check_backend_service()
    ai_ok = check_ai_service()
    
    # 测试API
    if backend_ok:
        test_backend_api()
    
    if ai_ok:
        test_ai_service_api()
    
    print("\n" + "=" * 50)
    print("系统状态总结:")
    print(f"后端服务: {'✅ 正常' if backend_ok else '❌ 异常'}")
    print(f"AI服务: {'✅ 正常' if ai_ok else '❌ 异常'}")
    
    if backend_ok and ai_ok:
        print("\n🎉 所有服务正常运行！")
        print("💡 前端应该可以正常工作了")
    else:
        print("\n⚠️  部分服务异常，请检查服务状态")
    
    print("\n服务地址:")
    print(f"后端服务: {BACKEND_URL}")
    print(f"AI服务: {AI_SERVICE_URL}")
    print(f"前端界面: {BACKEND_URL} (如果配置了静态文件)")

if __name__ == "__main__":
    main() 