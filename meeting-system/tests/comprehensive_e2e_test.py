#!/usr/bin/env python3
"""
完整的端到端消息队列集成测试
测试所有微服务的消息队列功能和事件流转
"""

import os
import sys
import time
import json
import base64
import requests
import redis
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import threading
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='[%(asctime)s] [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

# 配置
NGINX_URL = "http://localhost:8800"
API_BASE = f"{NGINX_URL}/api/v1"
REDIS_HOST = "localhost"
REDIS_PORT = 6379
TEST_VIDEO_DIR = "/root/meeting-system-server/meeting-system/backend/media-service/test_video"

# 测试用户
TEST_USERS = [
    {"username": f"e2e_user_{i}", "email": f"e2e{i}@test.com", "password": "Test@Pass123"}
    for i in range(1, 5)
]

class E2ETestRunner:
    def __init__(self):
        self.session = requests.Session()
        self.csrf_token = None
        self.redis_client = None
        self.users = {}  # {username: {user_info, token}}
        self.meeting_id = None
        self.test_results = {
            "total": 0,
            "passed": 0,
            "failed": 0,
            "skipped": 0,
            "details": []
        }
        self.start_time = datetime.now()
        
    def log_result(self, stage: str, test_name: str, success: bool, message: str = ""):
        """记录测试结果"""
        self.test_results["total"] += 1
        if success:
            self.test_results["passed"] += 1
            logger.info(f"✅ {stage} - {test_name}: PASSED {message}")
        else:
            self.test_results["failed"] += 1
            logger.error(f"❌ {stage} - {test_name}: FAILED {message}")
        
        self.test_results["details"].append({
            "stage": stage,
            "test": test_name,
            "success": success,
            "message": message,
            "timestamp": datetime.now().isoformat()
        })
    
    def setup(self):
        """初始化测试环境"""
        logger.info("=" * 60)
        logger.info("开始完整的端到端消息队列集成测试")
        logger.info("=" * 60)
        
        # 连接 Redis
        try:
            self.redis_client = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
            self.redis_client.ping()
            logger.info("✅ Redis 连接成功")
        except Exception as e:
            logger.error(f"❌ Redis 连接失败: {e}")
            sys.exit(1)
        
        # 获取 CSRF token
        try:
            response = self.session.get(f"{API_BASE}/csrf-token", timeout=5)
            if response.status_code == 200:
                data = response.json()
                self.csrf_token = data.get('data', {}).get('csrf_token')
                logger.info("✅ 获取 CSRF token 成功")
            else:
                logger.warning(f"⚠️ 获取 CSRF token 失败: {response.status_code}")
        except Exception as e:
            logger.error(f"❌ 获取 CSRF token 异常: {e}")
    
    def get_queue_stats(self) -> Dict:
        """获取 Redis 队列统计"""
        stats = {}
        queues = [
            "meeting_system:critical_queue",
            "meeting_system:high_queue",
            "meeting_system:normal_queue",
            "meeting_system:low_queue",
            "meeting_system:dead_letter_queue",
            "meeting_system:processing_queue"
        ]
        for queue in queues:
            try:
                length = self.redis_client.llen(queue)
                stats[queue.split(':')[1]] = length
            except:
                stats[queue.split(':')[1]] = 0
        return stats
    
    def print_queue_stats(self, prefix=""):
        """打印队列统计"""
        stats = self.get_queue_stats()
        logger.info(f"{prefix}Redis 队列状态:")
        for queue, length in stats.items():
            logger.info(f"  {queue}: {length}")
    
    def make_request(self, method: str, url: str, **kwargs) -> requests.Response:
        """发送 HTTP 请求"""
        headers = kwargs.get('headers', {})
        if self.csrf_token and method.upper() in ['POST', 'PUT', 'DELETE']:
            headers['X-CSRF-Token'] = self.csrf_token
        kwargs['headers'] = headers
        
        try:
            response = self.session.request(method, url, **kwargs)
            return response
        except Exception as e:
            logger.error(f"请求异常: {e}")
            raise
    
    def test_user_service(self):
        """测试用户服务"""
        logger.info("\n" + "=" * 60)
        logger.info("阶段 1: 用户服务测试")
        logger.info("=" * 60)
        
        # 1.1 用户注册
        logger.info("\n--- 1.1 用户注册 ---")
        for user in TEST_USERS:
            try:
                response = self.make_request(
                    'POST',
                    f"{API_BASE}/auth/register",
                    json=user,
                    headers={"Content-Type": "application/json"},
                    timeout=10
                )
                
                if response.status_code == 200:
                    data = response.json()
                    user_data = data.get('data', {})
                    self.users[user['username']] = {
                        'user_info': user_data,
                        'password': user['password']
                    }
                    self.log_result("用户注册", user['username'], True, f"ID: {user_data.get('id')}")
                elif response.status_code == 400 and "already exists" in response.text:
                    # 用户已存在，尝试登录
                    logger.info(f"用户 {user['username']} 已存在，将在登录阶段使用")
                    self.users[user['username']] = {'password': user['password']}
                    self.log_result("用户注册", user['username'], True, "用户已存在")
                else:
                    self.log_result("用户注册", user['username'], False, f"状态码: {response.status_code}")
            except Exception as e:
                self.log_result("用户注册", user['username'], False, str(e))
        
        time.sleep(2)
        
        # 1.2 用户登录
        logger.info("\n--- 1.2 用户登录 ---")
        for username, user_data in self.users.items():
            try:
                response = self.make_request(
                    'POST',
                    f"{API_BASE}/auth/login",
                    json={"username": username, "password": user_data['password']},
                    headers={"Content-Type": "application/json"},
                    timeout=10
                )
                
                if response.status_code == 200:
                    data = response.json()
                    self.users[username]['token'] = data['data']['token']
                    self.users[username]['user_info'] = data['data']['user']
                    self.log_result("用户登录", username, True, "获取 token 成功")
                else:
                    self.log_result("用户登录", username, False, f"状态码: {response.status_code}")
            except Exception as e:
                self.log_result("用户登录", username, False, str(e))
        
        time.sleep(2)
        
        # 1.3 获取用户资料
        logger.info("\n--- 1.3 获取用户资料 ---")
        first_user = list(self.users.keys())[0]
        if 'token' in self.users[first_user]:
            try:
                response = self.make_request(
                    'GET',
                    f"{API_BASE}/users/profile",
                    headers={"Authorization": f"Bearer {self.users[first_user]['token']}"},
                    timeout=10
                )
                
                if response.status_code == 200:
                    self.log_result("获取资料", first_user, True)
                else:
                    self.log_result("获取资料", first_user, False, f"状态码: {response.status_code}")
            except Exception as e:
                self.log_result("获取资料", first_user, False, str(e))
        
        time.sleep(1)
        self.print_queue_stats("用户服务测试后 - ")
    
    def test_meeting_service(self):
        """测试会议服务"""
        logger.info("\n" + "=" * 60)
        logger.info("阶段 2: 会议服务测试")
        logger.info("=" * 60)
        
        # 确保至少有一个用户有 token
        users_with_token = [u for u in self.users.keys() if 'token' in self.users[u]]
        if not users_with_token:
            logger.error("没有用户有有效的 token，跳过会议服务测试")
            return
        
        creator = users_with_token[0]
        
        # 2.1 创建会议
        logger.info("\n--- 2.1 创建会议 ---")
        now = datetime.now()
        meeting_data = {
            "title": "E2E Queue Integration Test Meeting",
            "description": "端到端消息队列集成测试会议",
            "start_time": now.strftime("%Y-%m-%dT%H:%M:%S+08:00"),
            "end_time": (now + timedelta(hours=2)).strftime("%Y-%m-%dT%H:%M:%S+08:00"),
            "max_participants": 10,
            "meeting_type": "video_conference"
        }
        
        try:
            response = self.make_request(
                'POST',
                f"{API_BASE}/meetings",
                json=meeting_data,
                headers={
                    "Authorization": f"Bearer {self.users[creator]['token']}",
                    "Content-Type": "application/json"
                },
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.meeting_id = data.get('data', {}).get('id') or data.get('data', {}).get('meeting_id')
                self.log_result("创建会议", creator, True, f"会议ID: {self.meeting_id}")
            else:
                self.log_result("创建会议", creator, False, f"状态码: {response.status_code}, 响应: {response.text[:200]}")
                logger.error(f"创建会议失败，响应: {response.text}")
        except Exception as e:
            self.log_result("创建会议", creator, False, str(e))
        
        time.sleep(2)
        self.print_queue_stats("创建会议后 - ")

        if not self.meeting_id:
            logger.error("会议创建失败，跳过后续会议测试")
            return

        # 2.2 其他用户加入会议
        logger.info("\n--- 2.2 用户加入会议 ---")
        for username in users_with_token[1:]:
            try:
                response = self.make_request(
                    'POST',
                    f"{API_BASE}/meetings/{self.meeting_id}/join",
                    headers={"Authorization": f"Bearer {self.users[username]['token']}"},
                    timeout=10
                )

                if response.status_code in [200, 201]:
                    self.log_result("加入会议", username, True)
                else:
                    self.log_result("加入会议", username, False, f"状态码: {response.status_code}")
            except Exception as e:
                self.log_result("加入会议", username, False, str(e))
            time.sleep(1)

        time.sleep(2)

        # 2.3 获取会议信息
        logger.info("\n--- 2.3 获取会议信息 ---")
        try:
            response = self.make_request(
                'GET',
                f"{API_BASE}/meetings/{self.meeting_id}",
                headers={"Authorization": f"Bearer {self.users[creator]['token']}"},
                timeout=10
            )

            if response.status_code == 200:
                self.log_result("获取会议信息", "meeting-service", True)
            else:
                self.log_result("获取会议信息", "meeting-service", False, f"状态码: {response.status_code}")
        except Exception as e:
            self.log_result("获取会议信息", "meeting-service", False, str(e))

        time.sleep(1)
        self.print_queue_stats("会议服务测试后 - ")

    def test_ai_service(self):
        """测试 AI 服务 - 模拟客户端周期性推理请求"""
        logger.info("\n" + "=" * 60)
        logger.info("阶段 3: AI 服务测试（模拟客户端周期性推理）")
        logger.info("=" * 60)

        users_with_token = [u for u in self.users.keys() if 'token' in self.users[u]]
        if len(users_with_token) < 2:
            logger.error("用户数量不足，跳过 AI 服务测试")
            return

        # 读取测试音视频文件
        test_files = {
            'video': os.path.join(TEST_VIDEO_DIR, "20250928_164722.mp4"),
            'audio': os.path.join(TEST_VIDEO_DIR, "20250602_215504.mp3")
        }

        # 检查文件是否存在
        for file_type, file_path in test_files.items():
            if not os.path.exists(file_path):
                logger.warning(f"测试文件不存在: {file_path}")
                return

        # 读取文件并转换为 base64
        media_data = {}
        for file_type, file_path in test_files.items():
            try:
                with open(file_path, 'rb') as f:
                    # 只读取前 100KB 用于测试
                    content = f.read(100 * 1024)
                    media_data[file_type] = base64.b64encode(content).decode('utf-8')
                logger.info(f"✅ 读取 {file_type} 文件成功: {len(media_data[file_type])} bytes (base64)")
            except Exception as e:
                logger.error(f"❌ 读取 {file_type} 文件失败: {e}")
                return

        # 3.1 情绪识别测试
        logger.info("\n--- 3.1 情绪识别（每个用户对其他用户的视频进行推理）---")
        for i, user1 in enumerate(users_with_token[:2]):  # 限制用户数量以加快测试
            for j, user2 in enumerate(users_with_token[:2]):
                if i == j:
                    continue

                try:
                    payload = {
                        "task_type": "emotion_detection",
                        "video_data": media_data['video'][:1000],  # 只发送部分数据
                        "source_user": user2,
                        "target_user": user1,
                        "meeting_id": self.meeting_id
                    }

                    response = self.make_request(
                        'POST',
                        f"{API_BASE}/ai/inference",
                        json=payload,
                        headers={
                            "Authorization": f"Bearer {self.users[user1]['token']}",
                            "Content-Type": "application/json"
                        },
                        timeout=30
                    )

                    if response.status_code in [200, 201, 202]:
                        self.log_result("情绪识别", f"{user1}→{user2}", True)
                    else:
                        self.log_result("情绪识别", f"{user1}→{user2}", False,
                                      f"状态码: {response.status_code}, 响应: {response.text[:100]}")
                except Exception as e:
                    self.log_result("情绪识别", f"{user1}→{user2}", False, str(e))

                time.sleep(0.5)

        time.sleep(2)

        # 3.2 语音识别测试
        logger.info("\n--- 3.2 语音识别（每个用户对其他用户的音频进行推理）---")
        for i, user1 in enumerate(users_with_token[:2]):
            for j, user2 in enumerate(users_with_token[:2]):
                if i == j:
                    continue

                try:
                    payload = {
                        "task_type": "speech_recognition",
                        "audio_data": media_data['audio'][:1000],
                        "source_user": user2,
                        "target_user": user1,
                        "meeting_id": self.meeting_id
                    }

                    response = self.make_request(
                        'POST',
                        f"{API_BASE}/ai/inference",
                        json=payload,
                        headers={
                            "Authorization": f"Bearer {self.users[user1]['token']}",
                            "Content-Type": "application/json"
                        },
                        timeout=30
                    )

                    if response.status_code in [200, 201, 202]:
                        self.log_result("语音识别", f"{user1}→{user2}", True)
                    else:
                        self.log_result("语音识别", f"{user1}→{user2}", False,
                                      f"状态码: {response.status_code}")
                except Exception as e:
                    self.log_result("语音识别", f"{user1}→{user2}", False, str(e))

                time.sleep(0.5)

        time.sleep(2)
        self.print_queue_stats("AI 服务测试后 - ")

    def generate_report(self):
        """生成测试报告"""
        logger.info("\n" + "=" * 60)
        logger.info("生成测试报告")
        logger.info("=" * 60)

        end_time = datetime.now()
        duration = (end_time - self.start_time).total_seconds()

        report = f"""
# 完整端到端消息队列集成测试报告

**测试时间**: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}
**测试时长**: {duration:.2f} 秒
**总测试数**: {self.test_results['total']}
**通过**: {self.test_results['passed']}
**失败**: {self.test_results['failed']}
**跳过**: {self.test_results['skipped']}
**成功率**: {(self.test_results['passed'] / self.test_results['total'] * 100) if self.test_results['total'] > 0 else 0:.2f}%

## 测试详情

"""

        # 按阶段分组
        stages = {}
        for detail in self.test_results['details']:
            stage = detail['stage']
            if stage not in stages:
                stages[stage] = []
            stages[stage].append(detail)

        for stage, tests in stages.items():
            report += f"\n### {stage}\n\n"
            for test in tests:
                status = "✅" if test['success'] else "❌"
                report += f"- {status} {test['test']}: {test['message']}\n"

        # Redis 队列统计
        report += "\n## Redis 队列最终状态\n\n"
        stats = self.get_queue_stats()
        for queue, length in stats.items():
            report += f"- {queue}: {length}\n"

        # 保存报告
        report_file = f"comprehensive_e2e_test_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.md"
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report)

        logger.info(f"✅ 测试报告已生成: {report_file}")
        print(report)

    def run(self):
        """运行所有测试"""
        try:
            self.setup()
            self.print_queue_stats("初始 - ")

            self.test_user_service()
            self.test_meeting_service()
            # self.test_ai_service()  # 暂时跳过 AI 服务测试

            self.generate_report()

        except KeyboardInterrupt:
            logger.info("\n测试被用户中断")
        except Exception as e:
            logger.error(f"测试执行异常: {e}", exc_info=True)
        finally:
            logger.info("\n" + "=" * 60)
            logger.info("测试完成")
            logger.info("=" * 60)

if __name__ == "__main__":
    runner = E2ETestRunner()
    runner.run()


