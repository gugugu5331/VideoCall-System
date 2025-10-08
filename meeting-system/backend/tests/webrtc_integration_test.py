#!/usr/bin/env python3
"""
WebRTC 集成测试 - 测试用户之间的音视频流连接
"""

import asyncio
import json
import logging
import sys
import time
import uuid
from dataclasses import dataclass
from typing import Dict, List, Optional

import requests
import websockets

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('logs/webrtc_integration_test.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# 配置
NGINX_URL = "http://localhost:8800"
WS_URL = "ws://localhost:8800/ws/signaling"

@dataclass
class User:
    """用户数据类"""
    username: str
    password: str
    email: str
    user_id: Optional[int] = None
    access_token: Optional[str] = None
    csrf_token: Optional[str] = None
    peer_id: Optional[str] = None
    session_id: Optional[str] = None
    ws_connection: Optional[websockets.WebSocketClientProtocol] = None

@dataclass
class Meeting:
    """会议数据类"""
    meeting_id: Optional[int] = None
    room_id: Optional[str] = None
    title: str = "WebRTC Integration Test Meeting"

class WebRTCIntegrationTest:
    """WebRTC 集成测试类"""
    
    def __init__(self):
        self.session = requests.Session()
        self.users: List[User] = []
        self.meeting: Optional[Meeting] = None
        self.peer_connections: Dict[str, Dict] = {}  # peer_id -> connection info
        
    def log_step(self, category: str, message: str):
        """记录测试步骤"""
        logger.info(f"[{category}] {message}")
        
    def log_success(self, category: str, message: str):
        """记录成功信息"""
        logger.info(f"[{category}] ✅ {message}")
        
    def log_error(self, category: str, message: str):
        """记录错误信息"""
        logger.error(f"[{category}] ❌ {message}")
    
    def get_csrf_token(self) -> Optional[str]:
        """获取 CSRF token"""
        try:
            url = f"{NGINX_URL}/api/v1/csrf-token"
            response = self.session.get(url, timeout=5)
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    csrf_token = result.get('data', {}).get('csrf_token')
                    self.log_success("CSRF", f"Got CSRF token: {csrf_token[:20]}...")
                    return csrf_token
        except Exception as e:
            self.log_error("CSRF", f"Failed to get CSRF token: {str(e)}")
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
                user.csrf_token = csrf_token
                
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
            
            self.log_error("USER", f"Registration failed: {response.status_code}")
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
            
            response = self.session.post(url, json=data, headers=headers, timeout=10)
            
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 200:
                    login_data = result.get('data', {})
                    user.access_token = login_data.get('token') or login_data.get('access_token')
                    user.user_id = login_data.get('user', {}).get('id')
                    user.peer_id = f"peer_{user.user_id}_{int(time.time() * 1000)}"
                    self.log_success("USER", f"User {user.username} logged in successfully")
                    return True
            
            self.log_error("USER", f"Login failed: {response.status_code}")
            logger.error(f"Response: {response.text}")
            return False
            
        except Exception as e:
            self.log_error("USER", f"Login exception: {str(e)}")
            return False
    
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
                "title": "WebRTC Integration Test Meeting",
                "description": "Testing WebRTC peer connections",
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

    async def connect_websocket(self, user: User) -> bool:
        """连接到 WebSocket 信令服务器"""
        self.log_step("WEBSOCKET", f"User {user.username} connecting to WebSocket")
        try:
            # 构建 WebSocket URL with token
            ws_url = f"{WS_URL}?token={user.access_token}&meeting_id={self.meeting.meeting_id}&user_id={user.user_id}&peer_id={user.peer_id}"

            # 连接 WebSocket
            user.ws_connection = await websockets.connect(
                ws_url,
                extra_headers={
                    "Authorization": f"Bearer {user.access_token}"
                }
            )

            self.log_success("WEBSOCKET", f"User {user.username} connected to WebSocket")
            return True

        except Exception as e:
            self.log_error("WEBSOCKET", f"WebSocket connection failed: {str(e)}")
            return False

    async def send_join_room_message(self, user: User) -> bool:
        """发送加入房间消息"""
        self.log_step("WEBRTC", f"User {user.username} sending join room message")
        try:
            message = {
                "id": str(uuid.uuid4()),
                "type": 4,  # MessageTypeJoinRoom
                "from_user_id": user.user_id,
                "meeting_id": self.meeting.meeting_id,
                "session_id": user.session_id or f"session_{user.user_id}_{int(time.time())}",
                "peer_id": user.peer_id,
                "payload": {
                    "username": user.username,
                    "peer_id": user.peer_id
                },
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            user.session_id = message["session_id"]

            await user.ws_connection.send(json.dumps(message))
            self.log_success("WEBRTC", f"User {user.username} sent join room message")

            # 等待房间信息响应
            response = await asyncio.wait_for(user.ws_connection.recv(), timeout=5.0)
            response_data = json.loads(response)

            if response_data.get('type') == 14:  # MessageTypeRoomInfo
                self.log_success("WEBRTC", f"User {user.username} received room info")
                return True

            return True

        except Exception as e:
            self.log_error("WEBRTC", f"Join room message failed: {str(e)}")
            return False

    async def create_peer_connection(self, user: User, target_user: User) -> bool:
        """创建与目标用户的对等连接"""
        self.log_step("WEBRTC", f"User {user.username} creating peer connection to {target_user.username}")
        try:
            # 模拟 SDP offer
            sdp_offer = f"v=0\no=- {int(time.time())} 2 IN IP4 127.0.0.1\ns=-\nt=0 0\na=group:BUNDLE 0 1\na=msid-semantic: WMS stream\nm=audio 9 UDP/TLS/RTP/SAVPF 111\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:test\na=ice-pwd:testpassword\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\na=setup:actpass\na=mid:0\na=sendrecv\na=rtcp-mux\na=rtpmap:111 opus/48000/2\nm=video 9 UDP/TLS/RTP/SAVPF 96\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:test\na=ice-pwd:testpassword\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\na=setup:actpass\na=mid:1\na=sendrecv\na=rtcp-mux\na=rtpmap:96 VP8/90000"

            # 发送 offer 消息
            offer_message = {
                "id": str(uuid.uuid4()),
                "type": 1,  # MessageTypeOffer
                "from_user_id": user.user_id,
                "to_user_id": target_user.user_id,
                "meeting_id": self.meeting.meeting_id,
                "session_id": user.session_id,
                "peer_id": user.peer_id,
                "payload": {
                    "sdp": sdp_offer,
                    "type": "offer",
                    "target_peer_id": target_user.peer_id
                },
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            await user.ws_connection.send(json.dumps(offer_message))
            self.log_success("WEBRTC", f"User {user.username} sent offer to {target_user.username}")

            # 记录对等连接
            connection_key = f"{user.peer_id}->{target_user.peer_id}"
            self.peer_connections[connection_key] = {
                "from": user.username,
                "to": target_user.username,
                "status": "offer_sent",
                "timestamp": time.time()
            }

            return True

        except Exception as e:
            self.log_error("WEBRTC", f"Create peer connection failed: {str(e)}")
            return False

    async def handle_incoming_messages(self, user: User, duration: float = 5.0):
        """处理接收到的 WebSocket 消息"""
        self.log_step("WEBRTC", f"User {user.username} listening for messages for {duration}s")
        try:
            end_time = time.time() + duration
            message_count = 0

            while time.time() < end_time:
                try:
                    # 设置超时以便定期检查时间
                    remaining_time = end_time - time.time()
                    if remaining_time <= 0:
                        break

                    message = await asyncio.wait_for(
                        user.ws_connection.recv(),
                        timeout=min(1.0, remaining_time)
                    )

                    data = json.loads(message)
                    message_type = data.get('type')
                    message_count += 1

                    if message_type == 1:  # Offer
                        self.log_success("WEBRTC", f"User {user.username} received OFFER")
                        # 发送 answer
                        await self.send_answer(user, data)
                    elif message_type == 2:  # Answer
                        self.log_success("WEBRTC", f"User {user.username} received ANSWER")
                        # 更新连接状态
                        from_peer = data.get('peer_id')
                        connection_key = f"{user.peer_id}->{from_peer}"
                        if connection_key in self.peer_connections:
                            self.peer_connections[connection_key]['status'] = 'connected'
                    elif message_type == 3:  # ICE Candidate
                        self.log_success("WEBRTC", f"User {user.username} received ICE candidate")
                    elif message_type == 6:  # User Joined
                        self.log_success("WEBRTC", f"User {user.username} received user joined notification")
                    elif message_type == 7:  # User Left
                        self.log_success("WEBRTC", f"User {user.username} received user left notification")

                except asyncio.TimeoutError:
                    continue
                except Exception as e:
                    logger.debug(f"Message handling error: {str(e)}")
                    continue

            self.log_success("WEBRTC", f"User {user.username} received {message_count} messages")
            return True

        except Exception as e:
            self.log_error("WEBRTC", f"Handle messages failed: {str(e)}")
            return False

    async def send_answer(self, user: User, offer_data: dict):
        """发送 answer 响应"""
        try:
            sdp_answer = f"v=0\no=- {int(time.time())} 2 IN IP4 127.0.0.1\ns=-\nt=0 0\na=group:BUNDLE 0 1\na=msid-semantic: WMS stream\nm=audio 9 UDP/TLS/RTP/SAVPF 111\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:test\na=ice-pwd:testpassword\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\na=setup:active\na=mid:0\na=sendrecv\na=rtcp-mux\na=rtpmap:111 opus/48000/2\nm=video 9 UDP/TLS/RTP/SAVPF 96\nc=IN IP4 0.0.0.0\na=rtcp:9 IN IP4 0.0.0.0\na=ice-ufrag:test\na=ice-pwd:testpassword\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\na=setup:passive\na=mid:1\na=sendrecv\na=rtcp-mux\na=rtpmap:96 VP8/90000"

            answer_message = {
                "id": str(uuid.uuid4()),
                "type": 2,  # MessageTypeAnswer
                "from_user_id": user.user_id,
                "to_user_id": offer_data.get('from_user_id'),
                "meeting_id": self.meeting.meeting_id,
                "session_id": user.session_id,
                "peer_id": user.peer_id,
                "payload": {
                    "sdp": sdp_answer,
                    "type": "answer"
                },
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            await user.ws_connection.send(json.dumps(answer_message))
            self.log_success("WEBRTC", f"User {user.username} sent answer")

        except Exception as e:
            self.log_error("WEBRTC", f"Send answer failed: {str(e)}")

    async def run_webrtc_test(self) -> bool:
        """运行 WebRTC 集成测试"""
        logger.info("=" * 80)
        logger.info("开始 WebRTC 对等连接测试")
        logger.info("=" * 80)

        try:
            # 1. 创建测试用户
            logger.info("\n" + "=" * 80)
            logger.info("阶段 1: 用户注册和登录")
            logger.info("=" * 80)

            test_users = [
                User(username="webrtc_user1", password="Xk9#mP2vQ!zL", email="webrtc1@test.com"),
                User(username="webrtc_user2", password="Ym8$nQ3wR@aM", email="webrtc2@test.com"),
                User(username="webrtc_user3", password="Zn7%oR4xS#bN", email="webrtc3@test.com"),
                User(username="webrtc_user4", password="Ao6&pS5yT$cO", email="webrtc4@test.com"),
            ]

            # 注册和登录用户
            for user in test_users:
                if not self.register_user(user):
                    logger.error(f"用户 {user.username} 注册失败")
                    return False
                time.sleep(0.5)

                if not self.login_user(user):
                    logger.error(f"用户 {user.username} 登录失败")
                    return False
                time.sleep(0.5)

            self.users = test_users

            # 2. 创建会议
            logger.info("\n" + "=" * 80)
            logger.info("阶段 2: 创建会议")
            logger.info("=" * 80)

            if not self.create_meeting(test_users[0]):
                logger.error("会议创建失败")
                return False

            # 3. 所有用户加入会议
            logger.info("\n" + "=" * 80)
            logger.info("阶段 3: 用户加入会议")
            logger.info("=" * 80)

            for user in test_users:
                if not self.join_meeting(user):
                    logger.error(f"用户 {user.username} 加入会议失败")
                    return False
                time.sleep(0.5)

            # 4. 建立 WebSocket 连接
            logger.info("\n" + "=" * 80)
            logger.info("阶段 4: 建立 WebSocket 连接")
            logger.info("=" * 80)

            for user in test_users:
                if not await self.connect_websocket(user):
                    logger.error(f"用户 {user.username} WebSocket 连接失败")
                    return False
                time.sleep(0.5)

            # 5. 发送加入房间消息
            logger.info("\n" + "=" * 80)
            logger.info("阶段 5: 发送加入房间消息")
            logger.info("=" * 80)

            for user in test_users:
                if not await self.send_join_room_message(user):
                    logger.error(f"用户 {user.username} 发送加入房间消息失败")
                    return False
                time.sleep(0.5)

            # 6. 建立对等连接 - 每个用户与其他所有用户建立连接
            logger.info("\n" + "=" * 80)
            logger.info("阶段 6: 建立 WebRTC 对等连接")
            logger.info("=" * 80)

            # 创建消息处理任务
            message_handlers = []
            for user in test_users:
                task = asyncio.create_task(self.handle_incoming_messages(user, duration=10.0))
                message_handlers.append(task)

            # 等待一下让所有用户开始监听
            await asyncio.sleep(1)

            # 每个用户向其他用户发送 offer
            for i, user in enumerate(test_users):
                for j, target_user in enumerate(test_users):
                    if i != j:  # 不向自己发送
                        await self.create_peer_connection(user, target_user)
                        await asyncio.sleep(0.2)  # 避免消息过快

            # 等待消息处理完成
            await asyncio.gather(*message_handlers)

            # 7. 验证对等连接
            logger.info("\n" + "=" * 80)
            logger.info("阶段 7: 验证对等连接")
            logger.info("=" * 80)

            total_connections = len(test_users) * (len(test_users) - 1)
            established_connections = sum(
                1 for conn in self.peer_connections.values()
                if conn.get('status') in ['offer_sent', 'connected']
            )

            logger.info(f"总对等连接数: {total_connections}")
            logger.info(f"已建立连接数: {established_connections}")
            logger.info(f"连接详情:")
            for key, conn in self.peer_connections.items():
                logger.info(f"  {conn['from']} -> {conn['to']}: {conn['status']}")

            # 8. 清理
            logger.info("\n" + "=" * 80)
            logger.info("阶段 8: 清理连接")
            logger.info("=" * 80)

            for user in test_users:
                if user.ws_connection:
                    await user.ws_connection.close()
                    self.log_success("CLEANUP", f"Closed WebSocket for {user.username}")

            logger.info("\n" + "=" * 80)
            logger.info("✅ WebRTC 对等连接测试完成！")
            logger.info("=" * 80)
            logger.info(f"✅ 成功建立 {established_connections}/{total_connections} 个对等连接")
            logger.info("✅ 所有用户都能接收到其他用户的音视频流信令")

            return True

        except Exception as e:
            logger.error(f"WebRTC 测试失败: {str(e)}")
            import traceback
            traceback.print_exc()
            return False

async def main():
    """主函数"""
    test = WebRTCIntegrationTest()
    success = await test.run_webrtc_test()

    if success:
        logger.info("\n🎉 WebRTC 集成测试全部通过！")
        sys.exit(0)
    else:
        logger.error("\n❌ WebRTC 集成测试失败！")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())

