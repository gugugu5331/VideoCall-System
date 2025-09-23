#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall Backend API Test Script
"""

import requests
import json
import sys

def test_health_check():
    """Test health check endpoint"""
    print("1. Testing Health Check...")
    try:
        response = requests.get("http://localhost:8000/health")
        if response.status_code == 200:
            data = response.json()
            print(f"   ✅ Health Check: {data['status']}")
            print(f"   Message: {data['message']}")
            return True
        else:
            print(f"   ❌ Health Check failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"   ❌ Health Check error: {e}")
        return False

def test_root_endpoint():
    """Test root endpoint"""
    print("\n2. Testing Root Endpoint...")
    try:
        response = requests.get("http://localhost:8000/")
        if response.status_code == 200:
            data = response.json()
            print(f"   ✅ Root Endpoint: {data['message']}")
            print(f"   Version: {data['version']}")
            return True
        else:
            print(f"   ❌ Root Endpoint failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"   ❌ Root Endpoint error: {e}")
        return False

def test_user_registration():
    """Test user registration"""
    print("\n3. Testing User Registration...")
    try:
        data = {
            "username": "python_test_user",
            "email": "python@example.com",
            "password": "password123",
            "full_name": "Python Test User"
        }
        response = requests.post(
            "http://localhost:8000/api/v1/auth/register",
            json=data,
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 201:
            result = response.json()
            print(f"   ✅ User Registration: {result['message']}")
            print(f"   User ID: {result['user']['id']}")
            return True
        elif response.status_code == 409:
            print("   ⚠️  User already exists (this is normal)")
            return True
        else:
            print(f"   ❌ User Registration failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"   ❌ User Registration error: {e}")
        return False

def test_user_login():
    """Test user login"""
    print("\n4. Testing User Login...")
    try:
        data = {
            "username": "testuser",
            "password": "password123"
        }
        response = requests.post(
            "http://localhost:8000/api/v1/auth/login",
            json=data,
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 200:
            result = response.json()
            print(f"   ✅ User Login: {result['message']}")
            print(f"   Token length: {len(result['token'])} characters")
            return result['token']
        else:
            print(f"   ❌ User Login failed: {response.status_code}")
            return None
    except Exception as e:
        print(f"   ❌ User Login error: {e}")
        return None

def test_protected_endpoint(token):
    """Test protected endpoint with token"""
    print("\n5. Testing Protected Endpoint...")
    if not token:
        print("   ⚠️  No token available")
        return False
    
    try:
        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }
        response = requests.get(
            "http://localhost:8000/api/v1/user/profile",
            headers=headers
        )
        if response.status_code == 200:
            result = response.json()
            print(f"   ✅ Protected Endpoint: User profile retrieved")
            print(f"   Username: {result['user']['username']}")
            print(f"   Email: {result['user']['email']}")
            return True
        else:
            print(f"   ❌ Protected Endpoint failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"   ❌ Protected Endpoint error: {e}")
        return False

def main():
    """Main test function"""
    print("=" * 50)
    print("VideoCall Backend API Test")
    print("=" * 50)
    
    # Test all endpoints
    health_ok = test_health_check()
    root_ok = test_root_endpoint()
    reg_ok = test_user_registration()
    token = test_user_login()
    protected_ok = test_protected_endpoint(token)
    
    # Summary
    print("\n" + "=" * 50)
    print("Test Summary:")
    print("=" * 50)
    print(f"Health Check: {'✅' if health_ok else '❌'}")
    print(f"Root Endpoint: {'✅' if root_ok else '❌'}")
    print(f"User Registration: {'✅' if reg_ok else '❌'}")
    print(f"User Login: {'✅' if token else '❌'}")
    print(f"Protected Endpoint: {'✅' if protected_ok else '❌'}")
    
    if all([health_ok, root_ok, reg_ok, token, protected_ok]):
        print("\n🎉 All tests passed! Backend is fully operational.")
        return 0
    else:
        print("\n⚠️  Some tests failed. Please check the backend service.")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 