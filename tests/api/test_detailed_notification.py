#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
详细呼叫通知测试脚本
测试呼叫发起后，被呼叫方是否能收到通知
"""

import requests
import json
import time
import asyncio
import websockets
from datetime import datetime

# 配置
BACKEND_URL = "http://localhost:8000"
WS_URL = "ws://localhost:8000"

class DetailedNotificationTester:
    def __init__(self):
        self.session = requests.Session()
        self.caller_token = None
        self.callee_token = None
        self.caller_uuid = None
        self.callee_uuid = None
        self.callee_messages = []
        self.caller_messages = []
        
    def print_step(self, step, message):
        print(f"\n{'='*50}")
        print(f"步骤 {step}: {message}")
        print(f"{'='*50}")
        
    def print_success(self, message):
        print(f"✅ {message}")
        
    def print_error(self, message):
        print(f"❌ {message}")
        
    def print_info(self, message):
        print(f"ℹ️  {message}")
        
    def print_warning(self, message):
        print(f"⚠️  {message}")
        
    def login_user(self, username, password):
        """登录用户并返回token和uuid"""
        login_data = {
            "username": username,
            "password": password
        }
        
        try:
            response = self.session.post(f"{BACKEND_URL}/api/v1/auth/login", json=login_data)
            if response.status_code == 200:
                data = response.json()
                token = data.get("token")
                user_uuid = data.get("user", {}).get("uuid")
                self.print_success(f"用户 {username} 登录成功")
                self.print_info(f"UUID: {user_uuid}")
                return token, user_uuid
            else:
                self.print_error(f"用户 {username} 登录失败: {response.status_code}")
                return None, None
        except Exception as e:
            self.print_error(f"登录用户 {username} 时出错: {e}")
            return None, None
            
    def setup_users(self):
        """设置呼叫方和被叫方"""
        self.print_step(1, "设置用户")
        
        # 登录呼叫方 (alice)
        self.caller_token, self.caller_uuid = self.login_user("alice", "password123")
        if not self.caller_token:
            return False
            
        # 登录被叫方 (bob)
        self.callee_token, self.callee_uuid = self.login_user("bob", "password123")
        if not self.callee_token:
            return False
            
        return True
        
    def start_call(self):
        """发起通话"""
        self.print_step(2, "发起通话")
        
        call_data = {
            "callee_username": "bob",
            "call_type": "video"
        }
        
        headers = {"Authorization": f"Bearer {self.caller_token}"}
        
        try:
            response = self.session.post(f"{BACKEND_URL}/api/v1/calls/start", 
                                       json=call_data, headers=headers)
            print(f"发起通话响应状态码: {response.status_code}")
            print(f"发起通话响应内容: {response.text}")
            
            if response.status_code == 201:
                data = response.json()
                call_info = data.get("call", {})
                call_uuid = call_info.get("uuid")
                self.print_success(f"成功发起通话: alice -> bob")
                self.print_info(f"通话UUID: {call_uuid}")
                return call_uuid
            else:
                self.print_error(f"发起通话失败: {response.status_code}")
                return None
        except Exception as e:
            self.print_error(f"发起通话时出错: {e}")
            return None
            
    async def listen_for_messages(self, websocket, user_type):
        """监听WebSocket消息"""
        try:
            while True:
                message = await websocket.recv()
                data = json.loads(message)
                timestamp = datetime.now().strftime("%H:%M:%S")
                
                if user_type == "callee":
                    self.callee_messages.append(data)
                else:
                    self.caller_messages.append(data)
                    
                self.print_info(f"[{timestamp}] {user_type} 收到消息: {data.get('type', 'unknown')}")
                if data.get('type') == 'join':
                    self.print_success(f"[{timestamp}] {user_type} 收到用户加入通知!")
                    
        except websockets.exceptions.ConnectionClosed:
            self.print_info(f"{user_type} WebSocket连接关闭")
        except Exception as e:
            self.print_error(f"{user_type} 监听消息时出错: {e}")
            
    async def test_call_notification(self, call_uuid):
        """测试呼叫通知"""
        self.print_step(3, "测试呼叫通知")
        
        # 首先让被叫方连接WebSocket并开始监听
        callee_ws_url = f"{WS_URL}/ws/call/{call_uuid}"
        callee_headers = {"Authorization": f"Bearer {self.callee_token}"}
        
        self.print_info(f"被叫方连接WebSocket: {callee_ws_url}")
        
        try:
            callee_websocket = await websockets.connect(callee_ws_url, additional_headers=callee_headers)
            self.print_success("被叫方WebSocket连接成功")
            
            # 开始监听被叫方消息
            callee_listener = asyncio.create_task(self.listen_for_messages(callee_websocket, "callee"))
            
            # 等待一秒，然后让呼叫方连接
            await asyncio.sleep(1)
            
            caller_ws_url = f"{WS_URL}/ws/call/{call_uuid}"
            caller_headers = {"Authorization": f"Bearer {self.caller_token}"}
            
            self.print_info(f"呼叫方连接WebSocket: {caller_ws_url}")
            
            caller_websocket = await websockets.connect(caller_ws_url, additional_headers=caller_headers)
            self.print_success("呼叫方WebSocket连接成功")
            
            # 开始监听呼叫方消息
            caller_listener = asyncio.create_task(self.listen_for_messages(caller_websocket, "caller"))
            
            # 等待消息
            await asyncio.sleep(3)
            
            # 取消监听任务
            callee_listener.cancel()
            caller_listener.cancel()
            
            # 关闭连接
            await callee_websocket.close()
            await caller_websocket.close()
            
            # 分析结果
            self.print_step(4, "分析通知结果")
            
            self.print_info(f"被叫方收到的消息数量: {len(self.callee_messages)}")
            for i, msg in enumerate(self.callee_messages):
                self.print_info(f"  消息 {i+1}: {msg.get('type', 'unknown')}")
                
            self.print_info(f"呼叫方收到的消息数量: {len(self.caller_messages)}")
            for i, msg in enumerate(self.caller_messages):
                self.print_info(f"  消息 {i+1}: {msg.get('type', 'unknown')}")
                
            # 检查是否有join消息
            callee_join_messages = [msg for msg in self.callee_messages if msg.get('type') == 'join']
            if callee_join_messages:
                self.print_success("✅ 被叫方收到了用户加入通知!")
                return True
            else:
                self.print_warning("⚠️  被叫方没有收到用户加入通知")
                return False
                
        except Exception as e:
            self.print_error(f"WebSocket连接失败: {e}")
            return False
        
    async def run_test(self):
        """运行完整测试"""
        print("🚀 开始详细呼叫通知测试")
        print("=" * 60)
        
        # 设置用户
        if not self.setup_users():
            return False
            
        # 发起通话
        call_uuid = self.start_call()
        if not call_uuid:
            return False
            
        # 测试呼叫通知
        success = await self.test_call_notification(call_uuid)
        
        print("\n" + "=" * 60)
        if success:
            print("🎉 呼叫通知测试通过！")
            print("💡 被叫方成功收到了呼叫通知")
        else:
            print("⚠️  呼叫通知测试失败")
            print("💡 被叫方没有收到呼叫通知，需要检查通知机制")
            
        return success

async def main():
    tester = DetailedNotificationTester()
    await tester.run_test()

if __name__ == "__main__":
    asyncio.run(main()) 