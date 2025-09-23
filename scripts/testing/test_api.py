#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import requests
import json
import time

# 配置
BASE_URL = "http://localhost:8000"
AI_URL = "http://localhost:5001"

def test_health_check():
    """测试健康检查"""
    print("测试健康检查...")
    try:
        response = requests.get(f"{BASE_URL}/health")
        if response.status_code == 200:
            print("✓ 后端服务健康检查通过")
            return True
        else:
            print(f"✗ 后端服务健康检查失败: {response.status_code}")
            return False
    except Exception as e:
        print(f"✗ 后端服务连接失败: {e}")
        return False

def test_ai_health_check():
    """测试AI服务健康检查"""
    print("测试AI服务健康检查...")
    try:
        response = requests.get(f"{AI_URL}/health")
        if response.status_code == 200:
            print("✓ AI服务健康检查通过")
            return True
        else:
            print(f"✗ AI服务健康检查失败: {response.status_code}")
            return False
    except Exception as e:
        print(f"✗ AI服务连接失败: {e}")
        return False

def test_user_registration():
    """测试用户注册"""
    print("测试用户注册...")
    try:
        data = {
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123",
            "full_name": "测试用户"
        }
        response = requests.post(f"{BASE_URL}/api/v1/auth/register", json=data)
        if response.status_code == 201:
            print("✓ 用户注册成功")
            return True
        elif response.status_code == 409:
            print("✓ 用户已存在（预期结果）")
            return True
        else:
            print(f"✗ 用户注册失败: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ 用户注册请求失败: {e}")
        return False

def test_user_login():
    """测试用户登录"""
    print("测试用户登录...")
    try:
        data = {
            "username": "testuser",
            "password": "password123"
        }
        response = requests.post(f"{BASE_URL}/api/v1/auth/login", json=data)
        if response.status_code == 200:
            result = response.json()
            if "token" in result:
                print("✓ 用户登录成功")
                return result["token"]
            else:
                print("✗ 登录响应中没有token")
                return None
        else:
            print(f"✗ 用户登录失败: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"✗ 用户登录请求失败: {e}")
        return None

def test_protected_endpoint(token):
    """测试受保护的端点"""
    print("测试受保护的端点...")
    try:
        headers = {"Authorization": f"Bearer {token}"}
        response = requests.get(f"{BASE_URL}/api/v1/user/profile", headers=headers)
        if response.status_code == 200:
            print("✓ 受保护端点访问成功")
            return True
        else:
            print(f"✗ 受保护端点访问失败: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ 受保护端点请求失败: {e}")
        return False

def test_ai_detection():
    """测试AI检测服务"""
    print("测试AI检测服务...")
    try:
        data = {
            "detection_id": "test-detection-001",
            "detection_type": "voice_spoofing",
            "call_id": "test-call-001",
            "audio_data": "dGVzdCBhdWRpbyBkYXRh",  # base64编码的测试数据
            "metadata": {"test": True}
        }
        response = requests.post(f"{AI_URL}/detect", json=data)
        if response.status_code == 200:
            result = response.json()
            print("✓ AI检测服务正常")
            print(f"  检测结果: 风险评分={result.get('risk_score', 'N/A')}, 置信度={result.get('confidence', 'N/A')}")
            return True
        else:
            print(f"✗ AI检测服务失败: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ AI检测请求失败: {e}")
        return False

def main():
    """主测试函数"""
    print("==========================================")
    print("音视频通话系统 - API测试")
    print("==========================================")
    
    # 等待服务启动
    print("等待服务启动...")
    time.sleep(5)
    
    tests = [
        ("后端健康检查", test_health_check),
        ("AI服务健康检查", test_ai_health_check),
        ("用户注册", test_user_registration),
        ("AI检测服务", test_ai_detection),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n--- {test_name} ---")
        if test_func():
            passed += 1
        time.sleep(1)
    
    # 测试登录和受保护端点
    print(f"\n--- 用户登录 ---")
    token = test_user_login()
    if token:
        passed += 1
        print(f"\n--- 受保护端点测试 ---")
        if test_protected_endpoint(token):
            passed += 1
        total += 1
    total += 1
    
    print("\n==========================================")
    print(f"测试完成: {passed}/{total} 通过")
    print("==========================================")
    
    if passed == total:
        print("🎉 所有测试通过！系统运行正常。")
    else:
        print("⚠️  部分测试失败，请检查服务状态。")

if __name__ == "__main__":
    main() 