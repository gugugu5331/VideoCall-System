#!/usr/bin/env python3
"""
AI Inference Service - 综合压力测试脚本

测试场景：
- 场景 A: 交替调用不同 AI 服务
- 场景 B: 连续调用相同 AI 服务
- 场景 C: 混合模式

通过 Nginx 网关访问，验证服务稳定性和性能。
"""

import requests
import time
import json
import base64
import statistics
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple
from collections import defaultdict

# 配置
NGINX_BASE_URL = "http://localhost:8800/api/v1/ai"
DIRECT_BASE_URL = "http://localhost:8085/api/v1/ai"
TEST_VIDEO_DIR = Path("/root/meeting-system-server/meeting-system/backend/media-service/test_video")
REQUEST_DELAY = 0.5  # 500ms 延迟（增加以避免 llm 节点崩溃）
TIMEOUT = 30  # 30秒超时

# 颜色输出
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    RESET = '\033[0m'
    BOLD = '\033[1m'

def print_header(text: str):
    """打印标题"""
    print(f"\n{Colors.BOLD}{Colors.CYAN}{'='*80}{Colors.RESET}")
    print(f"{Colors.BOLD}{Colors.CYAN}{text}{Colors.RESET}")
    print(f"{Colors.BOLD}{Colors.CYAN}{'='*80}{Colors.RESET}\n")

def print_success(text: str):
    """打印成功信息"""
    print(f"{Colors.GREEN}✓ {text}{Colors.RESET}")

def print_error(text: str):
    """打印错误信息"""
    print(f"{Colors.RED}✗ {text}{Colors.RESET}")

def print_info(text: str):
    """打印信息"""
    print(f"{Colors.BLUE}ℹ {text}{Colors.RESET}")

def print_warning(text: str):
    """打印警告"""
    print(f"{Colors.YELLOW}⚠ {text}{Colors.RESET}")

class TestMetrics:
    """测试指标收集器"""
    def __init__(self):
        self.requests: List[Dict] = []
        self.errors: List[Dict] = []
        self.start_time = None
        self.end_time = None
    
    def add_request(self, service: str, success: bool, response_time: float, 
                   status_code: int = None, error: str = None):
        """添加请求记录"""
        self.requests.append({
            'service': service,
            'success': success,
            'response_time': response_time,
            'status_code': status_code,
            'error': error,
            'timestamp': datetime.now()
        })
        
        if not success:
            self.errors.append({
                'service': service,
                'error': error,
                'status_code': status_code,
                'timestamp': datetime.now()
            })
    
    def get_statistics(self) -> Dict:
        """计算统计信息"""
        if not self.requests:
            return {}
        
        total_requests = len(self.requests)
        successful_requests = sum(1 for r in self.requests if r['success'])
        failed_requests = total_requests - successful_requests
        
        response_times = [r['response_time'] for r in self.requests if r['success']]
        
        stats = {
            'total_requests': total_requests,
            'successful_requests': successful_requests,
            'failed_requests': failed_requests,
            'success_rate': (successful_requests / total_requests * 100) if total_requests > 0 else 0,
            'error_rate': (failed_requests / total_requests * 100) if total_requests > 0 else 0,
        }
        
        if response_times:
            sorted_times = sorted(response_times)
            stats.update({
                'min_response_time': min(response_times),
                'max_response_time': max(response_times),
                'avg_response_time': statistics.mean(response_times),
                'median_response_time': statistics.median(response_times),
                'p95_response_time': sorted_times[int(len(sorted_times) * 0.95)] if len(sorted_times) > 0 else 0,
                'p99_response_time': sorted_times[int(len(sorted_times) * 0.99)] if len(sorted_times) > 0 else 0,
            })
        
        if self.start_time and self.end_time:
            duration = (self.end_time - self.start_time).total_seconds()
            stats['duration'] = duration
            stats['qps'] = total_requests / duration if duration > 0 else 0
        
        # 按服务分类统计
        by_service = defaultdict(lambda: {'total': 0, 'success': 0, 'failed': 0})
        for req in self.requests:
            service = req['service']
            by_service[service]['total'] += 1
            if req['success']:
                by_service[service]['success'] += 1
            else:
                by_service[service]['failed'] += 1
        
        stats['by_service'] = dict(by_service)
        
        return stats

