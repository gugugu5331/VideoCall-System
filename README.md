# VideoCall System

## 项目概述

基于深度学习的音视频通话系统，包含伪造检测功能。

## 目录结构

```
videocall-system/
├── 📁 core/                    # 核心服务
│   ├── 📁 backend/            # Golang后端服务
│   ├── 📁 ai-service/         # Python AI服务
│   └── 📁 database/           # 数据库相关
├── 📁 scripts/                # 脚本工具
│   ├── 📁 startup/           # 启动脚本
│   ├── 📁 management/        # 管理脚本
│   ├── 📁 testing/           # 测试脚本
│   └── 📁 utilities/         # 工具脚本
├── 📁 docs/                   # 文档
│   ├── 📁 guides/            # 使用指南
│   ├── 📁 api/               # API文档
│   └── 📁 status/            # 状态文档
├── 📁 config/                 # 配置文件
└── 📁 temp/                   # 临时文件
```

## 快速开始

### 启动系统
```bash
# 快速启动
scripts/startup/start_system_simple.bat

# 完整启动（包含测试）
scripts/startup/start_system.bat
```

### 管理服务
```bash
# 系统管理菜单
scripts/management/manage_system.bat

# 停止所有服务
scripts/management/stop_services_simple.bat
```

### 运行测试
```bash
# 完整测试
scripts/testing/run_all_tests.py

# 快速测试
scripts/testing/test_api.py
```

## 文档

- [启动指南](docs/guides/STARTUP_GUIDE.md)
- [服务管理](docs/guides/SERVICE_MANAGEMENT.md)
- [本地开发](docs/guides/LOCAL_DEVELOPMENT.md)
- [项目组织](docs/guides/PROJECT_ORGANIZATION.md)

## 技术栈

- **后端**: Golang + Gin + GORM
- **AI服务**: Python + FastAPI + PyTorch
- **数据库**: PostgreSQL + Redis
- **前端**: Qt C++ (计划中)
- **部署**: Docker + Docker Compose

## 开发状态

✅ 后端服务 - 完成
✅ AI服务 - 完成  
✅ 数据库 - 完成
✅ 启动脚本 - 完成
✅ 管理脚本 - 完成
🔄 前端界面 - 开发中
🔄 深度学习模型 - 开发中

## 许可证

MIT License
