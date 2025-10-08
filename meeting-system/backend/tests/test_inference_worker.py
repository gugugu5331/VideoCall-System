#!/usr/bin/env python3
"""
测试推理 Worker 是否正常工作
"""

import zmq
import json
import base64
import time

def test_inference_worker():
    """测试推理 worker"""
    print("连接到推理 worker...")
    
    context = zmq.Context()
    socket = context.socket(zmq.REQ)
    socket.connect("tcp://localhost:5010")
    
    # 测试 setup 请求
    print("\n1. 测试 Setup 请求...")
    setup_request = {
        "request_id": f"test_setup_{int(time.time())}",
        "work_id": "test_work_123",
        "action": "setup",
        "object": "model.setup",
        "data": {
            "model": "emotion-detection",
            "inference_type": "emotion_detection",
            "response_format": "json"
        }
    }
    
    socket.send_string(json.dumps(setup_request))
    response_str = socket.recv_string()
    response = json.loads(response_str)
    print(f"Setup Response: {json.dumps(response, indent=2)}")
    
    if response.get("error", {}).get("code") != 0:
        print(f"❌ Setup failed: {response.get('error')}")
        return False
    
    print("✅ Setup successful")
    
    # 测试 inference 请求
    print("\n2. 测试 Inference 请求...")
    
    # 创建测试图像数据
    test_image_data = b"test_image_data_placeholder"
    image_base64 = base64.b64encode(test_image_data).decode('utf-8')
    
    inference_request = {
        "request_id": f"test_inference_{int(time.time())}",
        "work_id": "test_work_123",
        "action": "inference",
        "object": "emotion_detection",
        "data": {
            "image_data": image_base64,
            "image_format": "jpg"
        }
    }
    
    socket.send_string(json.dumps(inference_request))
    response_str = socket.recv_string()
    response = json.loads(response_str)
    print(f"Inference Response: {json.dumps(response, indent=2)[:500]}")
    
    if response.get("error", {}).get("code") != 0:
        print(f"❌ Inference failed: {response.get('error')}")
        return False
    
    print("✅ Inference successful")
    
    socket.close()
    context.term()
    
    return True

if __name__ == "__main__":
    try:
        success = test_inference_worker()
        if success:
            print("\n🎉 推理 Worker 测试通过!")
        else:
            print("\n❌ 推理 Worker 测试失败!")
            exit(1)
    except Exception as e:
        print(f"\n❌ 测试异常: {e}")
        import traceback
        traceback.print_exc()
        exit(1)

