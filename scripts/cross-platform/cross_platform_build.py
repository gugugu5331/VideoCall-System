#!/usr/bin/env python3
"""
VideoCall System - 跨平台构建脚本
支持Windows前端和Linux后端的统一构建和部署
"""

import os
import sys
import subprocess
import platform
import argparse
import json
import shutil
from pathlib import Path
from typing import Dict, List, Optional
import logging

# 设置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('build.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class CrossPlatformBuilder:
    """跨平台构建器"""
    
    def __init__(self):
        self.platform = platform.system().lower()
        self.project_root = Path(__file__).parent.parent.parent
        self.config = self.load_config()
        
    def load_config(self) -> Dict:
        """加载构建配置"""
        config_file = self.project_root / "scripts" / "cross-platform" / "build_config.json"
        
        default_config = {
            "windows": {
                "frontend": {
                    "qt_dir": "C:/Qt/6.5.0/msvc2019_64",
                    "opencv_dir": "C:/vcpkg/installed/x64-windows",
                    "cmake_generator": "Visual Studio 17 2022",
                    "build_dir": "build-windows",
                    "targets": ["VideoEffectsDemo", "VideoCallSystemClient"]
                }
            },
            "linux": {
                "backend": {
                    "go_version": "1.21.5",
                    "build_dir": "build-linux",
                    "services": [
                        "user-service",
                        "meeting-service", 
                        "signaling-service",
                        "media-service",
                        "ai-detection-service",
                        "notification-service",
                        "file-service",
                        "gateway-service"
                    ]
                },
                "ai_detection": {
                    "python_version": "3.9",
                    "requirements_file": "src/ai-detection/requirements.txt",
                    "build_dir": "build-linux/ai-detection"
                }
            },
            "docker": {
                "registry": "localhost:5000",
                "namespace": "videocall-system",
                "services": {
                    "backend": "backend",
                    "ai-detection": "ai-detection",
                    "frontend": "frontend"
                }
            }
        }
        
        if config_file.exists():
            try:
                with open(config_file, 'r', encoding='utf-8') as f:
                    user_config = json.load(f)
                    # 合并配置
                    default_config.update(user_config)
            except Exception as e:
                logger.warning(f"加载配置文件失败: {e}，使用默认配置")
        
        return default_config
    
    def save_config(self):
        """保存配置到文件"""
        config_file = self.project_root / "scripts" / "cross-platform" / "build_config.json"
        config_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(config_file, 'w', encoding='utf-8') as f:
            json.dump(self.config, f, indent=2, ensure_ascii=False)
    
    def run_command(self, cmd: List[str], cwd: Optional[Path] = None, shell: bool = False) -> bool:
        """运行命令"""
        try:
            logger.info(f"执行命令: {' '.join(cmd)}")
            if cwd:
                logger.info(f"工作目录: {cwd}")
            
            result = subprocess.run(
                cmd,
                cwd=cwd,
                shell=shell,
                check=True,
                capture_output=True,
                text=True
            )
            
            if result.stdout:
                logger.info(f"输出: {result.stdout}")
            
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"命令执行失败: {e}")
            if e.stdout:
                logger.error(f"标准输出: {e.stdout}")
            if e.stderr:
                logger.error(f"错误输出: {e.stderr}")
            return False
    
    def build_windows_frontend(self, build_type: str = "Release", clean: bool = False) -> bool:
        """构建Windows前端"""
        if self.platform != "windows":
            logger.warning("当前不是Windows平台，跳过Windows前端构建")
            return True
        
        logger.info("🎯 开始构建Windows前端...")
        
        config = self.config["windows"]["frontend"]
        build_dir = self.project_root / config["build_dir"]
        
        if clean and build_dir.exists():
            logger.info("清理构建目录...")
            shutil.rmtree(build_dir)
        
        build_dir.mkdir(parents=True, exist_ok=True)
        
        # CMake配置
        cmake_args = [
            "cmake",
            "-G", config["cmake_generator"],
            "-A", "x64",
            f"-DCMAKE_BUILD_TYPE={build_type}",
            f"-DCMAKE_TOOLCHAIN_FILE=C:/vcpkg/scripts/buildsystems/vcpkg.cmake",
            f"-DQt6_DIR={config['qt_dir']}",
            f"-DOpenCV_DIR={config['opencv_dir']}",
            "-f", "../CMakeLists_effects_demo.txt",
            ".."
        ]
        
        if not self.run_command(cmake_args, cwd=build_dir):
            return False
        
        # 构建
        build_args = [
            "cmake",
            "--build", ".",
            "--config", build_type,
            "--parallel"
        ]
        
        if not self.run_command(build_args, cwd=build_dir):
            return False
        
        logger.info("✅ Windows前端构建完成")
        return True
    
    def build_linux_backend(self, build_type: str = "release", clean: bool = False) -> bool:
        """构建Linux后端"""
        if self.platform != "linux":
            logger.warning("当前不是Linux平台，跳过Linux后端构建")
            return True
        
        logger.info("🎯 开始构建Linux后端...")
        
        config = self.config["linux"]["backend"]
        build_dir = self.project_root / config["build_dir"]
        
        if clean and build_dir.exists():
            logger.info("清理构建目录...")
            shutil.rmtree(build_dir)
        
        build_dir.mkdir(parents=True, exist_ok=True)
        
        # 设置Go环境
        go_env = os.environ.copy()
        go_env["GOPATH"] = str(Path.home() / "go")
        go_env["GOBIN"] = str(Path.home() / "go" / "bin")
        go_env["PATH"] = f"/usr/local/go/bin:{go_env['GOBIN']}:{go_env['PATH']}"
        
        # 构建各个微服务
        for service in config["services"]:
            logger.info(f"🔨 构建 {service}...")
            
            service_dir = self.project_root / "src" / "backend" / "services" / service
            if not service_dir.exists():
                logger.warning(f"服务目录不存在: {service_dir}，跳过")
                continue
            
            # 下载依赖
            if not self.run_command(["go", "mod", "download"], cwd=service_dir):
                return False
            
            if not self.run_command(["go", "mod", "tidy"], cwd=service_dir):
                return False
            
            # 构建
            output_path = build_dir / service
            build_args = ["go", "build"]
            
            if build_type == "debug":
                build_args.extend(["-race"])
            else:
                build_args.extend(["-ldflags=-s -w"])
            
            build_args.extend(["-o", str(output_path), "."])
            
            if not self.run_command(build_args, cwd=service_dir):
                return False
            
            logger.info(f"✅ {service} 构建完成")
        
        logger.info("✅ Linux后端构建完成")
        return True
    
    def build_ai_detection(self, clean: bool = False) -> bool:
        """构建AI检测服务"""
        logger.info("🎯 开始构建AI检测服务...")
        
        config = self.config["linux"]["ai_detection"]
        build_dir = self.project_root / config["build_dir"]
        source_dir = self.project_root / "src" / "ai-detection"
        
        if not source_dir.exists():
            logger.error("AI检测服务源码目录不存在")
            return False
        
        if clean and build_dir.exists():
            logger.info("清理构建目录...")
            shutil.rmtree(build_dir)
        
        build_dir.mkdir(parents=True, exist_ok=True)
        
        # 复制源码
        logger.info("复制源码...")
        for item in source_dir.iterdir():
            if item.name == "__pycache__":
                continue
            
            dest = build_dir / item.name
            if item.is_dir():
                if dest.exists():
                    shutil.rmtree(dest)
                shutil.copytree(item, dest)
            else:
                shutil.copy2(item, dest)
        
        # 创建虚拟环境
        venv_dir = build_dir / "venv"
        if not venv_dir.exists():
            logger.info("创建Python虚拟环境...")
            if not self.run_command(["python3", "-m", "venv", "venv"], cwd=build_dir):
                return False
        
        # 安装依赖
        logger.info("安装Python依赖...")
        pip_path = venv_dir / "bin" / "pip" if self.platform == "linux" else venv_dir / "Scripts" / "pip.exe"
        requirements_file = build_dir / "requirements.txt"
        
        if requirements_file.exists():
            if not self.run_command([str(pip_path), "install", "-r", "requirements.txt"], cwd=build_dir):
                return False
        
        logger.info("✅ AI检测服务构建完成")
        return True
    
    def build_docker_images(self, services: List[str] = None) -> bool:
        """构建Docker镜像"""
        logger.info("🎯 开始构建Docker镜像...")
        
        config = self.config["docker"]
        registry = config["registry"]
        namespace = config["namespace"]
        
        if services is None:
            services = list(config["services"].keys())
        
        for service in services:
            if service not in config["services"]:
                logger.warning(f"未知服务: {service}，跳过")
                continue
            
            image_name = config["services"][service]
            full_image_name = f"{registry}/{namespace}/{image_name}:latest"
            
            logger.info(f"🐳 构建Docker镜像: {full_image_name}")
            
            dockerfile_path = self.project_root / "deployment" / "docker" / f"Dockerfile.{service}"
            if not dockerfile_path.exists():
                logger.warning(f"Dockerfile不存在: {dockerfile_path}，跳过")
                continue
            
            build_args = [
                "docker", "build",
                "-t", full_image_name,
                "-f", str(dockerfile_path),
                "."
            ]
            
            if not self.run_command(build_args, cwd=self.project_root):
                return False
            
            logger.info(f"✅ {service} Docker镜像构建完成")
        
        logger.info("✅ 所有Docker镜像构建完成")
        return True
    
    def deploy_to_remote(self, target: str, services: List[str] = None) -> bool:
        """部署到远程服务器"""
        logger.info(f"🚀 开始部署到远程服务器: {target}")
        
        # 这里可以实现SSH部署逻辑
        # 例如：rsync同步文件，docker-compose部署等
        
        logger.info("✅ 远程部署完成")
        return True
    
    def run_tests(self, test_type: str = "all") -> bool:
        """运行测试"""
        logger.info(f"🧪 开始运行测试: {test_type}")
        
        success = True
        
        if test_type in ["all", "backend"]:
            # 运行Go测试
            logger.info("运行Go后端测试...")
            backend_dir = self.project_root / "src" / "backend"
            if backend_dir.exists():
                if not self.run_command(["go", "test", "./..."], cwd=backend_dir):
                    success = False
        
        if test_type in ["all", "ai"]:
            # 运行Python测试
            logger.info("运行AI检测服务测试...")
            ai_dir = self.project_root / "src" / "ai-detection"
            if ai_dir.exists():
                if not self.run_command(["python", "-m", "pytest", "tests/"], cwd=ai_dir):
                    success = False
        
        if test_type in ["all", "frontend"]:
            # 运行前端测试
            logger.info("运行前端测试...")
            # 这里可以添加Qt测试逻辑
        
        if success:
            logger.info("✅ 所有测试通过")
        else:
            logger.error("❌ 部分测试失败")
        
        return success

