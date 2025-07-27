# 音视频通话系统

一个基于深度学习的音视频通话系统，具备实时伪造检测功能。

## 项目架构

```
音视频通话系统/
├── frontend/          # Qt C++ 前端
├── backend/           # Golang 后端
├── ai-service/        # Python AI服务
├── database/          # 数据库脚本
├── docker/            # Docker配置
└── docs/              # 文档
```

## 技术栈

- **前端**: Qt 6.x, C++, QML
- **后端**: Golang, Gin, GORM
- **AI服务**: Python, PyTorch, FastAPI
- **数据库**: PostgreSQL, Redis
- **部署**: Docker, Docker Compose

## 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.19+
- Python 3.8+
- Qt 6.x (前端开发)

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 开发环境

```bash
# 后端开发
cd backend
go mod tidy
go run main.go

# AI服务开发
cd ai-service
pip install -r requirements.txt
python main.py

# 前端开发
cd frontend
# 使用Qt Creator打开项目
```

## API文档

启动服务后访问: http://localhost:8000/swagger/index.html

## 许可证

MIT License 