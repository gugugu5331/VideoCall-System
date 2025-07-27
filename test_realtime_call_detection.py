#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
实时来电检测功能测试脚本
"""

import requests
import json
import time
import threading
import websocket
from datetime import datetime

# 基础URL
base_url = "http://localhost:8000"

class RealtimeCallDetectionTest:
    def __init__(self):
        self.user1_token = None
        self.user2_token = None
        self.user1_info = None
        self.user2_info = None
        self.notification_received = False
        self.notification_data = None
        
    def setup_users(self):
        """设置测试用户"""
        print("🔧 设置测试用户...")
        
        # 创建或登录用户1
        user1_data = {
            "username": "caller_realtime",
            "password": "password123",
            "email": "caller_realtime@example.com",
            "full_name": "主叫用户(实时)"
        }
        
        # 创建或登录用户2
        user2_data = {
            "username": "callee_realtime", 
            "password": "password123",
            "email": "callee_realtime@example.com",
            "full_name": "被叫用户(实时)"
        }
        
        # 注册/登录用户1
        try:
            response = requests.post(f"{base_url}/api/v1/auth/register", json=user1_data)
            if response.status_code in [201, 400, 409]:
                print("✅ 用户1注册/存在成功")
            else:
                print(f"❌ 用户1注册失败: {response.status_code}")
                return False
        except Exception as e:
            print(f"❌ 用户1注册异常: {e}")
            return False
        
        # 注册/登录用户2
        try:
            response = requests.post(f"{base_url}/api/v1/auth/register", json=user2_data)
            if response.status_code in [201, 400, 409]:
                print("✅ 用户2注册/存在成功")
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
    
    def test_notification_websocket(self):
        """测试通知WebSocket连接"""
        print("\n🔌 测试通知WebSocket连接...")
        
        try:
            # 构建通知WebSocket URL
            ws_url = f"ws://localhost:8000/ws/notifications?user_id={self.user2_info['uuid']}"
            print(f"   通知WebSocket URL: {ws_url}")
            
            # 创建WebSocket连接
            ws = websocket.create_connection(ws_url, timeout=10)
            
            # 发送订阅消息
            subscribe_message = {
                "type": "subscribe",
                "user_id": self.user2_info['uuid'],
                "event": "incoming_call"
            }
            ws.send(json.dumps(subscribe_message))
            print("   ✅ 订阅消息已发送")
            
            # 等待连接确认消息
            response = ws.recv()
            response_data = json.loads(response)
            print(f"   ✅ 收到连接确认: {response_data.get('type')}")
            
            # 启动监听线程
            def listen_for_notifications():
                try:
                    while True:
                        message = ws.recv()
                        data = json.loads(message)
                        print(f"   📨 收到通知: {data}")
                        
                        if data.get('type') == 'incoming_call':
                            self.notification_received = True
                            self.notification_data = data.get('data')
                            print(f"   ✅ 收到来电通知: {self.notification_data}")
                            break
                except Exception as e:
                    print(f"   ❌ 监听通知异常: {e}")
            
            # 启动监听线程
            listener_thread = threading.Thread(target=listen_for_notifications)
            listener_thread.daemon = True
            listener_thread.start()
            
            # 等待一下让WebSocket连接稳定
            time.sleep(2)
            
            return ws
            
        except Exception as e:
            print(f"❌ 通知WebSocket连接失败: {e}")
            return None
    
    def test_call_initiation_with_notification(self, ws):
        """测试发起通话并验证通知"""
        print("\n📞 测试发起通话并验证实时通知...")
        
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
                
                # 等待通知
                print("   ⏳ 等待实时通知...")
                timeout = 10  # 10秒超时
                start_time = time.time()
                
                while not self.notification_received and (time.time() - start_time) < timeout:
                    time.sleep(0.5)
                
                if self.notification_received:
                    print("   ✅ 实时通知接收成功")
                    print(f"   通知数据: {self.notification_data}")
                    return call_info
                else:
                    print("   ❌ 未收到实时通知")
                    return None
            else:
                print(f"❌ 通话发起失败: {response.status_code}")
                return None
                
        except Exception as e:
            print(f"❌ 通话发起异常: {e}")
            return None
        finally:
            if ws:
                ws.close()
    
    def test_call_history_check(self):
        """测试通话历史检查"""
        print("\n📋 测试通话历史检查...")
        
        try:
            headers = {"Authorization": f"Bearer {self.user2_token}"}
            response = requests.get(
                f"{base_url}/api/v1/calls/history?page=1&limit=10",
                headers=headers
            )
            
            if response.status_code == 200:
                history = response.json()
                calls = history["calls"]
                print(f"✅ 通话历史获取成功")
                print(f"   总通话数: {len(calls)}")
                
                # 查找状态为initiated的通话
                incoming_calls = [call for call in calls if call["status"] == "initiated"]
                if incoming_calls:
                    latest_call = incoming_calls[0]
                    print(f"   发现未接来电: ID={latest_call['id']}, 主叫={latest_call.get('caller_username', '未知')}")
                    return True
                else:
                    print("   没有发现未接来电")
                    return False
            else:
                print(f"❌ 通话历史获取失败: {response.status_code}")
                return False
                
        except Exception as e:
            print(f"❌ 通话历史异常: {e}")
            return False
    
    def run_test(self):
        """运行完整测试"""
        print("实时来电检测功能测试")
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
        
        print("🚀 开始实时来电检测功能测试")
        print("=" * 50)
        
        # 设置用户
        if not self.setup_users():
            print("❌ 用户设置失败")
            return
        
        # 测试通知WebSocket连接
        ws = self.test_notification_websocket()
        if not ws:
            print("❌ 通知WebSocket连接失败")
            return
        
        # 测试发起通话并验证通知
        call_info = self.test_call_initiation_with_notification(ws)
        if not call_info:
            print("❌ 通话发起或通知接收失败")
            return
        
        # 等待一下
        print("⏳ 等待3秒...")
        time.sleep(3)
        
        # 测试通话历史检查
        if not self.test_call_history_check():
            print("❌ 通话历史检查失败")
            return
        
        print("=" * 50)
        print("✅ 实时来电检测功能测试完成")
        print("💡 测试结果:")
        print(f"   - 通知WebSocket连接: ✅ 成功")
        print(f"   - 实时通知接收: {'✅ 成功' if self.notification_received else '❌ 失败'}")
        print(f"   - 通话历史检查: ✅ 成功")
        print("💡 下一步测试建议:")
        print("1. 打开两个浏览器窗口")
        print("2. 分别使用 caller_realtime 和 callee_realtime 登录")
        print("3. 在 caller_realtime 中搜索并呼叫 callee_realtime")
        print("4. 在 callee_realtime 中应该立即看到来电通知")
        print("5. 验证实时通知的响应速度和准确性")

if __name__ == "__main__":
    test = RealtimeCallDetectionTest()
    test.run_test() 