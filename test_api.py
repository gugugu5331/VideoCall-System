import requests
import json

# 基础URL
base_url = "http://localhost:8000"

# 登录获取token
login_data = {
    "username": "testuser",
    "password": "password123"
}

print("正在登录...")
response = requests.post(f"{base_url}/api/v1/auth/login", json=login_data)
if response.status_code == 200:
    login_result = response.json()
    token = login_result["token"]
    user_uuid = login_result["user"]["uuid"]
    print(f"登录成功，用户UUID: {user_uuid}")
    
    # 测试通话API
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    call_data = {
        "callee_id": user_uuid,
        "call_type": "video"
    }
    
    print("正在测试通话API...")
    response = requests.post(f"{base_url}/api/v1/calls/start", json=call_data, headers=headers)
    print(f"状态码: {response.status_code}")
    print(f"响应: {response.text}")
    
else:
    print(f"登录失败: {response.status_code}")
    print(response.text) 