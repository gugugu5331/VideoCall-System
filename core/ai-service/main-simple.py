#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Simple AI Service for Testing
"""

import os
import uvicorn
import asyncio
import multiprocessing
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import psutil
import time
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时执行
    logger.info(f"AI Service started successfully with {multiprocessing.cpu_count()} CPU cores")
    
    yield
    
    # 关闭时执行
    logger.info("AI Service stopped")

# 创建FastAPI应用
app = FastAPI(
    title="VideoCall AI Service - Simple Version",
    description="基于深度学习的音视频伪造检测服务 - 简化版本",
    version="1.0.0",
    lifespan=lifespan
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
        "message": "VideoCall AI Service - Simple Version",
        "version": "1.0.0",
        "status": "running",
        "cpu_cores": multiprocessing.cpu_count()
    }

@app.get("/health")
async def health_check():
    # 检查系统资源
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    
    return {
        "status": "healthy",
        "service": "ai-service",
        "cpu_usage": cpu_percent,
        "memory_usage": memory.percent,
        "memory_available": memory.available
    }

@app.get("/metrics")
async def get_metrics():
    """获取系统指标"""
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    
    return {
        "timestamp": time.time(),
        "cpu": {
            "usage_percent": cpu_percent,
            "cores": multiprocessing.cpu_count()
        },
        "memory": {
            "total": memory.total,
            "available": memory.available,
            "used": memory.used,
            "percent": memory.percent
        }
    }

@app.post("/detect")
async def detect_spoofing():
    """
    模拟伪造检测
    """
    try:
        # 模拟检测过程
        await asyncio.sleep(1)
        
        return {
            "detection_id": "test_001",
            "risk_score": 0.15,
            "confidence": 0.85,
            "detection_type": "voice",
            "status": "completed",
            "timestamp": time.time()
        }
        
    except Exception as e:
        logger.error(f"Detection failed: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Detection failed: {str(e)}")

if __name__ == "__main__":
    # 配置uvicorn服务器
    uvicorn.run(
        "main-simple:app",
        host="0.0.0.0",
        port=5001,  # 使用5001端口
        reload=True,
        access_log=True,
        log_level="info"
    ) 