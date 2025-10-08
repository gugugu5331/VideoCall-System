#!/usr/bin/env python3
"""
Test script for AI Inference Node
Tests ASR, Emotion Detection, and Synthesis Detection
"""

import socket
import json
import argparse
import time


def create_tcp_connection(host, port, timeout=10):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(timeout)
    sock.connect((host, port))
    return sock


def send_json(sock, data):
    json_data = json.dumps(data, ensure_ascii=False) + '\n'
    sock.sendall(json_data.encode('utf-8'))


def receive_response(sock):
    response = ''
    while True:
        part = sock.recv(4096).decode('utf-8')
        response += part
        if '\n' in response:
            break
    return response.strip()


def close_connection(sock):
    if sock:
        sock.close()


def test_asr(host, port):
    """Test ASR (Automatic Speech Recognition)"""
    print("\n" + "=" * 60)
    print("Testing ASR (Automatic Speech Recognition)")
    print("=" * 60)
    
    sock = create_tcp_connection(host, port)
    
    try:
        # Setup ASR task
        print("\n[1] Setting up ASR task...")
        setup_data = {
            "request_id": "asr_001",
            "work_id": "llm",  # Changed from "asr" to "llm" - must match unit name
            "action": "setup",
            "object": "llm.setup",
            "data": {
                "model": "asr-model",
                "response_format": "llm.utf-8.stream",
                "input": "llm.utf-8.stream",
                "enoutput": True
            }
        }
        
        send_json(sock, setup_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Setup Response: {json.dumps(response_data, indent=2)}")
        
        if response_data.get('error', {}).get('code', 0) != 0:
            print("✗ ASR setup failed")
            return False
        
        work_id = response_data.get('work_id')
        print(f"✓ ASR task setup successful, work_id: {work_id}")
        
        # Perform inference
        print("\n[2] Performing ASR inference...")
        inference_data = {
            "request_id": "asr_002",
            "work_id": work_id,
            "action": "inference",
            "object": "llm.utf-8.stream",
            "data": {
                "delta": "sample audio data",
                "index": 0,
                "finish": True
            }
        }
        
        send_json(sock, inference_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Inference Response: {json.dumps(response_data, indent=2)}")
        
        # Exit task
        print("\n[3] Exiting ASR task...")
        exit_data = {
            "request_id": "asr_exit",
            "work_id": work_id,
            "action": "exit"
        }
        
        send_json(sock, exit_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Exit Response: {json.dumps(response_data, indent=2)}")
        
        print("\n✓ ASR test completed successfully")
        return True
        
    except Exception as e:
        print(f"\n✗ ASR test failed: {e}")
        return False
    finally:
        close_connection(sock)


def test_emotion(host, port):
    """Test Emotion Detection"""
    print("\n" + "=" * 60)
    print("Testing Emotion Detection")
    print("=" * 60)
    
    sock = create_tcp_connection(host, port)
    
    try:
        # Setup Emotion task
        print("\n[1] Setting up Emotion Detection task...")
        setup_data = {
            "request_id": "emotion_001",
            "work_id": "llm",  # Changed from "emotion" to "llm" - must match unit name
            "action": "setup",
            "object": "llm.setup",
            "data": {
                "model": "emotion-model",
                "response_format": "llm.utf-8.stream",
                "input": "llm.utf-8.stream",
                "enoutput": True
            }
        }
        
        send_json(sock, setup_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Setup Response: {json.dumps(response_data, indent=2)}")
        
        if response_data.get('error', {}).get('code', 0) != 0:
            print("✗ Emotion Detection setup failed")
            return False
        
        work_id = response_data.get('work_id')
        print(f"✓ Emotion Detection task setup successful, work_id: {work_id}")
        
        # Perform inference
        print("\n[2] Performing Emotion Detection inference...")
        inference_data = {
            "request_id": "emotion_002",
            "work_id": work_id,
            "action": "inference",
            "object": "llm.utf-8.stream",
            "data": {
                "delta": "I am very happy today!",
                "index": 0,
                "finish": True
            }
        }
        
        send_json(sock, inference_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Inference Response: {json.dumps(response_data, indent=2)}")
        
        # Exit task
        print("\n[3] Exiting Emotion Detection task...")
        exit_data = {
            "request_id": "emotion_exit",
            "work_id": work_id,
            "action": "exit"
        }
        
        send_json(sock, exit_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Exit Response: {json.dumps(response_data, indent=2)}")
        
        print("\n✓ Emotion Detection test completed successfully")
        return True
        
    except Exception as e:
        print(f"\n✗ Emotion Detection test failed: {e}")
        return False
    finally:
        close_connection(sock)


def test_synthesis(host, port):
    """Test Synthesis Detection"""
    print("\n" + "=" * 60)
    print("Testing Synthesis Detection (Deepfake Detection)")
    print("=" * 60)
    
    sock = create_tcp_connection(host, port)
    
    try:
        # Setup Synthesis task
        print("\n[1] Setting up Synthesis Detection task...")
        setup_data = {
            "request_id": "synthesis_001",
            "work_id": "llm",  # Changed from "synthesis" to "llm" - must match unit name
            "action": "setup",
            "object": "llm.setup",
            "data": {
                "model": "synthesis-model",
                "response_format": "llm.utf-8.stream",
                "input": "llm.utf-8.stream",
                "enoutput": True
            }
        }
        
        send_json(sock, setup_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Setup Response: {json.dumps(response_data, indent=2)}")
        
        if response_data.get('error', {}).get('code', 0) != 0:
            print("✗ Synthesis Detection setup failed")
            return False
        
        work_id = response_data.get('work_id')
        print(f"✓ Synthesis Detection task setup successful, work_id: {work_id}")
        
        # Perform inference
        print("\n[2] Performing Synthesis Detection inference...")
        inference_data = {
            "request_id": "synthesis_002",
            "work_id": work_id,
            "action": "inference",
            "object": "llm.utf-8.stream",
            "data": {
                "delta": "sample audio data for deepfake detection",
                "index": 0,
                "finish": True
            }
        }
        
        send_json(sock, inference_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Inference Response: {json.dumps(response_data, indent=2)}")
        
        # Exit task
        print("\n[3] Exiting Synthesis Detection task...")
        exit_data = {
            "request_id": "synthesis_exit",
            "work_id": work_id,
            "action": "exit"
        }
        
        send_json(sock, exit_data)
        response = receive_response(sock)
        response_data = json.loads(response)
        print(f"Exit Response: {json.dumps(response_data, indent=2)}")
        
        print("\n✓ Synthesis Detection test completed successfully")
        return True
        
    except Exception as e:
        print(f"\n✗ Synthesis Detection test failed: {e}")
        return False
    finally:
        close_connection(sock)


def main(host, port):
    print("=" * 60)
    print("AI Inference Node Test Suite")
    print("=" * 60)
    print(f"Target: {host}:{port}")
    
    results = {
        "ASR": False,
        "Emotion Detection": False,
        "Synthesis Detection": False
    }
    
    # Test ASR
    results["ASR"] = test_asr(host, port)
    print("\n[Waiting 2 seconds before next test...]")
    time.sleep(2)

    # Test Emotion Detection
    results["Emotion Detection"] = test_emotion(host, port)
    print("\n[Waiting 2 seconds before next test...]")
    time.sleep(2)

    # Test Synthesis Detection
    results["Synthesis Detection"] = test_synthesis(host, port)
    
    # Print summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    for test_name, result in results.items():
        status = "✓ PASSED" if result else "✗ FAILED"
        print(f"{test_name}: {status}")
    
    all_passed = all(results.values())
    print("\n" + "=" * 60)
    if all_passed:
        print("✓ All tests passed!")
    else:
        print("✗ Some tests failed")
    print("=" * 60)
    
    return 0 if all_passed else 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Test AI Inference Node')
    parser.add_argument('--host', type=str, default='localhost', help='Server hostname (default: localhost)')
    parser.add_argument('--port', type=int, default=19001, help='Server port (default: 19001)')

    args = parser.parse_args()
    exit(main(args.host, args.port))

