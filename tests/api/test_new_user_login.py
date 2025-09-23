#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试新注册用户登录
"""

import requests
import json

# 配置
BASE_URL = "http://localhost:8000"
API_BASE = f"{BASE_URL}/api/v1"

def test_new_user_login():
    """测试新注册用户的登录"""
    print("🔍 测试新注册用户登录")
    print("=" * 50)
    
    # 先注册一个新用户
    register_data = {
        "username": "newuser_test",
        "email": "newuser@test.com",
        "password": "password123",
        "full_name": "New Test User"
    }
    
    print("1. 注册新用户...")
    try:
        response = requests.post(f"{API_BASE}/auth/register", json=register_data)
        if response.status_code == 201:
            print("✅ 用户注册成功")
            user_data = response.json()
            print(f"   用户ID: {user_data['user']['id']}")
        else:
            print(f"❌ 注册失败: {response.status_code}")
            print(f"   错误: {response.text}")
            return False
    except Exception as e:
        print(f"❌ 注册异常: {e}")
        return False
    
    print("\n2. 立即登录新用户...")
    login_data = {
        "username": "newuser_test",
        "password": "password123"
    }
    
    try:
        response = requests.post(f"{API_BASE}/auth/login", json=login_data)
        print(f"   状态码: {response.status_code}")
        
        if response.status_code == 200:
            print("✅ 登录成功")
            login_result = response.json()
            print(f"   Token: {login_result['token'][:50]}...")
            print(f"   用户: {login_result['user']['username']}")
            return True
        else:
            print("❌ 登录失败")
            print(f"   错误: {response.text}")
            return False
    except Exception as e:
        print(f"❌ 登录异常: {e}")
        return False

def test_existing_user_login():
    """测试现有用户登录"""
    print("\n🔍 测试现有用户登录")
    print("=" * 50)
    
    # 测试alice用户
    login_data = {
        "username": "alice",
        "password": "password123"
    }
    
    try:
        response = requests.post(f"{API_BASE}/auth/login", json=login_data)
        print(f"   状态码: {response.status_code}")
        
        if response.status_code == 200:
            print("✅ alice登录成功")
            return True
        else:
            print("❌ alice登录失败")
            print(f"   错误: {response.text}")
            return False
    except Exception as e:
        print(f"❌ alice登录异常: {e}")
        return False

def main():
    print("🚀 开始用户登录测试")
    print(f"目标服务器: {BASE_URL}")
    print("=" * 60)
    
    # 测试新用户
    new_user_success = test_new_user_login()
    
    # 测试现有用户
    existing_user_success = test_existing_user_login()
    
    print("\n" + "=" * 60)
    print("测试总结:")
    print(f"新用户登录: {'✅ 成功' if new_user_success else '❌ 失败'}")
    print(f"现有用户登录: {'✅ 成功' if existing_user_success else '❌ 失败'}")
    
    if new_user_success and not existing_user_success:
        print("\n💡 分析:")
        print("新用户登录成功但现有用户失败，可能的原因:")
        print("1. 现有用户的密码哈希格式不兼容")
        print("2. 现有用户的状态不是'active'")
        print("3. 现有用户的会话表有约束问题")
    elif not new_user_success:
        print("\n💡 分析:")
        print("新用户登录也失败，可能是系统级问题:")
        print("1. 数据库会话表约束问题")
        print("2. JWT token生成问题")
        print("3. 数据库连接问题")

if __name__ == "__main__":
    main() 