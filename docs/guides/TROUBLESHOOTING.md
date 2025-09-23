# VideoCall System - 问题解决指南

## 🚨 常见问题及解决方案

### 1. Go编译问题

#### 问题描述
```
go: golang.org/x/sync@v0.5.0: missing go.sum entry for go.mod file
Compilation failed!
```

#### 解决方案
1. **安装Go语言环境**
   ```bash
   # 下载Go: https://golang.org/dl/
   # 推荐版本: Go 1.21或更高版本
   ```

2. **检查Go环境**
   ```bash
   # 运行Go环境检查脚本
   .\scripts\utilities\check_go.bat
   ```

3. **使用基础版本**
   ```bash
   # 使用基础后端启动脚本
   .\core\backend\start-basic.bat
   ```

### 2. Python文件路径问题

#### 问题描述
```
python: can't open file 'D:\c++\音视频\testing\run_all_tests.py': [Errno 2] No such file or directory
python: can't open file 'D:\c++\音视频\release_ports.py': [Errno 2] No such file or directory
```

#### 解决方案
1. **运行路径验证脚本**
   ```bash
   .\scripts\utilities\verify_paths.bat
   ```

2. **已修复** - 所有脚本中的路径问题已修复
3. **使用正确的脚本**
   ```bash
   # 使用管理菜单
   .\quick_manage.bat
   
   # 或直接运行端口释放脚本
   .\scripts\management\release_ports.bat
   
   # 或运行测试脚本
   .\quick_test.bat
   ```

4. **手动检查文件位置**
   - 测试脚本: `scripts\testing\`
   - 后端脚本: `core\backend\`
   - AI服务脚本: `core\ai-service\`
   - 配置文件: `config\`

### 3. 服务启动失败

#### 后端服务启动失败
1. **检查Go环境**
   ```bash
   .\scripts\utilities\check_go.bat
   ```

2. **使用基础版本**
   ```bash
   .\core\backend\start-basic.bat
   ```

3. **检查依赖**
   ```bash
   cd core/backend
   go mod tidy
   go mod download
   ```

#### AI服务启动失败
1. **检查Python环境**
   ```bash
   python --version
   ```

2. **安装依赖**
   ```bash
   cd core/ai-service
   pip install -r requirements.txt
   ```

3. **检查端口占用**
   ```bash
   .\scripts\management\release_ports.py 5001
   ```

### 4. 数据库连接问题

#### PostgreSQL连接失败
1. **检查Docker服务**
   ```bash
   docker ps
   ```

2. **重启数据库服务**
   ```bash
   docker-compose --project-name videocall-system -f config/docker-compose.yml restart postgres
   ```

3. **检查端口占用**
   ```bash
   .\scripts\management\release_ports.py 5432
   ```

#### Redis连接失败
1. **检查Redis服务**
   ```bash
   docker ps | findstr redis
   ```

2. **重启Redis服务**
   ```bash
   docker-compose --project-name videocall-system -f config/docker-compose.yml restart redis
   ```

### 5. 端口占用问题

#### 释放所有端口
```bash
.\scripts\management\release_ports.bat
```

#### 释放特定端口
```bash
# 释放8000端口
.\scripts\management\release_ports.bat 8000

# 释放5001端口
.\scripts\management\release_ports.bat 5001
```

### 6. 编码问题

#### 中文乱码
1. **使用UTF-8编码**
   ```bash
   # 所有批处理脚本已添加
   chcp 65001 >nul
   ```

2. **使用英文菜单**
   - 管理菜单已改为英文
   - 避免中文字符编码问题

### 7. 依赖安装问题

#### Python依赖
```bash
# 安装基础依赖
pip install requests aiohttp asyncio-throttle

# 安装AI服务依赖
cd core/ai-service
pip install -r requirements.txt
```

#### Go依赖
```bash
# 下载Go依赖
cd core/backend
go mod download
go mod tidy
```

## 🔧 系统诊断

### 1. 环境检查脚本
```bash
# 检查Go环境
.\scripts\utilities\check_go.bat

# 检查系统状态
python scripts\testing\run_all_tests.py
```

### 2. 服务状态检查
```bash
# 检查所有服务
.\quick_test.bat

# 检查数据库
python scripts\testing\check_database.py

# 检查Docker
python scripts\testing\check_docker.py
```

### 3. 日志查看
- **后端日志**: 查看后端服务控制台输出
- **AI服务日志**: 查看AI服务控制台输出
- **Docker日志**: `docker logs videocall_postgres`

## 📋 启动顺序

### 推荐启动顺序
1. **检查环境**
   ```bash
   .\scripts\utilities\check_go.bat
   ```

2. **启动数据库**
   ```bash
   docker-compose --project-name videocall-system -f config/docker-compose.yml up -d postgres redis
   ```

3. **启动后端服务**
   ```bash
   .\core\backend\start-basic.bat
   ```

4. **启动AI服务**
   ```bash
   .\core\ai-service\start_ai_manual.bat
   ```

5. **测试系统**
   ```bash
   python scripts\testing\run_all_tests.py
   ```

### 一键启动
```bash
# 使用简化启动脚本
.\quick_start.bat

# 或使用管理菜单
.\quick_manage.bat
```

## 🆘 紧急恢复

### 完全重置系统
1. **停止所有服务**
   ```bash
   .\scripts\management\stop_all_services.bat
   ```

2. **释放所有端口**
   ```bash
   .\scripts\management\release_ports.bat
   ```

3. **清理Docker**
   ```bash
   docker-compose --project-name videocall-system -f config/docker-compose.yml down -v
   ```

4. **重新启动**
   ```bash
   .\quick_start.bat
   ```

## 📞 获取帮助

### 检查清单
- [ ] Go语言已安装 (1.21+)
- [ ] Python已安装 (3.8+)
- [ ] Docker已安装并运行
- [ ] 端口8000, 5001, 5432, 6379未被占用
- [ ] 所有依赖已安装

### 日志文件位置
- **项目根目录**: 查看各种日志文件
- **Docker日志**: `docker logs <container_name>`
- **服务控制台**: 查看启动脚本的输出

### 联系支持
如果问题仍然存在，请提供：
1. 错误信息截图
2. 系统环境信息
3. 已尝试的解决方案
4. 日志文件内容 