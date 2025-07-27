#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
快速API测试脚本
"""

import requests
import json

BASE_URL = "http://localhost:8000"
API_BASE = f"{BASE_URL}/api/v1"

def test_login():
    """测试登录功能"""
    print("测试登录功能...")
    
    login_data = {
        "username": "alice",
        "password": "password123"
    }
    
    try:
        response = requests.post(f"{API_BASE}/auth/login", json=login_data)
        print(f"状态码: {response.status_code}")
        print(f"响应: {response.text}")
        
        if response.status_code == 200:
            data = response.json()
            token = data.get("token")
            if token:
                print(f"✅ 登录成功，获取到token: {token[:50]}...")
                return token
            else:
                print("❌ 登录响应中没有token")
                return None
        else:
            print(f"❌ 登录失败: {response.status_code}")
            return None
    except Exception as e:
        print(f"❌ 登录请求失败: {e}")
        return None

def test_search_users(token):
    """测试用户搜索"""
    print("\n测试用户搜索...")
    
    headers = {"Authorization": f"Bearer {token}"}
    
    try:
        response = requests.get(f"{API_BASE}/users/search?query=alice&limit=10", headers=headers)
        print(f"状态码: {response.status_code}")
        print(f"响应: {response.text}")
        
        if response.status_code == 200:
            data = response.json()
            users = data.get("users", [])
            print(f"✅ 搜索成功，找到 {len(users)} 个用户")
            return True
        else:
            print(f"❌ 搜索失败: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ 搜索请求失败: {e}")
        return False

def test_call_by_username(token):
    """测试基于用户名的通话"""
    print("\n测试基于用户名的通话...")
    
    headers = {"Authorization": f"Bearer {token}"}
    call_data = {
        "callee_username": "bob",
        "call_type": "video"
    }
    
    try:
        response = requests.post(f"{API_BASE}/calls/start", json=call_data, headers=headers)
        print(f"状态码: {response.status_code}")
        print(f"响应: {response.text}")
        
        if response.status_code == 201:
            data = response.json()
            call_info = data.get("call", {})
            print(f"✅ 通话发起成功")
            print(f"通话ID: {call_info.get('id')}")
            print(f"房间ID: {call_info.get('room_id')}")
            return True
        else:
            print(f"❌ 通话发起失败: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ 通话请求失败: {e}")
        return False

def main():
    print("🚀 开始API测试")
    print(f"目标服务器: {BASE_URL}")
    
    # 测试登录
    token = test_login()
    if not token:
        print("❌ 登录失败，无法继续测试")
        return
    
    # 测试用户搜索
    test_search_users(token)
    
    # 测试基于用户名的通话
    test_call_by_username(token)
    
    print("\n✅ API测试完成")

if __name__ == "__main__":
    main() 