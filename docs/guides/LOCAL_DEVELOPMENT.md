# 本地开发指南

由于Docker网络问题，您可以选择在本地直接运行服务。

## 环境要求

### 1. 安装必要软件

#### Go 1.21+
```bash
# 下载并安装Go
# https://golang.org/dl/
```

#### Python 3.9+
```bash
# 下载并安装Python
# https://www.python.org/downloads/
```

#### PostgreSQL 15+
```bash
# Windows: 下载PostgreSQL安装包
# https://www.postgresql.org/download/windows/

# 或者使用包管理器
# chocolatey: choco install postgresql
```

#### Redis 7+
```bash
# Windows: 下载Redis for Windows
# https://github.com/microsoftarchive/redis/releases

# 或者使用包管理器
# chocolatey: choco install redis-64
```

## 启动步骤

### 1. 启动数据库

#### PostgreSQL
```bash
# 启动PostgreSQL服务
# Windows: 在服务管理器中启动PostgreSQL服务

# 创建数据库
psql -U postgres
CREATE DATABASE videocall;
CREATE USER admin WITH PASSWORD 'videocall123';
GRANT ALL PRIVILEGES ON DATABASE videocall TO admin;
\q

# 初始化数据库
psql -U admin -d videocall -f database/init.sql
```

#### Redis
```bash
# 启动Redis服务
redis-server

# 或者Windows服务
# 在服务管理器中启动Redis服务
```

### 2. 启动后端服务

```bash
cd backend

# 安装依赖
go mod tidy

# 设置环境变量
set DB_HOST=localhost
set DB_PORT=5432
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=videocall
set DB_USER=admin
set DB_PASSWORD=videocall123
set REDIS_HOST=localhost
set REDIS_PORT=6379

# 启动服务
go run main.go
```

### 3. 启动AI服务

```bash
cd ai-service

# 安装依赖
pip install -r requirements.txt

# 设置环境变量
set REDIS_HOST=localhost
set REDIS_PORT=6379

# 启动服务
python main.py
```

## 验证服务

### 1. 检查后端服务
```bash
curl http://localhost:8000/health
```

### 2. 检查AI服务
```bash
curl http://localhost:5000/health
```

### 3. 运行API测试
```bash
python test_api.py
```

## 访问地址

- 后端API: http://localhost:8000
- AI服务: http://localhost:5000
- API文档: http://localhost:8000/swagger/index.html

## 故障排除

### 1. 数据库连接问题
- 确保PostgreSQL服务正在运行
- 检查端口5432是否被占用
- 验证用户名和密码

### 2. Redis连接问题
- 确保Redis服务正在运行
- 检查端口6379是否被占用

### 3. 端口冲突
- 检查端口8000和5000是否被占用
- 修改配置文件中的端口设置

## 开发建议

1. 使用IDE（如VS Code、GoLand、PyCharm）
2. 启用热重载功能
3. 使用数据库管理工具（如pgAdmin）
4. 使用Redis客户端工具