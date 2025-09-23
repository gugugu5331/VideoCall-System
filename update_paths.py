#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
VideoCall System - Path Update Script
更新脚本中的路径引用
"""
import os
import re
from pathlib import Path

def update_file_paths(file_path, old_patterns, new_patterns):
    """更新文件中的路径引用"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        for old_pattern, new_pattern in zip(old_patterns, new_patterns):
            content = content.replace(old_pattern, new_pattern)
        
        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"✅ Updated: {file_path}")
            return True
        else:
            print(f"ℹ️  No changes needed: {file_path}")
            return False
            
    except Exception as e:
        print(f"❌ Error updating {file_path}: {e}")
        return False

def main():
    """主函数"""
    print("=" * 60)
    print("VideoCall System - Path Update Script")
    print("=" * 60)
    
    # 定义需要更新的路径模式
    path_updates = [
        # 启动脚本路径更新
        ("scripts/startup/start_system_simple.bat", [
            "start-full.bat",
            "start_ai_manual.bat",
            "python scripts/run_all_tests.py",
            "python test_api.py"
        ], [
            "..\\..\\core\\backend\\start-full.bat",
            "..\\..\\core\\ai-service\\start_ai_manual.bat", 
            "python ..\\testing\\run_all_tests.py",
            "python ..\\testing\\test_api.py"
        ]),
        
        ("scripts/startup/start_system.bat", [
            "start-full.bat",
            "start_ai_manual.bat",
            "python scripts/run_all_tests.py"
        ], [
            "..\\..\\core\\backend\\start-full.bat",
            "..\\..\\core\\ai-service\\start_ai_manual.bat",
            "python ..\\testing\\run_all_tests.py"
        ]),
        
        # 管理脚本路径更新
        ("scripts/management/manage_system.bat", [
            "python scripts/run_all_tests.py",
            "python test_api.py",
            "python check_database.py",
            "python check_docker.py",
            "python release_ports.py"
        ], [
            "python ..\\testing\\run_all_tests.py",
            "python ..\\testing\\test_api.py",
            "python ..\\testing\\check_database.py",
            "python ..\\testing\\check_docker.py",
            "python release_ports.py"
        ]),
        
        # 测试脚本路径更新
        ("scripts/testing/run_all_tests.py", [
            "import sys\nsys.path.append('..')",
            "import sys\nsys.path.append('..\\..')"
        ], [
            "import sys\nsys.path.append('..\\..\\core')",
            "import sys\nsys.path.append('..\\..\\core')"
        ]),
        
        # 其他脚本路径更新
        ("scripts/management/release_ports.py", [
            "import sys\nsys.path.append('..')"
        ], [
            "import sys\nsys.path.append('..\\..\\core')"
        ])
    ]
    
    updated_count = 0
    
    for file_path, old_patterns, new_patterns in path_updates:
        if os.path.exists(file_path):
            if update_file_paths(file_path, old_patterns, new_patterns):
                updated_count += 1
        else:
            print(f"⚠️  File not found: {file_path}")
    
    print(f"\n📊 Summary:")
    print(f"✅ Updated {updated_count} files")
    print(f"✅ Path references updated successfully")
    
    # 创建快速启动脚本
    create_quick_start_scripts()
    
    print("\n🎉 Path update completed!")

def create_quick_start_scripts():
    """创建快速启动脚本"""
    print("\n" + "=" * 40)
    print("Creating quick start scripts")
    print("=" * 40)
    
    # 创建根目录的快速启动脚本
    quick_start_content = """@echo off
chcp 65001 >nul
title VideoCall System - Quick Start

echo ==========================================
echo VideoCall System - Quick Start
echo ==========================================
echo.

echo Starting system...
cd /d "%~dp0"
call scripts\\startup\\start_system_simple.bat

echo.
echo Quick start completed!
pause
"""
    
    with open("quick_start.bat", 'w', encoding='utf-8') as f:
        f.write(quick_start_content)
    print("✅ Created: quick_start.bat")
    
    # 创建快速管理脚本
    quick_manage_content = """@echo off
chcp 65001 >nul
title VideoCall System - Quick Manage

echo ==========================================
echo VideoCall System - Quick Manage
echo ==========================================
echo.

echo Opening management menu...
cd /d "%~dp0"
call scripts\\management\\manage_system.bat

echo.
echo Management completed!
pause
"""
    
    with open("quick_manage.bat", 'w', encoding='utf-8') as f:
        f.write(quick_manage_content)
    print("✅ Created: quick_manage.bat")
    
    # 创建快速测试脚本
    quick_test_content = """@echo off
chcp 65001 >nul
title VideoCall System - Quick Test

echo ==========================================
echo VideoCall System - Quick Test
echo ==========================================
echo.

echo Running system tests...
cd /d "%~dp0"
python scripts\\testing\\run_all_tests.py

echo.
echo Test completed!
pause
"""
    
    with open("quick_test.bat", 'w', encoding='utf-8') as f:
        f.write(quick_test_content)
    print("✅ Created: quick_test.bat")

if __name__ == "__main__":
    main() 