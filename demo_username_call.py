#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
基于用户名的通话功能演示
不依赖数据库，使用内存存储进行演示
"""

import json
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional

class MockUser:
    """模拟用户类"""
    users: Dict[str, "MockUser"] = {} # 存储所有用户

    def __init__(self, username: str, email: str, full_name: str, password: str):
        self.id = len(MockUser.users) + 1
        self.uuid = str(uuid.uuid4())
        self.username = username
        self.email = email
        self.full_name = full_name
        self.password_hash = self._hash_password(password)
        self.status = "active"
        self.created_at = datetime.now()
        
    def _hash_password(self, password: str) -> str:
        """简单的密码哈希（仅用于演示）"""
        return f"hash_{password}"
        
    def to_dict(self):
        return {
            "id": self.id,
            "uuid": self.uuid,
            "username": self.username,
            "email": self.email,
            "full_name": self.full_name,
            "status": self.status,
            "created_at": self.created_at.isoformat()
        }

class MockCall:
    """模拟通话类"""
    calls: List["MockCall"] = [] # 存储所有通话

    def __init__(self, caller: MockUser, callee: MockUser, call_type: str):
        self.id = len(MockCall.calls) + 1
        self.uuid = str(uuid.uuid4())
        self.caller = caller
        self.callee = callee
        self.call_type = call_type
        self.status = "initiated"
        self.start_time = datetime.now()
        self.room_id = str(uuid.uuid4())
        
    def to_dict(self):
        return {
            "id": self.id,
            "uuid": self.uuid,
            "call_type": self.call_type,
            "status": self.status,
            "room_id": self.room_id,
            "start_time": self.start_time.isoformat(),
            "caller": self.caller.to_dict(),
            "callee": self.callee.to_dict()
        }

class UsernameCallDemo:
    """基于用户名的通话功能演示"""
    
    def __init__(self):
        self.users: Dict[str, MockUser] = {}
        self.calls: List[MockCall] = []
        self.sessions: Dict[str, MockUser] = {}
        self._init_demo_data()
        
    def _init_demo_data(self):
        """初始化演示数据"""
        # 创建演示用户
        demo_users = [
            ("alice", "alice@example.com", "Alice Johnson", "password123"),
            ("bob", "bob@example.com", "Bob Smith", "password123"),
            ("charlie", "charlie@example.com", "Charlie Brown", "password123"),
            ("diana", "diana@example.com", "Diana Prince", "password123"),
            ("edward", "edward@example.com", "Edward Norton", "password123")
        ]
        
        for username, email, full_name, password in demo_users:
            user = MockUser(username, email, full_name, password)
            self.users[username] = user
            
        print("✅ 演示数据初始化完成")
        print(f"创建了 {len(self.users)} 个演示用户")
        
    def register_user(self, username: str, email: str, full_name: str, password: str) -> Dict:
        """注册用户"""
        if username in self.users:
            return {"error": "用户名已存在"}
            
        user = MockUser(username, email, full_name, password)
        self.users[username] = user
        
        return {
            "message": "用户注册成功",
            "user": user.to_dict()
        }
        
    def login_user(self, username: str, password: str) -> Dict:
        """用户登录"""
        if username not in self.users:
            return {"error": "用户不存在"}
            
        user = self.users[username]
        if user.password_hash != f"hash_{password}":
            return {"error": "密码错误"}
            
        # 创建会话token
        token = str(uuid.uuid4())
        self.sessions[token] = user
        
        return {
            "message": "登录成功",
            "token": token,
            "user": user.to_dict()
        }
        
    def search_users(self, query: str, current_user: MockUser, limit: int = 10) -> Dict:
        """搜索用户"""
        results = []
        query_lower = query.lower()
        
        for user in self.users.values():
            if user.id == current_user.id:
                continue  # 排除自己
                
            if (query_lower in user.username.lower() or 
                query_lower in user.full_name.lower()):
                results.append(user.to_dict())
                
        return {
            "users": results[:limit],
            "count": len(results)
        }
        
    def start_call(self, caller: MockUser, callee_username: str, call_type: str) -> Dict:
        """发起通话"""
        if callee_username not in self.users:
            return {"error": "被叫用户不存在"}
            
        callee = self.users[callee_username]
        if callee.id == caller.id:
            return {"error": "不能呼叫自己"}
            
        call = MockCall(caller, callee, call_type)
        self.calls.append(call)
        
        return {
            "message": "通话已发起",
            "call": call.to_dict()
        }
        
    def get_call_history(self, user: MockUser, page: int = 1, limit: int = 10) -> Dict:
        """获取通话历史"""
        user_calls = []
        for call in self.calls:
            if call.caller.id == user.id or call.callee.id == user.id:
                user_calls.append(call.to_dict())
                
        # 简单的分页
        start = (page - 1) * limit
        end = start + limit
        paginated_calls = user_calls[start:end]
        
        return {
            "calls": paginated_calls,
            "total": len(user_calls),
            "page": page,
            "limit": limit
        }
        
    def get_active_calls(self, user: MockUser) -> Dict:
        """获取活跃通话"""
        active_calls = []
        for call in self.calls:
            if (call.caller.id == user.id or call.callee.id == user.id) and call.status == "initiated":
                active_calls.append(call.to_dict())
                
        return {
            "calls": active_calls,
            "count": len(active_calls)
        }

def print_step(step: int, title: str):
    """打印步骤标题"""
    print(f"\n{'='*60}")
    print(f"步骤 {step}: {title}")
    print(f"{'='*60}")

def print_success(message: str):
    """打印成功消息"""
    print(f"✅ {message}")

def print_error(message: str):
    """打印错误消息"""
    print(f"❌ {message}")

def print_info(message: str):
    """打印信息消息"""
    print(f"ℹ️  {message}")

def demo_username_call_system():
    """演示基于用户名的通话系统"""
    print("🚀 基于用户名的通话系统演示")
    print("="*60)
    
    # 初始化演示系统
    demo = UsernameCallDemo()
    
    # 步骤1: 用户注册
    print_step(1, "用户注册演示")
    
    new_user = demo.register_user("testuser", "test@example.com", "Test User", "password123")
    if "error" not in new_user:
        print_success("新用户注册成功")
        print_info(f"用户名: {new_user['user']['username']}")
        print_info(f"全名: {new_user['user']['full_name']}")
    else:
        print_error(f"注册失败: {new_user['error']}")
    
    # 步骤2: 用户登录
    print_step(2, "用户登录演示")
    
    login_result = demo.login_user("alice", "password123")
    if "error" not in login_result:
        print_success("用户登录成功")
        current_user = demo.users["alice"]
        print_info(f"当前用户: {current_user.username} ({current_user.full_name})")
    else:
        print_error(f"登录失败: {login_result['error']}")
        return
    
    # 步骤3: 用户搜索演示
    print_step(3, "用户搜索演示")
    
    search_queries = ["bob", "charlie", "john", "smith", "test"]
    for query in search_queries:
        search_result = demo.search_users(query, current_user, 5)
        print_info(f"搜索 '{query}' 找到 {search_result['count']} 个用户")
        
        if search_result['users']:
            for user in search_result['users']:
                print(f"  - {user['username']} ({user['full_name']})")
        else:
            print(f"  - 未找到匹配的用户")
    
    # 步骤4: 基于用户名的通话演示
    print_step(4, "基于用户名的通话演示")
    
    # 演示不同类型的搜索和通话
    call_scenarios = [
        ("bob", "视频通话"),
        ("charlie", "音频通话"),
        ("diana", "视频通话")
    ]
    
    for callee_username, call_type in call_scenarios:
        call_result = demo.start_call(current_user, callee_username, call_type)
        if "error" not in call_result:
            call_info = call_result['call']
            print_success(f"成功发起{call_type}: {current_user.username} -> {callee_username}")
            print_info(f"通话ID: {call_info['id']}")
            print_info(f"房间ID: {call_info['room_id']}")
            print_info(f"状态: {call_info['status']}")
        else:
            print_error(f"发起{call_type}失败: {call_result['error']}")
    
    # 步骤5: 通话历史演示
    print_step(5, "通话历史演示")
    
    history_result = demo.get_call_history(current_user, 1, 10)
    print_info(f"通话历史总数: {history_result['total']}")
    
    for call in history_result['calls']:
        caller_name = call['caller']['username']
        callee_name = call['callee']['username']
        call_type = call['call_type']
        status = call['status']
        print(f"  - {caller_name} -> {callee_name} ({call_type}) - {status}")
    
    # 步骤6: 活跃通话演示
    print_step(6, "活跃通话演示")
    
    active_result = demo.get_active_calls(current_user)
    print_info(f"活跃通话数量: {active_result['count']}")
    
    for call in active_result['calls']:
        caller_name = call['caller']['username']
        callee_name = call['callee']['username']
        call_type = call['call_type']
        print(f"  - {caller_name} -> {callee_name} ({call_type})")
    
    # 步骤7: API接口演示
    print_step(7, "API接口演示")
    
    print_info("模拟的API接口:")
    print("  POST /api/v1/auth/register - 用户注册")
    print("  POST /api/v1/auth/login - 用户登录")
    print("  GET /api/v1/users/search - 用户搜索")
    print("  POST /api/v1/calls/start - 发起通话")
    print("  GET /api/v1/calls/history - 通话历史")
    print("  GET /api/v1/calls/active - 活跃通话")
    
    # 步骤8: 前端界面演示
    print_step(8, "前端界面演示")
    
    print_info("前端界面功能:")
    print("  📱 用户搜索界面 - 实时搜索用户")
    print("  👥 搜索结果展示 - 显示用户信息")
    print("  📞 一键通话按钮 - 快速发起通话")
    print("  🎥 视频通话界面 - WebRTC连接")
    print("  📊 通话状态显示 - 实时状态更新")
    print("  📋 通话历史记录 - 完整通话记录")
    
    print("\n" + "="*60)
    print("🎉 基于用户名的通话系统演示完成！")
    print("="*60)
    
    print("\n💡 功能特点:")
    print("  ✅ 支持通过用户名搜索用户")
    print("  ✅ 支持基于用户名发起通话")
    print("  ✅ 完整的用户认证系统")
    print("  ✅ 实时通话状态管理")
    print("  ✅ 通话历史记录功能")
    print("  ✅ 用户友好的界面设计")
    
    print("\n🔧 技术实现:")
    print("  🎯 后端: Go + Gin + WebRTC")
    print("  🌐 前端: HTML5 + JavaScript + WebRTC")
    print("  🗄️ 数据库: PostgreSQL + Redis")
    print("  🔐 安全: JWT认证 + 权限控制")
    print("  📡 通信: WebSocket + HTTP API")

if __name__ == "__main__":
    demo_username_call_system() 