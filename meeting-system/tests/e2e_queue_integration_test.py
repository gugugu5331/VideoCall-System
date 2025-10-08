#!/usr/bin/env python3
"""
端到端消息队列集成测试脚本
测试场景：三个用户注册、加入同一会议室并调用 AI 服务
"""

import requests
import json
import time
import redis
from datetime import datetime
from typing import Dict, Optional, List
import sys

# 配置
NGINX_URL = "http://localhost:8800"
API_BASE = f"{NGINX_URL}/api/v1"
REDIS_HOST = "localhost"
REDIS_PORT = 6379

# 测试用户数据
USERS = [
    {"username": "test_user_1", "email": "user1@test.com", "password": "Test@Pass123"},
    {"username": "test_user_2", "email": "user2@test.com", "password": "Test@Pass123"},
    {"username": "test_user_3", "email": "user3@test.com", "password": "Test@Pass123"},
]

# 颜色输出
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'

class E2ETest:
    def __init__(self):
        self.redis_client = None
        self.tokens = {}
        self.meeting_id = None
        self.test_results = []
        self.start_time = datetime.now()
        self.log_file = f"e2e_test_{self.start_time.strftime('%Y%m%d_%H%M%S')}.log"
        self.report_file = f"e2e_test_report_{self.start_time.strftime('%Y%m%d_%H%M%S')}.md"
        self.session = requests.Session()  # 使用 session 保持 cookies
        self.csrf_token = None
        
    def log(self, message: str, level: str = "INFO"):
        """记录日志"""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        color = {
            "INFO": Colors.BLUE,
            "SUCCESS": Colors.GREEN,
            "ERROR": Colors.RED,
            "WARNING": Colors.YELLOW
        }.get(level, Colors.NC)
        
        log_msg = f"[{timestamp}] [{level}] {message}"
        print(f"{color}{log_msg}{Colors.NC}")
        
        with open(self.log_file, 'a') as f:
            f.write(log_msg + '\n')
    
    def get_csrf_token(self) -> bool:
        """获取 CSRF token"""
        try:
            response = self.session.get(f"{API_BASE}/csrf-token", timeout=5)
            if response.status_code == 200:
                data = response.json()
                self.csrf_token = data.get('data', {}).get('csrf_token')
                if self.csrf_token:
                    self.log(f"获取 CSRF token 成功", "SUCCESS")
                    return True
            self.log("获取 CSRF token 失败", "WARNING")
            return False
        except Exception as e:
            self.log(f"获取 CSRF token 异常: {e}", "WARNING")
            return False

    def check_services(self) -> bool:
        """检查服务状态"""
        self.log("检查服务状态...")

        # 检查 Redis
        try:
            self.redis_client = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
            self.redis_client.ping()
            self.log("Redis 运行正常", "SUCCESS")
        except Exception as e:
            self.log(f"Redis 连接失败: {e}", "ERROR")
            return False

        # 检查 Nginx/API
        try:
            response = self.session.get(f"{NGINX_URL}/health", timeout=5)
            self.log("Nginx 运行正常", "SUCCESS")
        except Exception as e:
            self.log(f"Nginx 可能未运行: {e}", "WARNING")

        # 获取 CSRF token
        self.get_csrf_token()

        return True
    
    def check_redis_queues(self) -> Dict[str, int]:
        """检查 Redis 队列状态"""
        self.log("检查 Redis 队列状态...")
        
        queues = {
            "critical_queue": self.redis_client.llen("meeting_system:critical_queue") or 0,
            "high_queue": self.redis_client.llen("meeting_system:high_queue") or 0,
            "normal_queue": self.redis_client.llen("meeting_system:normal_queue") or 0,
            "low_queue": self.redis_client.llen("meeting_system:low_queue") or 0,
            "dead_letter_queue": self.redis_client.llen("meeting_system:dead_letter_queue") or 0,
        }
        
        for queue_name, length in queues.items():
            self.log(f"  {queue_name}: {length}")
        
        return queues
    
    def register_user(self, user: Dict) -> bool:
        """注册用户"""
        self.log(f"注册用户: {user['username']}")

        try:
            headers = {"Content-Type": "application/json"}
            if self.csrf_token:
                headers["X-CSRF-Token"] = self.csrf_token

            response = self.session.post(
                f"{API_BASE}/auth/register",
                json=user,
                headers=headers,
                timeout=10
            )

            self.log(f"  Response: {response.status_code} - {response.text[:200]}")

            if response.status_code in [200, 201]:
                self.log(f"用户 {user['username']} 注册成功", "SUCCESS")
                self.test_results.append({
                    "step": "register",
                    "user": user['username'],
                    "status": "success",
                    "response": response.json() if response.text else {}
                })
                return True
            else:
                self.log(f"用户 {user['username']} 注册失败: {response.text}", "WARNING")
                self.test_results.append({
                    "step": "register",
                    "user": user['username'],
                    "status": "failed",
                    "error": response.text
                })
                return False
        except Exception as e:
            self.log(f"注册用户异常: {e}", "ERROR")
            return False
    
    def login_user(self, user: Dict) -> Optional[str]:
        """用户登录"""
        self.log(f"用户登录: {user['username']}")

        try:
            headers = {"Content-Type": "application/json"}
            if self.csrf_token:
                headers["X-CSRF-Token"] = self.csrf_token

            response = self.session.post(
                f"{API_BASE}/auth/login",
                json={"username": user['username'], "password": user['password']},
                headers=headers,
                timeout=10
            )

            self.log(f"  Response: {response.status_code} - {response.text[:200]}")

            if response.status_code == 200:
                data = response.json()
                token = data.get('token') or data.get('data', {}).get('token')

                if token:
                    self.log(f"用户 {user['username']} 登录成功", "SUCCESS")
                    self.tokens[user['username']] = token
                    self.test_results.append({
                        "step": "login",
                        "user": user['username'],
                        "status": "success"
                    })
                    return token

            self.log(f"用户 {user['username']} 登录失败: {response.text}", "WARNING")
            return None
        except Exception as e:
            self.log(f"登录用户异常: {e}", "ERROR")
            return None
    
    def create_meeting(self, token: str, title: str) -> Optional[int]:
        """创建会议"""
        self.log(f"创建会议: {title}")
        
        try:
            response = requests.post(
                f"{API_BASE}/meetings",
                json={
                    "title": title,
                    "description": "E2E Test Meeting",
                    "start_time": datetime.utcnow().isoformat() + "Z"
                },
                headers={"Authorization": f"Bearer {token}"},
                timeout=10
            )
            
            self.log(f"  Response: {response.status_code} - {response.text[:200]}")
            
            if response.status_code in [200, 201]:
                data = response.json()
                meeting_id = data.get('id') or data.get('data', {}).get('id')
                
                if meeting_id:
                    self.log(f"会议创建成功，ID: {meeting_id}", "SUCCESS")
                    self.meeting_id = meeting_id
                    self.test_results.append({
                        "step": "create_meeting",
                        "status": "success",
                        "meeting_id": meeting_id
                    })
                    return meeting_id
            
            self.log(f"会议创建失败: {response.text}", "WARNING")
            return None
        except Exception as e:
            self.log(f"创建会议异常: {e}", "ERROR")
            return None
    
    def join_meeting(self, token: str, meeting_id: int, username: str) -> bool:
        """加入会议"""
        self.log(f"用户 {username} 加入会议 {meeting_id}")
        
        try:
            response = requests.post(
                f"{API_BASE}/meetings/{meeting_id}/join",
                headers={"Authorization": f"Bearer {token}"},
                timeout=10
            )
            
            self.log(f"  Response: {response.status_code} - {response.text[:200]}")
            
            if response.status_code in [200, 201]:
                self.log(f"用户 {username} 成功加入会议", "SUCCESS")
                self.test_results.append({
                    "step": "join_meeting",
                    "user": username,
                    "meeting_id": meeting_id,
                    "status": "success"
                })
                return True
            else:
                self.log(f"用户 {username} 加入会议响应: {response.text}", "WARNING")
                return False
        except Exception as e:
            self.log(f"加入会议异常: {e}", "ERROR")
            return False
    
    def call_ai_service(self, token: str, service_type: str, username: str) -> bool:
        """调用 AI 服务"""
        self.log(f"用户 {username} 调用 AI 服务: {service_type}")
        
        endpoints = {
            "speech_recognition": f"{API_BASE}/ai/speech-recognition",
            "emotion_detection": f"{API_BASE}/ai/emotion-detection",
            "audio_denoising": f"{API_BASE}/ai/audio-denoising"
        }
        
        payload = {
            "audio_data": "base64_encoded_audio_data",
            "language": "zh-CN"
        }
        
        try:
            response = requests.post(
                endpoints.get(service_type, endpoints["speech_recognition"]),
                json=payload,
                headers={"Authorization": f"Bearer {token}"},
                timeout=10
            )
            
            self.log(f"  Response: {response.status_code} - {response.text[:200]}")
            
            if response.status_code in [200, 201]:
                self.log(f"AI 服务 {service_type} 调用成功", "SUCCESS")
                self.test_results.append({
                    "step": "ai_service",
                    "user": username,
                    "service": service_type,
                    "status": "success"
                })
                return True
            else:
                self.log(f"AI 服务 {service_type} 响应: {response.text}", "WARNING")
                return False
        except Exception as e:
            self.log(f"调用 AI 服务异常: {e}", "ERROR")
            return False
    
    def generate_report(self):
        """生成测试报告"""
        self.log("生成测试报告...")
        
        end_time = datetime.now()
        duration = (end_time - self.start_time).total_seconds()
        
        # 统计结果
        total_tests = len(self.test_results)
        success_tests = len([r for r in self.test_results if r.get('status') == 'success'])
        
        report = f"""# 端到端消息队列集成测试报告

**测试时间**: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}
**测试时长**: {duration:.2f} 秒
**测试结果**: {success_tests}/{total_tests} 成功

## 测试场景

三个用户注册、加入同一会议室并调用 AI 服务

## 测试步骤和结果

### 1. 用户注册阶段
"""
        
        for user in USERS:
            result = next((r for r in self.test_results if r.get('step') == 'register' and r.get('user') == user['username']), None)
            status = "✅" if result and result.get('status') == 'success' else "❌"
            report += f"- {status} {user['username']} ({user['email']})\n"
        
        report += "\n### 2. 用户登录阶段\n"
        for user in USERS:
            result = next((r for r in self.test_results if r.get('step') == 'login' and r.get('user') == user['username']), None)
            status = "✅" if result and result.get('status') == 'success' else "❌"
            report += f"- {status} {user['username']}\n"
        
        report += f"\n### 3. 创建会议室阶段\n"
        meeting_result = next((r for r in self.test_results if r.get('step') == 'create_meeting'), None)
        if meeting_result:
            status = "✅" if meeting_result.get('status') == 'success' else "❌"
            report += f"- {status} 会议 ID: {self.meeting_id}\n"
        
        report += "\n### 4. 用户加入会议阶段\n"
        for user in USERS:
            result = next((r for r in self.test_results if r.get('step') == 'join_meeting' and r.get('user') == user['username']), None)
            status = "✅" if result and result.get('status') == 'success' else "❌"
            report += f"- {status} {user['username']}\n"
        
        report += "\n### 5. 调用 AI 服务阶段\n"
        ai_services = ["speech_recognition", "emotion_detection", "audio_denoising"]
        for i, user in enumerate(USERS):
            service = ai_services[i] if i < len(ai_services) else ai_services[0]
            result = next((r for r in self.test_results if r.get('step') == 'ai_service' and r.get('user') == user['username']), None)
            status = "✅" if result and result.get('status') == 'success' else "❌"
            report += f"- {status} {user['username']}: {service}\n"
        
        # Redis 队列统计
        final_queues = self.check_redis_queues()
        report += f"\n## Redis 队列统计\n\n```\n"
        for queue_name, length in final_queues.items():
            report += f"{queue_name}: {length}\n"
        report += "```\n"
        
        report += f"\n## 测试结论\n\n"
        report += f"- 总测试数: {total_tests}\n"
        report += f"- 成功: {success_tests}\n"
        report += f"- 失败: {total_tests - success_tests}\n"
        report += f"- 成功率: {(success_tests/total_tests*100):.2f}%\n\n"
        
        report += "## 建议\n\n"
        report += "1. 检查各服务日志，确认消息队列系统正常工作\n"
        report += "2. 验证事件流转是否符合预期\n"
        report += "3. 监控 Redis 队列长度和死信队列\n"
        report += f"\n详细日志请查看: {self.log_file}\n"
        
        with open(self.report_file, 'w') as f:
            f.write(report)
        
        self.log(f"测试报告已生成: {self.report_file}", "SUCCESS")
        print(f"\n{report}")
    
    def run(self):
        """运行测试"""
        self.log("=" * 50)
        self.log("开始端到端消息队列集成测试")
        self.log("=" * 50)
        
        # 检查服务
        if not self.check_services():
            self.log("服务检查失败，退出测试", "ERROR")
            return
        
        # 初始队列状态
        self.log("=== 初始队列状态 ===")
        self.check_redis_queues()
        
        # 阶段 1: 用户注册
        self.log("=== 阶段 1: 用户注册 ===")
        for user in USERS:
            self.register_user(user)
            time.sleep(1)
        time.sleep(2)
        self.check_redis_queues()
        
        # 阶段 2: 用户登录
        self.log("=== 阶段 2: 用户登录 ===")
        for user in USERS:
            self.login_user(user)
            time.sleep(1)
        time.sleep(2)
        
        # 阶段 3: 创建会议
        self.log("=== 阶段 3: 创建会议 ===")
        if USERS[0]['username'] in self.tokens:
            self.create_meeting(self.tokens[USERS[0]['username']], "E2E Test Meeting")
            time.sleep(2)
            self.check_redis_queues()
        
        # 阶段 4: 加入会议
        self.log("=== 阶段 4: 用户加入会议 ===")
        if self.meeting_id:
            for user in USERS:
                if user['username'] in self.tokens:
                    self.join_meeting(self.tokens[user['username']], self.meeting_id, user['username'])
                    time.sleep(1)
            time.sleep(2)
            self.check_redis_queues()
        
        # 阶段 5: 调用 AI 服务
        self.log("=== 阶段 5: 调用 AI 服务 ===")
        ai_services = ["speech_recognition", "emotion_detection", "audio_denoising"]
        for i, user in enumerate(USERS):
            if user['username'] in self.tokens:
                service = ai_services[i] if i < len(ai_services) else ai_services[0]
                self.call_ai_service(self.tokens[user['username']], service, user['username'])
                time.sleep(1)
        time.sleep(2)
        
        # 最终队列状态
        self.log("=== 最终队列状态 ===")
        self.check_redis_queues()
        
        # 生成报告
        self.generate_report()
        
        self.log("=" * 50)
        self.log("测试完成！")
        self.log(f"日志文件: {self.log_file}")
        self.log(f"报告文件: {self.report_file}")
        self.log("=" * 50)

if __name__ == "__main__":
    test = E2ETest()
    test.run()

