#!/usr/bin/env python3
"""
Test script for AI Detection Node
Tests the new Edge-Model-Infra based AI detection service
"""

import socket
import json
import time
import argparse
import base64
import os

def create_tcp_connection(host, port):
    """Create TCP connection to unit-manager"""
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((host, port))
    return sock

def send_json(sock, data):
    """Send JSON data over TCP"""
    json_data = json.dumps(data, ensure_ascii=False) + '\n'
    sock.sendall(json_data.encode('utf-8'))

def receive_response(sock):
    """Receive JSON response"""
    response = ''
    while True:
        part = sock.recv(4096).decode('utf-8')
        response += part
        if '\n' in response:
            break
    return response.strip()

def test_detection_setup(sock):
    """Test detection service setup"""
    print("Testing detection setup...")
    
    setup_request = {
        "request_id": "setup_001",
        "work_id": "detection",
        "action": "setup",
        "data": {
            "detector_type": "face_swap",
            "model_path": "/app/models/face_swap_detector.pb"
        }
    }
    
    send_json(sock, setup_request)
    response = receive_response(sock)
    
    try:
        response_data = json.loads(response)
        print(f"Setup Response: {response_data}")
        return response_data.get('data', {}).get('status') == 'success'
    except json.JSONDecodeError:
        print(f"Invalid JSON response: {response}")
        return False

def test_image_detection(sock, image_path=None):
    """Test image detection"""
    print("Testing image detection...")
    
    # Use a dummy image path if none provided
    if not image_path:
        image_path = "/tmp/test_image.jpg"
    
    detect_request = {
        "request_id": "detect_001",
        "work_id": "detection",
        "action": "detect",
        "data": {
            "file_path": image_path,
            "file_type": "image"
        }
    }
    
    send_json(sock, detect_request)
    response = receive_response(sock)
    
    try:
        response_data = json.loads(response)
        print(f"Detection Response: {response_data}")
        
        # Extract task ID for status checking
        task_data = response_data.get('data', {})
        if isinstance(task_data, str):
            task_data = json.loads(task_data)
        
        task_id = task_data.get('task_id')
        return task_id
    except json.JSONDecodeError:
        print(f"Invalid JSON response: {response}")
        return None

def test_status_check(sock, task_id):
    """Test task status checking"""
    print(f"Checking status for task: {task_id}")
    
    status_request = {
        "request_id": "status_001",
        "work_id": "detection",
        "action": "status",
        "data": {
            "task_id": task_id
        }
    }
    
    send_json(sock, status_request)
    response = receive_response(sock)
    
    try:
        response_data = json.loads(response)
        print(f"Status Response: {response_data}")
        return response_data
    except json.JSONDecodeError:
        print(f"Invalid JSON response: {response}")
        return None

def test_llm_compatibility(sock):
    """Test LLM-style API compatibility"""
    print("Testing LLM-style API compatibility...")
    
    # Test setup similar to LLM setup
    llm_setup = {
        "request_id": "llm_001",
        "work_id": "detection",
        "action": "setup",
        "object": "detection.setup",
        "data": {
            "model": "face_swap_detector",
            "response_format": "detection.json",
            "max_token_len": 1023
        }
    }
    
    send_json(sock, llm_setup)
    response = receive_response(sock)
    
    try:
        response_data = json.loads(response)
        print(f"LLM Setup Response: {response_data}")
        return True
    except json.JSONDecodeError:
        print(f"Invalid JSON response: {response}")
        return False

def main():
    parser = argparse.ArgumentParser(description='Test AI Detection Node')
    parser.add_argument('--host', default='localhost', help='Unit manager host')
    parser.add_argument('--port', type=int, default=10001, help='Unit manager port')
    parser.add_argument('--image', help='Path to test image file')
    
    args = parser.parse_args()
    
    try:
        print(f"Connecting to {args.host}:{args.port}...")
        sock = create_tcp_connection(args.host, args.port)
        
        # Test sequence
        print("\n=== AI Detection Node Test ===")
        
        # 1. Test setup
        setup_success = test_detection_setup(sock)
        if not setup_success:
            print("Setup failed, continuing with other tests...")
        
        time.sleep(1)
        
        # 2. Test image detection
        task_id = test_image_detection(sock, args.image)
        if task_id:
            time.sleep(2)  # Wait for processing
            
            # 3. Test status check
            test_status_check(sock, task_id)
        
        time.sleep(1)
        
        # 4. Test LLM compatibility
        test_llm_compatibility(sock)
        
        print("\n=== Test completed ===")
        
    except Exception as e:
        print(f"Test failed: {e}")
    finally:
        if 'sock' in locals():
            sock.close()

if __name__ == "__main__":
    main()
