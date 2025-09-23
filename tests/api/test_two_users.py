#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试两个用户连接
"""

import asyncio
import websockets
import json

async def test_two_users():
    uri = "ws://localhost:8000/ws/call/test-call-456"
    
    print("=== 测试两个用户连接 ===")
    
    # 第一个用户连接
    print("\n1. 第一个用户连接...")
    try:
        async with websockets.connect(uri) as websocket1:
            print("✅ 第一个用户连接成功")
            
            # 等待第一个用户的消息
            message1 = await asyncio.wait_for(websocket1.recv(), timeout=5.0)
            data1 = json.loads(message1)
            print(f"第一个用户收到: {data1.get('type')}")
            
            # 第二个用户连接
            print("\n2. 第二个用户连接...")
            async with websockets.connect(uri) as websocket2:
                print("✅ 第二个用户连接成功")
                
                # 等待第二个用户的消息
                message2 = await asyncio.wait_for(websocket2.recv(), timeout=5.0)
                data2 = json.loads(message2)
                print(f"第二个用户收到: {data2.get('type')}")
                
                # 等待第一个用户是否收到join消息
                print("\n3. 等待第一个用户是否收到join消息...")
                try:
                    join_message = await asyncio.wait_for(websocket1.recv(), timeout=5.0)
                    join_data = json.loads(join_message)
                    print(f"🎉 第一个用户收到join消息: {join_data.get('type')}")
                    print(f"   新用户ID: {join_data.get('user_id')}")
                except asyncio.TimeoutError:
                    print("❌ 第一个用户没有收到join消息")
                
                # 等待一段时间
                await asyncio.sleep(2)
                
    except Exception as e:
        print(f"❌ 测试失败: {e}")

if __name__ == "__main__":
    asyncio.run(test_two_users()) 