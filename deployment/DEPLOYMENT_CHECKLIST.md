# VideoCall System - 部署检查清单

## 🚀 部署前准备

### 📋 环境检查清单

#### Windows 前端环境
- [ ] Windows 10/11 (64-bit)
- [ ] 8GB+ RAM (推荐16GB)
- [ ] 支持OpenGL 3.3+的显卡
- [ ] 2GB+ 可用存储空间
- [ ] 稳定的网络连接

#### Linux 后端环境
- [ ] Ubuntu 20.04+ / CentOS 8+ / RHEL 8+
- [ ] 4核心+ CPU (推荐8核心)
- [ ] 8GB+ RAM (推荐16GB)
- [ ] 50GB+ 可用存储空间
- [ ] 固定IP地址
- [ ] 开放的网络端口 (80, 443, 8080-8087)

### 🔧 依赖检查

#### Windows 依赖
- [ ] Visual Studio Build Tools 2019+
- [ ] Qt6 (6.5.0+)
- [ ] OpenCV (4.8.0+)
- [ ] CMake (3.20+)
- [ ] Git

#### Linux 依赖
- [ ] Go (1.21.5+)
- [ ] Python (3.9+)
- [ ] Node.js (LTS)
- [ ] PostgreSQL (15+)
- [ ] Redis (7.0+)
- [ ] MongoDB (6.0+)
- [ ] Docker (可选)

## 📝 部署步骤

### 第一阶段：环境准备

#### 1. Windows 前端环境设置
```powershell
# 运行自动化设置脚本
.\scripts\cross-platform\setup_development_environment.ps1 -All

# 验证安装
qt6-config --version
cmake --version
git --version
```

#### 2. Linux 后端环境设置
```bash
# 运行自动化设置脚本
chmod +x scripts/cross-platform/setup_backend_linux.sh
./scripts/cross-platform/setup_backend_linux.sh --all

# 验证安装
go version
python3 --version
docker --version
```

### 第二阶段：后端部署

#### 1. 数据库初始化
```bash
# PostgreSQL
sudo -u postgres createdb videocall_system
sudo -u postgres createuser videocall_user
psql -U videocall_user -d videocall_system -f config/database/init.sql

# Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# MongoDB
sudo systemctl start mongod
sudo systemctl enable mongod
```

#### 2. 构建后端服务
```bash
# 使用构建脚本
./scripts/cross-platform/setup_backend_linux.sh --build

# 或使用Python构建脚本
python3 scripts/cross-platform/cross_platform_build.py --platform linux --component backend
```

#### 3. 启动后端服务
```bash
cd build-linux
./start-all-services.sh

# 检查服务状态
./status.sh
```

### 第三阶段：前端部署

#### 1. 构建前端应用
```powershell
# 设置环境变量
$env:Qt6_DIR = "C:\Qt\6.5.0\msvc2019_64"
$env:OpenCV_DIR = "C:\vcpkg\installed\x64-windows"

# 构建项目
cd src\frontend\qt-client-new
.\scripts\build_effects_demo.sh --release

# 或使用Python构建脚本
python scripts\cross-platform\cross_platform_build.py --platform windows --component frontend
```

#### 2. 测试前端应用
```powershell
cd build-windows\Release
.\VideoEffectsDemo.exe
```

### 第四阶段：集成测试

#### 1. 网络连通性测试
```bash
# 测试后端API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/status

# 测试数据库连接
psql -U videocall_user -d videocall_system -c "SELECT version();"
redis-cli ping
mongo --eval "db.runCommand('ping')"
```

#### 2. 功能测试
```powershell
# 前端功能测试
.\VideoEffectsDemo.exe --test

# 后端API测试
python quick_test_api.py
```

## 🔍 验证检查

### 后端服务验证
- [ ] 所有微服务正常启动
- [ ] 数据库连接正常
- [ ] API接口响应正常
- [ ] 日志输出正常

### 前端应用验证
- [ ] 应用正常启动
- [ ] 摄像头访问正常
- [ ] 视频特效功能正常
- [ ] 网络通信正常

### 集成验证
- [ ] 前后端通信正常
- [ ] 视频通话功能正常
- [ ] AI检测功能正常
- [ ] 文件上传下载正常

## 🚨 常见问题排查

### Windows 前端问题
1. **Qt6找不到**
   - 检查Qt6_DIR环境变量
   - 确认Qt6安装路径正确

2. **OpenCV链接错误**
   - 检查OpenCV_DIR环境变量
   - 确认vcpkg安装正确

3. **构建失败**
   - 检查Visual Studio Build Tools
   - 确认CMake版本兼容

### Linux 后端问题
1. **服务启动失败**
   - 检查端口占用：`netstat -tuln | grep 808`
   - 查看服务日志：`journalctl -u videocall-*`

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数和权限

3. **Go模块下载失败**
   - 设置Go代理：`go env -w GOPROXY=https://goproxy.cn`
   - 检查网络连接

## 📊 性能监控

### 系统资源监控
```bash
# CPU和内存使用
htop

# 磁盘使用
df -h

# 网络连接
ss -tuln
```

### 应用监控
```bash
# 服务状态
./status.sh

# 日志监控
tail -f /var/log/videocall-system/*.log

# 性能指标
curl http://localhost:8080/metrics
```

## 🔒 安全配置

### 防火墙设置
```bash
# Ubuntu
sudo ufw allow 22,80,443,8080:8087/tcp
sudo ufw enable

# CentOS
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080-8087/tcp
sudo firewall-cmd --reload
```

### SSL证书配置
```bash
# Let's Encrypt证书
sudo certbot --nginx -d yourdomain.com

# 或自签名证书（开发环境）
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/ssl/private/videocall.key \
    -out /etc/ssl/certs/videocall.crt
```

## 📞 技术支持

如果遇到问题：
1. 查看相关日志文件
2. 检查系统资源使用
3. 验证网络连接
4. 参考故障排除文档
5. 联系技术支持团队

---

**部署完成后，您将拥有一个功能完整的跨平台智能视频会议系统！** 🎉
