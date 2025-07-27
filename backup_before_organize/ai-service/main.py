import os
import uvicorn
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import redis.asyncio as redis
from dotenv import load_dotenv

from app.core.config import settings
from app.core.database import init_redis
from app.api.routes import api_router
from app.models.detection import DetectionRequest, DetectionResponse

# 加载环境变量
load_dotenv()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时执行
    app.state.redis = await init_redis()
    print("AI Service started successfully")
    
    yield
    
    # 关闭时执行
    if hasattr(app.state, 'redis'):
        await app.state.redis.close()
    print("AI Service stopped")

# 创建FastAPI应用
app = FastAPI(
    title="VideoCall AI Service",
    description="基于深度学习的音视频伪造检测服务",
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

# 包含API路由
app.include_router(api_router, prefix="/api/v1")

@app.get("/")
async def root():
    return {
        "message": "VideoCall AI Service",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "ai-service"
    }

@app.post("/detect", response_model=DetectionResponse)
async def detect_spoofing(request: DetectionRequest):
    """
    执行伪造检测
    """
    try:
        # 这里将实现具体的检测逻辑
        # 暂时返回模拟结果
        result = {
            "detection_id": request.detection_id,
            "detection_type": request.detection_type,
            "risk_score": 0.15,  # 模拟风险评分
            "confidence": 0.85,  # 模拟置信度
            "status": "completed",
            "details": {
                "model_version": "v1.0.0",
                "processing_time": 0.5,
                "features_analyzed": ["spectral", "temporal", "prosodic"]
            }
        }
        
        return DetectionResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Detection failed: {str(e)}")

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=5000,
        reload=True if os.getenv("ENVIRONMENT") == "development" else False
    ) 