import os
from typing import Optional
from pydantic import BaseSettings

class Settings(BaseSettings):
    # 应用配置
    app_name: str = "VideoCall AI Service"
    app_version: str = "1.0.0"
    environment: str = "development"
    
    # 服务器配置
    host: str = "0.0.0.0"
    port: int = 5000
    debug: bool = True
    
    # Redis配置
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: Optional[str] = None
    redis_db: int = 0
    
    # 模型配置
    model_path: str = "/app/models"
    voice_model_name: str = "voice_anti_spoofing_v1.0.0.pth"
    video_model_name: str = "video_deepfake_v1.0.0.pth"
    
    # 检测配置
    voice_threshold: float = 0.7
    video_threshold: float = 0.8
    max_processing_time: int = 30
    
    # 后端服务配置
    backend_url: str = "http://localhost:8000"
    
    class Config:
        env_file = ".env"
        case_sensitive = False

# 创建全局设置实例
settings = Settings() 