#!/usr/bin/env python3
"""
Stress Test for AI Inference Node
Tests system stability under high load and complex scenarios
"""

import socket
import json
import time
import random
import argparse
from datetime import datetime
from typing import List, Tuple, Dict


class AIInferenceClient:
    """Client for AI Inference Node"""
    
    def __init__(self, host='localhost', port=19001, timeout=10):
        self.host = host
        self.port = port
        self.timeout = timeout
        self.stats = {
            'total': 0,
            'success': 0,
            'failed': 0,
            'timeouts': 0,
            'errors': {}
        }
        self.response_times = []
    
    def _create_connection(self):
        """Create TCP connection"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(self.timeout)
        sock.connect((self.host, self.port))
        return sock
    
    def _send_request(self, sock, request):
        """Send JSON request and receive response"""
        request_str = json.dumps(request) + '\n'
        sock.sendall(request_str.encode('utf-8'))
        
        response = b''
        while True:
            chunk = sock.recv(4096)
            if not chunk:
                break
            response += chunk
            if b'\n' in response:
                break
        
        return json.loads(response.decode('utf-8').strip())
    
    def run_task(self, model_type, test_data, test_id):
        """Run a complete task: setup → inference → exit"""
        start_time = time.time()
        self.stats['total'] += 1
        
        sock = None
        work_id = None
        
        try:
            sock = self._create_connection()
            
            # Setup
            setup_request = {
                "request_id": f"{model_type}_{test_id}_setup",
                "work_id": "llm",
                "action": "setup",
                "object": "llm.setup",
                "data": {
                    "model": f"{model_type}-model",
                    "response_format": "llm.utf-8.stream",
                    "input": "llm.utf-8.stream",
                    "enoutput": True
                }
            }
            
            setup_response = self._send_request(sock, setup_request)
            
            if setup_response.get('error', {}).get('code', -1) != 0:
                error_msg = setup_response.get('error', {}).get('message', 'Unknown error')
                self.stats['failed'] += 1
                self.stats['errors'][f"setup_{model_type}"] = self.stats['errors'].get(f"setup_{model_type}", 0) + 1
                return False, f"Setup failed: {error_msg}", 0
            
            work_id = setup_response.get('work_id')
            
            # Inference
            inference_request = {
                "request_id": f"{model_type}_{test_id}_inference",
                "work_id": work_id,
                "action": "inference",
                "object": "llm.utf-8.stream",
                "data": {
                    "delta": test_data,
                    "index": 0,
                    "finish": True
                }
            }
            
            inference_response = self._send_request(sock, inference_request)
            
            if inference_response.get('error', {}).get('code', -1) != 0:
                error_msg = inference_response.get('error', {}).get('message', 'Unknown error')
                self.stats['failed'] += 1
                self.stats['errors'][f"inference_{model_type}"] = self.stats['errors'].get(f"inference_{model_type}", 0) + 1
                return False, f"Inference failed: {error_msg}", 0
            
            # Exit
            exit_request = {
                "request_id": f"{model_type}_{test_id}_exit",
                "work_id": work_id,
                "action": "exit"
            }
            
            exit_response = self._send_request(sock, exit_request)
            
            elapsed_time = time.time() - start_time
            self.response_times.append(elapsed_time)
            self.stats['success'] += 1
            
            return True, f"Success (work_id: {work_id})", elapsed_time
            
        except socket.timeout:
            self.stats['timeouts'] += 1
            self.stats['failed'] += 1
            elapsed_time = time.time() - start_time
            return False, "Timeout", elapsed_time
            
        except Exception as e:
            self.stats['failed'] += 1
            self.stats['errors'][str(type(e).__name__)] = self.stats['errors'].get(str(type(e).__name__), 0) + 1
            elapsed_time = time.time() - start_time
            return False, f"Error: {e}", elapsed_time
            
        finally:
            if sock:
                sock.close()
    
    def print_stats(self):
        """Print statistics"""
        print("\n" + "=" * 70)
        print("STRESS TEST STATISTICS")
        print("=" * 70)
        print(f"Total Requests:    {self.stats['total']}")
        print(f"Successful:        {self.stats['success']} ({self.stats['success']/max(1,self.stats['total'])*100:.1f}%)")
        print(f"Failed:            {self.stats['failed']} ({self.stats['failed']/max(1,self.stats['total'])*100:.1f}%)")
        print(f"Timeouts:          {self.stats['timeouts']}")
        
        if self.response_times:
            avg_time = sum(self.response_times) / len(self.response_times)
            min_time = min(self.response_times)
            max_time = max(self.response_times)
            print(f"\nResponse Times:")
            print(f"  Average:         {avg_time:.3f}s")
            print(f"  Min:             {min_time:.3f}s")
            print(f"  Max:             {max_time:.3f}s")
        
        if self.stats['errors']:
            print(f"\nError Breakdown:")
            for error_type, count in sorted(self.stats['errors'].items(), key=lambda x: x[1], reverse=True):
                print(f"  {error_type}: {count}")
        
        print("=" * 70)


def test_sequential_same_model(client: AIInferenceClient, model_type: str, count: int, delay: float = 0.5):
    """Test: Sequential calls to the same model"""
    print(f"\n{'='*70}")
    print(f"TEST 1: Sequential {model_type.upper()} calls ({count} times)")
    print(f"{'='*70}")
    
    for i in range(count):
        test_data = f"test data {i+1}"
        success, message, elapsed = client.run_task(model_type, test_data, f"seq_{i+1}")
        
        status = "✓" if success else "✗"
        print(f"[{i+1}/{count}] {status} {model_type}: {message} ({elapsed:.3f}s)")
        
        if i < count - 1 and delay > 0:
            time.sleep(delay)


def test_interleaved_models(client: AIInferenceClient, sequence: List[Tuple[str, str]], delay: float = 0.5):
    """Test: Interleaved calls to different models"""
    print(f"\n{'='*70}")
    print(f"TEST 2: Interleaved model calls ({len(sequence)} calls)")
    print(f"{'='*70}")
    
    for i, (model_type, test_data) in enumerate(sequence):
        success, message, elapsed = client.run_task(model_type, test_data, f"inter_{i+1}")
        
        status = "✓" if success else "✗"
        print(f"[{i+1}/{len(sequence)}] {status} {model_type}: {message} ({elapsed:.3f}s)")
        
        if i < len(sequence) - 1 and delay > 0:
            time.sleep(delay)


def test_rapid_fire(client: AIInferenceClient, model_type: str, count: int):
    """Test: Rapid consecutive calls with minimal delay"""
    print(f"\n{'='*70}")
    print(f"TEST 3: Rapid-fire {model_type.upper()} calls ({count} times, no delay)")
    print(f"{'='*70}")
    
    for i in range(count):
        test_data = f"rapid test {i+1}"
        success, message, elapsed = client.run_task(model_type, test_data, f"rapid_{i+1}")
        
        status = "✓" if success else "✗"
        print(f"[{i+1}/{count}] {status} {model_type}: {message} ({elapsed:.3f}s)")


def test_random_pattern(client: AIInferenceClient, models: List[str], count: int, delay: float = 0.3):
    """Test: Random model selection"""
    print(f"\n{'='*70}")
    print(f"TEST 4: Random pattern ({count} calls)")
    print(f"{'='*70}")
    
    for i in range(count):
        model_type = random.choice(models)
        test_data = f"random test {i+1}"
        success, message, elapsed = client.run_task(model_type, test_data, f"random_{i+1}")
        
        status = "✓" if success else "✗"
        print(f"[{i+1}/{count}] {status} {model_type}: {message} ({elapsed:.3f}s)")
        
        if i < count - 1 and delay > 0:
            time.sleep(delay)


def main(host='localhost', port=19001):
    """Main stress test"""
    print("=" * 70)
    print("AI INFERENCE NODE - STRESS TEST")
    print("=" * 70)
    print(f"Target: {host}:{port}")
    print(f"Start Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 70)
    
    client = AIInferenceClient(host, port, timeout=15)
    
    # Test 1: Sequential same model (ASR)
    test_sequential_same_model(client, "asr", count=5, delay=0.5)
    
    # Test 2: Interleaved models
    interleaved_sequence = [
        ("asr", "audio sample 1"),
        ("emotion", "I am very happy today!"),
        ("asr", "audio sample 2"),
        ("synthesis", "voice sample 1"),
        ("emotion", "I feel sad and lonely"),
        ("asr", "audio sample 3"),
        ("synthesis", "voice sample 2"),
        ("emotion", "This is amazing!"),
    ]
    test_interleaved_models(client, interleaved_sequence, delay=0.5)
    
    # Test 3: Rapid-fire (Emotion)
    test_rapid_fire(client, "emotion", count=5)
    
    # Test 4: Random pattern
    test_random_pattern(client, ["asr", "emotion", "synthesis"], count=10, delay=0.3)
    
    # Print final statistics
    client.print_stats()
    
    # Final summary
    print(f"\nEnd Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    if client.stats['failed'] == 0:
        print("\n✓ ALL TESTS PASSED! System is stable under stress.")
        return 0
    else:
        print(f"\n✗ {client.stats['failed']} TESTS FAILED! Please check logs.")
        return 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Stress Test for AI Inference Node')
    parser.add_argument('--host', type=str, default='localhost', help='Server hostname (default: localhost)')
    parser.add_argument('--port', type=int, default=19001, help='Server port (default: 19001)')
    
    args = parser.parse_args()
    exit(main(args.host, args.port))

