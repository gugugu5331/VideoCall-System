import os
import uvicorn
import asyncio
import multiprocessing
from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import redis.asyncio as redis
from dotenv import load_dotenv
import psutil
import time
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor
import logging

from app.core.config import settings
from app.core.database import init_redis
from app.api.routes import api_router
from app.models.detection import DetectionRequest, DetectionResponse
from app.services.detection_service import DetectionService

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 加载环境变量
load_dotenv()

# 全局变量
detection_service = None
thread_pool = None
process_pool = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时执行
    global detection_service, thread_pool, process_pool
    
    # 初始化Redis连接
    app.state.redis = await init_redis()
    
    # 初始化检测服务
    detection_service = DetectionService()
    
    # 创建线程池和进程池
    thread_pool = ThreadPoolExecutor(max_workers=multiprocessing.cpu_count() * 2)
    process_pool = ProcessPoolExecutor(max_workers=multiprocessing.cpu_count())
    
    # 启动监控协程
    asyncio.create_task(monitor_system_resources())
    
    logger.info(f"AI Service started successfully with {multiprocessing.cpu_count()} CPU cores")
    
    yield
    
    # 关闭时执行
    if hasattr(app.state, 'redis'):
        await app.state.redis.close()
    
    if thread_pool:
        thread_pool.shutdown(wait=True)
    
    if process_pool:
        process_pool.shutdown(wait=True)
    
    logger.info("AI Service stopped")

# 创建FastAPI应用
app = FastAPI(
    title="VideoCall AI Service",
    description="基于深度学习的音视频伪造检测服务 - 高并发版本",
    version="2.0.0",
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
        "message": "VideoCall AI Service - High Concurrency Version",
        "version": "2.0.0",
        "status": "running",
        "cpu_cores": multiprocessing.cpu_count(),
        "concurrent_workers": multiprocessing.cpu_count() * 2
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
        "memory_available": memory.available,
        "active_threads": thread_pool._max_workers if thread_pool else 0,
        "active_processes": process_pool._max_workers if process_pool else 0
    }

@app.get("/metrics")
async def get_metrics():
    """获取系统指标"""
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    disk = psutil.disk_usage('/')
    
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
        },
        "disk": {
            "total": disk.total,
            "used": disk.used,
            "free": disk.free,
            "percent": (disk.used / disk.total) * 100
        },
        "pools": {
            "thread_pool_size": thread_pool._max_workers if thread_pool else 0,
            "process_pool_size": process_pool._max_workers if process_pool else 0
        }
    }

@app.post("/detect", response_model=DetectionResponse)
async def detect_spoofing(request: DetectionRequest, background_tasks: BackgroundTasks):
    """
    执行伪造检测 - 异步处理版本
    """
    try:
        # 使用线程池处理检测任务
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(
            thread_pool,
            detection_service.detect_sync,
            request.detection_id,
            request.detection_type,
            request.audio_data,
            request.video_data,
            request.metadata
        )
        
        # 添加后台任务用于日志记录
        background_tasks.add_task(log_detection_result, request.detection_id, result)
        
        return result
        
    except Exception as e:
        logger.error(f"Detection failed for {request.detection_id}: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Detection failed: {str(e)}")

@app.post("/detect/batch")
async def detect_spoofing_batch(requests: list[DetectionRequest]):
    """
    批量检测 - 并发处理多个检测请求
    """
    try:
        # 创建检测任务
        tasks = []
        for request in requests:
            task = asyncio.create_task(
                detect_single_request(request)
            )
            tasks.append(task)
        
        # 并发执行所有任务
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # 处理结果
        successful_results = []
        failed_results = []
        
        for i, result in enumerate(results):
            if isinstance(result, Exception):
                failed_results.append({
                    "detection_id": requests[i].detection_id,
                    "error": str(result)
                })
            else:
                successful_results.append(result)
        
        return {
            "total_requests": len(requests),
            "successful": len(successful_results),
            "failed": len(failed_results),
            "results": successful_results,
            "errors": failed_results
        }
        
    except Exception as e:
        logger.error(f"Batch detection failed: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Batch detection failed: {str(e)}")

async def detect_single_request(request: DetectionRequest):
    """处理单个检测请求"""
    try:
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(
            thread_pool,
            detection_service.detect_sync,
            request.detection_id,
            request.detection_type,
            request.audio_data,
            request.video_data,
            request.metadata
        )
        return result
    except Exception as e:
        raise e

async def log_detection_result(detection_id: str, result: DetectionResponse):
    """后台记录检测结果"""
    try:
        logger.info(f"Detection completed: {detection_id}, Risk: {result.risk_score}, Confidence: {result.confidence}")
    except Exception as e:
        logger.error(f"Failed to log detection result: {str(e)}")

async def monitor_system_resources():
    """监控系统资源"""
    while True:
        try:
            cpu_percent = psutil.cpu_percent(interval=5)
            memory = psutil.virtual_memory()
            
            # 如果资源使用率过高，记录警告
            if cpu_percent > 80:
                logger.warning(f"High CPU usage: {cpu_percent}%")
            
            if memory.percent > 80:
                logger.warning(f"High memory usage: {memory.percent}%")
            
            # 每30秒记录一次资源使用情况
            await asyncio.sleep(30)
            
        except Exception as e:
            logger.error(f"Resource monitoring error: {str(e)}")
            await asyncio.sleep(60)

if __name__ == "__main__":
    # 配置uvicorn服务器
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=5001,  # 使用5001端口
        reload=True if os.getenv("ENVIRONMENT") == "development" else False,
        workers=1,  # 使用单进程，通过asyncio处理并发
        loop="asyncio",
        access_log=True,
        log_level="info"
    ) 