#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
通话通知功能测试脚本
"""

import requests
import json
import time
import threading
from datetime import datetime

# 基础URL
base_url = "http://localhost:8000"

class CallNotificationTest:
    def __init__(self):
        self.user1_token = None
        self.user2_token = None
        self.user1_info = None
        self.user2_info = None
        
    def setup_users(self):
        """设置测试用户"""
        print("🔧 设置测试用户...")
        
        # 创建或登录用户1
        user1_data = {
            "username": "caller_user",
            "password": "password123",
            "email": "caller@example.com",
            "full_name": "主叫用户"
        }
        
        # 创建或登录用户2
        user2_data = {
            "username": "callee_user", 
            "password": "password123",
            "email": "callee@example.com",
            "full_name": "被叫用户"
        }
        
        # 注册/登录用户1
        try:
            response = requests.post(f"{base_url}/api/v1/auth/register", json=user1_data)
            if response.status_code == 201:
                print("✅ 用户1注册成功")
            elif response.status_code == 400 and "already exists" in response.text:
                print("✅ 用户1已存在")
            elif response.status_code == 409:
                print("✅ 用户1已存在")
            else:
                print(f"❌ 用户1注册失败: {response.status_code}")
                return False
        except Exception as e:
            print(f"❌ 用户1注册异常: {e}")
            return False
        
        # 注册/登录用户2
        try:
            response = requests.post(f"{base_url}/api/v1/auth/register", json=user2_data)
            if response.status_code == 201:
                print("✅ 用户2注册成功")
            elif response.status_code == 400 and "already exists" in response.text:
                print("✅ 用户2已存在")
            elif response.status_code == 409:
                print("✅ 用户2已存在")
            else:
                print(f"❌ 用户2注册失败: {response.status_code}")
                return False
        except Exception as e:
            print(f"❌ 用户2注册异常: {e}")
            return False
        
        # 登录用户1
        try:
            login_data = {
                "username": user1_data["username"],
                "password": user1_data["password"]
            }
            response = requests.post(f"{base_url}/api/v1/auth/login", json=login_data)
            if response.status_code == 200:
                self.user1_token = response.json()["token"]
                self.user1_info = response.json()["user"]
                print(f"✅ 用户1登录成功: {self.user1_info['username']}")
            else:
                print(f"❌ 用户1登录失败: {response.status_code}")
                return False
        except Exception as e:
            print(f"❌ 用户1登录异常: {e}")
            return False
        
        # 登录用户2
        try:
            login_data = {
                "username": user2_data["username"],
                "password": user2_data["password"]
            }
            response = requests.post(f"{base_url}/api/v1/auth/login", json=login_data)
            if response.status_code == 200:
                self.user2_token = response.json()["token"]
                self.user2_info = response.json()["user"]
                print(f"✅ 用户2登录成功: {self.user2_info['username']}")
            else:
                print(f"❌ 用户2登录失败: {response.status_code}")
                return False
        except Exception as e:
            print(f"❌ 用户2登录异常: {e}")
            return False
        
        return True
    
    def test_call_initiation(self):
        """测试通话发起"""
        print("\n📞 测试发起通话...")
        
        try:
            headers = {"Authorization": f"Bearer {self.user1_token}"}
            call_data = {
                "callee_id": self.user2_info["uuid"],
                "callee_username": self.user2_info["username"],
                "call_type": "video"
            }
            
            response = requests.post(
                f"{base_url}/api/v1/calls/start",
                json=call_data,
                headers=headers
            )
            
            if response.status_code == 201:
                call_info = response.json()["call"]
                print(f"✅ 通话发起成功")
                print(f"   通话ID: {call_info['id']}")
                print(f"   通话UUID: {call_info['uuid']}")
                print(f"   房间ID: {call_info['room_id']}")
                print(f"   被叫用户: {call_info['callee_username']}")
                return call_info
            else:
                print(f"❌ 通话发起失败: {response.status_code}")
                print(f"   错误信息: {response.text}")
                return None
                
        except Exception as e:
            print(f"❌ 通话发起异常: {e}")
            return None
    
    def test_call_history_for_callee(self):
        """测试被叫方的通话历史"""
        print("\n📋 测试被叫方通话历史...")
        
        try:
            headers = {"Authorization": f"Bearer {self.user2_token}"}
            response = requests.get(
                f"{base_url}/api/v1/calls/history?page=1&limit=10",
                headers=headers
            )
            
            if response.status_code == 200:
                history = response.json()
                calls = history["calls"]
                print(f"✅ 被叫方通话历史获取成功")
                print(f"   总通话数: {len(calls)}")
                
                # 查找状态为initiated的通话
                incoming_calls = [call for call in calls if call["status"] == "initiated"]
                if incoming_calls:
                    latest_call = incoming_calls[0]
                    print(f"   发现未接来电: ID={latest_call['id']}, 主叫={latest_call.get('caller_username', '未知')}")
                    return latest_call
                else:
                    print("   没有发现未接来电")
                    return None
            else:
                print(f"❌ 通话历史获取失败: {response.status_code}")
                return None
                
        except Exception as e:
            print(f"❌ 通话历史异常: {e}")
            return None
    
    def test_call_details(self, call_uuid):
        """测试获取通话详情"""
        print(f"\n📄 测试获取通话详情: {call_uuid}")
        
        try:
            headers = {"Authorization": f"Bearer {self.user2_token}"}
            response = requests.get(
                f"{base_url}/api/v1/calls/{call_uuid}",
                headers=headers
            )
            
            if response.status_code == 200:
                call_details = response.json()["call"]
                print(f"✅ 通话详情获取成功")
                print(f"   通话UUID: {call_details['uuid']}")
                print(f"   主叫用户: {call_details.get('caller_username', '未知')}")
                print(f"   被叫用户: {call_details.get('callee_username', '未知')}")
                print(f"   通话状态: {call_details['status']}")
                return call_details
            else:
                print(f"❌ 通话详情获取失败: {response.status_code}")
                return None
                
        except Exception as e:
            print(f"❌ 通话详情异常: {e}")
            return None
    
    def test_websocket_connection(self, call_info):
        """测试WebSocket连接"""
        print(f"\n🔌 测试WebSocket连接...")
        
        if not call_info:
            print("❌ 通话信息为空，跳过WebSocket测试")
            return False
        
        try:
            # 构建WebSocket URL
            ws_url = f"ws://localhost:8000/ws/call/{call_info['uuid']}?user_id={self.user1_info['uuid']}"
            print(f"   主叫方WebSocket URL: {ws_url}")
            
            ws_url_callee = f"ws://localhost:8000/ws/call/{call_info['uuid']}?user_id={self.user2_info['uuid']}"
            print(f"   被叫方WebSocket URL: {ws_url_callee}")
            
            print("   ✅ WebSocket URL构建成功")
            return True
            
        except Exception as e:
            print(f"❌ WebSocket URL构建失败: {e}")
            return False
    
    def run_test(self):
        """运行完整测试"""
        print("通话通知功能测试")
        print("=" * 60)
        
        # 检查后端服务
        try:
            response = requests.get(f"{base_url}/health")
            if response.status_code == 200:
                print("✅ 后端服务运行正常")
            else:
                print("❌ 后端服务异常")
                return
        except Exception as e:
            print(f"❌ 后端服务连接失败: {e}")
            return
        
        print("🚀 开始通话通知功能测试")
        print("=" * 50)
        
        # 设置用户
        if not self.setup_users():
            print("❌ 用户设置失败")
            return
        
        # 测试通话发起
        call_info = self.test_call_initiation()
        if not call_info:
            print("❌ 通话发起失败")
            return
        
        # 等待一下
        print("⏳ 等待3秒...")
        time.sleep(3)
        
        # 测试被叫方通话历史
        incoming_call = self.test_call_history_for_callee()
        if not incoming_call:
            print("❌ 被叫方没有收到通话通知")
            return
        
        # 测试获取通话详情
        call_details = self.test_call_details(incoming_call['uuid'])
        if not call_details:
            print("❌ 获取通话详情失败")
            return
        
        # 测试WebSocket连接
        if not self.test_websocket_connection(call_info):
            print("❌ WebSocket连接测试失败")
            return
        
        print("=" * 50)
        print("✅ 通话通知功能测试完成")
        print("💡 下一步测试建议:")
        print("1. 打开两个浏览器窗口")
        print("2. 分别使用 caller_user 和 callee_user 登录")
        print("3. 在 caller_user 中搜索并呼叫 callee_user")
        print("4. 在 callee_user 中应该看到来电通知")
        print("5. 点击接听按钮测试WebRTC连接")
        print("6. 检查浏览器控制台日志，确认WebRTC连接建立")

if __name__ == "__main__":
    test = CallNotificationTest()
    test.run_test() 