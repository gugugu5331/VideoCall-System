#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall System - Organized System Test
验证整理后的系统功能
"""
import os
import subprocess
import sys
from datetime import datetime

def print_header(title):
    """打印标题"""
    print("=" * 60)
    print(f" {title}")
    print("=" * 60)

def print_step(step, description):
    """打印步骤"""
    print(f"\n[{step}] {description}")
    print("-" * 40)

def check_file_exists(file_path, description):
    """检查文件是否存在"""
    if os.path.exists(file_path):
        print(f"✅ {description}: {file_path}")
        return True
    else:
        print(f"❌ {description}: {file_path} (NOT FOUND)")
        return False

def check_directory_exists(dir_path, description):
    """检查目录是否存在"""
    if os.path.exists(dir_path) and os.path.isdir(dir_path):
        print(f"✅ {description}: {dir_path}")
        return True
    else:
        print(f"❌ {description}: {dir_path} (NOT FOUND)")
        return False

def run_command(command, description):
    """运行命令并检查结果"""
    print(f"Running: {command}")
    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='ignore'
        )
        if result.returncode == 0:
            print(f"✅ {description}: SUCCESS")
            return True
        else:
            print(f"❌ {description}: FAILED")
            print(f"Error: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ {description}: ERROR - {e}")
        return False

def main():
    """主函数"""
    print_header("VideoCall System - Organized System Test")
    print(f"Test time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # 检查目录结构
    print_step("1", "Checking directory structure")
    
    core_dirs = [
        ("core/backend", "Backend service directory"),
        ("core/ai-service", "AI service directory"),
        ("core/database", "Database directory"),
    ]
    
    scripts_dirs = [
        ("scripts/startup", "Startup scripts directory"),
        ("scripts/management", "Management scripts directory"),
        ("scripts/testing", "Testing scripts directory"),
        ("scripts/utilities", "Utilities scripts directory"),
    ]
    
    docs_dirs = [
        ("docs/guides", "Guides documentation directory"),
        ("docs/status", "Status documentation directory"),
        ("docs/api", "API documentation directory"),
    ]
    
    config_dirs = [
        ("config", "Configuration directory"),
        ("temp", "Temporary files directory"),
    ]
    
    all_dirs = core_dirs + scripts_dirs + docs_dirs + config_dirs
    
    dir_success = 0
    for dir_path, description in all_dirs:
        if check_directory_exists(dir_path, description):
            dir_success += 1
    
    # 检查关键文件
    print_step("2", "Checking key files")
    
    key_files = [
        ("core/backend/main.go", "Backend main file"),
        ("core/backend/start-full.bat", "Backend startup script"),
        ("core/ai-service/main.py", "AI service main file"),
        ("core/ai-service/start_ai_manual.bat", "AI service startup script"),
        ("scripts/startup/start_system_simple.bat", "System startup script"),
        ("scripts/management/manage_system.bat", "System management script"),
        ("scripts/testing/run_all_tests.py", "Test runner script"),
        ("config/docker-compose.yml", "Docker Compose file"),
        ("quick_start.bat", "Quick start script"),
        ("quick_manage.bat", "Quick manage script"),
        ("quick_test.bat", "Quick test script"),
        ("README.md", "Project README"),
    ]
    
    file_success = 0
    for file_path, description in key_files:
        if check_file_exists(file_path, description):
            file_success += 1
    
    # 检查脚本功能
    print_step("3", "Testing script functionality")
    
    script_tests = [
        ("python --version", "Python environment"),
        ("docker --version", "Docker environment"),
        ("python scripts/testing/run_all_tests.py --help", "Test script help"),
    ]
    
    script_success = 0
    for command, description in script_tests:
        if run_command(command, description):
            script_success += 1
    
    # 总结
    print_step("4", "Test Summary")
    
    total_dirs = len(all_dirs)
    total_files = len(key_files)
    total_scripts = len(script_tests)
    
    print(f"Directory structure: {dir_success}/{total_dirs} ✅")
    print(f"Key files: {file_success}/{total_files} ✅")
    print(f"Script functionality: {script_success}/{total_scripts} ✅")
    
    overall_success = dir_success + file_success + script_success
    total_tests = total_dirs + total_files + total_scripts
    
    print(f"\nOverall: {overall_success}/{total_tests} tests passed")
    
    if overall_success == total_tests:
        print("\n🎉 All tests passed! System is properly organized.")
        print("\n📝 Next steps:")
        print("1. Run: .\\quick_start.bat (to start all services)")
        print("2. Run: .\\quick_test.bat (to test system)")
        print("3. Run: .\\quick_manage.bat (to manage services)")
    else:
        print(f"\n⚠️  {total_tests - overall_success} tests failed.")
        print("Please check the failed items above.")
    
    print("\n" + "=" * 60)

if __name__ == "__main__":
    main() 