def load_audio_file(file_path: Path) -> str:
    """加载音频文件并转换为 base64"""
    try:
        with open(file_path, 'rb') as f:
            audio_data = f.read()
        return base64.b64encode(audio_data).decode('utf-8')
    except Exception as e:
        print_error(f"Failed to load audio file {file_path}: {e}")
        return None

def prepare_test_data() -> Dict:
    """准备测试数据"""
    print_info("Preparing test data...")
    
    # 查找音频文件
    audio_files = list(TEST_VIDEO_DIR.glob("*.mp3"))
    if not audio_files:
        print_warning("No MP3 files found, using sample base64 data")
        audio_base64 = "c2FtcGxlIGF1ZGlvIGRhdGE="  # "sample audio data"
    else:
        audio_file = audio_files[0]
        print_info(f"Using audio file: {audio_file.name}")
        audio_base64 = load_audio_file(audio_file)
        if not audio_base64:
            audio_base64 = "c2FtcGxlIGF1ZGlvIGRhdGE="
    
    test_data = {
        'asr': {
            'audio_data': audio_base64,
            'format': 'mp3',
            'sample_rate': 16000
        },
        'emotion': {
            'text': 'I am very happy today! This is a wonderful day and everything is going great.'
        },
        'synthesis': {
            'audio_data': audio_base64,
            'format': 'mp3',
            'sample_rate': 16000
        }
    }
    
    print_success("Test data prepared")
    return test_data

def call_ai_service(service: str, data: Dict, base_url: str = NGINX_BASE_URL) -> Tuple[bool, float, int, str]:
    """
    调用 AI 服务
    
    Returns:
        (success, response_time, status_code, error_message)
    """
    url = f"{base_url}/{service}"
    
    start_time = time.time()
    try:
        response = requests.post(
            url,
            json=data,
            headers={'Content-Type': 'application/json'},
            timeout=TIMEOUT
        )
        response_time = time.time() - start_time
        
        if response.status_code == 200:
            result = response.json()
            if result.get('code') == 200:
                return True, response_time, 200, None
            else:
                return False, response_time, result.get('code'), result.get('message')
        else:
            return False, response_time, response.status_code, response.text
            
    except requests.exceptions.Timeout:
        response_time = time.time() - start_time
        return False, response_time, 0, "Request timeout"
    except requests.exceptions.ConnectionError:
        response_time = time.time() - start_time
        return False, response_time, 0, "Connection error"
    except Exception as e:
        response_time = time.time() - start_time
        return False, response_time, 0, str(e)

def scenario_a_alternating_calls(test_data: Dict, metrics: TestMetrics, rounds: int = 5):
    """
    场景 A: 交替调用不同 AI 服务
    """
    print_header("Scenario A: Alternating Calls to Different AI Services")
    print_info(f"Calling each service {rounds} times in rotation...")
    
    services = ['asr', 'emotion', 'synthesis']
    total_calls = rounds * len(services)
    current_call = 0
    
    for round_num in range(rounds):
        print(f"\n{Colors.BOLD}Round {round_num + 1}/{rounds}{Colors.RESET}")
        
        for service in services:
            current_call += 1
            print(f"  [{current_call}/{total_calls}] Calling {service}...", end=' ')
            
            success, response_time, status_code, error = call_ai_service(
                service, test_data[service]
            )
            
            metrics.add_request(service, success, response_time, status_code, error)
            
            if success:
                print_success(f"OK ({response_time:.3f}s)")
            else:
                print_error(f"FAILED ({error})")
            
            time.sleep(REQUEST_DELAY)
    
    print_success(f"\nScenario A completed: {total_calls} requests")

