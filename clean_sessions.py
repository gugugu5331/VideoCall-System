#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
清理过期会话脚本
"""

import psycopg2
from datetime import datetime, timedelta

# 数据库配置
DB_CONFIG = {
    'host': 'localhost',
    'port': 5432,
    'database': 'videocall',
    'user': 'admin',
    'password': 'videocall123'
}

def clean_expired_sessions():
    """清理过期的会话"""
    print("🔍 清理过期会话")
    print("=" * 50)
    
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查过期会话数量
        cursor.execute("""
            SELECT COUNT(*) FROM user_sessions 
            WHERE expires_at < NOW();
        """)
        expired_count = cursor.fetchone()[0]
        print(f"过期会话数量: {expired_count}")
        
        # 检查总会话数量
        cursor.execute("SELECT COUNT(*) FROM user_sessions;")
        total_count = cursor.fetchone()[0]
        print(f"总会话数量: {total_count}")
        
        if expired_count > 0:
            # 删除过期会话
            cursor.execute("""
                DELETE FROM user_sessions 
                WHERE expires_at < NOW();
            """)
            deleted_count = cursor.rowcount
            print(f"已删除过期会话: {deleted_count}个")
        else:
            print("没有过期会话需要清理")
        
        # 清理一些旧的活跃会话（保留最近的5个）
        cursor.execute("""
            DELETE FROM user_sessions 
            WHERE id NOT IN (
                SELECT id FROM user_sessions 
                ORDER BY created_at DESC 
                LIMIT 5
            );
        """)
        cleaned_count = cursor.rowcount
        print(f"清理旧会话: {cleaned_count}个")
        
        # 提交更改
        conn.commit()
        
        # 检查清理后的状态
        cursor.execute("SELECT COUNT(*) FROM user_sessions;")
        remaining_count = cursor.fetchone()[0]
        print(f"清理后剩余会话: {remaining_count}个")
        
        cursor.close()
        conn.close()
        
        return True
    except Exception as e:
        print(f"❌ 清理会话失败: {e}")
        return False

def check_session_health():
    """检查会话健康状态"""
    print("\n🔍 检查会话健康状态")
    print("=" * 50)
    
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查会话分布
        cursor.execute("""
            SELECT 
                COUNT(*) as total,
                COUNT(CASE WHEN expires_at < NOW() THEN 1 END) as expired,
                COUNT(CASE WHEN is_active = true THEN 1 END) as active
            FROM user_sessions;
        """)
        
        stats = cursor.fetchone()
        print(f"会话统计:")
        print(f"  总数: {stats[0]}")
        print(f"  过期: {stats[1]}")
        print(f"  活跃: {stats[2]}")
        
        # 检查最近的会话
        cursor.execute("""
            SELECT user_id, created_at, expires_at, is_active
            FROM user_sessions
            ORDER BY created_at DESC
            LIMIT 5;
        """)
        
        recent_sessions = cursor.fetchall()
        print(f"\n最近5个会话:")
        for session in recent_sessions:
            print(f"  用户ID:{session[0]} 创建:{session[1]} 过期:{session[2]} 活跃:{session[3]}")
        
        cursor.close()
        conn.close()
        
        return True
    except Exception as e:
        print(f"❌ 检查会话状态失败: {e}")
        return False

def main():
    print("🚀 开始会话清理")
    print("=" * 50)
    
    # 检查会话健康状态
    check_session_health()
    
    # 清理过期会话
    if clean_expired_sessions():
        print("\n✅ 会话清理完成")
        
        # 再次检查状态
        check_session_health()
    else:
        print("\n❌ 会话清理失败")
    
    print("\n" + "=" * 50)
    print("会话清理完成")

if __name__ == "__main__":
    main() 