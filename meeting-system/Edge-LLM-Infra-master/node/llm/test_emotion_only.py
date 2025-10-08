#!/usr/bin/env python3
"""Test only Emotion Detection"""

import socket
import json
import time

def send_request(sock, request):
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

def test_emotion():
    """Test Emotion Detection"""
    print("\n" + "="*60)
    print("Testing Emotion Detection (Standalone)")
    print("="*60)
    
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(10)  # 10 second timeout
    
    try:
        print("\n[1] Connecting to localhost:19001...")
        sock.connect(('localhost', 19001))
        print("✓ Connected")
        
        # Setup
        print("\n[2] Setting up Emotion Detection task...")
        setup_request = {
            "request_id": "emotion_001",
            "work_id": "llm",
            "action": "setup",
            "object": "llm.setup",
            "data": {
                "model": "emotion-model",
                "response_format": "llm.utf-8.stream",
                "input": "llm.utf-8.stream",
                "enoutput": True
            }
        }
        
        print(f"Sending: {json.dumps(setup_request, indent=2)}")
        setup_response = send_request(sock, setup_request)
        print(f"Setup Response: {json.dumps(setup_response, indent=2)}")
        
        if setup_response.get('error', {}).get('code', -1) != 0:
            print(f"✗ Setup failed: {setup_response.get('error', {}).get('message', 'Unknown error')}")
            return False
        
        work_id = setup_response.get('work_id', '')
        print(f"✓ Emotion task setup successful, work_id: {work_id}")
        
        # Inference
        print("\n[3] Performing Emotion Detection inference...")
        inference_request = {
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
        
        inference_response = send_request(sock, inference_request)
        print(f"Inference Response: {json.dumps(inference_response, indent=2)}")
        
        # Exit
        print("\n[4] Exiting Emotion Detection task...")
        exit_request = {
            "request_id": "emotion_exit",
            "work_id": work_id,
            "action": "exit"
        }
        
        exit_response = send_request(sock, exit_request)
        print(f"Exit Response: {json.dumps(exit_response, indent=2)}")
        
        print("\n✓ Emotion Detection test completed successfully")
        return True
        
    except socket.timeout:
        print("\n✗ Timeout waiting for response")
        return False
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False
    finally:
        sock.close()

if __name__ == "__main__":
    success = test_emotion()
    exit(0 if success else 1)

