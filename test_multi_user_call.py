#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
多用户通话功能测试脚本
"""

import requests
import json
import time
import threading
from datetime import datetime

# 基础URL
base_url = "http://localhost:8000"

class MultiUserCallTest:
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
            "username": "testuser1",
            "password": "password123",
            "email": "test1@example.com",
            "full_name": "测试用户1"
        }
        
        # 创建或登录用户2
        user2_data = {
            "username": "testuser2", 
            "password": "password123",
            "email": "test2@example.com",
            "full_name": "测试用户2"
        }
        
        # 尝试登录用户1
        login_response = requests.post(f"{base_url}/api/v1/auth/login", json={
            "username": user1_data["username"],
            "password": user1_data["password"]
        })
        
        if login_response.status_code == 200:
            self.user1_token = login_response.json()["token"]
            self.user1_info = login_response.json()["user"]
            print(f"✅ 用户1登录成功: {self.user1_info['username']}")
        else:
            # 注册用户1
            register_response = requests.post(f"{base_url}/api/v1/auth/register", json=user1_data)
            if register_response.status_code == 201:
                login_response = requests.post(f"{base_url}/api/v1/auth/login", json={
                    "username": user1_data["username"],
                    "password": user1_data["password"]
                })
                self.user1_token = login_response.json()["token"]
                self.user1_info = login_response.json()["user"]
                print(f"✅ 用户1注册并登录成功: {self.user1_info['username']}")
            else:
                print(f"❌ 用户1注册失败: {register_response.text}")
                return False
        
        # 尝试登录用户2
        login_response = requests.post(f"{base_url}/api/v1/auth/login", json={
            "username": user2_data["username"],
            "password": user2_data["password"]
        })
        
        if login_response.status_code == 200:
            self.user2_token = login_response.json()["token"]
            self.user2_info = login_response.json()["user"]
            print(f"✅ 用户2登录成功: {self.user2_info['username']}")
        else:
            # 注册用户2
            register_response = requests.post(f"{base_url}/api/v1/auth/register", json=user2_data)
            if register_response.status_code == 201:
                login_response = requests.post(f"{base_url}/api/v1/auth/login", json={
                    "username": user2_data["username"],
                    "password": user2_data["password"]
                })
                self.user2_token = login_response.json()["token"]
                self.user2_info = login_response.json()["user"]
                print(f"✅ 用户2注册并登录成功: {self.user2_info['username']}")
            else:
                print(f"❌ 用户2注册失败: {register_response.text}")
                return False
        
        return True
    
    def test_user_search(self):
        """测试用户搜索功能"""
        print("\n🔍 测试用户搜索功能...")
        
        headers = {
            "Authorization": f"Bearer {self.user1_token}",
            "Content-Type": "application/json"
        }
        
        # 搜索用户2
        search_response = requests.get(
            f"{base_url}/api/v1/users/search?query=testuser2&limit=10",
            headers=headers
        )
        
        if search_response.status_code == 200:
            search_result = search_response.json()
            print(f"✅ 用户搜索成功，找到 {search_result['count']} 个用户")
            
            if search_result['users']:
                user2_found = search_result['users'][0]
                print(f"   找到用户: {user2_found['username']} (UUID: {user2_found['uuid']})")
                return user2_found
            else:
                print("❌ 未找到目标用户")
                return None
        else:
            print(f"❌ 用户搜索失败: {search_response.status_code}")
            return None
    
    def test_start_call(self, callee_user):
        """测试发起通话"""
        print(f"\n📞 测试发起通话: {self.user1_info['username']} -> {callee_user['username']}")
        
        headers = {
            "Authorization": f"Bearer {self.user1_token}",
            "Content-Type": "application/json"
        }
        
        call_data = {
            "callee_id": callee_user["uuid"],
            "callee_username": callee_user["username"],
            "call_type": "video"
        }
        
        call_response = requests.post(
            f"{base_url}/api/v1/calls/start",
            json=call_data,
            headers=headers
        )
        
        if call_response.status_code == 201:
            call_result = call_response.json()
            print(f"✅ 通话发起成功")
            print(f"   通话ID: {call_result['call']['id']}")
            print(f"   通话UUID: {call_result['call']['uuid']}")
            print(f"   房间ID: {call_result['call']['room_id']}")
            print(f"   被叫用户: {call_result['call']['callee']['username']}")
            return call_result['call']
        else:
            print(f"❌ 通话发起失败: {call_response.status_code}")
            print(f"   错误信息: {call_response.text}")
            return None
    
    def test_websocket_connection(self, call_info):
        """测试WebSocket连接"""
        print(f"\n🔌 测试WebSocket连接...")
        
        # 这里只是测试连接URL的构建，实际的WebSocket连接需要在前端进行
        ws_url = f"ws://localhost:8000/ws/call/{call_info['uuid']}?user_id={self.user1_info['uuid']}"
        print(f"   WebSocket URL: {ws_url}")
        print("   ✅ WebSocket URL构建成功")
        return True
    
    def test_call_history(self):
        """测试通话历史"""
        print(f"\n📋 测试通话历史...")
        
        headers = {
            "Authorization": f"Bearer {self.user1_token}",
            "Content-Type": "application/json"
        }
        
        history_response = requests.get(
            f"{base_url}/api/v1/calls/history",
            headers=headers
        )
        
        if history_response.status_code == 200:
            history_result = history_response.json()
            print(f"✅ 通话历史获取成功")
            print(f"   总通话数: {history_result['pagination']['total']}")
            
            if history_result['calls']:
                latest_call = history_result['calls'][0]
                print(f"   最新通话: ID={latest_call['id']}, 状态={latest_call['status']}")
            
            return True
        else:
            print(f"❌ 通话历史获取失败: {history_response.status_code}")
            return False
    
    def test_end_call(self, call_info):
        """测试结束通话"""
        print(f"\n📴 测试结束通话...")
        
        headers = {
            "Authorization": f"Bearer {self.user1_token}",
            "Content-Type": "application/json"
        }
        
        end_data = {
            "call_id": call_info['id']
        }
        
        end_response = requests.post(
            f"{base_url}/api/v1/calls/end",
            json=end_data,
            headers=headers
        )
        
        if end_response.status_code == 200:
            print(f"✅ 通话结束成功")
            return True
        else:
            print(f"❌ 通话结束失败: {end_response.status_code}")
            print(f"   错误信息: {end_response.text}")
            return False
    
    def run_all_tests(self):
        """运行所有测试"""
        print("🚀 开始多用户通话功能测试")
        print("=" * 50)
        
        # 设置用户
        if not self.setup_users():
            print("❌ 用户设置失败，测试终止")
            return False
        
        # 测试用户搜索
        callee_user = self.test_user_search()
        if not callee_user:
            print("❌ 用户搜索失败，测试终止")
            return False
        
        # 测试发起通话
        call_info = self.test_start_call(callee_user)
        if not call_info:
            print("❌ 通话发起失败，测试终止")
            return False
        
        # 测试WebSocket连接
        self.test_websocket_connection(call_info)
        
        # 等待一段时间
        print("\n⏳ 等待5秒...")
        time.sleep(5)
        
        # 测试通话历史
        self.test_call_history()
        
        # 测试结束通话
        self.test_end_call(call_info)
        
        print("\n" + "=" * 50)
        print("✅ 多用户通话功能测试完成")
        print("\n💡 下一步测试建议:")
        print("1. 打开两个浏览器窗口")
        print("2. 分别使用 testuser1 和 testuser2 登录")
        print("3. 在 testuser1 中搜索并呼叫 testuser2")
        print("4. 在 testuser2 中接受通话")
        print("5. 验证WebRTC视频通话功能")
        
        return True

def main():
    """主函数"""
    print("智能视频通话系统 - 多用户通话功能测试")
    print("=" * 60)
    
    # 检查后端服务状态
    try:
        health_response = requests.get(f"{base_url}/health", timeout=5)
        if health_response.status_code == 200:
            print("✅ 后端服务运行正常")
        else:
            print("❌ 后端服务状态异常")
            return
    except requests.exceptions.RequestException as e:
        print(f"❌ 无法连接到后端服务: {e}")
        print("请确保后端服务正在运行: docker-compose up -d")
        return
    
    # 运行测试
    tester = MultiUserCallTest()
    tester.run_all_tests()

if __name__ == "__main__":
    main() 