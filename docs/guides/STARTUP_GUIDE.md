# VideoCall System - 启动脚本使用指南

## 🎉 启动脚本修复完成

### 问题解决 ✅
- **编码问题**: 移除了特殊中文字符和Unicode符号
- **语法错误**: 修复了批处理文件调用方式
- **路径问题**: 确保正确的文件路径和调用

## 可用的启动脚本

### 1. 快速启动（推荐） 🚀
```bash
.\start_system_simple.bat
```
**特点:**
- 3步快速启动
- 不包含自动测试
- 适合开发和日常使用
- 启动时间短

### 2. 完整启动（包含测试） 📊
```bash
.\start_system.bat
```
**特点:**
- 6步完整启动流程
- 包含环境检查
- 自动运行系统测试
- 详细状态报告

### 3. 系统管理菜单 🛠️
```bash
.\manage_system.bat
```
**特点:**
- 交互式菜单界面
- 9个管理选项
- 灵活的服务控制
- 状态检查和测试

## 启动流程说明

### 快速启动流程
```
[1/3] Starting database services...
[2/3] Starting backend service...
[3/3] Starting AI service...
System startup completed!
```

### 完整启动流程
```
[1/6] Checking Python environment...
[2/6] Checking Docker environment...
[3/6] Starting database services...
[4/6] Waiting for database services to be ready...
[5/6] Starting backend service...
[6/6] Starting AI service...
Running system tests...
System startup completed!
```

## 服务状态

### 启动后的服务
| 服务 | 端口 | 状态 | 访问地址 |
|------|------|------|----------|
| 后端服务 | 8000 | ✅ 运行中 | http://localhost:8000 |
| AI服务 | 5001 | ✅ 运行中 | http://localhost:5001 |
| PostgreSQL | 5432 | ✅ 运行中 | localhost:5432 |
| Redis | 6379 | ✅ 运行中 | localhost:6379 |

### 测试结果
```
============================================================
 测试结果统计
============================================================
✅ 后端健康检查: 服务正常
✅ 后端根端点: 服务信息正常
✅ AI服务健康检查: 服务正常
✅ AI服务根端点: 服务信息正常
✅ 用户注册: 用户已存在（预期结果）
✅ 用户登录: 登录成功
✅ 受保护端点: 访问成功
✅ AI检测服务: 检测成功 (风险评分=0.15, 置信度=0.85)

总计: 8/8 测试通过
🎉 所有测试通过！系统运行正常。
```

## 使用方法

### 首次使用
1. **确保环境准备**:
   - Python 3.8+ 已安装
   - Docker Desktop 已安装并运行
   - Go 1.19+ 已安装

2. **选择启动方式**:
   ```bash
   # 推荐：快速启动
   .\start_system_simple.bat
   
   # 或：完整启动（包含测试）
   .\start_system.bat
   ```

3. **验证系统状态**:
   ```bash
   # 运行完整测试
   python scripts/run_all_tests.py
   
   # 或快速测试
   python test_api.py
   ```

### 日常使用
```bash
# 快速启动所有服务
.\start_system_simple.bat

# 使用管理菜单
.\manage_system.bat

# 单独测试
python scripts/run_all_tests.py
```

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -an | findstr :8000
   netstat -an | findstr :5001
   ```

2. **服务未启动**
   ```bash
   # 检查Docker容器
   docker ps
   
   # 检查服务日志
   docker-compose logs postgres
   docker-compose logs redis
   ```

3. **编码问题**
   - 所有脚本已修复编码问题
   - 使用ASCII字符替代特殊符号
   - 支持中文路径

### 调试命令
```bash
# 检查数据库状态
python check_database.py

# 检查Docker状态
python check_docker.py

# 快速状态检查
.\status.bat
```

## 脚本特性

### ✅ 已修复的问题
- **编码问题**: 移除特殊字符，使用ASCII
- **语法错误**: 修复批处理文件调用
- **路径问题**: 确保正确的文件路径
- **错误处理**: 完善的错误检查和提示

### ✅ 功能特性
- **一键启动**: 简化的启动流程
- **环境检查**: 自动检查依赖环境
- **状态监控**: 实时服务状态检查
- **错误恢复**: 优雅的错误处理
- **用户友好**: 清晰的提示信息

## 管理命令

### 服务管理
```bash
# 启动所有服务
.\manage_system.bat  # 选择选项1

# 启动数据库
.\manage_system.bat  # 选择选项2

# 启动后端
.\manage_system.bat  # 选择选项3

# 启动AI服务
.\manage_system.bat  # 选择选项4
```

### 测试管理
```bash
# 运行完整测试
.\manage_system.bat  # 选择选项5

# 运行快速测试
.\manage_system.bat  # 选择选项6

# 检查数据库状态
.\manage_system.bat  # 选择选项7

# 检查Docker状态
.\manage_system.bat  # 选择选项8
```

### 停止服务
```bash
# 停止所有服务
.\manage_system.bat  # 选择选项9
```

## 总结

✅ **启动脚本完全可用** - 所有编码和语法问题已修复
✅ **一键启动功能** - 提供快速和完整两种启动方式
✅ **系统管理工具** - 交互式管理菜单
✅ **测试覆盖完整** - 8个核心功能测试
✅ **错误处理完善** - 优雅的错误处理和恢复

**推荐使用**: `.\start_system_simple.bat` 进行日常快速启动 