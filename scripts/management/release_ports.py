#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall System - Port Release Utility
智能端口释放工具
"""
import subprocess
import sys
import time
import psutil
from datetime import datetime

def run_command(command, capture_output=True):
    """运行命令并返回结果"""
    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=capture_output,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        return result
    except Exception as e:
        print(f"Error running command '{command}': {e}")
        return None

def get_process_info(pid):
    """获取进程信息"""
    try:
        process = psutil.Process(pid)
        return {
            'name': process.name(),
            'cmdline': ' '.join(process.cmdline()),
            'status': process.status(),
            'create_time': datetime.fromtimestamp(process.create_time()).strftime('%Y-%m-%d %H:%M:%S')
        }
    except (psutil.NoSuchProcess, psutil.AccessDenied):
        return None

def check_port_status(port):
    """检查端口状态"""
    result = run_command(f'netstat -ano | findstr :{port}')
    if result and result.stdout.strip():
        lines = result.stdout.strip().split('\n')
        processes = []
        for line in lines:
            if 'LISTENING' in line:
                parts = line.split()
                if len(parts) >= 5:
                    pid = parts[-1]
                    process_info = get_process_info(int(pid))
                    if process_info:
                        processes.append({
                            'pid': pid,
                            'info': process_info
                        })
        return processes
    return []

def kill_process(pid, force=False):
    """终止进程"""
    try:
        process = psutil.Process(int(pid))
        if force:
            process.kill()
        else:
            process.terminate()
        return True
    except (psutil.NoSuchProcess, psutil.AccessDenied) as e:
        print(f"Failed to kill process {pid}: {e}")
        return False

def release_port(port, force=False):
    """释放指定端口"""
    print(f"\nChecking port {port}...")
    processes = check_port_status(port)
    
    if not processes:
        print(f"✅ Port {port} is free")
        return True
    
    print(f"⚠️  Found {len(processes)} process(es) using port {port}:")
    
    for proc in processes:
        pid = proc['pid']
        info = proc['info']
        print(f"   PID: {pid}")
        print(f"   Name: {info['name']}")
        print(f"   Command: {info['cmdline'][:100]}...")
        print(f"   Status: {info['status']}")
        print(f"   Created: {info['create_time']}")
        
        # 询问是否终止进程
        if force:
            should_kill = True
        else:
            response = input(f"   Kill this process? (y/n): ").lower().strip()
            should_kill = response in ['y', 'yes']
        
        if should_kill:
            print(f"   Terminating process {pid}...")
            if kill_process(pid, force):
                print(f"   ✅ Process {pid} terminated")
            else:
                print(f"   ❌ Failed to terminate process {pid}")
        else:
            print(f"   Skipping process {pid}")
    
    # 等待进程完全终止
    time.sleep(2)
    
    # 再次检查端口状态
    remaining_processes = check_port_status(port)
    if not remaining_processes:
        print(f"✅ Port {port} is now free")
        return True
    else:
        print(f"❌ Port {port} is still in use by {len(remaining_processes)} process(es)")
        return False

def main():
    """主函数"""
    print("=" * 60)
    print("VideoCall System - Port Release Utility")
    print("=" * 60)
    print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    # 检查是否以管理员权限运行
    try:
        is_admin = subprocess.run(['net', 'session'], capture_output=True).returncode == 0
        if not is_admin:
            print("⚠️  Warning: Not running as administrator")
            print("   Some processes may not be killable")
            print()
    except:
        pass
    
    # 定义需要检查的端口
    ports = [
        (8000, "Backend Service"),
        (5001, "AI Service"),
        (5432, "PostgreSQL"),
        (6379, "Redis")
    ]
    
    # 检查命令行参数
    force_mode = '--force' in sys.argv or '-f' in sys.argv
    specific_ports = []
    
    for arg in sys.argv[1:]:
        if arg.isdigit():
            specific_ports.append(int(arg))
    
    if specific_ports:
        ports_to_check = [(port, f"Port {port}") for port in specific_ports]
    else:
        ports_to_check = ports
    
    if force_mode:
        print("🔧 Force mode enabled - will kill processes without asking")
        print()
    
    # 检查并释放端口
    results = []
    for port, description in ports_to_check:
        success = release_port(port, force_mode)
        results.append((port, description, success))
    
    # 总结
    print("\n" + "=" * 60)
    print("Summary:")
    print("=" * 60)
    
    all_success = True
    for port, description, success in results:
        status = "✅ FREE" if success else "❌ IN USE"
        print(f"{description} (Port {port}): {status}")
        if not success:
            all_success = False
    
    print()
    if all_success:
        print("🎉 All ports are now free!")
    else:
        print("⚠️  Some ports are still in use")
        print("   You may need to:")
        print("   1. Run this script as administrator")
        print("   2. Manually close the applications")
        print("   3. Restart your computer")
    
    print("\nPress Enter to exit...")
    input()

if __name__ == "__main__":
    main() 