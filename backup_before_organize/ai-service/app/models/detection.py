from typing import Dict, Any, Optional
from pydantic import BaseModel, Field
from enum import Enum

class DetectionType(str, Enum):
    VOICE_SPOOFING = "voice_spoofing"
    VIDEO_DEEPFAKE = "video_deepfake"
    FACE_SWAP = "face_swap"

class DetectionStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"

class DetectionRequest(BaseModel):
    detection_id: str = Field(..., description="检测ID")
    detection_type: DetectionType = Field(..., description="检测类型")
    call_id: str = Field(..., description="通话ID")
    audio_data: Optional[str] = Field(None, description="音频数据（base64编码）")
    video_data: Optional[str] = Field(None, description="视频数据（base64编码）")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")

class DetectionResponse(BaseModel):
    detection_id: str = Field(..., description="检测ID")
    detection_type: DetectionType = Field(..., description="检测类型")
    risk_score: float = Field(..., ge=0, le=1, description="风险评分")
    confidence: float = Field(..., ge=0, le=1, description="置信度")
    status: DetectionStatus = Field(..., description="检测状态")
    details: Dict[str, Any] = Field(..., description="详细信息")
    processing_time: Optional[float] = Field(None, description="处理时间（秒）")

class DetectionResult(BaseModel):
    detection_id: str
    detection_type: DetectionType
    risk_score: float
    confidence: float
    status: DetectionStatus
    details: Dict[str, Any]
    processing_time: float
    model_version: str
    created_at: str 