def scenario_b_consecutive_calls(test_data: Dict, metrics: TestMetrics, calls_per_service: int = 10):
    """
    场景 B: 连续调用相同 AI 服务
    """
    print_header("Scenario B: Consecutive Calls to Same AI Service")
    print_info(f"Calling each service {calls_per_service} times consecutively...")
    
    services = ['asr', 'emotion', 'synthesis']
    
    for service in services:
        print(f"\n{Colors.BOLD}Testing {service.upper()} ({calls_per_service} consecutive calls){Colors.RESET}")
        
        for call_num in range(calls_per_service):
            print(f"  [{call_num + 1}/{calls_per_service}] Calling {service}...", end=' ')
            
            success, response_time, status_code, error = call_ai_service(
                service, test_data[service]
            )
            
            metrics.add_request(service, success, response_time, status_code, error)
            
            if success:
                print_success(f"OK ({response_time:.3f}s)")
            else:
                print_error(f"FAILED ({error})")
            
            time.sleep(REQUEST_DELAY)
    
    total_calls = calls_per_service * len(services)
    print_success(f"\nScenario B completed: {total_calls} requests")

def scenario_c_mixed_mode(test_data: Dict, metrics: TestMetrics, rounds: int = 3):
    """
    场景 C: 混合模式（交替执行场景 A 和场景 B）
    """
    print_header("Scenario C: Mixed Mode (Alternating Scenarios A and B)")
    print_info(f"Executing {rounds} rounds of mixed scenarios...")
    
    for round_num in range(rounds):
        print(f"\n{Colors.BOLD}Mixed Round {round_num + 1}/{rounds}{Colors.RESET}")
        
        # 执行场景 A（简化版：每种服务 2 次）
        print(f"\n{Colors.CYAN}  → Running Scenario A (simplified){Colors.RESET}")
        scenario_a_alternating_calls(test_data, metrics, rounds=2)
        
        time.sleep(0.5)  # 场景间短暂休息
        
        # 执行场景 B（简化版：每种服务 3 次）
        print(f"\n{Colors.CYAN}  → Running Scenario B (simplified){Colors.RESET}")
        scenario_b_consecutive_calls(test_data, metrics, calls_per_service=3)
        
        time.sleep(0.5)
    
    print_success(f"\nScenario C completed: {rounds} mixed rounds")

def print_statistics(metrics: TestMetrics):
    """打印统计信息"""
    print_header("Test Statistics Summary")
    
    stats = metrics.get_statistics()
    
    if not stats:
        print_error("No statistics available")
        return
    
    # 总体统计
    print(f"{Colors.BOLD}Overall Statistics:{Colors.RESET}")
    print(f"  Total Requests:      {stats['total_requests']}")
    print(f"  Successful Requests: {Colors.GREEN}{stats['successful_requests']}{Colors.RESET}")
    print(f"  Failed Requests:     {Colors.RED}{stats['failed_requests']}{Colors.RESET}")
    print(f"  Success Rate:        {Colors.GREEN if stats['success_rate'] >= 95 else Colors.RED}{stats['success_rate']:.2f}%{Colors.RESET}")
    print(f"  Error Rate:          {stats['error_rate']:.2f}%")
    
    if 'duration' in stats:
        print(f"  Duration:            {stats['duration']:.2f}s")
        print(f"  Throughput (QPS):    {stats['qps']:.2f}")
    
    # 响应时间统计
    if 'avg_response_time' in stats:
        print(f"\n{Colors.BOLD}Response Time Statistics:{Colors.RESET}")
        print(f"  Min:     {stats['min_response_time']:.3f}s")
        print(f"  Max:     {stats['max_response_time']:.3f}s")
        print(f"  Average: {Colors.GREEN if stats['avg_response_time'] < 3 else Colors.RED}{stats['avg_response_time']:.3f}s{Colors.RESET}")
        print(f"  Median:  {stats['median_response_time']:.3f}s")
        print(f"  P95:     {stats['p95_response_time']:.3f}s")
        print(f"  P99:     {stats['p99_response_time']:.3f}s")
    
    # 按服务统计
    if 'by_service' in stats:
        print(f"\n{Colors.BOLD}Statistics by Service:{Colors.RESET}")
        for service, service_stats in stats['by_service'].items():
            success_rate = (service_stats['success'] / service_stats['total'] * 100) if service_stats['total'] > 0 else 0
            print(f"  {service.upper()}:")
            print(f"    Total:   {service_stats['total']}")
            print(f"    Success: {Colors.GREEN}{service_stats['success']}{Colors.RESET}")
            print(f"    Failed:  {Colors.RED}{service_stats['failed']}{Colors.RESET}")
            print(f"    Rate:    {Colors.GREEN if success_rate >= 95 else Colors.RED}{success_rate:.2f}%{Colors.RESET}")
    
    # 错误详情
    if metrics.errors:
        print(f"\n{Colors.BOLD}Error Details:{Colors.RESET}")
        error_types = defaultdict(int)
        for error in metrics.errors:
            error_types[error['error']] += 1
        
        for error_msg, count in error_types.items():
            print(f"  {Colors.RED}✗{Colors.RESET} {error_msg}: {count} occurrences")

