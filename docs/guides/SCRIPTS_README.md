# VideoCall System - 测试脚本说明

## 概述

本目录包含VideoCall系统的所有测试脚本和管理工具。

## 主要脚本

### 1. 统一测试脚本
- **`run_all_tests.py`** - 完整的系统测试套件
  - 测试所有服务组件
  - 提供详细的测试报告
  - 支持错误处理和结果统计

### 2. 启动脚本
- **`start_system.bat`** - 一键式完整启动（包含测试）
- **`start_system_simple.bat`** - 快速启动（不包含测试）
- **`manage_system.bat`** - 系统管理菜单

## 测试脚本分类

### 系统级测试
- **`run_all_tests.py`** - 完整系统测试
- **`test_api.py`** - API集成测试
- **`test_backend.py`** - 后端服务测试
- **`test_ai_simple.py`** - AI服务测试

### 组件测试
- **`check_database.py`** - 数据库连接测试
- **`check_docker.py`** - Docker容器状态测试

### 状态检查
- **`status.bat`** - 快速状态检查
- **`manage_db.bat`** - 数据库管理

## 使用方法

### 快速启动
```bash
# 一键启动所有服务（包含测试）
.\start_system.bat

# 快速启动（不包含测试）
.\start_system_simple.bat

# 系统管理菜单
.\manage_system.bat
```

### 运行测试
```bash
# 完整系统测试
python scripts/run_all_tests.py

# 快速API测试
python test_api.py

# 后端服务测试
python test_backend.py

# AI服务测试
python test_ai_simple.py
```

### 状态检查
```bash
# 数据库状态检查
python check_database.py

# Docker状态检查
python check_docker.py

# 快速状态检查
.\status.bat
```

## 测试结果说明

### 测试项目
1. **后端健康检查** - 验证后端服务状态
2. **后端根端点** - 验证服务信息
3. **AI服务健康检查** - 验证AI服务状态
4. **AI服务根端点** - 验证AI服务信息
5. **用户注册** - 测试用户注册功能
6. **用户登录** - 测试用户登录功能
7. **受保护端点** - 测试JWT认证
8. **AI检测服务** - 测试伪造检测功能

### 结果格式
- ✅ 测试通过
- ❌ 测试失败
- 📊 详细统计信息

## 服务端口

- **后端服务**: http://localhost:8000
- **AI服务**: http://localhost:5001
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 故障排除

### 常见问题
1. **端口冲突** - 检查端口是否被占用
2. **服务未启动** - 确认Docker和Python环境
3. **连接超时** - 增加超时时间或检查网络

### 调试命令
```bash
# 检查端口占用
netstat -an | findstr :8000
netstat -an | findstr :5001

# 检查Docker容器
docker ps

# 检查服务日志
docker-compose logs postgres
docker-compose logs redis
```

## 开发说明

### 添加新测试
1. 在`run_all_tests.py`中添加新的测试函数
2. 遵循`TestResult`类的格式
3. 在`run_all_tests()`函数中调用新测试

### 修改配置
- 修改脚本顶部的URL配置
- 调整超时时间设置
- 更新测试数据

## 版本信息

- **版本**: 1.0.0
- **更新日期**: 2025-07-27
- **兼容性**: Windows 10/11, Python 3.8+ 