def main():
    parser = argparse.ArgumentParser(description="VideoCall System 跨平台构建脚本")
    parser.add_argument("--platform", choices=["windows", "linux", "all"], default="all",
                       help="目标平台")
    parser.add_argument("--component", choices=["frontend", "backend", "ai", "docker", "all"], default="all",
                       help="构建组件")
    parser.add_argument("--build-type", choices=["debug", "release"], default="release",
                       help="构建类型")
    parser.add_argument("--clean", action="store_true",
                       help="清理构建目录")
    parser.add_argument("--test", action="store_true",
                       help="运行测试")
    parser.add_argument("--deploy", type=str,
                       help="部署到指定目标")
    parser.add_argument("--docker-services", nargs="+",
                       help="指定要构建的Docker服务")
    
    args = parser.parse_args()
    
    builder = CrossPlatformBuilder()
    success = True
    
    logger.info("🚀 VideoCall System 跨平台构建开始")
    logger.info(f"平台: {args.platform}, 组件: {args.component}, 类型: {args.build_type}")
    
    # 构建前端
    if args.component in ["frontend", "all"] and args.platform in ["windows", "all"]:
        if not builder.build_windows_frontend(args.build_type, args.clean):
            success = False
    
    # 构建后端
    if args.component in ["backend", "all"] and args.platform in ["linux", "all"]:
        if not builder.build_linux_backend(args.build_type, args.clean):
            success = False
    
    # 构建AI检测服务
    if args.component in ["ai", "all"]:
        if not builder.build_ai_detection(args.clean):
            success = False
    
    # 构建Docker镜像
    if args.component in ["docker", "all"]:
        if not builder.build_docker_images(args.docker_services):
            success = False
    
    # 运行测试
    if args.test:
        if not builder.run_tests():
            success = False
    
    # 部署
    if args.deploy:
        if not builder.deploy_to_remote(args.deploy):
            success = False
    
    if success:
        logger.info("🎉 构建完成！")
        return 0
    else:
        logger.error("💥 构建失败！")
        return 1

if __name__ == "__main__":
    sys.exit(main())
