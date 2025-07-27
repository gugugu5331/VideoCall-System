#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Simple AI Service for Testing
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

# 创建FastAPI应用
app = FastAPI(
    title="VideoCall AI Service (Simple)",
    description="简化的AI服务用于测试",
    version="1.0.0"
)

# 添加CORS中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
async def root():
    return {
        "message": "VideoCall AI Service (Simple)",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "ai-service-simple"
    }

@app.post("/detect")
async def detect_spoofing():
    """简化的检测端点"""
    return {
        "detection_id": "test_001",
        "detection_type": "voice",
        "risk_score": 0.15,
        "confidence": 0.85,
        "status": "completed",
        "details": {
            "model_version": "v1.0.0-simple",
            "processing_time": 0.1,
            "features_analyzed": ["spectral", "temporal"]
        }
    }

if __name__ == "__main__":
    print("启动简化AI服务...")
    print("服务地址: http://localhost:5000")
    print("按 Ctrl+C 停止服务")
    
    uvicorn.run(
        "main-simple:app",
        host="0.0.0.0",
        port=5001,
        reload=False,
        log_level="info"
    ) 