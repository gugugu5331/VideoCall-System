#!/usr/bin/env python3
"""
Test script for AI Inference Service
Tests ASR, Emotion Detection, and Synthesis Detection via HTTP API
"""

import requests
import json
import base64
import time
import argparse


class AIInferenceServiceTester:
    def __init__(self, base_url):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json'
        })

    def test_health_check(self):
        """测试健康检查"""
        print("\n" + "=" * 60)
        print("Testing Health Check")
        print("=" * 60)

        try:
            response = self.session.get(f"{self.base_url}/api/v1/ai/health")
            response.raise_for_status()
            data = response.json()
            print(f"✓ Health check passed: {json.dumps(data, indent=2)}")
            return True
        except Exception as e:
            print(f"✗ Health check failed: {e}")
            return False

    def test_service_info(self):
        """测试服务信息"""
        print("\n" + "=" * 60)
        print("Testing Service Info")
        print("=" * 60)

        try:
            response = self.session.get(f"{self.base_url}/api/v1/ai/info")
            response.raise_for_status()
            data = response.json()
            print(f"✓ Service info retrieved: {json.dumps(data, indent=2)}")
            return True
        except Exception as e:
            print(f"✗ Service info failed: {e}")
            return False

    def test_asr(self):
        """测试语音识别"""
        print("\n" + "=" * 60)
        print("Testing ASR (Speech Recognition)")
        print("=" * 60)

        # 创建测试音频数据（Base64 编码）
        test_audio = b"sample audio data for testing"
        audio_base64 = base64.b64encode(test_audio).decode('utf-8')

        request_data = {
            "audio_data": audio_base64,
            "format": "wav",
            "sample_rate": 16000,
            "language": "en"
        }

        try:
            start_time = time.time()
            response = self.session.post(
                f"{self.base_url}/api/v1/ai/asr",
                json=request_data
            )
            elapsed = (time.time() - start_time) * 1000

            response.raise_for_status()
            data = response.json()

            print(f"✓ ASR completed in {elapsed:.2f}ms")
            print(f"Response: {json.dumps(data, indent=2)}")

            if data.get('code') == 200:
                result = data.get('data', {})
                print(f"\nRecognized Text: {result.get('text')}")
                print(f"Confidence: {result.get('confidence')}")
                print(f"Language: {result.get('language')}")
                return True
            else:
                print(f"✗ ASR failed: {data.get('message')}")
                return False

        except Exception as e:
            print(f"✗ ASR test failed: {e}")
            return False

    def test_emotion_detection(self):
        """测试情感检测"""
        print("\n" + "=" * 60)
        print("Testing Emotion Detection")
        print("=" * 60)

        request_data = {
            "text": "I am very happy today! This is wonderful news."
        }

        try:
            start_time = time.time()
            response = self.session.post(
                f"{self.base_url}/api/v1/ai/emotion",
                json=request_data
            )
            elapsed = (time.time() - start_time) * 1000

            response.raise_for_status()
            data = response.json()

            print(f"✓ Emotion detection completed in {elapsed:.2f}ms")
            print(f"Response: {json.dumps(data, indent=2)}")

            if data.get('code') == 200:
                result = data.get('data', {})
                print(f"\nDetected Emotion: {result.get('emotion')}")
                print(f"Confidence: {result.get('confidence')}")
                print(f"All Emotions: {result.get('emotions')}")
                return True
            else:
                print(f"✗ Emotion detection failed: {data.get('message')}")
                return False

        except Exception as e:
            print(f"✗ Emotion detection test failed: {e}")
            return False

    def test_synthesis_detection(self):
        """测试深度伪造检测"""
        print("\n" + "=" * 60)
        print("Testing Synthesis Detection")
        print("=" * 60)

        # 创建测试音频数据（Base64 编码）
        test_audio = b"sample audio data for synthesis detection"
        audio_base64 = base64.b64encode(test_audio).decode('utf-8')

        request_data = {
            "audio_data": audio_base64,
            "format": "wav",
            "sample_rate": 16000
        }

        try:
            start_time = time.time()
            response = self.session.post(
                f"{self.base_url}/api/v1/ai/synthesis",
                json=request_data
            )
            elapsed = (time.time() - start_time) * 1000

            response.raise_for_status()
            data = response.json()

            print(f"✓ Synthesis detection completed in {elapsed:.2f}ms")
            print(f"Response: {json.dumps(data, indent=2)}")

            if data.get('code') == 200:
                result = data.get('data', {})
                print(f"\nIs Synthetic: {result.get('is_synthetic')}")
                print(f"Confidence: {result.get('confidence')}")
                print(f"Score: {result.get('score')}")
                return True
            else:
                print(f"✗ Synthesis detection failed: {data.get('message')}")
                return False

        except Exception as e:
            print(f"✗ Synthesis detection test failed: {e}")
            return False

    def test_batch_inference(self):
        """测试批量推理"""
        print("\n" + "=" * 60)
        print("Testing Batch Inference")
        print("=" * 60)

        test_audio = b"sample audio data"
        audio_base64 = base64.b64encode(test_audio).decode('utf-8')

        request_data = {
            "tasks": [
                {
                    "task_id": "task_1",
                    "type": "asr",
                    "data": {
                        "audio_data": audio_base64,
                        "format": "wav",
                        "sample_rate": 16000
                    }
                },
                {
                    "task_id": "task_2",
                    "type": "emotion",
                    "data": {
                        "text": "I am feeling great today!"
                    }
                },
                {
                    "task_id": "task_3",
                    "type": "synthesis",
                    "data": {
                        "audio_data": audio_base64,
                        "format": "wav",
                        "sample_rate": 16000
                    }
                }
            ]
        }

        try:
            start_time = time.time()
            response = self.session.post(
                f"{self.base_url}/api/v1/ai/batch",
                json=request_data
            )
            elapsed = (time.time() - start_time) * 1000

            response.raise_for_status()
            data = response.json()

            print(f"✓ Batch inference completed in {elapsed:.2f}ms")
            print(f"Response: {json.dumps(data, indent=2)}")

            if data.get('code') == 200:
                result = data.get('data', {})
                print(f"\nTotal tasks: {result.get('total')}")
                print(f"Results: {len(result.get('results', []))}")
                return True
            else:
                print(f"✗ Batch inference failed: {data.get('message')}")
                return False

        except Exception as e:
            print(f"✗ Batch inference test failed: {e}")
            return False

    def run_all_tests(self):
        """运行所有测试"""
        print("\n" + "=" * 60)
        print("AI Inference Service Test Suite")
        print("=" * 60)
        print(f"Base URL: {self.base_url}")

        results = {
            "Health Check": self.test_health_check(),
            "Service Info": self.test_service_info(),
            "ASR": self.test_asr(),
            "Emotion Detection": self.test_emotion_detection(),
            "Synthesis Detection": self.test_synthesis_detection(),
            "Batch Inference": self.test_batch_inference(),
        }

        # 打印总结
        print("\n" + "=" * 60)
        print("Test Summary")
        print("=" * 60)

        passed = sum(1 for v in results.values() if v)
        total = len(results)

        for test_name, result in results.items():
            status = "✓ PASSED" if result else "✗ FAILED"
            print(f"{test_name}: {status}")

        print(f"\nTotal: {passed}/{total} tests passed")

        if passed == total:
            print("\n🎉 All tests passed!")
            return 0
        else:
            print(f"\n❌ {total - passed} test(s) failed")
            return 1


def main():
    parser = argparse.ArgumentParser(description='Test AI Inference Service')
    parser.add_argument('--host', default='localhost', help='Service host')
    parser.add_argument('--port', type=int, default=8085, help='Service port')
    args = parser.parse_args()

    base_url = f"http://{args.host}:{args.port}"
    tester = AIInferenceServiceTester(base_url)

    exit_code = tester.run_all_tests()
    exit(exit_code)


if __name__ == "__main__":
    main()

