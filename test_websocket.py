#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
WebSocket连接测试脚本
"""

import asyncio
import websockets
import json
import requests

# 配置
BACKEND_URL = "http://localhost:8000"
WS_URL = "ws://localhost:8000"

async def test_websocket_connection():
    """测试WebSocket连接"""
    print("🔍 测试WebSocket连接")
    print("=" * 50)
    
    # 首先登录获取token
    login_data = {
        "username": "alice",
        "password": "password123"
    }
    
    try:
        response = requests.post(f"{BACKEND_URL}/api/v1/auth/login", json=login_data)
        print(f"   登录响应状态码: {response.status_code}")
        if response.status_code != 200:
            print(f"❌ 登录失败: {response.text}")
            return False
        
        login_result = response.json()
        token = login_result['token']
        user_uuid = login_result['user']['uuid']
        
        print(f"✅ 登录成功")
        print(f"   用户UUID: {user_uuid}")
        print(f"   Token: {token[:50]}...")
        
    except Exception as e:
        print(f"❌ 登录异常: {e}")
        return False
    
    # 发起通话获取call_id
    try:
        call_data = {
            "callee_username": "bob",
            "call_type": "video"
        }
        
        headers = {"Authorization": f"Bearer {token}"}
        response = requests.post(f"{BACKEND_URL}/api/v1/calls/start", 
                               json=call_data, headers=headers)
        
        if response.status_code != 201:
            print("❌ 发起通话失败")
            return False
        
        call_result = response.json()
        call_uuid = call_result['call']['uuid']
        
        print(f"✅ 发起通话成功")
        print(f"   通话UUID: {call_uuid}")
        
    except Exception as e:
        print(f"❌ 发起通话异常: {e}")
        return False
    
    # 测试WebSocket连接
    ws_url = f"{WS_URL}/ws/call/{call_uuid}"
    headers = {"Authorization": f"Bearer {token}"}
    
    print(f"\n🔗 连接WebSocket: {ws_url}")
    
    try:
        async with websockets.connect(ws_url, additional_headers=headers) as websocket:
            print("✅ WebSocket连接成功")
            
            # 等待连接消息
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                data = json.loads(message)
                print(f"✅ 收到连接消息: {data.get('type', 'unknown')}")
                
                if data.get('type') == 'connection':
                    print("✅ WebSocket连接验证成功")
                    return True
                else:
                    print(f"⚠️  收到意外消息类型: {data.get('type')}")
                    return False
                    
            except asyncio.TimeoutError:
                print("❌ 等待连接消息超时")
                return False
                
    except Exception as e:
        print(f"❌ WebSocket连接失败: {e}")
        return False

async def test_websocket_without_auth():
    """测试无认证的WebSocket连接（应该失败）"""
    print("\n🔍 测试无认证的WebSocket连接")
    print("=" * 50)
    
    # 使用一个假的call_id
    fake_call_id = "test-call-123"
    ws_url = f"{WS_URL}/ws/call/{fake_call_id}"
    
    try:
        async with websockets.connect(ws_url) as websocket:
            print("❌ 无认证连接成功（这不应该发生）")
            return False
    except Exception as e:
        print(f"✅ 无认证连接被正确拒绝: {e}")
        return True

def main():
    print("🚀 开始WebSocket连接测试")
    print("=" * 60)
    
    # 测试有认证的连接
    auth_success = asyncio.run(test_websocket_connection())
    
    # 测试无认证的连接
    no_auth_success = asyncio.run(test_websocket_without_auth())
    
    print("\n" + "=" * 60)
    print("WebSocket测试总结:")
    print(f"有认证连接: {'✅ 成功' if auth_success else '❌ 失败'}")
    print(f"无认证连接: {'✅ 正确拒绝' if no_auth_success else '❌ 意外成功'}")
    
    if auth_success and no_auth_success:
        print("\n🎉 WebSocket连接测试通过！")
        print("💡 WebSocket认证和连接功能正常工作")
    else:
        print("\n⚠️  WebSocket连接测试失败")
        print("💡 请检查WebSocket配置和认证中间件")

if __name__ == "__main__":
    main() 