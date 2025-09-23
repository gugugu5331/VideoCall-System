# VideoCall System - 项目代码整理方案

## 🎯 整理目标

### 主要目标
1. **优化目录结构** - 清晰的模块化组织
2. **分类管理文件** - 按功能和类型分组
3. **清理冗余文件** - 移除过时和重复文件
4. **统一命名规范** - 一致的命名约定
5. **完善文档体系** - 系统化的文档管理

## 📁 建议的目录结构

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
├── 📁 docker/                 # Docker配置
├── 📁 config/                 # 配置文件
└── 📁 temp/                   # 临时文件
```

## 🔄 整理计划

### 阶段1: 目录重组
1. **创建新的目录结构**
2. **移动核心服务文件**
3. **整理脚本文件**
4. **组织文档文件**

### 阶段2: 文件分类
1. **启动脚本分类**
2. **管理脚本分类**
3. **测试脚本分类**
4. **文档分类**

### 阶段3: 清理优化
1. **移除冗余文件**
2. **统一命名规范**
3. **更新引用路径**
4. **完善文档**

## 📋 文件分类清单

### 核心服务文件
```
core/backend/           # 后端服务
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── env.example
├── config/
├── database/
├── handlers/
├── middleware/
├── models/
├── routes/
└── auth/

core/ai-service/        # AI服务
├── main.py
├── main-simple.py
├── requirements.txt
├── Dockerfile
└── app/

core/database/          # 数据库
├── init.sql
└── migrations/
```

### 脚本文件
```
scripts/startup/        # 启动脚本
├── start_system.bat
├── start_system_simple.bat
├── start-full.bat
├── start_ai_manual.bat
└── start.sh

scripts/management/     # 管理脚本
├── manage_system.bat
├── stop_services_simple.bat
├── stop_all_services.bat
├── release_ports.bat
├── release_ports.py
└── manage_db.bat

scripts/testing/        # 测试脚本
├── run_all_tests.py
├── test_api.py
├── test_backend.py
├── test_ai_simple.py
├── check_database.py
└── check_docker.py

scripts/utilities/      # 工具脚本
├── status.bat
├── start_ai_debug.py
└── check-status.ps1
```

### 文档文件
```
docs/guides/           # 使用指南
├── STARTUP_GUIDE.md
├── SERVICE_MANAGEMENT.md
├── LOCAL_DEVELOPMENT.md
└── PROJECT_ORGANIZATION.md

docs/api/              # API文档
├── backend_api.md
├── ai_service_api.md
└── database_schema.md

docs/status/           # 状态文档
├── SYSTEM_STATUS.md
├── BACKEND_STATUS.md
├── AI_SERVICE_STATUS.md
├── DATABASE_STATUS.md
└── FINAL_STATUS.md
```

### 配置文件
```
config/
├── docker-compose.yml
├── docker-compose-local.yml
├── docker.env
└── nginx.conf

docker/
└── nginx/
```

## 🗑️ 需要清理的文件

### 过时文件
- `start-dev.bat` → 合并到管理脚本
- `start-backend.bat` → 合并到启动脚本
- `start-simple.bat` → 合并到启动脚本
- `fix-docker.bat` → 功能已集成
- `start_ai_service.bat` → 重复文件
- `start_ai_debug.py` → 移动到工具脚本

### 重复文件
- `test-api.ps1` → 使用Python测试脚本
- `test-api-en.ps1` → 使用Python测试脚本
- `check-status.ps1` → 使用Python测试脚本
- `项目状态.md` → 合并到状态文档

### 临时文件
- `Proxies` → 删除
- `*.exe` → 移动到temp目录
- `__pycache__/` → 删除

## 📝 命名规范

### 文件命名
- **启动脚本**: `start_*.bat`
- **停止脚本**: `stop_*.bat`
- **管理脚本**: `manage_*.bat`
- **测试脚本**: `test_*.py`
- **检查脚本**: `check_*.py`
- **工具脚本**: `*_utility.py`

### 目录命名
- **核心服务**: `core/`
- **脚本工具**: `scripts/`
- **文档**: `docs/`
- **配置**: `config/`
- **临时文件**: `temp/`

## 🔧 实施步骤

### 步骤1: 创建新目录结构
```bash
mkdir core
mkdir core\backend
mkdir core\ai-service
mkdir core\database
mkdir scripts\startup
mkdir scripts\management
mkdir scripts\testing
mkdir scripts\utilities
mkdir docs\guides
mkdir docs\api
mkdir docs\status
mkdir config
mkdir temp
```

### 步骤2: 移动核心文件
```bash
# 移动后端文件
move backend\* core\backend\

# 移动AI服务文件
move ai-service\* core\ai-service\

# 移动数据库文件
move database\* core\database\
```

### 步骤3: 整理脚本文件
```bash
# 移动启动脚本
move start_*.bat scripts\startup\
move start.sh scripts\startup\

# 移动管理脚本
move manage_*.bat scripts\management\
move stop_*.bat scripts\management\
move release_*.bat scripts\management\
move release_ports.py scripts\management\

# 移动测试脚本
move test_*.py scripts\testing\
move check_*.py scripts\testing\
move run_all_tests.py scripts\testing\

# 移动工具脚本
move *_debug.py scripts\utilities\
move status.bat scripts\utilities\
```

### 步骤4: 整理文档文件
```bash
# 移动指南文档
move *GUIDE.md docs\guides\
move *MANAGEMENT.md docs\guides\
move LOCAL_DEVELOPMENT.md docs\guides\

# 移动状态文档
move *STATUS.md docs\status\

# 移动API文档
move *API.md docs\api\
```

### 步骤5: 整理配置文件
```bash
# 移动配置文件
move docker-compose*.yml config\
move docker.env config\
move docker\ config\
```

### 步骤6: 清理临时文件
```bash
# 移动可执行文件
move *.exe temp\

# 删除缓存文件
rmdir /s __pycache__

# 删除过时文件
del Proxies
del fix-docker.bat
del start-dev.bat
del start-backend.bat
del start-simple.bat
```

## 📊 整理效果

### 整理前
- 文件分散在根目录
- 命名不规范
- 功能重复
- 文档混乱

### 整理后
- 清晰的目录结构
- 统一的命名规范
- 功能模块化
- 文档系统化

## 🎯 后续维护

### 定期清理
- 每月清理临时文件
- 每季度更新文档
- 每年重构代码

### 版本控制
- 使用Git管理代码
- 创建版本标签
- 维护更新日志

### 自动化
- 创建自动化脚本
- 设置CI/CD流程
- 自动化测试

---

**总结**: 通过系统性的项目整理，将大大提高代码的可维护性、可读性和可扩展性，为后续开发奠定良好基础。 