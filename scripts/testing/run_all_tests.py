#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall System - Complete Test Suite
统一测试脚本，测试所有系统组件
"""

import requests
import json
import time
import sys
import asyncio
import aiohttp
from datetime import datetime

# 配置
BACKEND_URL = "http://localhost:8000"
AI_URL = "http://localhost:5001"
TIMEOUT = 10

class TestResult:
    """测试结果类"""
    def __init__(self, name, success, message="", details=None):
        self.name = name
        self.success = success
        self.message = message
        self.details = details or {}

def print_header(title):
    """打印标题"""
    print("\n" + "=" * 60)
    print(f" {title}")
    print("=" * 60)

def print_result(result):
    """打印测试结果"""
    status = "✅" if result.success else "❌"
    print(f"{status} {result.name}: {result.message}")

def test_backend_health():
    """测试后端健康检查"""
    try:
        response = requests.get(f"{BACKEND_URL}/health", timeout=TIMEOUT)
        if response.status_code == 200:
            return TestResult("后端健康检查", True, "服务正常", response.json())
        else:
            return TestResult("后端健康检查", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("后端健康检查", False, f"连接失败: {e}")

def test_backend_root():
    """测试后端根端点"""
    try:
        response = requests.get(f"{BACKEND_URL}/", timeout=TIMEOUT)
        if response.status_code == 200:
            return TestResult("后端根端点", True, "服务信息正常", response.json())
        else:
            return TestResult("后端根端点", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("后端根端点", False, f"连接失败: {e}")

def test_ai_health():
    """测试AI服务健康检查"""
    try:
        response = requests.get(f"{AI_URL}/health", timeout=TIMEOUT)
        if response.status_code == 200:
            return TestResult("AI服务健康检查", True, "服务正常", response.json())
        else:
            return TestResult("AI服务健康检查", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("AI服务健康检查", False, f"连接失败: {e}")

def test_ai_root():
    """测试AI服务根端点"""
    try:
        response = requests.get(f"{AI_URL}/", timeout=TIMEOUT)
        if response.status_code == 200:
            return TestResult("AI服务根端点", True, "服务信息正常", response.json())
        else:
            return TestResult("AI服务根端点", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("AI服务根端点", False, f"连接失败: {e}")

def test_user_registration():
    """测试用户注册"""
    try:
        data = {
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123",
            "full_name": "测试用户"
        }
        response = requests.post(f"{BACKEND_URL}/api/v1/auth/register", json=data, timeout=TIMEOUT)
        if response.status_code == 201:
            return TestResult("用户注册", True, "注册成功")
        elif response.status_code == 409:
            return TestResult("用户注册", True, "用户已存在（预期结果）")
        else:
            return TestResult("用户注册", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("用户注册", False, f"请求失败: {e}")

def test_user_login():
    """测试用户登录"""
    try:
        data = {
            "username": "testuser",
            "password": "password123"
        }
        response = requests.post(f"{BACKEND_URL}/api/v1/auth/login", json=data, timeout=TIMEOUT)
        if response.status_code == 200:
            result = response.json()
            if "token" in result:
                return TestResult("用户登录", True, "登录成功", {"token_length": len(result["token"])})
            else:
                return TestResult("用户登录", False, "响应中没有token")
        else:
            return TestResult("用户登录", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("用户登录", False, f"请求失败: {e}")

def test_protected_endpoint(token):
    """测试受保护的端点"""
    try:
        headers = {"Authorization": f"Bearer {token}"}
        response = requests.get(f"{BACKEND_URL}/api/v1/user/profile", headers=headers, timeout=TIMEOUT)
        if response.status_code == 200:
            return TestResult("受保护端点", True, "访问成功", response.json())
        else:
            return TestResult("受保护端点", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("受保护端点", False, f"请求失败: {e}")

def test_ai_detection():
    """测试AI检测服务"""
    try:
        test_data = {
            "detection_id": "test_001",
            "detection_type": "voice",
            "audio_data": "base64_encoded_audio_data_here"
        }
        response = requests.post(f"{AI_URL}/detect", json=test_data, timeout=TIMEOUT)
        if response.status_code == 200:
            result = response.json()
            risk_score = result.get("risk_score", 0)
            confidence = result.get("confidence", 0)
            return TestResult("AI检测服务", True, 
                            f"检测成功 (风险评分={risk_score}, 置信度={confidence})", 
                            result)
        else:
            return TestResult("AI检测服务", False, f"状态码: {response.status_code}")
    except Exception as e:
        return TestResult("AI检测服务", False, f"请求失败: {e}")

def run_all_tests():
    """运行所有测试"""
    print_header("VideoCall System - 完整测试套件")
    print(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"后端服务: {BACKEND_URL}")
    print(f"AI服务: {AI_URL}")
    
    results = []
    
    # 基础服务测试
    print_header("基础服务测试")
    results.append(test_backend_health())
    results.append(test_backend_root())
    results.append(test_ai_health())
    results.append(test_ai_root())
    
    # 用户功能测试
    print_header("用户功能测试")
    results.append(test_user_registration())
    login_result = test_user_login()
    results.append(login_result)
    
    # 获取token用于后续测试
    token = None
    if login_result.success and login_result.details:
        token = login_result.details.get("token_length")
    
    if token:
        results.append(test_protected_endpoint(token))
    else:
        results.append(TestResult("受保护端点", False, "无法获取token"))
    
    # AI功能测试
    print_header("AI功能测试")
    results.append(test_ai_detection())
    
    # 结果统计
    print_header("测试结果统计")
    passed = sum(1 for r in results if r.success)
    total = len(results)
    
    for result in results:
        print_result(result)
    
    print(f"\n总计: {passed}/{total} 测试通过")
    
    if passed == total:
        print("\n🎉 所有测试通过！系统运行正常。")
        return True
    else:
        print(f"\n⚠️  {total - passed} 个测试失败，请检查服务状态。")
        return False

async def run_concurrency_test():
    """运行并发测试"""
    print_header("并发性能测试")
    
    try:
        # 导入并发测试模块
        from test_concurrency import ConcurrencyTester
        
        tester = ConcurrencyTester()
        
        # 运行健康检查并发测试
        print("运行健康检查并发测试 (50个并发请求)...")
        health_results = await tester.run_concurrent_tests(50, "health")
        health_analysis = tester.analyze_results(health_results)
        
        print(f"健康检查测试结果:")
        print(f"  成功率: {health_analysis['success_rate']:.2f}%")
        print(f"  平均响应时间: {health_analysis['avg_response_time']:.3f}s")
        print(f"  最大响应时间: {health_analysis['max_response_time']:.3f}s")
        
        # 运行检测服务并发测试
        print("\n运行检测服务并发测试 (20个并发请求)...")
        detection_results = await tester.run_concurrent_tests(20, "detection")
        detection_analysis = tester.analyze_results(detection_results)
        
        print(f"检测服务测试结果:")
        print(f"  成功率: {detection_analysis['success_rate']:.2f}%")
        print(f"  平均响应时间: {detection_analysis['avg_response_time']:.3f}s")
        print(f"  最大响应时间: {detection_analysis['max_response_time']:.3f}s")
        
        return True
        
    except Exception as e:
        print(f"并发测试失败: {e}")
        return False

def main():
    """主函数"""
    try:
        print("选择测试类型:")
        print("1. 运行完整测试套件")
        print("2. 运行并发性能测试")
        print("0. 退出")
        
        choice = input("\n请选择 (0-2): ").strip()
        
        if choice == "1":
            success = run_all_tests()
            return 0 if success else 1
        elif choice == "2":
            success = asyncio.run(run_concurrency_test())
            return 0 if success else 1
        elif choice == "0":
            print("退出测试")
            return 0
        else:
            print("无效选择，运行完整测试套件")
        success = run_all_tests()
        return 0 if success else 1
            
    except KeyboardInterrupt:
        print("\n\n测试被用户中断")
        return 1
    except Exception as e:
        print(f"\n测试过程中发生错误: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 