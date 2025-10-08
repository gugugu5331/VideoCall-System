#!/usr/bin/env python3
"""Simple test script to debug AI inference node"""

import socket
import json
import sys

def create_tcp_connection(host, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(10)  # 10 second timeout
    sock.connect((host, port))
    return sock

def send_json(sock, data):
    json_data = json.dumps(data, ensure_ascii=False) + '\n'
    print(f"Sending: {json_data}")
    sock.sendall(json_data.encode('utf-8'))

def receive_response(sock):
    response = ''
    while True:
        part = sock.recv(4096).decode('utf-8')
        response += part
        if '\n' in response:
            break
    return response.strip()

def main():
    host = 'localhost'
    port = 19001
    
    print("="*60)
    print("Simple AI Inference Node Test")
    print("="*60)
    
    try:
        print(f"\nConnecting to {host}:{port}...")
        sock = create_tcp_connection(host, port)
        print("✓ Connected successfully")
        
        # Setup request
        print("\nSending setup request...")
        setup_data = {
            "request_id": "test_001",
            "work_id": "llm",
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
        
        print("\nWaiting for response...")
        response = receive_response(sock)
        print(f"Received: {response}")
        
        response_data = json.loads(response)
        print(f"\nParsed response:")
        print(json.dumps(response_data, indent=2))
        
        error = response_data.get('error', {})
        if error.get('code', 0) != 0:
            print(f"\n✗ Error: {error.get('message')}")
            return 1
        
        work_id = response_data.get('work_id')
        print(f"\n✓ Setup successful! work_id: {work_id}")
        
        # Exit request
        print("\nSending exit request...")
        exit_data = {
            "request_id": "test_exit",
            "work_id": work_id,
            "action": "exit"
        }
        
        send_json(sock, exit_data)
        response = receive_response(sock)
        print(f"Exit response: {response}")
        
        sock.close()
        print("\n✓ Test completed successfully!")
        return 0
        
    except socket.timeout:
        print("\n✗ Socket timeout - no response from server")
        return 1
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())

