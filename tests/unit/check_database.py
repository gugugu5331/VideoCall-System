#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
数据库检查脚本
"""

import psycopg2
import json

# 数据库配置
DB_CONFIG = {
    'host': 'localhost',
    'port': 5432,
    'database': 'videocall',
    'user': 'admin',
    'password': 'videocall123'
}

def check_database_connection():
    """检查数据库连接"""
    print("🔍 检查数据库连接...")
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查数据库版本
        cursor.execute("SELECT version();")
        version = cursor.fetchone()
        print(f"✅ 数据库连接成功")
        print(f"   版本: {version[0]}")
        
        cursor.close()
        conn.close()
        return True
    except Exception as e:
        print(f"❌ 数据库连接失败: {e}")
        return False

def check_user_sessions():
    """检查用户会话表"""
    print("\n🔍 检查用户会话表...")
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查会话表结构
        cursor.execute("""
            SELECT column_name, data_type, is_nullable, column_default
            FROM information_schema.columns
            WHERE table_name = 'user_sessions'
            ORDER BY ordinal_position;
        """)
        
        columns = cursor.fetchall()
        print("   表结构:")
        for col in columns:
            print(f"     {col[0]}: {col[1]} (nullable: {col[2]}, default: {col[3]})")
        
        # 检查会话数量
        cursor.execute("SELECT COUNT(*) FROM user_sessions;")
        count = cursor.fetchone()[0]
        print(f"   当前会话数量: {count}")
        
        # 检查活跃会话
        cursor.execute("SELECT COUNT(*) FROM user_sessions WHERE is_active = true;")
        active_count = cursor.fetchone()[0]
        print(f"   活跃会话数量: {active_count}")
        
        # 检查是否有重复的token
        cursor.execute("""
            SELECT session_token, COUNT(*) as count
            FROM user_sessions
            GROUP BY session_token
            HAVING COUNT(*) > 1;
        """)
        
        duplicates = cursor.fetchall()
        if duplicates:
            print(f"   ⚠️  发现重复的session_token: {len(duplicates)}个")
            for dup in duplicates:
                print(f"     Token: {dup[0][:20]}... (出现{dup[1]}次)")
        else:
            print("   ✅ 没有重复的session_token")
        
        cursor.close()
        conn.close()
        return True
    except Exception as e:
        print(f"❌ 检查会话表失败: {e}")
        return False

def check_users():
    """检查用户表"""
    print("\n🔍 检查用户表...")
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查用户数量
        cursor.execute("SELECT COUNT(*) FROM users;")
        count = cursor.fetchone()[0]
        print(f"   用户总数: {count}")
        
        # 检查用户状态
        cursor.execute("""
            SELECT status, COUNT(*) as count
            FROM users
            GROUP BY status;
        """)
        
        statuses = cursor.fetchall()
        print("   用户状态分布:")
        for status in statuses:
            print(f"     {status[0]}: {status[1]}个")
        
        # 检查特定用户
        cursor.execute("""
            SELECT id, username, email, status, created_at
            FROM users
            WHERE username IN ('alice', 'bob', 'charlie', 'newuser_test')
            ORDER BY username;
        """)
        
        users = cursor.fetchall()
        print("   测试用户:")
        for user in users:
            print(f"     ID:{user[0]} {user[1]} ({user[2]}) - {user[3]} - {user[4]}")
        
        cursor.close()
        conn.close()
        return True
    except Exception as e:
        print(f"❌ 检查用户表失败: {e}")
        return False

def check_constraints():
    """检查约束"""
    print("\n🔍 检查数据库约束...")
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        # 检查唯一约束
        cursor.execute("""
            SELECT conname, contype, pg_get_constraintdef(oid) as definition
            FROM pg_constraint
            WHERE conrelid = 'user_sessions'::regclass;
        """)
        
        constraints = cursor.fetchall()
        print("   用户会话表约束:")
        for constraint in constraints:
            print(f"     {constraint[0]}: {constraint[2]}")
        
        cursor.close()
        conn.close()
        return True
    except Exception as e:
        print(f"❌ 检查约束失败: {e}")
        return False

def main():
    print("🚀 开始数据库诊断")
    print("=" * 50)
    
    # 检查数据库连接
    if not check_database_connection():
        return
    
    # 检查用户表
    check_users()
    
    # 检查会话表
    check_user_sessions()
    
    # 检查约束
    check_constraints()
    
    print("\n" + "=" * 50)
    print("数据库诊断完成")

if __name__ == "__main__":
    main() 