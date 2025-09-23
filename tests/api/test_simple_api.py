#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简单API测试脚本
"""

import requests
import json
import time

# 配置
BASE_URL = "http://localhost:8000"

def test_health_check():
    """测试健康检查"""
    print("🔍 测试健康检查...")
    try:
        response = requests.get(f"{BASE_URL}/health", timeout=5)
        if response.status_code == 200:
            print("✅ 后端服务正常运行")
            print(f"   响应: {response.json()}")
            return True
        else:
            print(f"❌ 后端服务异常: {response.status_code}")
            return False
    except requests.exceptions.ConnectionError:
        print("❌ 无法连接到后端服务")
        print("   请确保后端服务正在运行在 http://localhost:8000")
        return False
    except Exception as e:
        print(f"❌ 连接失败: {e}")
        return False

def test_root_endpoint():
    """测试根端点"""
    print("\n🔍 测试根端点...")
    try:
        response = requests.get(f"{BASE_URL}/", timeout=5)
        if response.status_code == 200:
            print("✅ 根端点正常")
            print(f"   响应: {response.json()}")
            return True
        else:
            print(f"❌ 根端点异常: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ 根端点测试失败: {e}")
        return False

def test_swagger_docs():
    """测试Swagger文档"""
    print("\n🔍 测试Swagger文档...")
    try:
        response = requests.get(f"{BASE_URL}/swagger/index.html", timeout=5)
        if response.status_code == 200:
            print("✅ Swagger文档可访问")
            return True
        else:
            print(f"❌ Swagger文档异常: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Swagger文档测试失败: {e}")
        return False

def test_register_endpoint():
    """测试注册端点（不依赖数据库）"""
    print("\n🔍 测试注册端点...")
    try:
        test_user = {
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123",
            "full_name": "Test User"
        }
        response = requests.post(f"{BASE_URL}/api/v1/auth/register", 
                               json=test_user, timeout=5)
        print(f"   状态码: {response.status_code}")
        if response.status_code in [201, 409]:  # 成功或用户已存在
            print("✅ 注册端点响应正常")
            return True
        else:
            print(f"❌ 注册端点异常: {response.status_code}")
            if response.text:
                print(f"   错误信息: {response.text}")
            return False
    except Exception as e:
        print(f"❌ 注册端点测试失败: {e}")
        return False

def main():
    print("🚀 开始简单API测试")
    print(f"目标服务器: {BASE_URL}")
    print("=" * 50)
    
    tests = [
        test_health_check,
        test_root_endpoint,
        test_swagger_docs,
        test_register_endpoint
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        if test():
            passed += 1
        time.sleep(1)  # 短暂延迟
    
    print("\n" + "=" * 50)
    print(f"测试完成: {passed}/{total} 通过")
    
    if passed == total:
        print("🎉 所有测试通过！系统运行正常")
    else:
        print("⚠️  部分测试失败，请检查服务状态")
        
    print("\n💡 建议:")
    if passed < 2:
        print("   - 确保后端服务正在运行")
        print("   - 检查端口8000是否被占用")
        print("   - 运行: cd core/backend && go run main.go")
    elif passed < 4:
        print("   - 检查数据库连接")
        print("   - 确保PostgreSQL和Redis正在运行")
        print("   - 运行: docker-compose up -d postgres redis")

if __name__ == "__main__":
    main() 