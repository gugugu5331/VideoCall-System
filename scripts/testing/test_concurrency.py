#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall System - Concurrency Test
并发性能测试脚本
"""
import asyncio
import aiohttp
import time
import json
import statistics
from concurrent.futures import ThreadPoolExecutor
import threading
from typing import List, Dict, Any
import argparse

class ConcurrencyTester:
    def __init__(self, base_url: str = "http://localhost:8000", ai_url: str = "http://localhost:5001"):
        self.base_url = base_url
        self.ai_url = ai_url
        self.results = []
        self.lock = threading.Lock()
        
    async def test_backend_health(self, session: aiohttp.ClientSession) -> Dict[str, Any]:
        """测试后端健康检查"""
        start_time = time.time()
        try:
            async with session.get(f"{self.base_url}/health") as response:
                duration = time.time() - start_time
                return {
                    "endpoint": "/health",
                    "status": response.status,
                    "duration": duration,
                    "success": response.status == 200
                }
        except Exception as e:
            duration = time.time() - start_time
            return {
                "endpoint": "/health",
                "status": 0,
                "duration": duration,
                "success": False,
                "error": str(e)
            }
    
    async def test_ai_health(self, session: aiohttp.ClientSession) -> Dict[str, Any]:
        """测试AI服务健康检查"""
        start_time = time.time()
        try:
            async with session.get(f"{self.ai_url}/health") as response:
                duration = time.time() - start_time
                return {
                    "endpoint": "/ai/health",
                    "status": response.status,
                    "duration": duration,
                    "success": response.status == 200
                }
        except Exception as e:
            duration = time.time() - start_time
            return {
                "endpoint": "/ai/health",
                "status": 0,
                "duration": duration,
                "success": False,
                "error": str(e)
            }
    
    async def test_ai_detection(self, session: aiohttp.ClientSession, detection_id: str) -> Dict[str, Any]:
        """测试AI检测服务"""
        start_time = time.time()
        payload = {
            "detection_id": detection_id,
            "detection_type": "voice_spoofing",
            "audio_data": "dGVzdCBhdWRpbyBkYXRh",  # base64 encoded "test audio data"
            "metadata": {"test": True}
        }
        
        try:
            async with session.post(f"{self.ai_url}/detect", json=payload) as response:
                duration = time.time() - start_time
                return {
                    "endpoint": "/detect",
                    "detection_id": detection_id,
                    "status": response.status,
                    "duration": duration,
                    "success": response.status == 200
                }
        except Exception as e:
            duration = time.time() - start_time
            return {
                "endpoint": "/detect",
                "detection_id": detection_id,
                "status": 0,
                "duration": duration,
                "success": False,
                "error": str(e)
            }
    
    async def test_batch_detection(self, session: aiohttp.ClientSession, batch_size: int) -> Dict[str, Any]:
        """测试批量检测"""
        start_time = time.time()
        requests = []
        
        for i in range(batch_size):
            requests.append({
                "detection_id": f"batch_test_{i}",
                "detection_type": "voice_spoofing",
                "audio_data": "dGVzdCBhdWRpbyBkYXRh",
                "metadata": {"batch_test": True, "index": i}
            })
        
        try:
            async with session.post(f"{self.ai_url}/detect/batch", json=requests) as response:
                duration = time.time() - start_time
                return {
                    "endpoint": "/detect/batch",
                    "batch_size": batch_size,
                    "status": response.status,
                    "duration": duration,
                    "success": response.status == 200
                }
        except Exception as e:
            duration = time.time() - start_time
            return {
                "endpoint": "/detect/batch",
                "batch_size": batch_size,
                "status": 0,
                "duration": duration,
                "success": False,
                "error": str(e)
            }
    
    async def run_concurrent_tests(self, num_requests: int, test_type: str = "health") -> List[Dict[str, Any]]:
        """运行并发测试"""
        print(f"Running {num_requests} concurrent {test_type} tests...")
        
        connector = aiohttp.TCPConnector(limit=100, limit_per_host=50)
        timeout = aiohttp.ClientTimeout(total=30)
        
        async with aiohttp.ClientSession(connector=connector, timeout=timeout) as session:
            tasks = []
            
            if test_type == "health":
                for i in range(num_requests):
                    if i % 2 == 0:
                        tasks.append(self.test_backend_health(session))
                    else:
                        tasks.append(self.test_ai_health(session))
            elif test_type == "detection":
                for i in range(num_requests):
                    tasks.append(self.test_ai_detection(session, f"concurrent_test_{i}"))
            elif test_type == "batch":
                batch_size = min(10, num_requests)
                for i in range(0, num_requests, batch_size):
                    current_batch_size = min(batch_size, num_requests - i)
                    tasks.append(self.test_batch_detection(session, current_batch_size))
            
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # 处理结果
            processed_results = []
            for result in results:
                if isinstance(result, Exception):
                    processed_results.append({
                        "endpoint": "unknown",
                        "status": 0,
                        "duration": 0,
                        "success": False,
                        "error": str(result)
                    })
                else:
                    processed_results.append(result)
            
            return processed_results
    
    def analyze_results(self, results: List[Dict[str, Any]]) -> Dict[str, Any]:
        """分析测试结果"""
        if not results:
            return {"error": "No results to analyze"}
        
        successful = [r for r in results if r.get("success", False)]
        failed = [r for r in results if not r.get("success", False)]
        
        durations = [r.get("duration", 0) for r in successful]
        
        analysis = {
            "total_requests": len(results),
            "successful_requests": len(successful),
            "failed_requests": len(failed),
            "success_rate": len(successful) / len(results) * 100 if results else 0,
            "avg_response_time": statistics.mean(durations) if durations else 0,
            "min_response_time": min(durations) if durations else 0,
            "max_response_time": max(durations) if durations else 0,
            "median_response_time": statistics.median(durations) if durations else 0,
            "p95_response_time": statistics.quantiles(durations, n=20)[18] if len(durations) >= 20 else max(durations) if durations else 0,
            "p99_response_time": statistics.quantiles(durations, n=100)[98] if len(durations) >= 100 else max(durations) if durations else 0,
        }
        
        return analysis
    
    def print_results(self, results: List[Dict[str, Any]], analysis: Dict[str, Any]):
        """打印测试结果"""
        print("\n" + "="*60)
        print("CONCURRENCY TEST RESULTS")
        print("="*60)
        
        print(f"\n📊 SUMMARY:")
        print(f"   Total Requests: {analysis['total_requests']}")
        print(f"   Successful: {analysis['successful_requests']}")
        print(f"   Failed: {analysis['failed_requests']}")
        print(f"   Success Rate: {analysis['success_rate']:.2f}%")
        
        print(f"\n⏱️  RESPONSE TIMES:")
        print(f"   Average: {analysis['avg_response_time']:.3f}s")
        print(f"   Median: {analysis['median_response_time']:.3f}s")
        print(f"   Min: {analysis['min_response_time']:.3f}s")
        print(f"   Max: {analysis['max_response_time']:.3f}s")
        print(f"   95th Percentile: {analysis['p95_response_time']:.3f}s")
        print(f"   99th Percentile: {analysis['p99_response_time']:.3f}s")
        
        if analysis['failed_requests'] > 0:
            print(f"\n❌ FAILED REQUESTS:")
            for i, result in enumerate(results):
                if not result.get("success", False):
                    print(f"   {i+1}. {result.get('endpoint', 'unknown')} - {result.get('error', 'Unknown error')}")
        
        print("\n" + "="*60)

async def main():
    parser = argparse.ArgumentParser(description="Concurrency Test for VideoCall System")
    parser.add_argument("--requests", type=int, default=100, help="Number of concurrent requests")
    parser.add_argument("--type", choices=["health", "detection", "batch"], default="health", help="Test type")
    parser.add_argument("--backend-url", default="http://localhost:8000", help="Backend service URL")
    parser.add_argument("--ai-url", default="http://localhost:5001", help="AI service URL")
    
    args = parser.parse_args()
    
    tester = ConcurrencyTester(args.backend_url, args.ai_url)
    
    print("🚀 Starting Concurrency Test")
    print(f"   Backend URL: {args.backend_url}")
    print(f"   AI Service URL: {args.ai_url}")
    print(f"   Test Type: {args.type}")
    print(f"   Concurrent Requests: {args.requests}")
    
    start_time = time.time()
    results = await tester.run_concurrent_tests(args.requests, args.type)
    total_time = time.time() - start_time
    
    analysis = tester.analyze_results(results)
    analysis["total_test_time"] = total_time
    analysis["requests_per_second"] = args.requests / total_time if total_time > 0 else 0
    
    tester.print_results(results, analysis)
    
    print(f"\n🎯 PERFORMANCE METRICS:")
    print(f"   Total Test Time: {total_time:.2f}s")
    print(f"   Requests/Second: {analysis['requests_per_second']:.2f} RPS")
    
    # 保存结果到文件
    timestamp = int(time.time())
    filename = f"concurrency_test_{args.type}_{timestamp}.json"
    
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump({
            "test_config": {
                "requests": args.requests,
                "type": args.type,
                "backend_url": args.backend_url,
                "ai_url": args.ai_url
            },
            "results": results,
            "analysis": analysis
        }, f, indent=2, ensure_ascii=False)
    
    print(f"\n💾 Results saved to: {filename}")

if __name__ == "__main__":
    asyncio.run(main()) 