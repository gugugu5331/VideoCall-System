import redis.asyncio as redis
from app.core.config import settings

async def init_redis() -> redis.Redis:
    """
    初始化Redis连接
    """
    redis_client = redis.Redis(
        host=settings.redis_host,
        port=settings.redis_port,
        password=settings.redis_password,
        db=settings.redis_db,
        decode_responses=True,
        socket_connect_timeout=5,
        socket_timeout=5,
        retry_on_timeout=True
    )
    
    # 测试连接
    try:
        await redis_client.ping()
        print("Redis connection established")
    except Exception as e:
        print(f"Failed to connect to Redis: {e}")
        raise
    
    return redis_client

async def get_redis() -> redis.Redis:
    """
    获取Redis客户端实例
    """
    return redis.Redis(
        host=settings.redis_host,
        port=settings.redis_port,
        password=settings.redis_password,
        db=settings.redis_db,
        decode_responses=True
    ) 