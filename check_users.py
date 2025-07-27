import requests
import json

# 基础URL
base_url = "http://localhost:8000"

# 登录获取token
login_data = {
    "username": "testuser",
    "password": "password123"
}

print("正在尝试登录现有用户...")
response = requests.post(f"{base_url}/api/v1/auth/login", json=login_data)
if response.status_code == 200:
    login_result = response.json()
    token = login_result["token"]
    user_uuid = login_result["user"]["uuid"]
    print(f"✅ 登录成功！")
    print(f"用户UUID: {user_uuid}")
    print(f"用户名: {login_result['user']['username']}")
    print(f"邮箱: {login_result['user']['email']}")
    
    # 获取用户资料
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    response = requests.get(f"{base_url}/api/v1/user/profile", headers=headers)
    if response.status_code == 200:
        profile = response.json()
        print(f"✅ 用户资料获取成功")
        if 'full_name' in profile:
            print(f"完整姓名: {profile['full_name']}")
        if 'last_login' in profile:
            print(f"最后登录: {profile['last_login']}")
    
    print("\n💡 解决方案：")
    print("1. 使用现有用户登录（用户名: testuser, 密码: password123）")
    print("2. 或者使用不同的用户名注册新用户")
    print("3. 或者清理数据库中的测试用户")
    
else:
    print(f"❌ 登录失败: {response.status_code}")
    print(response.text)
    print("\n💡 建议：使用不同的用户名注册新用户") 