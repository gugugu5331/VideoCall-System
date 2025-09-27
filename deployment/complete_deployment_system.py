#!/usr/bin/env python3
"""
VideoCall System - 完整部署系统
确保与现有组件完全兼容的集成部署方案
"""

import os
import sys
import subprocess
import platform
import json
import shutil
import time
import logging
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import yaml
import requests
import psutil

# 设置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('deployment.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class VideoCallSystemDeployer:
    """VideoCall System完整部署器"""
    
    def __init__(self):
        self.platform = platform.system().lower()
        self.project_root = Path(__file__).parent.parent
        self.deployment_config = self.load_deployment_config()
        self.service_status = {}
        
    def load_deployment_config(self) -> Dict:
        """加载部署配置"""
        config_file = self.project_root / "deployment" / "deployment_config.yaml"
        
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return yaml.safe_load(f)
        
        # 默认配置
        return {
            "project": {
                "name": "VideoCall System",
                "version": "1.0.0"
            },
            "backend": {
                "services": [
                    {"name": "user", "port": 8081, "path": "src/backend/services/user"},
                    {"name": "meeting", "port": 8082, "path": "src/backend/services/meeting"},
                    {"name": "signaling", "port": 8083, "path": "src/backend/services/signaling"},
                    {"name": "media", "port": 8084, "path": "src/backend/services/media"},
                    {"name": "detection", "port": 8085, "path": "src/backend/services/detection"},
                    {"name": "notification", "port": 8086, "path": "src/backend/services/notification"},
                    {"name": "record", "port": 8087, "path": "src/backend/services/record"},
                    {"name": "smart-editing", "port": 8088, "path": "src/backend/services/smart-editing"},
                    {"name": "gateway", "port": 8080, "path": "src/backend/services/gateway"}
                ]
            },
            "frontend": {
                "qt_client": {
                    "path": "src/frontend/qt-client-new",
                    "build_dir": "build-qt",
                    "executable": "VideoCallSystemClient"
                },
                "web_interface": {
                    "path": "src/frontend/web_interface",
                    "port": 3000
                },
                "video_effects": {
                    "path": "src/video-processing",
                    "build_dir": "build",
                    "executable": "VideoEffectsDemo"
                }
            },
            "ai_detection": {
                "path": "src/ai-detection",
                "port": 8085,
                "models_dir": "models"
            },
            "databases": {
                "postgresql": {"port": 5432, "database": "videocall_system"},
                "redis": {"port": 6379},
                "mongodb": {"port": 27017, "database": "videocall_files"}
            }
        }
    
    def check_system_requirements(self) -> bool:
        """检查系统要求"""
        logger.info("🔍 检查系统要求...")
        
        # 检查内存
        memory_gb = psutil.virtual_memory().total / (1024**3)
        if memory_gb < 8:
            logger.warning(f"内存不足: {memory_gb:.1f}GB (推荐8GB+)")
        
        # 检查磁盘空间
        disk_usage = psutil.disk_usage(str(self.project_root))
        free_gb = disk_usage.free / (1024**3)
        if free_gb < 10:
            logger.error(f"磁盘空间不足: {free_gb:.1f}GB (需要10GB+)")
            return False
        
        # 检查网络连接
        try:
            requests.get("https://www.baidu.com", timeout=5)
            logger.info("✅ 网络连接正常")
        except:
            logger.warning("⚠️ 网络连接可能有问题")
        
        logger.info("✅ 系统要求检查完成")
        return True
    
    def check_dependencies(self) -> Dict[str, bool]:
        """检查依赖项"""
        logger.info("🔍 检查依赖项...")
        
        dependencies = {}
        
        # 检查基础工具
        basic_tools = ["git", "cmake", "python3", "node", "npm"]
        for tool in basic_tools:
            try:
                result = subprocess.run([tool, "--version"], 
                                      capture_output=True, text=True, timeout=10)
                dependencies[tool] = result.returncode == 0
                if dependencies[tool]:
                    logger.info(f"✅ {tool} 已安装")
                else:
                    logger.error(f"❌ {tool} 未安装或版本不兼容")
            except:
                dependencies[tool] = False
                logger.error(f"❌ {tool} 未找到")
        
        # 检查Go环境
        try:
            result = subprocess.run(["go", "version"], 
                                  capture_output=True, text=True, timeout=10)
            dependencies["go"] = result.returncode == 0 and "go1.21" in result.stdout
            if dependencies["go"]:
                logger.info("✅ Go 1.21+ 已安装")
            else:
                logger.error("❌ Go 1.21+ 未安装")
        except:
            dependencies["go"] = False
            logger.error("❌ Go 未找到")
        
        # Windows特定检查
        if self.platform == "windows":
            # 检查Qt6
            qt_paths = [
                "C:/Qt/6.5.0/msvc2019_64/bin/qmake.exe",
                "C:/Qt/6.6.0/msvc2019_64/bin/qmake.exe"
            ]
            dependencies["qt6"] = any(Path(p).exists() for p in qt_paths)
            
            # 检查OpenCV
            opencv_paths = [
                "C:/vcpkg/installed/x64-windows/include/opencv2",
                "C:/opencv/build/include/opencv2"
            ]
            dependencies["opencv"] = any(Path(p).exists() for p in opencv_paths)
            
            # 检查Visual Studio
            vs_paths = [
                "C:/Program Files (x86)/Microsoft Visual Studio/2019",
                "C:/Program Files (x86)/Microsoft Visual Studio/2022"
            ]
            dependencies["visual_studio"] = any(Path(p).exists() for p in vs_paths)
        
        # Linux特定检查
        elif self.platform == "linux":
            # 检查数据库
            db_services = ["postgresql", "redis-server", "mongod"]
            for service in db_services:
                try:
                    result = subprocess.run(["systemctl", "is-active", service], 
                                          capture_output=True, text=True)
                    dependencies[service] = result.stdout.strip() == "active"
                except:
                    dependencies[service] = False
        
        return dependencies
    
    def setup_environment(self) -> bool:
        """设置环境"""
        logger.info("🔧 设置环境...")
        
        # 创建必要目录
        dirs_to_create = [
            "logs", "storage/uploads", "storage/media", "storage/detection",
            "build-qt", "build-linux", "deployment/temp"
        ]
        
        for dir_path in dirs_to_create:
            full_path = self.project_root / dir_path
            full_path.mkdir(parents=True, exist_ok=True)
            logger.info(f"📁 创建目录: {dir_path}")
        
        # 设置环境变量
        if self.platform == "windows":
            self.setup_windows_environment()
        elif self.platform == "linux":
            self.setup_linux_environment()
        
        return True
    
    def setup_windows_environment(self):
        """设置Windows环境"""
        logger.info("🖥️ 设置Windows环境...")
        
        # 查找Qt6路径
        qt_paths = [
            "C:/Qt/6.5.0/msvc2019_64",
            "C:/Qt/6.6.0/msvc2019_64"
        ]
        
        for qt_path in qt_paths:
            if Path(qt_path).exists():
                os.environ["Qt6_DIR"] = qt_path
                logger.info(f"✅ 设置Qt6_DIR: {qt_path}")
                break
        
        # 查找OpenCV路径
        opencv_paths = [
            "C:/vcpkg/installed/x64-windows",
            "C:/opencv/build"
        ]
        
        for opencv_path in opencv_paths:
            if Path(opencv_path).exists():
                os.environ["OpenCV_DIR"] = opencv_path
                logger.info(f"✅ 设置OpenCV_DIR: {opencv_path}")
                break
        
        # 设置CMake工具链
        vcpkg_toolchain = "C:/vcpkg/scripts/buildsystems/vcpkg.cmake"
        if Path(vcpkg_toolchain).exists():
            os.environ["CMAKE_TOOLCHAIN_FILE"] = vcpkg_toolchain
            logger.info(f"✅ 设置CMAKE_TOOLCHAIN_FILE: {vcpkg_toolchain}")
    
    def setup_linux_environment(self):
        """设置Linux环境"""
        logger.info("🐧 设置Linux环境...")
        
        # 设置Go环境
        go_path = "/usr/local/go/bin"
        if Path(go_path).exists():
            current_path = os.environ.get("PATH", "")
            if go_path not in current_path:
                os.environ["PATH"] = f"{go_path}:{current_path}"
                logger.info(f"✅ 添加Go到PATH: {go_path}")
        
        # 设置GOPATH
        go_workspace = Path.home() / "go"
        go_workspace.mkdir(exist_ok=True)
        os.environ["GOPATH"] = str(go_workspace)
        os.environ["GOBIN"] = str(go_workspace / "bin")
        
        logger.info(f"✅ 设置GOPATH: {go_workspace}")
    
    def build_backend_services(self) -> bool:
        """构建后端服务"""
        logger.info("🔨 构建后端服务...")
        
        if self.platform != "linux":
            logger.warning("后端服务需要在Linux环境中构建")
            return True
        
        build_dir = self.project_root / "build-linux"
        build_dir.mkdir(exist_ok=True)
        
        success_count = 0
        total_services = len(self.deployment_config["backend"]["services"])
        
        for service in self.deployment_config["backend"]["services"]:
            service_name = service["name"]
            service_path = self.project_root / service["path"]
            
            if not service_path.exists():
                logger.warning(f"⚠️ 服务路径不存在: {service_path}")
                continue
            
            logger.info(f"🔨 构建服务: {service_name}")
            
            try:
                # 检查go.mod文件
                go_mod_path = service_path / "go.mod"
                if not go_mod_path.exists():
                    # 创建go.mod
                    subprocess.run(["go", "mod", "init", f"videocall/{service_name}"], 
                                 cwd=service_path, check=True)
                
                # 下载依赖
                subprocess.run(["go", "mod", "tidy"], cwd=service_path, check=True)
                
                # 构建服务
                output_path = build_dir / f"{service_name}-service"
                build_cmd = [
                    "go", "build", 
                    "-ldflags", "-s -w",
                    "-o", str(output_path),
                    "."
                ]
                
                subprocess.run(build_cmd, cwd=service_path, check=True)
                
                logger.info(f"✅ {service_name} 构建成功")
                success_count += 1
                
            except subprocess.CalledProcessError as e:
                logger.error(f"❌ {service_name} 构建失败: {e}")
        
        logger.info(f"🎯 后端服务构建完成: {success_count}/{total_services}")
        return success_count > 0
    
    def build_frontend_applications(self) -> bool:
        """构建前端应用"""
        logger.info("🎨 构建前端应用...")
        
        if self.platform != "windows":
            logger.warning("前端应用需要在Windows环境中构建")
            return True
        
        success = True
        
        # 构建Qt客户端
        if self.build_qt_client():
            logger.info("✅ Qt客户端构建成功")
        else:
            logger.error("❌ Qt客户端构建失败")
            success = False
        
        # 构建视频处理模块
        if self.build_video_processing():
            logger.info("✅ 视频处理模块构建成功")
        else:
            logger.error("❌ 视频处理模块构建失败")
            success = False
        
        return success
    
    def build_qt_client(self) -> bool:
        """构建Qt客户端"""
        logger.info("🖥️ 构建Qt客户端...")
        
        qt_client_path = self.project_root / "src/frontend/qt-client-new"
        build_dir = qt_client_path / "build-qt"
        
        if build_dir.exists():
            shutil.rmtree(build_dir)
        build_dir.mkdir(parents=True)
        
        try:
            # CMake配置
            cmake_cmd = [
                "cmake",
                "-G", "Visual Studio 17 2022",
                "-A", "x64",
                "-DCMAKE_BUILD_TYPE=Release"
            ]
            
            # 添加Qt6路径
            if "Qt6_DIR" in os.environ:
                cmake_cmd.extend(["-DQt6_DIR", os.environ["Qt6_DIR"]])
            
            # 添加OpenCV路径
            if "OpenCV_DIR" in os.environ:
                cmake_cmd.extend(["-DOpenCV_DIR", os.environ["OpenCV_DIR"]])
            
            # 添加工具链文件
            if "CMAKE_TOOLCHAIN_FILE" in os.environ:
                cmake_cmd.extend(["-DCMAKE_TOOLCHAIN_FILE", os.environ["CMAKE_TOOLCHAIN_FILE"]])
            
            cmake_cmd.append("..")
            
            subprocess.run(cmake_cmd, cwd=build_dir, check=True)
            
            # 构建
            build_cmd = [
                "cmake", "--build", ".", 
                "--config", "Release", 
                "--parallel"
            ]
            
            subprocess.run(build_cmd, cwd=build_dir, check=True)
            
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"Qt客户端构建失败: {e}")
            return False
    
    def build_video_processing(self) -> bool:
        """构建视频处理模块"""
        logger.info("🎥 构建视频处理模块...")
        
        video_processing_path = self.project_root / "src/video-processing"
        build_dir = video_processing_path / "build"
        
        if build_dir.exists():
            shutil.rmtree(build_dir)
        build_dir.mkdir(parents=True)
        
        try:
            # CMake配置
            cmake_cmd = [
                "cmake",
                "-G", "Visual Studio 17 2022",
                "-A", "x64",
                "-DCMAKE_BUILD_TYPE=Release"
            ]
            
            # 添加OpenCV路径
            if "OpenCV_DIR" in os.environ:
                cmake_cmd.extend(["-DOpenCV_DIR", os.environ["OpenCV_DIR"]])
            
            # 添加工具链文件
            if "CMAKE_TOOLCHAIN_FILE" in os.environ:
                cmake_cmd.extend(["-DCMAKE_TOOLCHAIN_FILE", os.environ["CMAKE_TOOLCHAIN_FILE"]])
            
            cmake_cmd.append("..")
            
            subprocess.run(cmake_cmd, cwd=build_dir, check=True)
            
            # 构建
            build_cmd = [
                "cmake", "--build", ".", 
                "--config", "Release", 
                "--parallel"
            ]
            
            subprocess.run(build_cmd, cwd=build_dir, check=True)
            
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"视频处理模块构建失败: {e}")
            return False

    def build_ai_detection_service(self) -> bool:
        """构建AI检测服务"""
        logger.info("🤖 构建AI检测服务...")

        ai_detection_path = self.project_root / "src/ai-detection"

        if not ai_detection_path.exists():
            logger.error("AI检测服务路径不存在")
            return False

        try:
            # 检查requirements.txt
            requirements_file = ai_detection_path / "requirements.txt"
            if not requirements_file.exists():
                logger.warning("requirements.txt不存在，创建默认文件")
                with open(requirements_file, 'w') as f:
                    f.write("""fastapi==0.100.0
uvicorn==0.23.0
opencv-python==4.8.0.76
numpy==1.24.3
torch==2.0.1
torchvision==0.15.2
Pillow==10.0.0
python-multipart==0.0.6
aiofiles==23.1.0
""")

            # 创建虚拟环境
            venv_path = ai_detection_path / "venv"
            if venv_path.exists():
                shutil.rmtree(venv_path)

            subprocess.run([sys.executable, "-m", "venv", "venv"],
                         cwd=ai_detection_path, check=True)

            # 安装依赖
            if self.platform == "windows":
                pip_path = venv_path / "Scripts" / "pip.exe"
                python_path = venv_path / "Scripts" / "python.exe"
            else:
                pip_path = venv_path / "bin" / "pip"
                python_path = venv_path / "bin" / "python"

            subprocess.run([str(pip_path), "install", "-r", "requirements.txt"],
                         cwd=ai_detection_path, check=True)

            # 验证安装
            subprocess.run([str(python_path), "-c", "import fastapi, torch, cv2"],
                         cwd=ai_detection_path, check=True)

            logger.info("✅ AI检测服务构建成功")
            return True

        except subprocess.CalledProcessError as e:
            logger.error(f"AI检测服务构建失败: {e}")
            return False

    def deploy_services(self) -> bool:
        """部署服务"""
        logger.info("🚀 部署服务...")

        success = True

        # 部署后端服务
        if self.platform == "linux":
            if not self.deploy_backend_services():
                success = False

        # 部署Edge-Model-Infra AI检测服务
        if not self.deploy_edge_model_infra():
            success = False

        # 部署前端应用
        if self.platform == "windows":
            if not self.deploy_frontend_applications():
                success = False

        return success

    def deploy_backend_services(self) -> bool:
        """部署后端服务"""
        logger.info("⚙️ 部署后端服务...")

        build_dir = self.project_root / "build-linux"
        if not build_dir.exists():
            logger.error("后端构建目录不存在，请先构建后端服务")
            return False

        # 创建systemd服务文件
        for service in self.deployment_config["backend"]["services"]:
            service_name = service["name"]
            service_port = service["port"]
            executable_path = build_dir / f"{service_name}-service"

            if not executable_path.exists():
                logger.warning(f"⚠️ 服务可执行文件不存在: {executable_path}")
                continue

            # 创建systemd服务文件
            service_file_content = f"""[Unit]
Description=VideoCall System {service_name.title()} Service
After=network.target postgresql.service redis-server.service mongod.service

[Service]
Type=simple
User=videocall
Group=videocall
WorkingDirectory={self.project_root}
ExecStart={executable_path}
Restart=always
RestartSec=5
Environment=GO_ENV=production
Environment=PORT={service_port}
Environment=DB_HOST=localhost
Environment=REDIS_HOST=localhost
Environment=MONGO_HOST=localhost

[Install]
WantedBy=multi-user.target
"""

            service_file_path = f"/etc/systemd/system/videocall-{service_name}.service"

            try:
                # 写入服务文件
                with open(service_file_path, 'w') as f:
                    f.write(service_file_content)

                # 重新加载systemd
                subprocess.run(["sudo", "systemctl", "daemon-reload"], check=True)

                # 启用服务
                subprocess.run(["sudo", "systemctl", "enable", f"videocall-{service_name}"], check=True)

                logger.info(f"✅ {service_name} 服务配置完成")

            except Exception as e:
                logger.error(f"❌ {service_name} 服务配置失败: {e}")

        return True

    def deploy_ai_detection_service(self) -> bool:
        """部署AI检测服务"""
        logger.info("🤖 部署AI检测服务...")

        ai_detection_path = self.project_root / "src/ai-detection"

        if not ai_detection_path.exists():
            logger.error("AI检测服务路径不存在")
            return False

        try:
            # 创建启动脚本
            start_script_content = f"""#!/bin/bash
cd {ai_detection_path}
source venv/bin/activate
export PYTHONPATH={ai_detection_path}:$PYTHONPATH
uvicorn app:app --host 0.0.0.0 --port 8085 --workers 4
"""

            start_script_path = ai_detection_path / "start_ai_service.sh"
            with open(start_script_path, 'w') as f:
                f.write(start_script_content)

            # 设置执行权限
            os.chmod(start_script_path, 0o755)

            # 创建systemd服务文件（Linux环境）
            if self.platform == "linux":
                service_file_content = f"""[Unit]
Description=VideoCall System AI Detection Service
After=network.target

[Service]
Type=simple
User=videocall
Group=videocall
WorkingDirectory={ai_detection_path}
ExecStart={start_script_path}
Restart=always
RestartSec=5
Environment=PYTHON_ENV=production

[Install]
WantedBy=multi-user.target
"""

                service_file_path = "/etc/systemd/system/videocall-ai-detection.service"

                with open(service_file_path, 'w') as f:
                    f.write(service_file_content)

                subprocess.run(["sudo", "systemctl", "daemon-reload"], check=True)
                subprocess.run(["sudo", "systemctl", "enable", "videocall-ai-detection"], check=True)

            logger.info("✅ AI检测服务部署完成")
            return True

        except Exception as e:
            logger.error(f"AI检测服务部署失败: {e}")
            return False

    def deploy_edge_model_infra(self) -> bool:
        """部署Edge-Model-Infra AI检测服务"""
        logger.info("🤖 部署Edge-Model-Infra AI检测服务...")

        try:
            # 导入Edge-Model-Infra集成器
            from edge_model_infra_integration import EdgeModelInfraIntegrator

            integrator = EdgeModelInfraIntegrator()

            # 运行完整集成
            if not integrator.run_full_integration():
                logger.error("Edge-Model-Infra集成失败")
                return False

            logger.info("✅ Edge-Model-Infra部署成功")
            return True

        except ImportError:
            logger.error("无法导入Edge-Model-Infra集成器")
            return False
        except Exception as e:
            logger.error(f"Edge-Model-Infra部署失败: {e}")
            return False

    def deploy_with_docker_wsl(self) -> bool:
        """在WSL中使用Docker部署后端服务"""
        logger.info("🐳 在WSL中使用Docker部署后端服务...")

        # 检查是否在WSL环境中
        if not self.is_wsl_environment():
            logger.error("当前不在WSL环境中，请在WSL中运行此部署")
            return False

        # 创建Docker配置
        if not self.create_docker_configurations():
            return False

        # 构建Docker镜像
        if not self.build_docker_images():
            return False

        # 部署Docker服务
        if not self.deploy_docker_services():
            return False

        # 配置网络访问
        if not self.configure_wsl_network():
            return False

        return True

    def is_wsl_environment(self) -> bool:
        """检查是否在WSL环境中"""
        try:
            with open('/proc/version', 'r') as f:
                version_info = f.read().lower()
                return 'microsoft' in version_info or 'wsl' in version_info
        except:
            return False

    def create_docker_configurations(self) -> bool:
        """创建Docker配置文件"""
        logger.info("📝 创建Docker配置文件...")

        # 创建Docker Compose文件
        docker_compose_content = self.generate_docker_compose_config()

        docker_compose_path = self.project_root / "deployment" / "docker-compose.wsl.yml"
        with open(docker_compose_path, 'w', encoding='utf-8') as f:
            f.write(docker_compose_content)

        # 创建各服务的Dockerfile
        self.create_service_dockerfiles()

        # 创建环境配置文件
        self.create_environment_files()

        logger.info("✅ Docker配置文件创建完成")
        return True

    def generate_docker_compose_config(self) -> str:
        """生成Docker Compose配置"""
        return """version: '3.8'

services:
  # 数据库服务
  postgres:
    image: postgres:15-alpine
    container_name: videocall-postgres
    environment:
      POSTGRES_DB: videocall_system
      POSTGRES_USER: videocall_user
      POSTGRES_PASSWORD: videocall_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/database/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - videocall-network
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: videocall-redis
    command: redis-server --requirepass videocall_redis_password
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - videocall-network
    restart: unless-stopped

  mongodb:
    image: mongo:6
    container_name: videocall-mongodb
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: videocall_mongo_password
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    networks:
      - videocall-network
    restart: unless-stopped

  # 后端微服务
  user-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.user-service
    container_name: videocall-user-service
    environment:
      - GO_ENV=production
      - PORT=8081
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=videocall_system
      - DB_USER=videocall_user
      - DB_PASSWORD=videocall_password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=videocall_redis_password
    ports:
      - "8081:8081"
    depends_on:
      - postgres
      - redis
    networks:
      - videocall-network
    restart: unless-stopped

  meeting-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.meeting-service
    container_name: videocall-meeting-service
    environment:
      - GO_ENV=production
      - PORT=8082
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=videocall_system
      - DB_USER=videocall_user
      - DB_PASSWORD=videocall_password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=videocall_redis_password
    ports:
      - "8082:8082"
    depends_on:
      - postgres
      - redis
    networks:
      - videocall-network
    restart: unless-stopped

  signaling-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.signaling-service
    container_name: videocall-signaling-service
    environment:
      - GO_ENV=production
      - PORT=8083
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=videocall_redis_password
    ports:
      - "8083:8083"
    depends_on:
      - redis
    networks:
      - videocall-network
    restart: unless-stopped

  media-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.media-service
    container_name: videocall-media-service
    environment:
      - GO_ENV=production
      - PORT=8084
      - MONGO_HOST=mongodb
      - MONGO_PORT=27017
      - MONGO_USER=admin
      - MONGO_PASSWORD=videocall_mongo_password
    ports:
      - "8084:8084"
    depends_on:
      - mongodb
    networks:
      - videocall-network
    restart: unless-stopped

  # Edge-Model-Infra Unit Manager
  edge-unit-manager:
    build:
      context: ./Edge-Model-Infra/unit-manager
      dockerfile: Dockerfile
    container_name: videocall-edge-unit-manager
    ports:
      - "10001:10001"
    volumes:
      - ./Edge-Model-Infra/unit-manager/master_config.json:/app/master_config.json
      - /tmp/llm:/tmp/llm
    networks:
      - videocall-network
    depends_on:
      - edge-ai-detection
    restart: unless-stopped

  # Edge-Model-Infra AI Detection Node
  edge-ai-detection:
    build:
      context: ./Edge-Model-Infra/node/ai-detection
      dockerfile: Dockerfile
    container_name: videocall-edge-ai-detection
    volumes:
      - ./Edge-Model-Infra/node/ai-detection/models:/app/models
      - ./Edge-Model-Infra/node/ai-detection/config:/app/config
      - /tmp/llm:/tmp/llm
      - /tmp/detection_uploads:/tmp/detection_uploads
      - ./storage/detection:/app/storage
    networks:
      - videocall-network
    restart: unless-stopped
    environment:
      - MODEL_PATH=/app/models
      - UPLOAD_PATH=/tmp/detection_uploads
      - STORAGE_PATH=/app/storage

  # Legacy AI Detection Service (Python) - Fallback
  ai-detection-legacy:
    build:
      context: ./src/ai-detection
      dockerfile: Dockerfile
    container_name: videocall-ai-detection-legacy
    environment:
      - PYTHON_ENV=production
      - PORT=8085
    ports:
      - "8085:8085"
    volumes:
      - ./src/ai-detection/models:/app/models
      - ./storage/detection:/app/storage
    networks:
      - videocall-network
    restart: unless-stopped
    profiles:
      - legacy

  notification-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.notification-service
    container_name: videocall-notification-service
    environment:
      - GO_ENV=production
      - PORT=8086
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=videocall_system
      - DB_USER=videocall_user
      - DB_PASSWORD=videocall_password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=videocall_redis_password
    ports:
      - "8086:8086"
    depends_on:
      - postgres
      - redis
    networks:
      - videocall-network
    restart: unless-stopped

  record-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.record-service
    container_name: videocall-record-service
    environment:
      - GO_ENV=production
      - PORT=8087
      - MONGO_HOST=mongodb
      - MONGO_PORT=27017
      - MONGO_USER=admin
      - MONGO_PASSWORD=videocall_mongo_password
    ports:
      - "8087:8087"
    depends_on:
      - mongodb
    volumes:
      - ./storage/media:/app/storage
    networks:
      - videocall-network
    restart: unless-stopped

  smart-editing-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.smart-editing-service
    container_name: videocall-smart-editing-service
    environment:
      - GO_ENV=production
      - PORT=8088
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=videocall_system
      - DB_USER=videocall_user
      - DB_PASSWORD=videocall_password
    ports:
      - "8088:8088"
    depends_on:
      - postgres
    networks:
      - videocall-network
    restart: unless-stopped

  gateway-service:
    build:
      context: .
      dockerfile: deployment/docker/Dockerfile.gateway-service
    container_name: videocall-gateway
    environment:
      - GO_ENV=production
      - PORT=8080
      - USER_SERVICE_URL=http://user-service:8081
      - MEETING_SERVICE_URL=http://meeting-service:8082
      - SIGNALING_SERVICE_URL=http://signaling-service:8083
      - MEDIA_SERVICE_URL=http://media-service:8084
      - AI_DETECTION_SERVICE_URL=http://edge-unit-manager:10001
      - EDGE_AI_DETECTION_URL=http://edge-ai-detection:5000
      - NOTIFICATION_SERVICE_URL=http://notification-service:8086
      - RECORD_SERVICE_URL=http://record-service:8087
      - SMART_EDITING_SERVICE_URL=http://smart-editing-service:8088
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - meeting-service
      - signaling-service
      - media-service
      - edge-unit-manager
      - notification-service
      - record-service
      - smart-editing-service
    networks:
      - videocall-network
    restart: unless-stopped

  # Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: videocall-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./deployment/nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./deployment/nginx/ssl:/etc/nginx/ssl
    depends_on:
      - gateway-service
    networks:
      - videocall-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  mongodb_data:

networks:
  videocall-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
"""
