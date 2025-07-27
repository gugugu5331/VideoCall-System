# 🎥 音视频通话系统 - 一键管理脚本使用指南

## 📋 概述

本管理脚本提供了一套完整的系统管理工具，可以一键启动、停止、测试和监控音视频通话系统的各个组件。

## 🚀 快速开始

### 方法一：图形化菜单（推荐）

1. **双击运行** `manage_system.bat`
2. **选择操作**：
   - `[1]` - 启动所有服务
   - `[2]` - 停止所有服务
   - `[3]` - 重启所有服务
   - `[4]` - 测试所有服务
   - `[5]` - 查看服务状态
   - `[6]` - 清理端口占用
   - `[7]` - 打开前端界面
   - `[8]` - 显示帮助信息
   - `[0]` - 退出

### 方法二：命令行操作

```powershell
# 进入scripts目录
cd scripts

# 启动所有服务
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" start

# 测试服务状态
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" test

# 查看服务状态
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" status

# 停止所有服务
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" stop
```

## 🔧 系统架构

### 服务组件

| 组件 | 技术栈 | 端口 | 状态 |
|------|--------|------|------|
| **后端服务** | Go + Gin | 8000 | ✅ 运行中 |
| **前端服务** | Python + HTTP Server | 8080 | ✅ 运行中 |
| **AI服务** | Python + FastAPI | 5001 | ⏸️ 待启动 |
| **数据库** | PostgreSQL | 5432 | ⏸️ 待启动 |
| **缓存** | Redis | 6379 | ⏸️ 待启动 |

### 访问地址

- **前端界面**: http://localhost:8080
- **后端API**: http://localhost:8000
- **健康检查**: http://localhost:8000/health
- **API状态**: http://localhost:8000/api/v1/status

## 📁 脚本文件说明

### 主要脚本

- **`manage_system.bat`** - 图形化菜单界面
- **`simple_manage.ps1`** - PowerShell核心脚本
- **`start_backend.bat`** - 快速启动后端
- **`start_frontend.bat`** - 快速启动前端

### 功能特性

✅ **自动端口管理** - 智能检测和释放端口冲突  
✅ **服务状态监控** - 实时检查服务运行状态  
✅ **API功能测试** - 自动测试接口响应  
✅ **错误处理** - 完善的错误提示和恢复机制  
✅ **彩色输出** - 友好的用户界面  
✅ **中文支持** - 完整的中文提示信息  

## 🎯 使用流程

### 首次使用

1. **清理环境**
   ```bash
   # 选择 [6] 清理端口占用
   ```

2. **启动服务**
   ```bash
   # 选择 [1] 启动所有服务
   ```

3. **验证服务**
   ```bash
   # 选择 [4] 测试所有服务
   ```

4. **访问系统**
   ```bash
   # 选择 [7] 打开前端界面
   ```

### 日常使用

- **启动系统**: 选择 `[1]` 启动所有服务
- **停止系统**: 选择 `[2]` 停止所有服务
- **重启系统**: 选择 `[3]` 重启所有服务
- **检查状态**: 选择 `[5]` 查看服务状态

## 🛠️ 故障排除

### 常见问题

#### 1. 端口被占用
```bash
# 解决方案：清理端口占用
选择 [6] 清理端口占用
```

#### 2. 服务启动失败
```bash
# 检查步骤：
1. 确认Go和Python已安装
2. 检查相关文件是否存在
3. 查看错误日志信息
```

#### 3. 权限问题
```powershell
# 设置PowerShell执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 4. 网络连接问题
```bash
# 检查防火墙设置
# 确保端口未被其他程序占用
```

### 调试模式

```powershell
# 查看详细日志
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" status

# 单独测试服务
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" test
```

## 📊 服务监控

### 健康检查

- **后端健康**: http://localhost:8000/health
- **API状态**: http://localhost:8000/api/v1/status
- **前端响应**: http://localhost:8080

### 监控指标

- ✅ 服务运行状态
- ✅ 端口占用情况
- ✅ API响应时间
- ✅ 错误日志记录

## 🔄 自动化部署

### 一键部署脚本

```bash
# 完整部署流程
1. 清理环境 (clean)
2. 启动服务 (start)
3. 等待启动 (sleep 10s)
4. 测试服务 (test)
5. 打开界面 (open)
```

### 定时任务

```powershell
# 创建定时重启任务
powershell -ExecutionPolicy Bypass -File "simple_manage.ps1" restart
```

## 📝 开发说明

### 脚本结构

```
scripts/
├── manage_system.bat      # 图形化菜单
├── simple_manage.ps1      # 核心管理脚本
├── start_backend.bat      # 后端启动脚本
├── start_frontend.bat     # 前端启动脚本
├── README.md             # 使用说明
└── USAGE_GUIDE.md        # 详细指南
```

### 扩展功能

- 🔧 支持自定义端口配置
- 🔧 支持服务依赖管理
- 🔧 支持日志记录功能
- 🔧 支持性能监控

## 🎉 成功案例

### 测试结果

```
================================================================
VideoCall System - Management Script
================================================================
Project Root: D:\c++\yspth
================================================================
Testing all services...
  Testing backend service...
    SUCCESS: Backend service OK - VideoCall Backend is running
  Testing frontend service...
    SUCCESS: Frontend service OK
  Testing API functionality...
    SUCCESS: API status OK - running
  Testing completed!
```

### 系统状态

- ✅ 后端服务正常运行 (端口: 8000)
- ✅ 前端服务正常运行 (端口: 8080)
- ✅ API接口响应正常
- ✅ 用户认证功能正常
- ✅ 通话管理功能正常

## 📞 技术支持

### 联系方式

- **项目地址**: 本地项目目录
- **文档位置**: `scripts/README.md`
- **日志位置**: 控制台输出

### 问题反馈

如遇到问题，请提供以下信息：
1. 操作系统版本
2. 错误日志信息
3. 服务状态输出
4. 网络连接情况

---

**🎯 祝您使用愉快！** 