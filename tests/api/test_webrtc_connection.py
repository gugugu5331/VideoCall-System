#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
WebRTC连接测试脚本
"""

import requests
import json
import time
import threading
from datetime import datetime

# 基础URL
base_url = "http://localhost:8000"

class WebRTCTest:
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
            "username": "webrtc_user1",
            "password": "password123",
            "email": "webrtc1@example.com",
            "full_name": "WebRTC测试用户1"
        }
        
        # 创建或登录用户2
        user2_data = {
            "username": "webrtc_user2", 
            "password": "password123",
            "email": "webrtc2@example.com",
            "full_name": "WebRTC测试用户2"
        }
        
        # 注册/登录用户1
        try:
            response = requests.post(f"{base_url}/api/v1/auth/register", json=user1_data)
            if response.status_code == 201:
                print("✅ 用户1注册成功")
            elif response.status_code == 400 and "already exists" in response.text:
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
    
    def test_user_search(self):
        """测试用户搜索功能"""
        print("\n🔍 测试用户搜索功能...")
        
        try:
            headers = {"Authorization": f"Bearer {self.user1_token}"}
            response = requests.get(
                f"{base_url}/api/v1/users/search?query=webrtc_user2&limit=10",
                headers=headers
            )
            
            if response.status_code == 200:
                users = response.json()["users"]
                print(f"✅ 用户搜索成功，找到 {len(users)} 个用户")
                for user in users:
                    print(f"   找到用户: {user['username']} (UUID: {user['uuid']})")
                return users
            else:
                print(f"❌ 用户搜索失败: {response.status_code}")
                return []
                
        except Exception as e:
            print(f"❌ 用户搜索异常: {e}")
            return []
    
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
                print(f"   房间ID: {call_info['uuid']}")
                print(f"   被叫用户: {call_info['callee_username']}")
                return call_info
            else:
                print(f"❌ 通话发起失败: {response.status_code}")
                print(f"   错误信息: {response.text}")
                return None
                
        except Exception as e:
            print(f"❌ 通话发起异常: {e}")
            return None
    
    def test_websocket_url(self, call_info):
        """测试WebSocket URL构建"""
        print("\n🔌 测试WebSocket连接...")
        
        if not call_info:
            print("❌ 通话信息为空，跳过WebSocket测试")
            return False
        
        try:
            # 构建WebSocket URL
            ws_url = f"ws://localhost:8000/ws/call/{call_info['uuid']}?user_id={self.user1_info['uuid']}"
            print(f"   WebSocket URL: {ws_url}")
            print("   ✅ WebSocket URL构建成功")
            return True
            
        except Exception as e:
            print(f"❌ WebSocket URL构建失败: {e}")
            return False
    
    def test_call_history(self):
        """测试通话历史"""
        print("\n📋 测试通话历史...")
        
        try:
            headers = {"Authorization": f"Bearer {self.user1_token}"}
            response = requests.get(
                f"{base_url}/api/v1/calls/history?page=1&limit=10",
                headers=headers
            )
            
            if response.status_code == 200:
                history = response.json()
                calls = history["calls"]
                print(f"✅ 通话历史获取成功")
                print(f"   总通话数: {len(calls)}")
                if calls:
                    latest_call = calls[0]
                    print(f"   最新通话: ID={latest_call['id']}, 状态={latest_call['status']}")
                return True
            else:
                print(f"❌ 通话历史获取失败: {response.status_code}")
                return False
                
        except Exception as e:
            print(f"❌ 通话历史异常: {e}")
            return False
    
    def test_end_call(self, call_info):
        """测试结束通话"""
        print("\n📴 测试结束通话...")
        
        if not call_info:
            print("❌ 通话信息为空，跳过结束通话测试")
            return False
        
        try:
            headers = {"Authorization": f"Bearer {self.user1_token}"}
            end_data = {
                "call_id": call_info["id"]
            }
            
            response = requests.post(
                f"{base_url}/api/v1/calls/end",
                json=end_data,
                headers=headers
            )
            
            if response.status_code == 200:
                print("✅ 通话结束成功")
                return True
            else:
                print(f"❌ 通话结束失败: {response.status_code}")
                return False
                
        except Exception as e:
            print(f"❌ 通话结束异常: {e}")
            return False
    
    def run_test(self):
        """运行完整测试"""
        print("WebRTC连接测试")
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
        
        print("🚀 开始WebRTC连接测试")
        print("=" * 50)
        
        # 设置用户
        if not self.setup_users():
            print("❌ 用户设置失败")
            return
        
        # 测试用户搜索
        users = self.test_user_search()
        if not users:
            print("❌ 用户搜索失败")
            return
        
        # 测试通话发起
        call_info = self.test_call_initiation()
        if not call_info:
            print("❌ 通话发起失败")
            return
        
        # 测试WebSocket URL
        if not self.test_websocket_url(call_info):
            print("❌ WebSocket URL测试失败")
            return
        
        # 等待一下
        print("⏳ 等待5秒...")
        time.sleep(5)
        
        # 测试通话历史
        if not self.test_call_history():
            print("❌ 通话历史测试失败")
            return
        
        # 测试结束通话
        if not self.test_end_call(call_info):
            print("❌ 结束通话测试失败")
            return
        
        print("=" * 50)
        print("✅ WebRTC连接测试完成")
        print("💡 下一步测试建议:")
        print("1. 打开两个浏览器窗口")
        print("2. 分别使用 webrtc_user1 和 webrtc_user2 登录")
        print("3. 在 webrtc_user1 中搜索并呼叫 webrtc_user2")
        print("4. 在 webrtc_user2 中接受通话")
        print("5. 验证WebRTC视频通话功能")
        print("6. 检查浏览器控制台日志，确认WebRTC连接建立")

if __name__ == "__main__":
    test = WebRTCTest()
    test.run_test() 