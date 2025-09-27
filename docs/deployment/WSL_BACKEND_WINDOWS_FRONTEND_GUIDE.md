# VideoCall System - WSL后端 + Windows前端部署指南

## 🎯 部署架构

本指南介绍如何在Windows环境中部署VideoCall System，采用以下架构：

- **后端服务**: 在WSL (Windows Subsystem for Linux) 中使用Docker部署
- **前端应用**: Windows Qt6客户端
- **AI检测**: 通过Edge-Model-Infra集成，部署在WSL后端
- **网络通信**: Windows前端通过WSL IP地址与后端通信

## 📋 系统要求

### Windows主机要求
- Windows 10 版本 2004 或更高版本 / Windows 11
- 至少 16GB RAM
- 至少 50GB 可用磁盘空间
- 支持虚拟化的CPU

### 软件依赖
- WSL 2
- Docker Desktop for Windows
- Git
- CMake 3.20+
- Qt6 (6.5+)
- Visual Studio 2022 Build Tools
- PowerShell 5.1+

## 🚀 快速部署

### 1. 一键部署（推荐）

```powershell
# 克隆项目
git clone <repository-url>
cd videocall-system

# 执行完整部署
.\scripts\deploy_complete_system.ps1
```

### 2. 分步部署

#### 步骤1: 环境准备

```powershell
# 安装WSL 2
wsl --install

# 安装Docker Desktop
# 下载并安装 Docker Desktop for Windows

# 安装开发工具
.\scripts\cross-platform\setup_development_environment.ps1 -All
```

#### 步骤2: 部署后端服务

```powershell
# 在WSL中部署后端
wsl bash scripts/deploy_wsl_backend.sh
```

#### 步骤3: 构建前端应用

```powershell
# 构建Qt客户端
cd src/frontend/qt-client-new
mkdir build && cd build
cmake .. -G "Visual Studio 17 2022" -A x64
cmake --build . --config Release
```

## 🔧 详细配置

### WSL配置

1. **启用WSL 2**
```powershell
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart
wsl --set-default-version 2
```

2. **安装Ubuntu发行版**
```powershell
wsl --install -d Ubuntu
```

3. **配置WSL资源限制**
创建 `%USERPROFILE%\.wslconfig`:
```ini
[wsl2]
memory=8GB
processors=4
swap=2GB
```

### Docker配置

1. **Docker Desktop设置**
   - 启用WSL 2集成
   - 分配足够的资源（至少8GB RAM）
   - 启用Kubernetes（可选）

2. **WSL中的Docker**
```bash
# 在WSL中验证Docker
docker --version
docker-compose --version
```

### 网络配置

1. **WSL网络**
   - WSL自动分配IP地址（通常在172.x.x.x范围）
   - Windows可以通过WSL IP访问WSL中的服务

2. **防火墙配置**
```powershell
# 允许WSL网络通信
New-NetFirewallRule -DisplayName "WSL" -Direction Inbound -InterfaceAlias "vEthernet (WSL)" -Action Allow
```

## 🌐 服务架构

### 后端服务 (WSL)

| 服务 | 端口 | 描述 |
|------|------|------|
| Nginx | 80 | 反向代理 |
| Gateway | 8080 | API网关 |
| User Service | 8081 | 用户管理 |
| Meeting Service | 8082 | 会议管理 |
| Signaling Service | 8083 | 信令服务 |
| Media Service | 8084 | 媒体处理 |
| AI Detection (Legacy) | 8085 | AI检测（备用） |
| Notification Service | 8086 | 通知服务 |
| Record Service | 8087 | 录制服务 |
| Smart Editing Service | 8088 | 智能编辑 |
| Edge Unit Manager | 10001 | Edge-Model-Infra管理器 |
| Edge AI Detection | 5000 | Edge AI检测节点 |

### 数据库服务 (WSL)

