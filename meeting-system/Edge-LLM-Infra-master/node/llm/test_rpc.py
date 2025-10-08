#!/usr/bin/env python3
"""Test RPC call to llm node"""

import zmq
import json
import sys

def main():
    context = zmq.Context()
    socket = context.socket(zmq.REQ)
    socket.setsockopt(zmq.RCVTIMEO, 5000)  # 5 second timeout
    socket.setsockopt(zmq.SNDTIMEO, 5000)
    
    try:
        print("Connecting to ipc:///tmp/rpc.llm...")
        socket.connect("ipc:///tmp/rpc.llm")
        print("✓ Connected")
        
        # Prepare RPC call - ZMQ multipart message
        # Part 1: action name
        action = "setup"

        # Part 2: parameters (zmq_url + raw JSON)
        zmq_url = "ipc:///tmp/test_zmq.sock"
        raw_json = json.dumps({
            "request_id": "test_rpc_001",
            "work_id": "llm",
            "action": "setup",
            "object": "llm.setup",
            "data": {
                "model": "asr-model",
                "response_format": "llm.utf-8.stream",
                "input": "llm.utf-8.stream",
                "enoutput": True
            }
        })

        # Combine parameters using set_param format (param0\nparam1)
        params = f"{zmq_url}\n{raw_json}"

        print("\nSending RPC request...")
        print(f"Action: {action}")
        print(f"Params: {params[:200]}...")

        # Send as multipart message
        socket.send_string(action, zmq.SNDMORE)
        socket.send_string(params)
        print("\n✓ Request sent, waiting for response...")

        raw_response = socket.recv()
        print(f"\n✓ Received raw response ({len(raw_response)} bytes):")
        print(f"Raw: {raw_response}")
        print(f"Decoded: {raw_response.decode('utf-8', errors='replace')}")

        try:
            response = json.loads(raw_response)
            print("\n✓ Parsed JSON response:")
            print(json.dumps(response, indent=2))
        except:
            print("\n⚠ Response is not JSON")
        
        return 0
        
    except zmq.error.Again:
        print("\n✗ Timeout - no response from RPC server")
        return 1
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return 1
    finally:
        socket.close()
        context.term()

if __name__ == "__main__":
    sys.exit(main())

