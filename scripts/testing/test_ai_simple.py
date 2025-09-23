#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Simple AI Service Test
"""

import requests
import time

def test_ai_service():
    """测试AI服务"""
    print("=" * 50)
    print("AI Service Test")
    print("=" * 50)
    
    # 测试健康检查
    print("1. Testing AI service health check...")
    try:
        response = requests.get("http://localhost:5001/health", timeout=5)
        if response.status_code == 200:
            print("   ✅ AI服务健康检查通过")
            print(f"   响应: {response.json()}")
        else:
            print(f"   ❌ AI服务健康检查失败: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("   ❌ AI服务连接失败 - 服务可能未启动")
    except Exception as e:
        print(f"   ❌ AI服务测试错误: {e}")
    
    # 测试根端点
    print("\n2. Testing AI service root endpoint...")
    try:
        response = requests.get("http://localhost:5001/", timeout=5)
        if response.status_code == 200:
            print("   ✅ AI服务根端点正常")
            print(f"   响应: {response.json()}")
        else:
            print(f"   ❌ AI服务根端点失败: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("   ❌ AI服务连接失败 - 服务可能未启动")
    except Exception as e:
        print(f"   ❌ AI服务测试错误: {e}")
    
    # 测试检测端点
    print("\n3. Testing AI detection endpoint...")
    try:
        test_data = {
            "detection_id": "test_001",
            "detection_type": "voice",
            "audio_data": "base64_encoded_audio_data_here"
        }
        response = requests.post("http://localhost:5001/detect", json=test_data, timeout=10)
        if response.status_code == 200:
            print("   ✅ AI检测服务正常")
            print(f"   响应: {response.json()}")
        else:
            print(f"   ❌ AI检测服务失败: {response.status_code}")
            print(f"   错误: {response.text}")
    except requests.exceptions.ConnectionError:
        print("   ❌ AI服务连接失败 - 服务可能未启动")
    except Exception as e:
        print(f"   ❌ AI服务测试错误: {e}")

if __name__ == "__main__":
    test_ai_service() 