| 服务 | 端口 | 用途 |
|------|------|------|
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存和会话 |
| MongoDB | 27017 | 媒体元数据 |

### 前端应用 (Windows)

- Qt6桌面应用
- 自动检测WSL IP地址
- 支持实时视频特效
- WebRTC P2P通信

## 🔗 网络通信

### 前端到后端通信

```
Windows Qt Client
       ↓ HTTP/WebSocket
WSL IP:80 (Nginx)
       ↓ 反向代理
Docker容器网络
       ↓ 微服务通信
各个后端服务
```

### WSL IP检测

前端应用自动检测WSL IP地址：

```cpp
// C++代码示例
QString WSLNetworkManager::detectWSLIP() {
    QProcess process;
    process.start("wsl", QStringList() << "hostname" << "-I");
    process.waitForFinished();
    QString output = process.readAllStandardOutput().trimmed();
    return output.split(' ').first();
}
```

## 🧪 测试和验证

### 1. 后端服务测试

```bash
# 在WSL中测试
curl http://localhost:80/health
curl http://localhost:80/api/v1/users/health
curl http://localhost:10001/health
```

### 2. 前端连接测试

```powershell
# 从Windows测试WSL服务
$wslIP = wsl hostname -I | ForEach-Object { $_.Trim().Split(' ')[0] }
Invoke-WebRequest "http://$wslIP:80/health"
```

### 3. 完整功能测试

1. 启动前端应用
2. 创建用户账户
3. 创建会议房间
4. 测试视频通话
5. 测试AI检测功能
6. 测试录制功能

## 🛠️ 故障排除

### 常见问题

1. **WSL IP地址变化**
   - 重启WSL后IP可能变化
   - 前端应用会自动重新检测

2. **Docker服务启动失败**
   ```bash
   # 检查Docker状态
   docker info
   
   # 重启Docker服务
   sudo service docker restart
   ```

3. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -tulpn | grep :80
   
   # 停止冲突服务
   docker-compose down
   ```

4. **防火墙阻止连接**
   ```powershell
   # 检查防火墙规则
   Get-NetFirewallRule | Where-Object {$_.DisplayName -like "*WSL*"}
   
   # 添加防火墙例外
   New-NetFirewallRule -DisplayName "VideoCall WSL" -Direction Inbound -Protocol TCP -LocalPort 80,8080,10001 -Action Allow
   ```

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f deployment/docker-compose.wsl.yml logs -f

# 查看特定服务日志
docker-compose -f deployment/docker-compose.wsl.yml logs -f gateway-service

# 查看Edge-Model-Infra日志
docker-compose -f Edge-Model-Infra/docker-compose.ai-detection.yml logs -f
```

## 📊 性能优化

### WSL性能优化

1. **资源分配**
   - 增加WSL内存限制
   - 分配更多CPU核心

2. **磁盘性能**
   - 使用WSL 2文件系统
   - 避免跨文件系统操作

### Docker优化

1. **镜像优化**
   - 使用多阶段构建
   - 最小化镜像大小

2. **容器资源**
   - 合理分配CPU和内存
   - 使用健康检查

## 🔄 更新和维护

### 更新后端服务

```bash
# 拉取最新代码
git pull origin main

# 重新构建和部署
docker-compose -f deployment/docker-compose.wsl.yml up --build -d
```

### 更新前端应用

```powershell
# 重新构建前端
cd src/frontend/qt-client-new/build
cmake --build . --config Release
```

### 数据备份

```bash
# 备份数据库
docker exec videocall-postgres pg_dump -U videocall_user videocall_system > backup.sql

# 备份媒体文件
tar -czf media_backup.tar.gz storage/
```

## 📞 技术支持

如遇到部署问题，请：

1. 检查系统要求是否满足
2. 查看相关日志文件
3. 参考故障排除章节
4. 提交Issue并附上详细错误信息

---

**注意**: 本部署方案适用于开发和测试环境。生产环境部署请参考生产部署指南。
