# 🚀 快速启动指南

## 📋 系统要求

- **Go**: 1.19+
- **Python**: 3.7+
- **浏览器**: Chrome/Firefox/Safari (支持WebRTC)
- **数据库**: PostgreSQL (已配置)
- **Redis**: (可选，用于缓存)

## ⚡ 快速启动

### 方法1: 使用启动脚本 (推荐)
```bash
# Windows
run_project.bat

# Linux/Mac
./run_project.sh
```

### 方法2: 手动启动
```bash
# 1. 启动后端服务
cd core/backend
go run main.go

# 2. 启动前端服务 (新终端)
cd web_interface
python -m http.server 3000

# 3. 打开浏览器
# 访问: http://localhost:3000
```

## 🧪 测试系统

### 运行自动化测试
```bash
python test_real_call.py
```

### 查看功能演示
```bash
python demo_real_call.py
```

## 📱 使用步骤

1. **打开浏览器**访问 http://localhost:3000
2. **注册/登录**用户账户
3. **点击"开始通话"**按钮
4. **允许摄像头和麦克风**权限
5. **测试音视频通话**功能

## 🔧 功能特性

### ✅ 已实现功能
- 📞 **真正的WebRTC P2P通话**
- 🔌 **WebSocket信令服务器**
- 🏠 **通话房间管理**
- 📊 **通话状态跟踪**
- 📋 **通话历史记录**
- 🔒 **AI安全检测**
- 👥 **用户认证管理**

### 🎯 技术亮点
- **真正的点对点连接**
- **实时信令传输**
- **自动ICE候选收集**
- **连接状态监控**
- **音视频质量控制**
- **安全风险检测**

## 📡 API端点

### 后端API
- **健康检查**: `GET /health`
- **用户注册**: `POST /api/v1/auth/register`
- **用户登录**: `POST /api/v1/auth/login`
- **开始通话**: `POST /api/v1/calls/start`
- **结束通话**: `POST /api/v1/calls/end`
- **通话历史**: `GET /api/v1/calls/history`

### WebSocket
- **信令连接**: `ws://localhost:8000/ws/call/{room_id}`

## 🐛 故障排除

### 常见问题

#### 1. 后端服务启动失败
```bash
# 检查Go环境
go version

# 检查依赖
cd core/backend
go mod tidy
go run main.go
```

#### 2. 前端页面无法访问
```bash
# 检查端口是否被占用
netstat -an | findstr :3000

# 使用不同端口
python -m http.server 3001
```

#### 3. WebRTC连接失败
- 确保浏览器支持WebRTC
- 检查摄像头和麦克风权限
- 确保网络连接正常

#### 4. 数据库连接失败
```bash
# 检查PostgreSQL服务
# 确保数据库配置正确
```

### 调试模式
```bash
# 启用详细日志
set GIN_MODE=debug
set LOG_LEVEL=debug
```

## 📊 系统状态

### 检查服务状态
```bash
# 后端服务
curl http://localhost:8000/health

# 前端服务
curl http://localhost:3000
```

### 查看日志
- 后端日志: 在运行Go程序的终端中查看
- 前端日志: 浏览器开发者工具控制台

## 🔒 安全配置

### 生产环境设置
```bash
# 设置环境变量
set GIN_MODE=release
set JWT_SECRET=your-secure-secret-key
set DB_PASSWORD=your-secure-db-password
```

### HTTPS配置
- 配置SSL证书
- 使用WSS (WebSocket Secure)
- 启用CORS安全策略

## 📈 性能优化

### 后端优化
- 启用连接池
- 配置缓存
- 优化数据库查询

### 前端优化
- 启用WebRTC优化
- 配置媒体流质量
- 实现错误重试机制

## 🎉 成功标志

当您看到以下内容时，说明系统运行正常：

1. ✅ 后端服务响应: `{"status":"ok","message":"VideoCall Backend is running"}`
2. ✅ 前端页面正常加载
3. ✅ 用户注册/登录成功
4. ✅ 通话创建成功
5. ✅ WebSocket连接建立
6. ✅ 音视频通话功能正常

## 📞 技术支持

如果遇到问题，请：

1. 查看浏览器控制台错误信息
2. 检查后端服务日志
3. 运行测试脚本验证功能
4. 检查网络连接状态
5. 确保所有依赖正确安装

---

**🎯 现在您可以开始使用真正的WebRTC通话系统了！** 