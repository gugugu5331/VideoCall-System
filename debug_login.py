#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
登录调试脚本
"""

import requests
import json

# 配置
BASE_URL = "http://localhost:8000"
API_BASE = f"{BASE_URL}/api/v1"

def test_login_detailed():
    """详细测试登录功能"""
    print("🔍 详细测试登录功能")
    print("=" * 50)
    
    # 测试数据
    login_data = {
        "username": "alice",
        "password": "password123"
    }
    
    print(f"登录数据: {json.dumps(login_data, ensure_ascii=False, indent=2)}")
    print(f"请求URL: {API_BASE}/auth/login")
    print()
    
    try:
        # 发送登录请求
        response = requests.post(
            f"{API_BASE}/auth/login",
            json=login_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"响应状态码: {response.status_code}")
        print(f"响应头: {dict(response.headers)}")
        print()
        
        if response.status_code == 200:
            print("✅ 登录成功")
            print(f"响应内容: {json.dumps(response.json(), ensure_ascii=False, indent=2)}")
            return True
        else:
            print("❌ 登录失败")
            print(f"响应内容: {response.text}")
            
            # 尝试解析JSON错误信息
            try:
                error_data = response.json()
                print(f"错误详情: {json.dumps(error_data, ensure_ascii=False, indent=2)}")
            except:
                print("无法解析错误响应为JSON")
            
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ 请求异常: {e}")
        return False
    except Exception as e:
        print(f"❌ 其他异常: {e}")
        return False

def test_register_detailed():
    """详细测试注册功能"""
    print("\n🔍 详细测试注册功能")
    print("=" * 50)
    
    # 测试数据
    register_data = {
        "username": "testuser_debug",
        "email": "testdebug@example.com",
        "password": "password123",
        "full_name": "Test Debug User"
    }
    
    print(f"注册数据: {json.dumps(register_data, ensure_ascii=False, indent=2)}")
    print(f"请求URL: {API_BASE}/auth/register")
    print()
    
    try:
        # 发送注册请求
        response = requests.post(
            f"{API_BASE}/auth/register",
            json=register_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        print(f"响应状态码: {response.status_code}")
        print(f"响应头: {dict(response.headers)}")
        print()
        
        if response.status_code == 201:
            print("✅ 注册成功")
            print(f"响应内容: {json.dumps(response.json(), ensure_ascii=False, indent=2)}")
            return True
        elif response.status_code == 409:
            print("ℹ️  用户已存在")
            print(f"响应内容: {response.text}")
            return True
        else:
            print("❌ 注册失败")
            print(f"响应内容: {response.text}")
            
            # 尝试解析JSON错误信息
            try:
                error_data = response.json()
                print(f"错误详情: {json.dumps(error_data, ensure_ascii=False, indent=2)}")
            except:
                print("无法解析错误响应为JSON")
            
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ 请求异常: {e}")
        return False
    except Exception as e:
        print(f"❌ 其他异常: {e}")
        return False

def main():
    print("🚀 开始详细登录调试")
    print(f"目标服务器: {BASE_URL}")
    print("=" * 60)
    
    # 先测试注册
    register_success = test_register_detailed()
    
    # 再测试登录
    login_success = test_login_detailed()
    
    print("\n" + "=" * 60)
    print("调试总结:")
    print(f"注册测试: {'✅ 成功' if register_success else '❌ 失败'}")
    print(f"登录测试: {'✅ 成功' if login_success else '❌ 失败'}")
    
    if not login_success:
        print("\n💡 登录失败可能的原因:")
        print("1. 数据库连接问题")
        print("2. 密码加密/验证问题")
        print("3. JWT token生成问题")
        print("4. 用户状态问题")
        print("\n建议检查后端日志获取更详细的错误信息")

if __name__ == "__main__":
    main() 