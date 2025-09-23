#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简单WebSocket测试
"""

import asyncio
import websockets
import json

async def test_websocket():
    # 连接WebSocket
    uri = "ws://localhost:8000/ws/call/test-call-123"
    
    print(f"连接WebSocket: {uri}")
    
    try:
        async with websockets.connect(uri) as websocket:
            print("✅ WebSocket连接成功")
            
            # 等待消息
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=10.0)
                data = json.loads(message)
                print(f"收到消息: {data}")
            except asyncio.TimeoutError:
                print("等待消息超时")
                
    except Exception as e:
        print(f"❌ WebSocket连接失败: {e}")

if __name__ == "__main__":
    asyncio.run(test_websocket()) 