#!/usr/bin/env python3
"""
真正的通话功能演示脚本
展示WebRTC通话系统的完整功能
"""

import requests
import json
import time
import webbrowser
from datetime import datetime

class RealCallDemo:
    def __init__(self):
        self.base_url = "http://localhost:8000"
        self.frontend_url = "http://localhost:3000"
        self.demo_users = []
        
    def print_header(self):
        """打印演示标题"""
        print("=" * 60)
        print("🎥 真正的通话功能演示")
        print("=" * 60)
        print("📞 基于WebRTC的P2P音视频通话系统")
        print("🔒 集成AI安全检测功能")
        print("🌐 支持实时信令传输")
        print("=" * 60)
        print()
    
    def check_services(self):
        """检查服务状态"""
        print("🔍 检查服务状态...")
        
        # 检查后端服务
        try:
            response = requests.get(f"{self.base_url}/health", timeout=5)
            if response.status_code == 200:
                print("✅ 后端服务正常运行")
            else:
                print("❌ 后端服务异常")
                return False
        except Exception as e:
            print(f"❌ 无法连接到后端服务: {e}")
            return False
        
        # 检查前端服务
        try:
            response = requests.get(f"{self.frontend_url}", timeout=5)
            if response.status_code == 200:
                print("✅ 前端服务正常运行")
            else:
                print("⚠️  前端服务未运行，将启动本地服务器")
        except Exception as e:
            print("⚠️  前端服务未运行，将启动本地服务器")
        
        print()
        return True
    
    def setup_demo_users(self):
        """设置演示用户"""
        print("👥 设置演示用户...")
        
        demo_users_data = [
            {
                "username": "demo_user1",
                "email": "user1@demo.com",
                "password": "demo123456",
                "full_name": "演示用户1"
            },
            {
                "username": "demo_user2",
                "email": "user2@demo.com", 
                "password": "demo123456",
                "full_name": "演示用户2"
            }
        ]
        
        for user_data in demo_users_data:
            try:
                # 注册用户
                response = requests.post(f"{self.base_url}/api/v1/auth/register", json=user_data)
                if response.status_code == 201:
                    print(f"✅ 用户 {user_data['username']} 注册成功")
                elif response.status_code == 409:
                    print(f"⚠️  用户 {user_data['username']} 已存在")
                else:
                    print(f"❌ 用户 {user_data['username']} 注册失败")
                    continue
                
                # 登录用户
                login_data = {
                    "username": user_data["username"],
                    "password": user_data["password"]
                }
                login_response = requests.post(f"{self.base_url}/api/v1/auth/login", json=login_data)
                if login_response.status_code == 200:
                    token = login_response.json()["token"]
                    user_info = login_response.json()["user"]
                    self.demo_users.append({
                        "username": user_data["username"],
                        "token": token,
                        "user_info": user_info
                    })
                    print(f"✅ 用户 {user_data['username']} 登录成功")
                else:
                    print(f"❌ 用户 {user_data['username']} 登录失败")
                    
            except Exception as e:
                print(f"❌ 设置用户 {user_data['username']} 时出错: {e}")
        
        print()
    
    def demo_call_features(self):
        """演示通话功能"""
        print("📞 演示通话功能...")
        
        if len(self.demo_users) < 1:
            print("❌ 没有可用的演示用户")
            return
        
        user = self.demo_users[0]
        headers = {"Authorization": f"Bearer {user['token']}"}
        
        # 1. 创建通话
        print("1️⃣ 创建视频通话...")
        call_data = {
            "callee_id": user["user_info"]["uuid"],  # 自测模式
            "call_type": "video"
        }
        
        try:
            response = requests.post(f"{self.base_url}/api/v1/calls/start", json=call_data, headers=headers)
            if response.status_code == 201:
                call_info = response.json()["call"]
                print(f"✅ 通话创建成功")
                print(f"   通话ID: {call_info['id']}")
                print(f"   通话类型: {call_info['call_type']}")
                print(f"   状态: {call_info['status']}")
            else:
                print(f"❌ 通话创建失败: {response.text}")
                return
        except Exception as e:
            print(f"❌ 通话创建时出错: {e}")
            return
        
        # 2. 获取活跃通话
        print("\n2️⃣ 获取活跃通话...")
        try:
            response = requests.get(f"{self.base_url}/api/v1/calls/active", headers=headers)
            if response.status_code == 200:
                active_calls = response.json()["active_calls"]
                print(f"✅ 当前活跃通话: {len(active_calls)} 个")
                for call in active_calls:
                    print(f"   - 房间ID: {call['id']}")
                    print(f"   - 类型: {call['call_type']}")
                    print(f"   - 状态: {call['status']}")
            else:
                print(f"❌ 获取活跃通话失败: {response.text}")
        except Exception as e:
            print(f"❌ 获取活跃通话时出错: {e}")
        
        # 3. 获取通话详情
        print("\n3️⃣ 获取通话详情...")
        try:
            response = requests.get(f"{self.base_url}/api/v1/calls/{call_info['id']}", headers=headers)
            if response.status_code == 200:
                call_details = response.json()["call"]
                print(f"✅ 通话详情获取成功")
                print(f"   主叫: {call_details.get('caller', {}).get('username', 'N/A')}")
                print(f"   被叫: {call_details.get('callee', {}).get('username', 'N/A')}")
                print(f"   开始时间: {call_details.get('start_time', 'N/A')}")
            else:
                print(f"❌ 获取通话详情失败: {response.text}")
        except Exception as e:
            print(f"❌ 获取通话详情时出错: {e}")
        
        # 4. 结束通话
        print("\n4️⃣ 结束通话...")
        try:
            end_data = {"call_id": call_info["id"]}
            response = requests.post(f"{self.base_url}/api/v1/calls/end", json=end_data, headers=headers)
            if response.status_code == 200:
                end_result = response.json()["call"]
                print(f"✅ 通话结束成功")
                print(f"   通话时长: {end_result.get('duration', 0)} 秒")
            else:
                print(f"❌ 结束通话失败: {response.text}")
        except Exception as e:
            print(f"❌ 结束通话时出错: {e}")
        
        print()
    
    def demo_call_history(self):
        """演示通话历史"""
        print("📋 演示通话历史...")
        
        if len(self.demo_users) < 1:
            print("❌ 没有可用的演示用户")
            return
        
        user = self.demo_users[0]
        headers = {"Authorization": f"Bearer {user['token']}"}
        
        try:
            response = requests.get(f"{self.base_url}/api/v1/calls/history", headers=headers)
            if response.status_code == 200:
                history = response.json()
                calls = history["calls"]
                pagination = history["pagination"]
                
                print(f"✅ 通话历史获取成功")
                print(f"   总记录数: {pagination['total']}")
                print(f"   当前页: {pagination['page']}")
                print(f"   每页数量: {pagination['limit']}")
                
                if calls:
                    print("\n📞 最近的通话记录:")
                    for i, call in enumerate(calls[:3], 1):
                        print(f"   {i}. {call['call_type']} 通话")
                        print(f"      状态: {call['status']}")
                        print(f"      时间: {call.get('start_time', 'N/A')}")
                        if call.get('duration'):
                            print(f"      时长: {call['duration']} 秒")
                        print()
                else:
                    print("   暂无通话记录")
            else:
                print(f"❌ 获取通话历史失败: {response.text}")
        except Exception as e:
            print(f"❌ 获取通话历史时出错: {e}")
        
        print()
    
    def demo_websocket_info(self):
        """演示WebSocket信息"""
        print("🔌 WebSocket信令服务器信息...")
        
        print("📡 WebSocket端点:")
        print(f"   ws://localhost:8000/ws/call/{{room_id}}")
        
        print("\n📨 支持的消息类型:")
        print("   • offer - WebRTC Offer消息")
        print("   • answer - WebRTC Answer消息") 
        print("   • ice_candidate - ICE候选消息")
        print("   • join - 用户加入消息")
        print("   • leave - 用户离开消息")
        
        print("\n🔧 消息格式:")
        print("   {")
        print('     "type": "offer",')
        print('     "call_id": "room-uuid",')
        print('     "user_id": "user-uuid",')
        print('     "data": {...},')
        print('     "timestamp": 1234567890')
        print("   }")
        
        print()
    
    def open_frontend(self):
        """打开前端界面"""
        print("🌐 打开前端界面...")
        
        try:
            webbrowser.open(self.frontend_url)
            print(f"✅ 已在浏览器中打开: {self.frontend_url}")
            print("\n📱 前端功能:")
            print("   • 用户注册/登录")
            print("   • 视频通话界面")
            print("   • 实时音视频通话")
            print("   • 静音/视频开关")
            print("   • 通话历史查看")
            print("   • 安全检测状态")
        except Exception as e:
            print(f"❌ 无法打开浏览器: {e}")
            print(f"请手动访问: {self.frontend_url}")
        
        print()
    
    def print_summary(self):
        """打印功能总结"""
        print("=" * 60)
        print("🎉 真正的通话功能演示完成！")
        print("=" * 60)
        print()
        print("✅ 已实现的功能:")
        print("   📞 WebRTC P2P音视频通话")
        print("   🔌 WebSocket信令服务器")
        print("   🏠 通话房间管理")
        print("   📊 通话状态跟踪")
        print("   📋 通话历史记录")
        print("   🔒 AI安全检测")
        print("   👥 用户认证管理")
        print()
        print("🚀 技术特性:")
        print("   • 真正的点对点连接")
        print("   • 实时信令传输")
        print("   • 自动ICE候选收集")
        print("   • 连接状态监控")
        print("   • 音视频质量控制")
        print("   • 安全风险检测")
        print()
        print("🔧 下一步:")
        print("   1. 在浏览器中测试完整通话流程")
        print("   2. 测试多人通话功能")
        print("   3. 测试安全检测功能")
        print("   4. 部署到生产环境")
        print()
        print("📞 技术支持:")
        print("   • 查看日志文件排查问题")
        print("   • 运行测试脚本验证功能")
        print("   • 检查网络连接和防火墙")
        print("   • 确保浏览器支持WebRTC")
        print("=" * 60)
    
    def run_demo(self):
        """运行完整演示"""
        self.print_header()
        
        # 检查服务状态
        if not self.check_services():
            print("❌ 服务检查失败，请确保后端服务正在运行")
            return
        
        # 设置演示用户
        self.setup_demo_users()
        
        # 演示通话功能
        self.demo_call_features()
        
        # 演示通话历史
        self.demo_call_history()
        
        # 演示WebSocket信息
        self.demo_websocket_info()
        
        # 打开前端界面
        self.open_frontend()
        
        # 打印总结
        self.print_summary()

if __name__ == "__main__":
    demo = RealCallDemo()
    demo.run_demo() 