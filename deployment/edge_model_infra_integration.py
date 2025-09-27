#!/usr/bin/env python3
"""
VideoCall System - Edge-Model-Infra AI检测服务集成部署
专门用于集成和部署Edge-Model-Infra AI检测服务
"""

import os
import sys
import subprocess
import json
import shutil
import logging
from pathlib import Path
from typing import Dict, List, Optional

# 设置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('edge_model_infra_deployment.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class EdgeModelInfraIntegrator:
    """Edge-Model-Infra AI检测服务集成器"""
    
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.edge_infra_path = self.project_root / "Edge-Model-Infra"
        self.ai_detection_node_path = self.edge_infra_path / "node" / "ai-detection"
        self.unit_manager_path = self.edge_infra_path / "unit-manager"
        
    def check_edge_infra_availability(self) -> bool:
        """检查Edge-Model-Infra是否可用"""
        logger.info("🔍 检查Edge-Model-Infra可用性...")
        
        required_paths = [
            self.edge_infra_path,
            self.ai_detection_node_path,
            self.unit_manager_path,
            self.ai_detection_node_path / "CMakeLists.txt",
            self.ai_detection_node_path / "Dockerfile",
            self.edge_infra_path / "docker-compose.ai-detection.yml"
        ]
        
        for path in required_paths:
            if not path.exists():
                logger.error(f"❌ 缺少必要文件: {path}")
                return False
            logger.info(f"✅ 找到: {path}")
        
        logger.info("✅ Edge-Model-Infra结构检查通过")
        return True
    
    def prepare_edge_infra_environment(self) -> bool:
        """准备Edge-Model-Infra环境"""
        logger.info("🔧 准备Edge-Model-Infra环境...")
        
        # 创建必要的目录
        dirs_to_create = [
            "/tmp/llm",
            "/tmp/detection_uploads",
            self.project_root / "storage" / "detection",
            self.ai_detection_node_path / "models",
            self.ai_detection_node_path / "build"
        ]
        
        for dir_path in dirs_to_create:
            Path(dir_path).mkdir(parents=True, exist_ok=True)
            logger.info(f"📁 创建目录: {dir_path}")
        
        # 检查并创建配置文件
        self.create_detection_config()
        self.create_unit_manager_config()
        
        return True
    
    def create_detection_config(self):
        """创建AI检测配置文件"""
        config_path = self.ai_detection_node_path / "config" / "detection_config.json"
        config_path.parent.mkdir(exist_ok=True)
        
        if not config_path.exists():
            config = {
                "detection": {
                    "face_swap": {
                        "enabled": True,
                        "model_path": "/app/models/face_swap_detection.onnx",
                        "threshold": 0.8,
                        "batch_size": 4
                    },
                    "voice_synthesis": {
                        "enabled": True,
                        "model_path": "/app/models/voice_synthesis_detection.pt",
                        "threshold": 0.7,
                        "sample_rate": 16000
                    },
                    "content_analysis": {
                        "enabled": True,
                        "max_file_size": "100MB",
                        "supported_formats": ["mp4", "avi", "mov", "wav", "mp3"]
                    }
                },
                "network": {
                    "zmq_port": 5555,
                    "http_port": 5000,
                    "max_connections": 100
                },
                "performance": {
                    "gpu_enabled": False,
                    "num_threads": 4,
                    "queue_size": 1000
                },
                "logging": {
                    "level": "INFO",
                    "file": "/app/logs/detection.log"
                }
            }
            
            with open(config_path, 'w') as f:
                json.dump(config, f, indent=2)
            
            logger.info(f"✅ 创建AI检测配置: {config_path}")
    
    def create_unit_manager_config(self):
        """创建Unit Manager配置文件"""
        config_path = self.unit_manager_path / "master_config.json"
        
        if not config_path.exists():
            config = {
                "master": {
                    "port": 10001,
                    "max_workers": 10,
                    "heartbeat_interval": 30
                },
                "nodes": [
                    {
                        "name": "ai-detection",
                        "type": "ai_detection",
                        "endpoint": "tcp://edge-ai-detection:5555",
                        "capabilities": ["face_swap_detection", "voice_synthesis_detection"],
                        "max_concurrent_tasks": 5
                    }
                ],
                "load_balancing": {
                    "strategy": "round_robin",
                    "health_check_interval": 10
                },
                "logging": {
                    "level": "INFO",
                    "file": "/app/logs/unit_manager.log"
                }
            }
            
            with open(config_path, 'w') as f:
                json.dump(config, f, indent=2)
            
            logger.info(f"✅ 创建Unit Manager配置: {config_path}")
    
    def build_edge_infra_components(self) -> bool:
        """构建Edge-Model-Infra组件"""
        logger.info("🔨 构建Edge-Model-Infra组件...")
        
        # 构建AI检测节点
        if not self.build_ai_detection_node():
            return False
        
        # 构建Unit Manager
        if not self.build_unit_manager():
            return False
        
        return True
    
    def build_ai_detection_node(self) -> bool:
        """构建AI检测节点"""
        logger.info("🤖 构建AI检测节点...")
        
        build_dir = self.ai_detection_node_path / "build"
        
        try:
            # 清理构建目录
            if build_dir.exists():
                shutil.rmtree(build_dir)
            build_dir.mkdir()
            
            # CMake配置
            cmake_cmd = [
                "cmake", "..", 
                "-DCMAKE_BUILD_TYPE=Release",
                "-DCMAKE_CXX_STANDARD=17"
            ]
            
            subprocess.run(cmake_cmd, cwd=build_dir, check=True)
            
            # 构建
            make_cmd = ["make", "-j", str(os.cpu_count() or 4)]
            subprocess.run(make_cmd, cwd=build_dir, check=True)
            
            logger.info("✅ AI检测节点构建成功")
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"❌ AI检测节点构建失败: {e}")
            return False
    
    def build_unit_manager(self) -> bool:
        """构建Unit Manager"""
        logger.info("⚙️ 构建Unit Manager...")
        
        build_dir = self.unit_manager_path / "build"
        
        try:
            # 清理构建目录
            if build_dir.exists():
                shutil.rmtree(build_dir)
            build_dir.mkdir()
            
            # CMake配置
            cmake_cmd = [
                "cmake", "..", 
                "-DCMAKE_BUILD_TYPE=Release",
                "-DCMAKE_CXX_STANDARD=17"
            ]
            
            subprocess.run(cmake_cmd, cwd=build_dir, check=True)
            
            # 构建
            make_cmd = ["make", "-j", str(os.cpu_count() or 4)]
            subprocess.run(make_cmd, cwd=build_dir, check=True)
            
            logger.info("✅ Unit Manager构建成功")
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"❌ Unit Manager构建失败: {e}")
            return False
    
    def deploy_with_docker(self) -> bool:
        """使用Docker部署Edge-Model-Infra"""
        logger.info("🐳 使用Docker部署Edge-Model-Infra...")
        
        try:
            # 使用Edge-Model-Infra的docker-compose文件
            compose_file = self.edge_infra_path / "docker-compose.ai-detection.yml"
            
            # 停止现有服务
            subprocess.run([
                "docker-compose", "-f", str(compose_file), "down"
            ], cwd=self.edge_infra_path)
            
            # 构建并启动服务
            subprocess.run([
                "docker-compose", "-f", str(compose_file), "up", "--build", "-d"
            ], cwd=self.edge_infra_path, check=True)
            
            logger.info("✅ Edge-Model-Infra Docker部署成功")
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"❌ Edge-Model-Infra Docker部署失败: {e}")
            return False
    
    def integrate_with_backend_services(self) -> bool:
        """与后端服务集成"""
        logger.info("🔗 与后端服务集成...")
        
        # 创建集成配置
        integration_config = {
            "edge_model_infra": {
                "unit_manager_url": "http://localhost:10001",
                "ai_detection_url": "http://localhost:5000",
                "enabled": True,
                "fallback_to_legacy": True
            },
            "detection_services": {
                "face_swap_detection": {
                    "endpoint": "/api/v1/detect/face-swap",
                    "timeout": 30,
                    "retry_count": 3
                },
                "voice_synthesis_detection": {
                    "endpoint": "/api/v1/detect/voice-synthesis",
                    "timeout": 45,
                    "retry_count": 3
                }
            }
        }
        
        # 保存集成配置
        config_path = self.project_root / "config" / "edge_model_infra_integration.json"
        config_path.parent.mkdir(exist_ok=True)
        
        with open(config_path, 'w') as f:
            json.dump(integration_config, f, indent=2)
        
        logger.info(f"✅ 集成配置已保存: {config_path}")
        return True
    
    def test_integration(self) -> bool:
        """测试集成"""
        logger.info("🧪 测试Edge-Model-Infra集成...")
        
        try:
            # 测试Unit Manager连接
            import requests
            
            unit_manager_url = "http://localhost:10001/health"
            response = requests.get(unit_manager_url, timeout=10)
            
            if response.status_code == 200:
                logger.info("✅ Unit Manager连接测试成功")
            else:
                logger.warning(f"⚠️ Unit Manager响应异常: {response.status_code}")
            
            # 测试AI检测节点
            ai_detection_url = "http://localhost:5000/health"
            response = requests.get(ai_detection_url, timeout=10)
            
            if response.status_code == 200:
                logger.info("✅ AI检测节点连接测试成功")
            else:
                logger.warning(f"⚠️ AI检测节点响应异常: {response.status_code}")
            
            return True
            
        except Exception as e:
            logger.error(f"❌ 集成测试失败: {e}")
            return False
    
    def create_startup_scripts(self):
        """创建启动脚本"""
        logger.info("📝 创建启动脚本...")
        
        # WSL启动脚本
        wsl_startup_script = self.project_root / "scripts" / "start_edge_infra_wsl.sh"
        wsl_startup_script.parent.mkdir(exist_ok=True)
        
        wsl_script_content = f"""#!/bin/bash

# VideoCall System - Edge-Model-Infra WSL启动脚本

set -e

echo "🚀 启动Edge-Model-Infra AI检测服务..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 进入Edge-Model-Infra目录
cd {self.edge_infra_path}

# 停止现有服务
echo "🛑 停止现有服务..."
docker-compose -f docker-compose.ai-detection.yml down

# 启动服务
echo "🚀 启动Edge-Model-Infra服务..."
docker-compose -f docker-compose.ai-detection.yml up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.ai-detection.yml ps

echo "✅ Edge-Model-Infra启动完成！"
echo "🌐 Unit Manager: http://localhost:10001"
echo "🤖 AI Detection: http://localhost:5000"
"""
        
        with open(wsl_startup_script, 'w') as f:
            f.write(wsl_script_content)
        
        os.chmod(wsl_startup_script, 0o755)
        
        # Windows PowerShell脚本
        ps_startup_script = self.project_root / "scripts" / "start_edge_infra_wsl.ps1"
        
        ps_script_content = f"""# VideoCall System - Edge-Model-Infra Windows启动脚本

Write-Host "🚀 启动Edge-Model-Infra AI检测服务..." -ForegroundColor Green

# 检查WSL是否可用
try {{
    wsl --list --running | Out-Null
    Write-Host "✅ WSL可用" -ForegroundColor Green
}} catch {{
    Write-Host "❌ WSL不可用，请先安装并启动WSL" -ForegroundColor Red
    exit 1
}}

# 在WSL中启动Edge-Model-Infra
Write-Host "🐧 在WSL中启动Edge-Model-Infra..." -ForegroundColor Cyan
wsl bash -c "cd {self.edge_infra_path} && ./scripts/start_edge_infra_wsl.sh"

Write-Host "✅ Edge-Model-Infra启动完成！" -ForegroundColor Green
Write-Host "🌐 Unit Manager: http://localhost:10001" -ForegroundColor Yellow
Write-Host "🤖 AI Detection: http://localhost:5000" -ForegroundColor Yellow
"""
        
        with open(ps_startup_script, 'w') as f:
            f.write(ps_script_content)
        
        logger.info("✅ 启动脚本创建完成")
    
    def run_full_integration(self) -> bool:
        """运行完整集成流程"""
        logger.info("🎯 开始Edge-Model-Infra完整集成...")
        
        steps = [
            ("检查Edge-Model-Infra可用性", self.check_edge_infra_availability),
            ("准备环境", self.prepare_edge_infra_environment),
            ("构建组件", self.build_edge_infra_components),
            ("Docker部署", self.deploy_with_docker),
            ("后端服务集成", self.integrate_with_backend_services),
            ("创建启动脚本", lambda: (self.create_startup_scripts(), True)[1]),
            ("测试集成", self.test_integration)
        ]
        
        for step_name, step_func in steps:
            logger.info(f"📋 执行步骤: {step_name}")
            try:
                if not step_func():
                    logger.error(f"❌ 步骤失败: {step_name}")
                    return False
                logger.info(f"✅ 步骤完成: {step_name}")
            except Exception as e:
                logger.error(f"❌ 步骤异常: {step_name} - {e}")
                return False
        
        logger.info("🎉 Edge-Model-Infra集成完成！")
        return True

def main():
    """主函数"""
    integrator = EdgeModelInfraIntegrator()
    
    if len(sys.argv) > 1:
        command = sys.argv[1]
        
        if command == "check":
            return integrator.check_edge_infra_availability()
        elif command == "build":
            return integrator.build_edge_infra_components()
        elif command == "deploy":
            return integrator.deploy_with_docker()
        elif command == "test":
            return integrator.test_integration()
        elif command == "full":
            return integrator.run_full_integration()
        else:
            print("用法: python edge_model_infra_integration.py [check|build|deploy|test|full]")
            return False
    else:
        # 默认运行完整集成
        return integrator.run_full_integration()

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
