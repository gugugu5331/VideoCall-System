from fastapi import APIRouter, HTTPException, Depends
from typing import List
import redis.asyncio as redis

from app.core.database import get_redis
from app.models.detection import (
    DetectionRequest, 
    DetectionResponse, 
    DetectionResult,
    DetectionType,
    DetectionStatus
)
from app.services.detection_service import DetectionService

router = APIRouter()
detection_service = DetectionService()

@router.get("/health")
async def health_check():
    """健康检查"""
    return {"status": "healthy", "service": "ai-detection"}

@router.post("/detect", response_model=DetectionResponse)
async def detect_spoofing(
    request: DetectionRequest,
    redis_client: redis.Redis = Depends(get_redis)
):
    """
    执行伪造检测
    """
    try:
        # 更新检测状态为处理中
        await redis_client.hset(
            f"detection:{request.detection_id}",
            mapping={
                "status": DetectionStatus.PROCESSING,
                "detection_type": request.detection_type
            }
        )
        
        # 执行检测
        result = await detection_service.detect(
            detection_id=request.detection_id,
            detection_type=request.detection_type,
            audio_data=request.audio_data,
            video_data=request.video_data,
            metadata=request.metadata
        )
        
        # 更新Redis中的结果
        await redis_client.hset(
            f"detection:{request.detection_id}",
            mapping={
                "status": DetectionStatus.COMPLETED,
                "risk_score": result.risk_score,
                "confidence": result.confidence,
                "details": str(result.details)
            }
        )
        
        return result
        
    except Exception as e:
        # 更新状态为失败
        await redis_client.hset(
            f"detection:{request.detection_id}",
            "status",
            DetectionStatus.FAILED
        )
        raise HTTPException(status_code=500, detail=f"Detection failed: {str(e)}")

@router.get("/status/{detection_id}")
async def get_detection_status(
    detection_id: str,
    redis_client: redis.Redis = Depends(get_redis)
):
    """
    获取检测状态
    """
    try:
        status_data = await redis_client.hgetall(f"detection:{detection_id}")
        
        if not status_data:
            raise HTTPException(status_code=404, detail="Detection not found")
        
        return {
            "detection_id": detection_id,
            "status": status_data.get("status", "unknown"),
            "detection_type": status_data.get("detection_type"),
            "risk_score": float(status_data.get("risk_score", 0)),
            "confidence": float(status_data.get("confidence", 0)),
            "details": status_data.get("details", {})
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get status: {str(e)}")

@router.get("/models")
async def get_available_models():
    """
    获取可用的模型列表
    """
    return {
        "models": [
            {
                "name": "voice_anti_spoofing",
                "version": "v1.0.0",
                "type": "voice",
                "description": "语音反欺骗检测模型"
            },
            {
                "name": "video_deepfake_detection",
                "version": "v1.0.0",
                "type": "video",
                "description": "视频深度伪造检测模型"
            }
        ]
    } 