#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Database Status Check Script
"""

import psycopg2
import redis
import sys
from datetime import datetime

def check_postgresql():
    """Check PostgreSQL connection and basic operations"""
    print("1. Checking PostgreSQL Database...")
    try:
        # Connect to PostgreSQL
        conn = psycopg2.connect(
            host="localhost",
            port="5432",
            database="videocall",
            user="admin",
            password="videocall123"
        )
        
        # Test connection
        cursor = conn.cursor()
        cursor.execute("SELECT version();")
        version = cursor.fetchone()
        print(f"   ✅ PostgreSQL Connected")
        print(f"   Version: {version[0]}")
        
        # Check if tables exist
        cursor.execute("""
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
            ORDER BY table_name;
        """)
        tables = cursor.fetchall()
        print(f"   Tables found: {len(tables)}")
        for table in tables:
            print(f"     - {table[0]}")
        
        # Check user count
        cursor.execute("SELECT COUNT(*) FROM users;")
        user_count = cursor.fetchone()[0]
        print(f"   Users in database: {user_count}")
        
        # Check recent activity
        cursor.execute("""
            SELECT username, created_at, last_login 
            FROM users 
            ORDER BY created_at DESC 
            LIMIT 3;
        """)
        recent_users = cursor.fetchall()
        print(f"   Recent users:")
        for user in recent_users:
            print(f"     - {user[0]} (created: {user[1]}, last login: {user[2]})")
        
        cursor.close()
        conn.close()
        return True
        
    except Exception as e:
        print(f"   ❌ PostgreSQL Error: {e}")
        return False

def check_redis():
    """Check Redis connection and basic operations"""
    print("\n2. Checking Redis Database...")
    try:
        # Connect to Redis
        r = redis.Redis(
            host="localhost",
            port="6379",
            db=0,
            decode_responses=True
        )
        
        # Test connection
        r.ping()
        print(f"   ✅ Redis Connected")
        
        # Check Redis info
        info = r.info()
        print(f"   Redis Version: {info.get('redis_version', 'Unknown')}")
        print(f"   Connected Clients: {info.get('connected_clients', 0)}")
        print(f"   Used Memory: {info.get('used_memory_human', 'Unknown')}")
        
        # Test basic operations
        r.set("test_key", "test_value")
        value = r.get("test_key")
        r.delete("test_key")
        
        if value == "test_value":
            print(f"   ✅ Redis read/write operations working")
        else:
            print(f"   ❌ Redis read/write operations failed")
            return False
        
        return True
        
    except Exception as e:
        print(f"   ❌ Redis Error: {e}")
        return False

def check_docker_containers():
    """Check Docker container status"""
    print("\n3. Checking Docker Containers...")
    try:
        import subprocess
        result = subprocess.run(
            ["docker-compose", "--project-name", "videocall-system", "ps"],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        
        if result.returncode == 0 and "Up" in result.stdout:
            print(f"   ✅ Docker containers are running")
            lines = result.stdout.strip().split('\n')
            for line in lines[1:]:  # Skip header
                if line.strip():
                    print(f"     {line.strip()}")
            return True
        else:
            print(f"   ❌ Docker containers not running")
            if result.stderr:
                print(f"   Error: {result.stderr}")
            return False
        
    except Exception as e:
        print(f"   ❌ Docker check error: {e}")
        return False

def main():
    """Main function"""
    print("=" * 60)
    print("Database Status Check")
    print("=" * 60)
    print(f"Check time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # Check all components
    postgres_ok = check_postgresql()
    redis_ok = check_redis()
    docker_ok = check_docker_containers()
    
    # Summary
    print("\n" + "=" * 60)
    print("Database Status Summary:")
    print("=" * 60)
    print(f"PostgreSQL: {'✅' if postgres_ok else '❌'}")
    print(f"Redis: {'✅' if redis_ok else '❌'}")
    print(f"Docker Containers: {'✅' if docker_ok else '❌'}")
    
    if all([postgres_ok, redis_ok, docker_ok]):
        print("\n🎉 All database services are operational!")
        return 0
    else:
        print("\n⚠️  Some database services have issues.")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 