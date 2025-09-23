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
    print(f"✅ 登录成功！")
    
    # 设置请求头
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    # 测试开始通话API
    start_call_data = {
        "callee_id": login_result["user"]["uuid"],
        "call_type": "video"
    }
    
    print("正在测试开始通话API...")
    response = requests.post(f"{base_url}/api/v1/calls/start", json=start_call_data, headers=headers)
    print(f"开始通话状态码: {response.status_code}")
    print(f"开始通话响应: {response.text}")
    
    if response.status_code == 201:
        call_result = response.json()
        call_id = call_result["call"]["id"]
        print(f"通话ID: {call_id}")
        
        # 测试结束通话API
        end_call_data = {
            "call_id": str(call_id)
        }
        
        print("正在测试结束通话API...")
        response = requests.post(f"{base_url}/api/v1/calls/end", json=end_call_data, headers=headers)
        print(f"结束通话状态码: {response.status_code}")
        print(f"结束通话响应: {response.text}")
        
        if response.status_code == 200:
            print("✅ 结束通话API测试成功！")
        else:
            print("❌ 结束通话API测试失败！")
    else:
        print("❌ 开始通话API测试失败！")
        
else:
    print(f"❌ 登录失败: {response.status_code}")
    print(response.text) 