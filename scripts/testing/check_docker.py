#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Docker Container Status Check
"""

import subprocess
import sys
from datetime import datetime

def check_docker_containers():
    """Check Docker container status using simple commands"""
    print("=" * 60)
    print("Docker Container Status Check")
    print("=" * 60)
    print(f"Check time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    try:
        # Check if docker is running
        print("\n1. Checking Docker service...")
        result = subprocess.run(
            ["docker", "version"],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        
        if result.returncode == 0:
            print("   ✅ Docker service is running")
        else:
            print("   ❌ Docker service is not running")
            print(f"   Error: {result.stderr}")
            return False
        
        # Check container status using docker ps
        print("\n2. Checking running containers...")
        result = subprocess.run(
            ["docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}"],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        
        if result.returncode == 0:
            output = result.stdout.strip()
            if "videocall" in output:
                print("   ✅ VideoCall containers found:")
                lines = output.split('\n')
                for line in lines:
                    if "videocall" in line:
                        print(f"     {line}")
                return True
            else:
                print("   ❌ No VideoCall containers found")
                print("   Available containers:")
                print(output)
                return False
        else:
            print("   ❌ Failed to check containers")
            print(f"   Error: {result.stderr}")
            return False
            
    except Exception as e:
        print(f"   ❌ Docker check error: {e}")
        return False

def check_docker_compose():
    """Check Docker Compose project status"""
    print("\n3. Checking Docker Compose project...")
    try:
        result = subprocess.run(
            ["docker-compose", "--project-name", "videocall-system", "ps"],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        
        if result.returncode == 0:
            output = result.stdout.strip()
            if "Up" in output:
                print("   ✅ Docker Compose project is running")
                lines = output.split('\n')
                for line in lines[1:]:  # Skip header
                    if line.strip():
                        print(f"     {line.strip()}")
                return True
            else:
                print("   ❌ Docker Compose project not running")
                print("   Output:", output)
                return False
        else:
            print("   ❌ Failed to check Docker Compose")
            print(f"   Error: {result.stderr}")
            return False
            
    except Exception as e:
        print(f"   ❌ Docker Compose check error: {e}")
        return False

def check_ports():
    """Check if required ports are open"""
    print("\n4. Checking required ports...")
    import socket
    
    ports_to_check = [
        ("PostgreSQL", "localhost", 5432),
        ("Redis", "localhost", 6379)
    ]
    
    all_ports_ok = True
    for service, host, port in ports_to_check:
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(2)
            result = sock.connect_ex((host, port))
            sock.close()
            
            if result == 0:
                print(f"   ✅ {service} port {port} is open")
            else:
                print(f"   ❌ {service} port {port} is closed")
                all_ports_ok = False
        except Exception as e:
            print(f"   ❌ {service} port {port} check failed: {e}")
            all_ports_ok = False
    
    return all_ports_ok

def main():
    """Main function"""
    docker_ok = check_docker_containers()
    compose_ok = check_docker_compose()
    ports_ok = check_ports()
    
    # Summary
    print("\n" + "=" * 60)
    print("Docker Status Summary:")
    print("=" * 60)
    print(f"Docker Service: {'✅' if docker_ok else '❌'}")
    print(f"Docker Compose: {'✅' if compose_ok else '❌'}")
    print(f"Required Ports: {'✅' if ports_ok else '❌'}")
    
    if all([docker_ok, compose_ok, ports_ok]):
        print("\n🎉 All Docker services are operational!")
        return 0
    else:
        print("\n⚠️  Some Docker services have issues.")
        return 1

if __name__ == "__main__":
    sys.exit(main()) 