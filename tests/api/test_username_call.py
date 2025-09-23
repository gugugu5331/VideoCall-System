#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试基于用户名的通话功能
"""

import requests
import json
import time
import sys
import os

# 配置
BASE_URL = "http://localhost:8000"
API_BASE = f"{BASE_URL}/api/v1"

class UsernameCallTester:
    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.users = []
        
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
        
    def test_health_check(self):
        """测试健康检查"""
        self.print_step(1, "健康检查")
        try:
            response = self.session.get(f"{BASE_URL}/health")
            if response.status_code == 200:
                self.print_success("后端服务正常运行")
                return True
            else:
                self.print_error(f"后端服务异常: {response.status_code}")
                return False
        except Exception as e:
            self.print_error(f"连接后端服务失败: {e}")
            return False
            
    def test_register_users(self):
        """测试用户注册"""
        self.print_step(2, "注册测试用户")
        
        test_users = [
            {
                "username": "alice",
                "email": "alice@example.com",
                "password": "password123",
                "full_name": "Alice Johnson"
            },
            {
                "username": "bob",
                "email": "bob@example.com", 
                "password": "password123",
                "full_name": "Bob Smith"
            },
            {
                "username": "charlie",
                "email": "charlie@example.com",
                "password": "password123", 
                "full_name": "Charlie Brown"
            }
        ]
        
        for user_data in test_users:
            try:
                response = self.session.post(f"{API_BASE}/auth/register", json=user_data)
                if response.status_code == 201:
                    self.print_success(f"用户 {user_data['username']} 注册成功")
                    self.users.append(user_data)
                elif response.status_code == 409:
                    self.print_info(f"用户 {user_data['username']} 已存在")
                    self.users.append(user_data)
                else:
                    self.print_error(f"用户 {user_data['username']} 注册失败: {response.status_code}")
            except Exception as e:
                self.print_error(f"注册用户 {user_data['username']} 时出错: {e}")
                
        return len(self.users) > 0
        
    def test_login(self):
        """测试用户登录"""
        self.print_step(3, "用户登录")
        
        if not self.users:
            self.print_error("没有可用的测试用户")
            return False
            
        # 使用第一个用户登录
        user = self.users[0]
        login_data = {
            "username": user["username"],
            "password": user["password"]
        }
        
        try:
            response = self.session.post(f"{API_BASE}/auth/login", json=login_data)
            if response.status_code == 200:
                data = response.json()
                self.token = data.get("token")
                if self.token:
                    self.session.headers.update({"Authorization": f"Bearer {self.token}"})
                    self.print_success(f"用户 {user['username']} 登录成功")
                    return True
                else:
                    self.print_error("登录响应中没有token")
                    return False
            else:
                self.print_error(f"登录失败: {response.status_code}")
                return False
        except Exception as e:
            self.print_error(f"登录时出错: {e}")
            return False
            
    def test_search_users(self):
        """测试用户搜索功能"""
        self.print_step(4, "测试用户搜索")
        
        if not self.token:
            self.print_error("未登录，无法测试搜索功能")
            return False
            
        search_queries = ["alice", "bob", "charlie", "john", "smith"]
        
        for query in search_queries:
            try:
                response = self.session.get(f"{API_BASE}/users/search?query={query}&limit=10")
                if response.status_code == 200:
                    data = response.json()
                    users = data.get("users", [])
                    count = data.get("count", 0)
                    self.print_success(f"搜索 '{query}' 找到 {count} 个用户")
                    
                    if users:
                        for user in users:
                            self.print_info(f"  - {user.get('username')} ({user.get('full_name', 'N/A')})")
                else:
                    self.print_error(f"搜索 '{query}' 失败: {response.status_code}")
            except Exception as e:
                self.print_error(f"搜索 '{query}' 时出错: {e}")
                
        return True
        
    def test_call_by_username(self):
        """测试基于用户名的通话功能"""
        self.print_step(5, "测试基于用户名的通话")
        
        if not self.token or len(self.users) < 2:
            self.print_error("需要至少两个用户才能测试通话功能")
            return False
            
        # 使用第一个用户呼叫第二个用户
        caller = self.users[0]
        callee = self.users[1]
        
        call_data = {
            "callee_username": callee["username"],
            "call_type": "video"
        }
        
        try:
            response = self.session.post(f"{API_BASE}/calls/start", json=call_data)
            if response.status_code == 201:
                data = response.json()
                call_info = data.get("call", {})
                self.print_success(f"成功发起通话: {caller['username']} -> {callee['username']}")
                self.print_info(f"通话ID: {call_info.get('id')}")
                self.print_info(f"通话UUID: {call_info.get('uuid')}")
                self.print_info(f"房间ID: {call_info.get('room_id')}")
                
                # 获取通话详情
                call_id = call_info.get("id")
                if call_id:
                    self.test_get_call_details(call_id)
                    
                return True
            else:
                self.print_error(f"发起通话失败: {response.status_code}")
                try:
                    error_data = response.json()
                    self.print_error(f"错误信息: {error_data.get('error', 'Unknown error')}")
                except:
                    pass
                return False
        except Exception as e:
            self.print_error(f"发起通话时出错: {e}")
            return False
            
    def test_get_call_details(self, call_id):
        """测试获取通话详情"""
        self.print_step(6, f"获取通话详情 (ID: {call_id})")
        
        try:
            response = self.session.get(f"{API_BASE}/calls/{call_id}")
            if response.status_code == 200:
                data = response.json()
                call = data.get("call", {})
                self.print_success("成功获取通话详情")
                self.print_info(f"状态: {call.get('status')}")
                self.print_info(f"类型: {call.get('call_type')}")
                self.print_info(f"开始时间: {call.get('start_time')}")
                
                # 获取调用者和被叫者信息
                caller = call.get("caller", {})
                callee = call.get("callee", {})
                self.print_info(f"调用者: {caller.get('username')} ({caller.get('full_name')})")
                self.print_info(f"被叫者: {callee.get('username')} ({callee.get('full_name')})")
                
                return True
            else:
                self.print_error(f"获取通话详情失败: {response.status_code}")
                return False
        except Exception as e:
            self.print_error(f"获取通话详情时出错: {e}")
            return False
            
    def test_get_active_calls(self):
        """测试获取活跃通话"""
        self.print_step(7, "获取活跃通话列表")
        
        try:
            response = self.session.get(f"{API_BASE}/calls/active")
            if response.status_code == 200:
                data = response.json()
                calls = data.get("calls", [])
                self.print_success(f"成功获取活跃通话列表，共 {len(calls)} 个通话")
                
                for call in calls:
                    self.print_info(f"  - 通话ID: {call.get('id')}, 状态: {call.get('status')}")
                    
                return True
            else:
                self.print_error(f"获取活跃通话失败: {response.status_code}")
                return False
        except Exception as e:
            self.print_error(f"获取活跃通话时出错: {e}")
            return False
            
    def test_call_history(self):
        """测试通话历史"""
        self.print_step(8, "获取通话历史")
        
        try:
            response = self.session.get(f"{API_BASE}/calls/history?page=1&limit=10")
            if response.status_code == 200:
                data = response.json()
                calls = data.get("calls", [])
                total = data.get("total", 0)
                self.print_success(f"成功获取通话历史，共 {total} 条记录")
                
                for call in calls[:3]:  # 只显示前3条
                    self.print_info(f"  - {call.get('call_type')} 通话: {call.get('status')} ({call.get('created_at')})")
                    
                return True
            else:
                self.print_error(f"获取通话历史失败: {response.status_code}")
                return False
        except Exception as e:
            self.print_error(f"获取通话历史时出错: {e}")
            return False
            
    def run_all_tests(self):
        """运行所有测试"""
        print("🚀 开始测试基于用户名的通话功能")
        print(f"目标服务器: {BASE_URL}")
        
        tests = [
            ("健康检查", self.test_health_check),
            ("用户注册", self.test_register_users),
            ("用户登录", self.test_login),
            ("用户搜索", self.test_search_users),
            ("基于用户名的通话", self.test_call_by_username),
            ("获取活跃通话", self.test_get_active_calls),
            ("通话历史", self.test_call_history)
        ]
        
        passed = 0
        total = len(tests)
        
        for test_name, test_func in tests:
            try:
                if test_func():
                    passed += 1
                else:
                    self.print_error(f"测试 '{test_name}' 失败")
            except Exception as e:
                self.print_error(f"测试 '{test_name}' 时发生异常: {e}")
                
        print(f"\n{'='*60}")
        print(f"测试完成: {passed}/{total} 通过")
        print(f"{'='*60}")
        
        if passed == total:
            self.print_success("所有测试通过！基于用户名的通话功能正常工作")
            return True
        else:
            self.print_error(f"有 {total - passed} 个测试失败")
            return False

def main():
    """主函数"""
    tester = UsernameCallTester()
    
    try:
        success = tester.run_all_tests()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print("\n\n⏹️  测试被用户中断")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n💥 测试过程中发生未预期的错误: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main() 