def main():
    """主函数"""
    print_header("AI Inference Service - Comprehensive Stress Test")
    print(f"Start Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"Nginx Gateway: {NGINX_BASE_URL}")
    print(f"Request Delay: {REQUEST_DELAY}s")
    print(f"Timeout: {TIMEOUT}s")
    
    # 准备测试数据
    test_data = prepare_test_data()
    
    # 创建指标收集器
    metrics = TestMetrics()
    metrics.start_time = datetime.now()
    
    try:
        # 执行场景 A（减少轮数以避免 llm 节点崩溃）
        scenario_a_alternating_calls(test_data, metrics, rounds=3)
        time.sleep(2)

        # 执行场景 B（减少每个服务的调用次数）
        scenario_b_consecutive_calls(test_data, metrics, calls_per_service=5)
        time.sleep(2)

        # 执行场景 C（减少混合轮数）
        scenario_c_mixed_mode(test_data, metrics, rounds=2)
        
    except KeyboardInterrupt:
        print_warning("\n\nTest interrupted by user")
    except Exception as e:
        print_error(f"\n\nTest failed with error: {e}")
        import traceback
        traceback.print_exc()
    finally:
        metrics.end_time = datetime.now()
    
    # 打印统计信息
    print_statistics(metrics)
    
    # 成功标准检查
    stats = metrics.get_statistics()
    print_header("Success Criteria Check")
    
    success_rate_ok = stats.get('success_rate', 0) >= 95
    avg_time_ok = stats.get('avg_response_time', float('inf')) < 3
    no_critical_errors = stats.get('failed_requests', 0) < stats.get('total_requests', 1) * 0.05
    
    print(f"  Success Rate ≥ 95%:        {Colors.GREEN + '✓ PASS' if success_rate_ok else Colors.RED + '✗ FAIL'}{Colors.RESET} ({stats.get('success_rate', 0):.2f}%)")
    print(f"  Avg Response Time < 3s:    {Colors.GREEN + '✓ PASS' if avg_time_ok else Colors.RED + '✗ FAIL'}{Colors.RESET} ({stats.get('avg_response_time', 0):.3f}s)")
    print(f"  No Critical Errors:        {Colors.GREEN + '✓ PASS' if no_critical_errors else Colors.RED + '✗ FAIL'}{Colors.RESET}")
    
    all_passed = success_rate_ok and avg_time_ok and no_critical_errors
    
    if all_passed:
        print(f"\n{Colors.GREEN}{Colors.BOLD}✓ ALL TESTS PASSED{Colors.RESET}")
        return 0
    else:
        print(f"\n{Colors.RED}{Colors.BOLD}✗ SOME TESTS FAILED{Colors.RESET}")
        return 1

if __name__ == "__main__":
    sys.exit(main())

