# VideoCall System - 服务管理功能指南

## 🎯 新增功能

### ✅ 服务停止功能
- **一键停止所有服务**
- **智能端口释放**
- **进程清理**
- **状态检查**

## 📋 可用的管理脚本

### 1. 停止所有服务（推荐）
```bash
.\stop_services_simple.bat
```
**功能:**
- 停止Docker容器
- 终止Python进程（AI服务）
- 终止Go进程（后端服务）
- 检查端口状态
- 提供清理建议

### 2. 智能端口释放
```bash
.\release_ports.bat
```
**功能:**
- 检查所有默认端口（8000, 5001, 5432, 6379）
- 显示占用进程详细信息
- 交互式确认终止进程
- 支持强制模式

### 3. Python端口释放工具
```bash
python release_ports.py [options] [port1] [port2] ...
```
**选项:**
- `--force, -f`: 强制终止进程，不询问
- `--help`: 显示帮助信息

**示例:**
```bash
# 检查所有默认端口
python release_ports.py

# 强制释放所有端口
python release_ports.py --force

# 检查特定端口
python release_ports.py 8000 5001

# 强制释放特定端口
python release_ports.py -f 8000
```

### 4. 系统管理菜单
```bash
.\manage_system.bat
```
**新增选项:**
- 选项10: 释放所有端口
- 选项11: 释放指定端口

## 🔧 使用场景

### 场景1: 正常停止服务
```bash
# 使用简化脚本
.\stop_services_simple.bat
```

### 场景2: 端口被占用
```bash
# 检查端口状态
python release_ports.py

# 或使用批处理脚本
.\release_ports.bat
```

### 场景3: 强制清理
```bash
# 强制释放所有端口
python release_ports.py --force

# 或使用批处理脚本
.\release_ports.bat --force
```

### 场景4: 特定端口问题
```bash
# 释放特定端口
python release_ports.py 8000

# 或通过管理菜单
.\manage_system.bat  # 选择选项11
```

## 📊 端口状态检查

### 默认检查的端口
| 端口 | 服务 | 说明 |
|------|------|------|
| 8000 | 后端服务 | Golang API服务 |
| 5001 | AI服务 | Python FastAPI服务 |
| 5432 | PostgreSQL | 数据库服务 |
| 6379 | Redis | 缓存服务 |

### 端口状态显示
```
============================================================
Summary:
============================================================
Backend Service (Port 8000): ✅ FREE
AI Service (Port 5001): ✅ FREE
PostgreSQL (Port 5432): ✅ FREE
Redis (Port 6379): ✅ FREE

🎉 All ports are now free!
```

## 🛠️ 进程管理

### 进程信息显示
```
⚠️  Found 1 process(es) using port 8000:
   PID: 12345
   Name: videocall.exe
   Command: D:\c++\音视频\backend\videocall.exe
   Status: running
   Created: 2025-07-27 07:15:30
```

### 进程终止选项
- **交互式**: 询问是否终止每个进程
- **强制模式**: 自动终止所有占用进程
- **选择性**: 只终止特定端口的进程

## 🔍 故障排除

### 常见问题

1. **端口仍被占用**
   ```bash
   # 以管理员权限运行
   python release_ports.py --force
   ```

2. **进程无法终止**
   ```bash
   # 检查进程权限
   tasklist /fi "pid eq 12345"
   
   # 手动终止
   taskkill /f /pid 12345
   ```

3. **Docker容器未停止**
   ```bash
   # 强制停止Docker容器
   docker-compose --project-name videocall-system down --remove-orphans
   ```

### 调试命令
```bash
# 检查端口占用
netstat -ano | findstr :8000

# 检查进程状态
tasklist | findstr python
tasklist | findstr videocall

# 检查Docker状态
docker ps
docker-compose ps
```

## 📝 脚本特性

### ✅ 安全特性
- **权限检查**: 自动检测管理员权限
- **进程确认**: 交互式确认终止进程
- **错误处理**: 优雅的错误处理和恢复
- **状态验证**: 操作后验证结果

### ✅ 功能特性
- **智能检测**: 自动检测占用进程
- **详细信息**: 显示进程详细信息
- **灵活选项**: 支持多种操作模式
- **用户友好**: 清晰的提示和反馈

### ✅ 兼容性
- **Windows支持**: 专为Windows环境优化
- **编码支持**: 支持中文路径和显示
- **权限适配**: 适配不同权限级别

## 🎯 最佳实践

### 日常使用
1. **启动服务**: `.\start_system_simple.bat`
2. **停止服务**: `.\stop_services_simple.bat`
3. **检查状态**: `python release_ports.py`

### 故障处理
1. **端口冲突**: `python release_ports.py --force`
2. **进程残留**: `.\stop_services_simple.bat`
3. **完全清理**: 重启计算机

### 开发调试
1. **快速重启**: 使用管理菜单
2. **端口检查**: 定期检查端口状态
3. **进程监控**: 监控服务进程状态

## 📈 性能优化

### 启动优化
- 使用快速启动脚本
- 并行启动服务
- 智能等待机制

### 停止优化
- 优雅停止进程
- 并行清理操作
- 状态验证机制

## 🔄 工作流程

### 完整工作流程
```
1. 启动服务 → 2. 开发/测试 → 3. 停止服务 → 4. 清理端口
```

### 快速工作流程
```
1. 快速启动 → 2. 快速停止 → 3. 状态检查
```

## 📞 支持

### 获取帮助
```bash
# 查看帮助信息
python release_ports.py --help
.\release_ports.bat --help
```

### 报告问题
- 检查端口状态
- 查看进程信息
- 运行调试命令
- 提供错误日志

---

**总结**: 新增的服务管理功能提供了完整的服务生命周期管理，包括启动、停止、监控和清理，确保系统资源的有效管理和释放。 