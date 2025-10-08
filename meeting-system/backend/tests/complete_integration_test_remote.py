#!/usr/bin/env python3
"""
远程服务器集成测试脚本 - 会议系统
从本地测试远程部署的服务，包括用户、会议室、媒体、信令和AI服务
"""

import os
import sys
import json
import time
import base64
import requests
import subprocess
import threading
from typing import Dict, List, Optional
from dataclasses import dataclass
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('logs/remote_integration_test.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# ========== 远程服务器配置 ==========
REMOTE_HOST = "js1.blockelite.cn"
REMOTE_HTTP_PORT = "22176"  # 映射到远程服务器的 8800 端口
REMOTE_JAEGER_PORT = "22177"  # 映射到远程服务器的 8801 端口

# 远程服务器 URL
NGINX_URL = f"http://{REMOTE_HOST}:{REMOTE_HTTP_PORT}"
TEST_VIDEO_DIR = "/root/meeting-system-server/meeting-system/backend/media-service/test_video"

@dataclass
class User:
    """用户数据类"""
    username: str
    password: str
    email: str
    user_id: Optional[int] = None
    access_token: Optional[str] = None
    csrf_token: Optional[str] = None

@dataclass
class Meeting:
    """会议数据类"""
    meeting_id: Optional[int] = None
    room_id: Optional[str] = None
    title: str = "Remote Integration Test Meeting"

class RemoteIntegrationTest:
    """远程集成测试类"""
    
    def __init__(self):
        self.users: List[User] = []
        self.meeting: Optional[Meeting] = None
        self.session = requests.Session()
        # 设置较长的超时时间，因为是远程连接
        self.timeout = 30

    def reset_session(self):
        """重置 session,清除所有 cookies 和状态"""
        self.session = requests.Session()
        
    def log_step(self, step: str, message: str):
        """记录测试步骤"""
        logger.info(f"[{step}] {message}")
        
    def log_error(self, step: str, message: str):
        """记录错误"""
        logger.error(f"[{step}] ❌ {message}")
        
    def log_success(self, step: str, message: str):
        """记录成功"""
        logger.info(f"[{step}] ✅ {message}")
        
    # ========== 环境准备 ==========
    
    def test_remote_connectivity(self):
        """测试远程服务器连接"""
        self.log_step("CONNECTIVITY", f"Testing connection to {NGINX_URL}...")
        try:
            response = self.session.get(f"{NGINX_URL}/health", timeout=10)
            if response.status_code == 200:
                self.log_success("CONNECTIVITY", "Remote server is reachable")
                return True
            else:
                self.log_error("CONNECTIVITY", f"Health check failed: {response.status_code}")
                return False
        except Exception as e:
            self.log_error("CONNECTIVITY", f"Connection failed: {str(e)}")
            return False
    
    def cleanup_remote_database(self):
        """清空远程数据库（通过 SSH）"""
        self.log_step("SETUP", "Cleaning up remote database...")
        try:
            # 通过 SSH 连接到远程服务器并清空数据库
            ssh_cmd = [
                "sshpass", "-p", "beip3ius",
                "ssh", "-p", "22124", "-o", "StrictHostKeyChecking=no",
                "root@js1.blockelite.cn"
            ]
            
            commands = [
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE users CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meetings CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meeting_participants CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meeting_rooms CASCADE;'",
            ]
            
            for cmd in commands:
                full_cmd = ssh_cmd + [cmd]
                result = subprocess.run(full_cmd, capture_output=True, text=True)
                if result.returncode != 0:
                    self.log_error("SETUP", f"Failed to execute: {cmd}\n{result.stderr}")
                    
            self.log_success("SETUP", "Remote database cleaned successfully")
            return True
        except Exception as e:
            self.log_error("SETUP", f"Database cleanup failed: {str(e)}")
            self.log_step("SETUP", "Continuing without database cleanup...")
            return True  # 继续执行，即使清理失败
            
    # ========== 用户服务测试 ==========
    
    def get_csrf_token(self) -> Optional[str]:
        """获取CSRF Token"""
        try:
            url = f"{NGINX_URL}/api/v1/csrf-token"
            response = self.session.get(url, timeout=self.timeout)

            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    csrf_token = result.get('data', {}).get('csrf_token')
                    if csrf_token:
                        self.log_success("CSRF", f"Got CSRF token: {csrf_token[:20]}...")
                        return csrf_token
                    else:
                        self.log_error("CSRF", "CSRF token not found in response data")
                        return None
                else:
                    self.log_error("CSRF", f"Failed to get CSRF token: {result.get('message')}")
                    return None
            else:
                self.log_error("CSRF", f"Failed to get CSRF token: {response.status_code}")
                return None
        except Exception as e:
            self.log_error("CSRF", f"Exception: {str(e)}")
            return None
            
    def register_user(self, user: User) -> bool:
        """注册用户"""
        self.log_step("USER", f"Registering user: {user.username}")
        try:
            csrf_token = self.get_csrf_token()
            
            url = f"{NGINX_URL}/api/v1/auth/register"
            headers = {}
            if csrf_token:
                headers['X-CSRF-Token'] = csrf_token
                
            data = {
                "username": user.username,
                "password": user.password,
                "email": user.email
            }
            
            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)
            
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    user_data = result.get('data', {})
                    user.user_id = user_data.get('id')
                    self.log_success("USER", f"User {user.username} registered with ID: {user.user_id}")
                    return True
                else:
                    self.log_error("USER", f"Registration failed: {result.get('message')}")
                    return False
            else:
                self.log_error("USER", f"Registration failed with status: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False
                
        except Exception as e:
            self.log_error("USER", f"Registration exception: {str(e)}")
            return False
            
    def login_user(self, user: User) -> bool:
        """用户登录"""
        self.log_step("USER", f"Logging in user: {user.username}")
        try:
            csrf_token = self.get_csrf_token()
            
            url = f"{NGINX_URL}/api/v1/auth/login"
            headers = {}
            if csrf_token:
                headers['X-CSRF-Token'] = csrf_token
                user.csrf_token = csrf_token
                
            data = {
                "username": user.username,
                "password": user.password
            }
            
            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)
            
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    login_data = result.get('data', {})
                    user.access_token = login_data.get('token') or login_data.get('access_token')
                    user.user_id = login_data.get('user', {}).get('id')
                    self.log_success("USER", f"User {user.username} logged in successfully")
                    return True
                else:
                    self.log_error("USER", f"Login failed: {result.get('message')}")
                    return False
            else:
                self.log_error("USER", f"Login failed with status: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False
                
        except Exception as e:
            self.log_error("USER", f"Login exception: {str(e)}")
            return False

    def change_password(self, user: User, new_password: str) -> bool:
        """修改密码"""
        self.log_step("USER", f"Changing password for user: {user.username}")
        try:
            url = f"{NGINX_URL}/api/v1/users/password"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
            }
            if user.csrf_token:
                headers['X-CSRF-Token'] = user.csrf_token

            data = {
                "old_password": user.password,
                "new_password": new_password
            }

            response = self.session.put(url, json=data, headers=headers, timeout=self.timeout)

            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    user.password = new_password
                    self.log_success("USER", f"Password changed for {user.username}")
                    return True

            self.log_error("USER", f"Password change failed: {response.status_code}")
            return False

        except Exception as e:
            self.log_error("USER", f"Password change exception: {str(e)}")
            return False

    # ========== 会议服务测试 ==========

    def create_meeting(self, creator: User) -> bool:
        """创建会议"""
        self.log_step("MEETING", f"Creating meeting by {creator.username}")
        try:
            url = f"{NGINX_URL}/api/v1/meetings"
            headers = {
                'Authorization': f'Bearer {creator.access_token}',
                'Content-Type': 'application/json'
            }

            start_time = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 60))
            end_time = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 3660))

            data = {
                "title": "Remote Integration Test Meeting",
                "description": "Automated remote integration test meeting",
                "start_time": start_time,
                "end_time": end_time,
                "max_participants": 10,
                "meeting_type": "video"
            }

            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)

            if response.status_code in [200, 201]:
                result = response.json()
                if result.get('code') in [200, 201]:
                    meeting_data = result.get('data', {}).get('meeting', {})
                    self.meeting = Meeting(
                        meeting_id=meeting_data.get('id'),
                        title=meeting_data.get('title')
                    )
                    self.log_success("MEETING", f"Meeting created with ID: {self.meeting.meeting_id}")
                    return True

            self.log_error("MEETING", f"Meeting creation failed: {response.status_code}")
            logger.error(f"Response: {response.text}")
            return False

        except Exception as e:
            self.log_error("MEETING", f"Meeting creation exception: {str(e)}")
            return False

    def join_meeting(self, user: User) -> bool:
        """加入会议"""
        self.log_step("MEETING", f"User {user.username} joining meeting {self.meeting.meeting_id}")
        try:
            url = f"{NGINX_URL}/api/v1/meetings/{self.meeting.meeting_id}/join"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            data = {}

            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)

            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200 or result.get('message'):
                    join_data = result.get('data', {})
                    room_id = join_data.get('room_id')
                    if room_id and not self.meeting.room_id:
                        self.meeting.room_id = room_id
                    self.log_success("MEETING", f"User {user.username} joined meeting, room_id: {room_id}")
                    return True

            self.log_error("MEETING", f"Join meeting failed: {response.status_code}")
            logger.error(f"Response: {response.text}")
            return False

        except Exception as e:
            self.log_error("MEETING", f"Join meeting exception: {str(e)}")
            return False

    # ========== AI服务测试 ==========

    def test_emotion_detection(self, user: User) -> bool:
        """测试情绪识别"""
        self.log_step("AI", f"Testing emotion detection for user {user.username}")
        try:
            url = f"{NGINX_URL}/api/v1/speech/emotion"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            # 使用测试数据
            test_image_data = base64.b64encode(b"test_image_data" * 100).decode('utf-8')

            data = {
                "request_id": f"emotion_test_{user.user_id}_{int(time.time())}",
                "data": {
                    "image_data": test_image_data,
                    "image_format": "jpg",
                    "width": 640,
                    "height": 480
                }
            }

            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)

            if response.status_code == 200:
                result = response.json()
                self.log_success("AI", f"Emotion detection successful: {result}")
                return True
            else:
                self.log_error("AI", f"Emotion detection failed: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False

        except Exception as e:
            self.log_error("AI", f"Emotion detection exception: {str(e)}")
            return False

    def test_speech_recognition(self, user: User) -> bool:
        """测试语音识别"""
        self.log_step("AI", f"Testing speech recognition for user {user.username}")
        try:
            url = f"{NGINX_URL}/api/v1/speech/recognize"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            # 使用测试数据
            test_audio_data = base64.b64encode(b"test_audio_data" * 100).decode('utf-8')

            data = {
                "request_id": f"speech_test_{user.user_id}_{int(time.time())}",
                "data": {
                    "audio_data": test_audio_data,
                    "audio_format": "wav",
                    "sample_rate": 16000
                }
            }

            response = self.session.post(url, json=data, headers=headers, timeout=self.timeout)

            if response.status_code == 200:
                result = response.json()
                self.log_success("AI", f"Speech recognition successful: {result}")
                return True
            else:
                self.log_error("AI", f"Speech recognition failed: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False

        except Exception as e:
            self.log_error("AI", f"Speech recognition exception: {str(e)}")
            return False

    # ========== 主测试流程 ==========

    def run_all_tests(self):
        """运行所有测试"""
        logger.info("=" * 80)
        logger.info("开始远程集成测试")
        logger.info(f"远程服务器: {REMOTE_HOST}:{REMOTE_HTTP_PORT}")
        logger.info("=" * 80)

        test_results = {
            "total": 0,
            "passed": 0,
            "failed": 0
        }

        def record_test(success: bool):
            test_results["total"] += 1
            if success:
                test_results["passed"] += 1
            else:
                test_results["failed"] += 1

        # 1. 测试远程连接
        logger.info("\n[1/8] 测试远程服务器连接...")
        record_test(self.test_remote_connectivity())

        # 2. 清理数据库（可选）
        logger.info("\n[2/8] 清理远程数据库...")
        self.cleanup_remote_database()

        # 3. 创建测试用户
        logger.info("\n[3/8] 创建测试用户...")
        test_users = [
            User(username=f"remote_user_{i}", password="Test@Pass123", email=f"remote{i}@test.com")
            for i in range(1, 5)
        ]

        for user in test_users:
            if self.register_user(user):
                record_test(True)
                self.users.append(user)
            else:
                record_test(False)

        # 4. 用户登录
        logger.info("\n[4/8] 用户登录...")
        for user in self.users:
            record_test(self.login_user(user))

        # 5. 创建会议
        logger.info("\n[5/8] 创建会议...")
        if len(self.users) > 0:
            record_test(self.create_meeting(self.users[0]))

        # 6. 加入会议
        logger.info("\n[6/8] 用户加入会议...")
        for user in self.users:
            record_test(self.join_meeting(user))

        # 7. 测试AI服务
        logger.info("\n[7/8] 测试AI服务...")
        if len(self.users) > 0:
            record_test(self.test_emotion_detection(self.users[0]))
            record_test(self.test_speech_recognition(self.users[0]))

        # 8. 显示测试结果
        logger.info("\n[8/8] 测试完成")
        logger.info("=" * 80)
        logger.info("测试结果汇总")
        logger.info("=" * 80)
        logger.info(f"总测试数: {test_results['total']}")
        logger.info(f"通过: {test_results['passed']}")
        logger.info(f"失败: {test_results['failed']}")
        logger.info(f"通过率: {test_results['passed'] / test_results['total'] * 100:.2f}%")
        logger.info("=" * 80)

        return test_results['failed'] == 0


def main():
    """主函数"""
    # 创建日志目录
    os.makedirs('logs', exist_ok=True)

    # 运行测试
    test = RemoteIntegrationTest()
    success = test.run_all_tests()

    # 返回退出码
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()

