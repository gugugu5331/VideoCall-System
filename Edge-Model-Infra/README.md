# Edge-Model-Infra

一个用于大语言模型（LLM）部署和管理的分布式边缘计算基础设施，专为边缘环境中的高效模型服务和通信而设计。

## 🚀 核心特性

- **分布式架构**：模块化设计，包含基础设施控制、单元管理和网络通信等独立组件
- **LLM 集成**：内置支持 LLM 模型部署和推理
- **高性能通信**：基于 ZeroMQ 的消息系统，实现低延迟组件间通信
- **Docker 支持**：容器化部署，预配置依赖项
- **TCP/JSON API**：类 RESTful API，便于与外部应用集成
- **事件驱动架构**：使用 eventpp 库进行异步事件处理
- **跨平台支持**：基于 Linux 的部署，具有完善的依赖管理

## 📁 项目结构

```
Edge-Model-Infra/
├── infra-controller/     # 基础设施控制和流程管理
├── unit-manager/         # 核心单元管理和协调
├── network/             # 网络通信层
├── hybrid-comm/         # 混合通信协议（ZMQ 封装）
├── node/               # 节点管理和 LLM 集成
├── sample/             # 示例实现和测试客户端
├── docker/             # Docker 配置和构建脚本
├── utils/              # 实用工具库和辅助函数
└── thirds/             # 第三方依赖
```

## 🛠️ 组件说明

### 基础设施控制器 (`infra-controller/`)
- **StackFlow**：事件驱动的工作流管理系统
- **通道管理**：通信通道抽象
- **流程控制**：请求/响应流程协调

### 单元管理器 (`unit-manager/`)
- **核心服务**：主要服务编排 (`main.cpp`)
- **远程操作**：RPC 风格的操作处理
- **会话管理**：客户端会话生命周期管理
- **ZMQ 总线**：消息总线实现
- **TCP 通信**：用于外部 API 的类 HTTP TCP 服务器

### 网络层 (`network/`)
- **事件循环**：高性能事件驱动网络
- **TCP 服务器/客户端**：健壮的 TCP 通信
- **连接管理**：连接池和生命周期管理
- **缓冲区管理**：高效的数据缓冲

### 混合通信 (`hybrid-comm/`)
- **pzmq**：增强功能的 ZeroMQ 封装
- **消息序列化**：高效的数据序列化/反序列化
- **协议抽象**：统一的通信接口

## 🔧 系统要求

- **操作系统**：Ubuntu 20.04 或兼容的 Linux 发行版
- **编译器**：支持 C++17 的 GCC/G++
- **CMake**：3.10 或更高版本
- **依赖项**：
  - libzmq3-dev
  - libgoogle-glog-dev
  - libboost-all-dev
  - libssl-dev
  - libbsd-dev
  - eventpp
  - simdjson

## 🚀 快速开始

### 1. 克隆仓库
```bash
git clone https://github.com/gugugu5331/Edge-Model-Infra.git
cd Edge-Model-Infra
```

### 2. 使用 Docker 构建（推荐）
```bash
# 构建 Docker 镜像
cd docker/scripts
./llm_docker_run.sh

# 进入容器
./llm_docker_into.sh
```

### 3. 手动构建
```bash
# 安装依赖
sudo ./build.sh

# 构建项目
mkdir build && cd build
cmake ..
make -j$(nproc)
```

### 4. 运行系统
```bash
# 启动单元管理器
cd unit-manager
./unit_manager

# 系统将在端口 10001 上启动（可在 master_config.json 中配置）
```

## 📖 使用示例

### Python 客户端示例
```python
import socket
import json

# 连接到服务
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('localhost', 10001))

# 设置 LLM
setup_data = {
    "request_id": "llm_001",
    "work_id": "llm",
    "action": "setup",
    "object": "llm.setup",
    "data": {
        "model": "DeepSeek-R1-Distill-Qwen-1.5B",
        "response_format": "llm.utf-8.stream",
        "max_token_len": 1023,
        "prompt": "你是一个有用的助手。"
    }
}

# 发送设置请求
sock.sendall((json.dumps(setup_data) + '\n').encode('utf-8'))
response = sock.recv(4096).decode('utf-8')
print("设置响应:", response)
```

### C++ RPC 示例
```cpp
#include "pzmq.hpp"
using namespace StackFlows;

// 创建 RPC 服务器
pzmq rpc_server("my_service");
rpc_server.register_rpc_action("process", [](pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
    return "已处理: " + msg->string();
});

// 创建 RPC 客户端
pzmq rpc_client("client");
auto result = rpc_client.rpc_call("my_service", "process", "Hello World");
```

## ⚙️ 配置说明

### 单元管理器配置 (`unit-manager/master_config.json`)
```json
{
    "config_tcp_server": 10001,
    "config_zmq_min_port": 5010,
    "config_zmq_max_port": 5555,
    "config_zmq_s_format": "ipc:///tmp/llm/%i.sock",
    "config_zmq_c_format": "ipc:///tmp/llm/%i.sock"
}
```

## 🧪 测试

运行包含的测试客户端：
```bash
cd sample
python3 test.py --host localhost --port 10001
```

运行 C++ 示例：
```bash
cd sample
# 终端 1：启动 RPC 服务器
./rpc_server

# 终端 2：运行 RPC 客户端
./rpc_call
```

## 🐳 Docker 部署

项目包含完整的 Docker 支持：

```bash
# 使用 Docker 构建和运行
cd docker/scripts
./llm_docker_run.sh    # 构建并启动容器
./llm_docker_into.sh   # 进入运行中的容器
```

## 📊 API 参考

### 设置 LLM 模型
```json
POST /
{
    "request_id": "唯一标识",
    "work_id": "llm",
    "action": "setup",
    "object": "llm.setup",
    "data": {
        "model": "模型名称",
        "response_format": "llm.utf-8.stream",
        "max_token_len": 1023
    }
}
```

### 推理请求
```json
POST /
{
    "request_id": "唯一标识",
    "work_id": "返回的工作ID",
    "action": "inference",
    "object": "llm.utf-8.stream",
    "data": {
        "delta": "用户输入",
        "index": 0,
        "finish": true
    }
}
```

## 🤝 贡献指南

1. Fork 本仓库
2. 创建您的功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。



## 🔗 相关链接

- [GitHub 仓库](https://github.com/gugugu5331/Edge-Model-Infra)
- [问题反馈](https://github.com/gugugu5331/Edge-Model-Infra/issues)
- [Pull Requests](https://github.com/gugugu5331/Edge-Model-Infra/pulls)

## 📞 技术支持

如需支持和咨询，请在 GitHub 上提交 issue 或联系开发团队。

---

**注意**：本项目正在积极开发中，某些功能可能是实验性的或可能会发生变化。
