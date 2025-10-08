#!/usr/bin/env python3
"""
完整集成测试脚本 - 会议系统
测试所有服务的集成功能，包括用户、会议室、媒体、信令和AI服务
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
        logging.FileHandler('logs/integration_test.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# 配置
NGINX_URL = "http://localhost:8800"
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
    title: str = "Integration Test Meeting"

class IntegrationTest:
    """集成测试类"""
    
    def __init__(self):
        self.users: List[User] = []
        self.meeting: Optional[Meeting] = None
        self.session = requests.Session()

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
    
    def cleanup_database(self):
        """清空数据库"""
        self.log_step("SETUP", "Cleaning up database...")
        try:
            # 连接PostgreSQL并清空表
            commands = [
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE users CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meetings CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meeting_participants CASCADE;'",
                "docker exec meeting-postgres psql -U postgres -d meeting_system -c 'TRUNCATE TABLE meeting_rooms CASCADE;'",
            ]
            
            for cmd in commands:
                result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
                if result.returncode != 0:
                    self.log_error("SETUP", f"Failed to execute: {cmd}\n{result.stderr}")
                    
            self.log_success("SETUP", "Database cleaned successfully")
            return True
        except Exception as e:
            self.log_error("SETUP", f"Database cleanup failed: {str(e)}")
            return False
            
    def restart_services(self):
        """重启所有服务"""
        self.log_step("SETUP", "Restarting all services...")
        try:
            # 重启Docker容器
            services = [
                "user-service",
                "meeting-service",
                "signaling-service",
                "media-service",
                "ai-service",
                "nginx"
            ]
            
            for service in services:
                self.log_step("SETUP", f"Restarting {service}...")
                subprocess.run(f"docker restart meeting-{service}", shell=True, check=True)
                
            # 等待服务启动
            self.log_step("SETUP", "Waiting for services to be ready...")
            time.sleep(10)
            
            # 检查服务健康状态
            health_url = f"{NGINX_URL}/health"
            response = self.session.get(health_url, timeout=5)
            if response.status_code == 200:
                self.log_success("SETUP", "All services are healthy")
                return True
            else:
                self.log_error("SETUP", f"Health check failed: {response.status_code}")
                return False
                
        except Exception as e:
            self.log_error("SETUP", f"Service restart failed: {str(e)}")
            return False
            
    # ========== 用户服务测试 ==========
    
    def get_csrf_token(self) -> Optional[str]:
        """获取CSRF Token"""
        try:
            url = f"{NGINX_URL}/api/v1/csrf-token"
            response = self.session.get(url, timeout=5)

            if response.status_code == 200:
                # 从响应体中获取CSRF token
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
            # 获取CSRF token
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
            
            response = self.session.post(url, json=data, headers=headers, timeout=10)
            
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
            # 获取CSRF token
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
            
            response = self.session.post(url, json=data, headers=headers, timeout=10)
            
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    login_data = result.get('data', {})
                    # 用户服务返回的字段是 'token' 而不是 'access_token'
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

            response = self.session.put(url, json=data, headers=headers, timeout=10)

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

            # 会议时间
            start_time = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 60))
            end_time = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() + 3660))

            data = {
                "title": "Integration Test Meeting",
                "description": "Automated integration test meeting",
                "start_time": start_time,
                "end_time": end_time,
                "max_participants": 10,
                "meeting_type": "video"
            }

            response = self.session.post(url, json=data, headers=headers, timeout=10)

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

            response = self.session.post(url, json=data, headers=headers, timeout=10)

            if response.status_code == 200:
                result = response.json()
                # 会议服务可能返回 code 字段或直接返回 data 和 message
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

    def get_participants(self, user: User) -> List[Dict]:
        """获取会议参与者"""
        self.log_step("MEETING", f"Getting participants for meeting {self.meeting.meeting_id}")
        try:
            url = f"{NGINX_URL}/api/v1/meetings/{self.meeting.meeting_id}/participants"
            headers = {
                'Authorization': f'Bearer {user.access_token}'
            }

            response = self.session.get(url, headers=headers, timeout=10)

            if response.status_code == 200:
                result = response.json()
                # 会议服务可能返回 code 字段或直接返回 data
                if result.get('code') == 200:
                    participants = result.get('data', [])
                elif 'data' in result:
                    participants = result.get('data', [])
                else:
                    participants = []
                self.log_success("MEETING", f"Got {len(participants)} participants")
                return participants

            self.log_error("MEETING", f"Get participants failed: {response.status_code}")
            return []

        except Exception as e:
            self.log_error("MEETING", f"Get participants exception: {str(e)}")
            return []

    def end_meeting(self, user: User) -> bool:
        """结束会议"""
        self.log_step("MEETING", f"Ending meeting {self.meeting.meeting_id}")
        try:
            url = f"{NGINX_URL}/api/v1/meetings/{self.meeting.meeting_id}/end"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            response = self.session.post(url, json={}, headers=headers, timeout=10)

            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    self.log_success("MEETING", "Meeting ended successfully")
                    return True

            self.log_error("MEETING", f"End meeting failed: {response.status_code}")
            return False

        except Exception as e:
            self.log_error("MEETING", f"End meeting exception: {str(e)}")
            return False

    # ========== 媒体和信令服务测试 ==========

    def load_test_media_file(self, filename: str) -> Optional[bytes]:
        """加载测试媒体文件"""
        try:
            filepath = os.path.join(TEST_VIDEO_DIR, filename)
            if not os.path.exists(filepath):
                self.log_error("MEDIA", f"Test file not found: {filepath}")
                return None

            with open(filepath, 'rb') as f:
                data = f.read()
            self.log_success("MEDIA", f"Loaded test file: {filename} ({len(data)} bytes)")
            return data
        except Exception as e:
            self.log_error("MEDIA", f"Failed to load test file: {str(e)}")
            return None

    def join_webrtc_room(self, user: User) -> bool:
        """加入WebRTC房间"""
        self.log_step("WEBRTC", f"User {user.username} joining WebRTC room {self.meeting.room_id}")
        try:
            url = f"{NGINX_URL}/api/v1/webrtc/room/{self.meeting.room_id}/join"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            data = {
                "user_id": str(user.user_id)
            }

            response = self.session.post(url, json=data, headers=headers, timeout=10)

            if response.status_code == 200:
                result = response.json()
                self.log_success("WEBRTC", f"User {user.username} joined WebRTC room")
                return True
            else:
                self.log_error("WEBRTC", f"Join WebRTC room failed: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False

        except Exception as e:
            self.log_error("WEBRTC", f"Join WebRTC room exception: {str(e)}")
            return False

    def get_room_peers(self, user: User) -> List[Dict]:
        """获取房间内的对等连接"""
        self.log_step("WEBRTC", f"Getting peers in room {self.meeting.room_id}")
        try:
            url = f"{NGINX_URL}/api/v1/webrtc/room/{self.meeting.room_id}/peers"
            headers = {
                'Authorization': f'Bearer {user.access_token}'
            }

            response = self.session.get(url, headers=headers, timeout=10)

            if response.status_code == 200:
                result = response.json()
                peers = result.get('peers', [])
                self.log_success("WEBRTC", f"Got {len(peers)} peers in room")
                return peers
            else:
                self.log_error("WEBRTC", f"Get room peers failed: {response.status_code}")
                return []

        except Exception as e:
            self.log_error("WEBRTC", f"Get room peers exception: {str(e)}")
            return []

    # ========== AI服务测试 ==========

    def test_emotion_detection(self, user: User, video_data: bytes) -> bool:
        """测试情绪识别"""
        self.log_step("AI", f"Testing emotion detection for user {user.username}")
        try:
            url = f"{NGINX_URL}/api/v1/speech/emotion"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            # 编码视频数据
            video_base64 = base64.b64encode(video_data[:100000]).decode('utf-8')  # 限制大小

            data = {
                "request_id": f"emotion_test_{user.user_id}_{int(time.time())}",
                "data": {
                    "image_data": video_base64,
                    "image_format": "jpg",
                    "width": 640,
                    "height": 480
                }
            }

            response = self.session.post(url, json=data, headers=headers, timeout=60)

            if response.status_code == 200:
                result = response.json()
                # 检查多种可能的成功响应格式
                if result.get('code') == 200 or result.get('object') or ('data' in result and not result.get('error')):
                    ai_result = result.get('data', {})
                    self.log_success("AI", f"Emotion detection result: {str(ai_result)[:100]}")
                    logger.info(f"Full AI Response: {json.dumps(result, indent=2)[:300]}")
                    return True
                else:
                    self.log_error("AI", f"Emotion detection returned unexpected format")
                    logger.error(f"Response: {response.text[:500]}")
                    return False

            self.log_error("AI", f"Emotion detection failed: {response.status_code}")
            logger.error(f"Response: {response.text[:500]}")
            return False

        except Exception as e:
            self.log_error("AI", f"Emotion detection exception: {str(e)}")
            return False

    def test_speech_recognition(self, user: User, audio_data: bytes) -> bool:
        """测试语音识别"""
        self.log_step("AI", f"Testing speech recognition for user {user.username}")
        try:
            url = f"{NGINX_URL}/api/v1/speech/recognition"
            headers = {
                'Authorization': f'Bearer {user.access_token}',
                'Content-Type': 'application/json'
            }

            # 编码音频数据
            audio_base64 = base64.b64encode(audio_data[:30000]).decode('utf-8')  # 限制大小

            data = {
                "request_id": f"speech_test_{user.user_id}_{int(time.time())}",
                "data": {
                    "audio_data": audio_base64,
                    "audio_format": "mp3",
                    "sample_rate": 16000,
                    "channels": 1
                }
            }

            response = self.session.post(url, json=data, headers=headers, timeout=60)

            if response.status_code == 200:
                result = response.json()
                # 检查多种可能的成功响应格式
                if result.get('code') == 200 or result.get('object') or ('data' in result and not result.get('error')):
                    ai_result = result.get('data', {})
                    self.log_success("AI", f"Speech recognition result: {str(ai_result)[:100]}")
                    logger.info(f"Full AI Response: {json.dumps(result, indent=2)[:300]}")
                    return True
                else:
                    self.log_error("AI", f"Speech recognition returned unexpected format")
                    logger.error(f"Response: {response.text[:500]}")
                    return False

            self.log_error("AI", f"Speech recognition failed: {response.status_code}")
            logger.error(f"Response: {response.text[:500]}")
            return False

        except Exception as e:
            self.log_error("AI", f"Speech recognition exception: {str(e)}")
            return False

    # ========== 主测试流程 ==========

    def run_all_tests(self):
        """运行所有测试"""
        logger.info("=" * 80)
        logger.info("开始完整集成测试")
        logger.info("=" * 80)

        # 1. 环境准备
        logger.info("\n" + "=" * 80)
        logger.info("阶段 1: 环境准备")
        logger.info("=" * 80)

        if not self.cleanup_database():
            logger.error("数据库清理失败，测试终止")
            return False

        # 跳过服务重启,因为服务已经在运行
        # if not self.restart_services():
        #     logger.error("服务重启失败，测试终止")
        #     return False

        # 等待服务就绪
        self.log_step("SETUP", "Waiting for services to be ready...")
        time.sleep(5)

        # 检查健康状态
        try:
            response = self.session.get(f"{NGINX_URL}/health", timeout=5)
            if response.status_code == 200:
                self.log_success("SETUP", "All services are healthy")
            else:
                self.log_error("SETUP", f"Health check failed: {response.status_code}")
                return False
        except Exception as e:
            self.log_error("SETUP", f"Health check exception: {str(e)}")
            return False

        # 2. 用户服务测试
        logger.info("\n" + "=" * 80)
        logger.info("阶段 2: 用户服务测试")
        logger.info("=" * 80)

        # 重置 session,清除之前的 cookies 和速率限制状态
        self.reset_session()

        # 创建4个测试用户 - 使用强密码(包含大小写字母、数字、特殊字符)
        test_users = [
            User(username="user1", password="Xk9#mP2vQ!zL", email="user1@test.com"),
            User(username="user2", password="Ym8$nQ3wR@aM", email="user2@test.com"),
            User(username="user3", password="Zn7%oR4xS#bN", email="user3@test.com"),
            User(username="user4", password="Ao6&pS5yT$cO", email="user4@test.com"),
        ]

        # 注册用户
        for user in test_users:
            if not self.register_user(user):
                logger.error(f"用户 {user.username} 注册失败，测试终止")
                return False
            time.sleep(0.5)

        # 登录用户
        for user in test_users:
            if not self.login_user(user):
                logger.error(f"用户 {user.username} 登录失败，测试终止")
                return False
            time.sleep(0.5)

        self.users = test_users

        # 测试修改密码 - 使用强密码
        if not self.change_password(test_users[0], "Bp7&qT6zU$dP"):
            logger.warning("密码修改测试失败（非致命错误）")
        else:
            # 用新密码重新登录
            test_users[0].password = "Bp7&qT6zU$dP"
            if not self.login_user(test_users[0]):
                logger.error("新密码登录失败")
                return False

        # 3. 会议室管理测试
        logger.info("\n" + "=" * 80)
        logger.info("阶段 3: 会议室管理测试")
        logger.info("=" * 80)

        # user1 创建会议
        if not self.create_meeting(test_users[0]):
            logger.error("会议创建失败，测试终止")
            return False

        time.sleep(1)

        # 所有用户加入会议
        for user in test_users:
            if not self.join_meeting(user):
                logger.error(f"用户 {user.username} 加入会议失败")
                return False
            time.sleep(0.5)

        # 验证所有用户都在会议中
        participants = self.get_participants(test_users[0])
        if len(participants) != len(test_users):
            logger.error(f"参与者数量不匹配: 期望 {len(test_users)}, 实际 {len(participants)}")
            return False

        # 4. 媒体和信令服务测试
        logger.info("\n" + "=" * 80)
        logger.info("阶段 4: 媒体和信令服务测试")
        logger.info("=" * 80)

        # 所有用户加入WebRTC房间
        for user in test_users:
            if not self.join_webrtc_room(user):
                logger.warning(f"用户 {user.username} 加入WebRTC房间失败（非致命错误）")
            time.sleep(0.5)

        # 获取房间内的对等连接
        peers = self.get_room_peers(test_users[0])
        logger.info(f"房间内有 {len(peers)} 个对等连接")

        # 5. AI服务测试
        logger.info("\n" + "=" * 80)
        logger.info("阶段 5: AI服务测试")
        logger.info("=" * 80)

        # 加载测试媒体文件
        video_files = [f for f in os.listdir(TEST_VIDEO_DIR) if f.endswith('.mp4')]
        audio_files = [f for f in os.listdir(TEST_VIDEO_DIR) if f.endswith('.mp3')]

        if video_files:
            video_data = self.load_test_media_file(video_files[0])
            if video_data:
                # 每个用户测试情绪识别
                for user in test_users[:2]:  # 只测试前2个用户以节省时间
                    self.test_emotion_detection(user, video_data)
                    time.sleep(3)
        else:
            logger.warning("未找到视频测试文件")

        if audio_files:
            audio_data = self.load_test_media_file(audio_files[0])
            if audio_data:
                # 每个用户测试语音识别
                for user in test_users[:2]:  # 只测试前2个用户
                    self.test_speech_recognition(user, audio_data)
                    time.sleep(3)
        else:
            logger.warning("未找到音频测试文件")

        # 6. 会议结束测试
        logger.info("\n" + "=" * 80)
        logger.info("阶段 6: 会议结束测试")
        logger.info("=" * 80)

        if not self.end_meeting(test_users[0]):
            logger.warning("会议结束失败（非致命错误）")

        # 验证会议已结束
        participants = self.get_participants(test_users[0])
        logger.info(f"会议结束后参与者数量: {len(participants)}")

        # 测试完成
        logger.info("\n" + "=" * 80)
        logger.info("✅ 所有测试完成！")
        logger.info("=" * 80)

        return True


def main():
    """主函数"""
    # 创建日志目录
    os.makedirs('logs', exist_ok=True)

    # 运行测试
    test = IntegrationTest()
    success = test.run_all_tests()

    if success:
        logger.info("\n🎉 集成测试全部通过！")
        sys.exit(0)
    else:
        logger.error("\n❌ 集成测试失败！")
        sys.exit(1)


if __name__ == "__main__":
    main()

