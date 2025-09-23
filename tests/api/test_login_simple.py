#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简化登录测试 - 只测试JWT生成
"""

import requests
import json

# 配置
BASE_URL = "http://localhost:8000"
API_BASE = f"{BASE_URL}/api/v1"

def test_simple_login():
    """测试简化登录（不创建会话）"""
    print("🔍 测试简化登录")
    print("=" * 50)
    
    # 测试数据
    login_data = {
        "username": "alice",
        "password": "password123"
    }
    
    print(f"登录数据: {json.dumps(login_data, ensure_ascii=False, indent=2)}")
    print(f"请求URL: {API_BASE}/auth/login")
    print()
    
    try:
        # 发送登录请求
        response = requests.post(
            f"{API_BASE}/auth/login",
            json=login_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"响应状态码: {response.status_code}")
        print(f"响应头: {dict(response.headers)}")
        print()
        
        if response.status_code == 200:
            print("✅ 登录成功")
            result = response.json()
            print(f"Token: {result['token'][:50]}...")
            print(f"用户: {result['user']['username']}")
            return True
        else:
            print("❌ 登录失败")
            print(f"响应内容: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
        return False

def test_with_session_cleanup():
    """测试前清理会话"""
    print("\n🔍 测试前清理会话")
    print("=" * 50)
    
    # 这里我们可以尝试清理一些旧的会话
    # 但由于我们没有直接的API，我们先测试登录
    
    return test_simple_login()

def main():
    print("🚀 开始简化登录测试")
    print(f"目标服务器: {BASE_URL}")
    print("=" * 60)
    
    # 测试简化登录
    success = test_simple_login()
    
    if not success:
        print("\n尝试清理会话后再次测试...")
        test_with_session_cleanup()
    
    print("\n" + "=" * 60)
    if success:
        print("✅ 登录测试成功！")
        print("💡 问题可能在于会话管理，而不是JWT生成")
    else:
        print("❌ 登录测试失败")
        print("💡 问题可能在于密码验证或JWT生成")

if __name__ == "__main__":
    main() 