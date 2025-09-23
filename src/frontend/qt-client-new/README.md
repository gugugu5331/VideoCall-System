# VideoCall System Qt Client

🚀 **功能完整的Qt6音视频会议客户端** - 集成WebRTC、AI检测、视频处理和Edge-Model-Infra

## ✨ 核心功能

### 🎥 **音视频会议**
- **WebRTC实时通信** - P2P音视频传输，低延迟高质量
- **多人会议支持** - 支持多达50人同时在线
- **屏幕共享** - 高清屏幕分享和远程协作
- **会议录制** - 本地录制和云端存储
- **聊天功能** - 实时文字聊天和文件传输

### 🤖 **AI智能检测**
- **换脸检测** - 实时检测Deepfake和换脸技术
- **语音合成检测** - 识别AI生成的语音内容
- **内容分析** - 智能分析会议内容和情绪
- **实时警报** - 检测到可疑内容时立即提醒
- **检测历史** - 完整的检测记录和统计分析

### 🎨 **视频处理**
- **实时滤镜** - 15+种专业滤镜效果
- **美颜功能** - 磨皮、美白、瘦脸、大眼
- **背景替换** - 智能背景分割和虚拟背景
- **贴纸特效** - 动态贴纸和3D模型
- **面部检测** - 68点面部关键点实时跟踪

### ⚡ **Edge-Model-Infra集成**
- **分布式推理** - C++高性能AI推理框架
- **模型管理** - 动态加载和卸载AI模型
- **任务调度** - 智能任务分配和负载均衡
- **性能监控** - 实时监控系统资源使用

## 🏗️ 技术架构

### **前端技术栈**
- **Qt6** - 现代C++跨平台UI框架
- **OpenCV** - 计算机视觉和图像处理
- **OpenGL** - 硬件加速图形渲染
- **WebRTC** - 实时音视频通信
- **ZeroMQ** - 高性能异步消息传递

### **后端集成**
- **Go Backend** - RESTful API和WebSocket服务
- **Python AI Services** - Flask AI检测服务
- **Edge-Model-Infra** - C++分布式推理框架
- **PostgreSQL + Redis** - 数据存储和缓存

### **通信协议**
- **WebSocket** - 信令服务器通信
- **HTTP/REST** - API服务调用
- **ZMQ** - Edge-Model-Infra通信
- **WebRTC** - P2P媒体传输

## 📦 快速开始

### **系统要求**
- **操作系统**: Windows 10+, macOS 10.15+, Ubuntu 18.04+
- **编译器**: GCC 9+, Clang 10+, MSVC 2019+
- **Qt版本**: Qt 6.2+
- **OpenCV**: 4.5+
- **CMake**: 3.16+

### **安装依赖**

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y \
    build-essential cmake pkg-config \
    qt6-base-dev qt6-multimedia-dev qt6-webengine-dev \
    qt6-websockets-dev qt6-charts-dev libqt6opengl6-dev \
    libopencv-dev libzmq3-dev libprotobuf-dev \
    libavcodec-dev libavformat-dev libgl1-mesa-dev
```

#### macOS
```bash
brew install cmake qt6 opencv zeromq protobuf ffmpeg
```

#### Windows
```powershell
# 使用vcpkg安装依赖
vcpkg install qt6 opencv zeromq protobuf ffmpeg
```

### **编译和运行**

#### 自动构建（推荐）
```bash
# 克隆项目
git clone https://github.com/gugugu5331/VideoCall-System.git
cd VideoCall-System/src/frontend/qt-client-new

# 给构建脚本执行权限
chmod +x build.sh

# 完整构建（安装依赖 + 编译 + 测试）
./build.sh --all

# 运行应用程序
./build-release/VideoCallSystemClient
```

#### 手动构建
```bash
# 创建构建目录
mkdir build && cd build

# 配置CMake
cmake -DCMAKE_BUILD_TYPE=Release ..

# 编译
cmake --build . --parallel $(nproc)

# 运行
./VideoCallSystemClient
```

### **Docker部署**
```bash
# 构建Docker镜像
docker build -t videocall-qt-client .

# 运行容器
docker run -it --rm \
    -e DISPLAY=$DISPLAY \
    -v /tmp/.X11-unix:/tmp/.X11-unix \
    videocall-qt-client
```

## 🎮 使用指南

### **基本操作**
1. **启动应用** - 运行可执行文件
2. **登录账户** - 输入用户名和服务器地址
3. **创建会议** - 点击"新建会议"按钮
4. **加入会议** - 输入会议ID或点击邀请链接
5. **开启摄像头** - 点击摄像头按钮
6. **开启麦克风** - 点击麦克风按钮

### **快捷键**
- `Ctrl+N` - 新建会议
- `Ctrl+J` - 加入会议
- `Ctrl+L` - 离开会议
- `Ctrl+M` - 静音/取消静音
- `Ctrl+D` - 开启/关闭摄像头
- `Ctrl+S` - 屏幕共享
- `Ctrl+R` - 开始/停止录制
- `F11` - 全屏模式
- `Esc` - 退出全屏

### **滤镜和特效**
1. **选择滤镜** - 在滤镜面板选择效果
2. **调整参数** - 使用滑块调整强度
3. **保存预设** - 保存常用滤镜组合
4. **背景替换** - 上传自定义背景图片
5. **贴纸特效** - 选择动态贴纸和3D模型

### **AI检测功能**
1. **启用检测** - 在设置中开启AI检测
2. **设置阈值** - 调整检测敏感度
3. **查看结果** - 在检测面板查看实时结果
4. **历史记录** - 查看检测历史和统计
5. **警报设置** - 配置检测警报规则

## ⚙️ 配置说明

### **服务器配置**
```json
{
  "server": {
    "host": "localhost",
    "port": 8080,
    "url": "http://localhost:8080"
  },
  "signaling": {
    "host": "localhost",
    "port": 8081,
    "url": "ws://localhost:8081"
  },
  "ai_service": {
    "host": "localhost",
    "port": 5000,
    "url": "http://localhost:5000"
  },
  "edge_infra": {
    "host": "localhost",
    "port": 9000,
    "url": "tcp://localhost:9000"
  }
}
```

### **媒体配置**
```json
{
  "video": {
    "width": 1280,
    "height": 720,
    "fps": 30,
    "bitrate": 1000000,
    "codec": "VP8"
  },
  "audio": {
    "sample_rate": 44100,
    "channels": 2,
    "bitrate": 128000,
    "codec": "OPUS"
  }
}
```

### **AI检测配置**
```json
{
  "detection": {
    "face_swap_enabled": true,
    "voice_synthesis_enabled": true,
    "content_analysis_enabled": true,
    "threshold": 0.7,
    "interval": 1000
  }
}
```

## 🔧 开发指南

### **项目结构**
```
src/frontend/qt-client-new/
├── include/                 # 头文件
│   ├── core/               # 核心组件
│   ├── ui/                 # UI组件
│   ├── network/            # 网络通信
│   ├── media/              # 媒体处理
│   └── data/               # 数据管理
├── src/                    # 源文件
├── ui/                     # UI文件
├── resources/              # 资源文件
├── shaders/                # OpenGL着色器
├── assets/                 # 静态资源
├── config/                 # 配置文件
└── tests/                  # 测试文件
```

### **添加新功能**
1. **创建头文件** - 在`include/`目录添加头文件
2. **实现源文件** - 在`src/`目录添加实现
3. **更新CMakeLists.txt** - 添加新文件到构建系统
4. **编写测试** - 在`tests/`目录添加单元测试
5. **更新文档** - 更新README和API文档

### **调试技巧**
- 使用`--debug`参数启动调试模式
- 查看日志文件：`~/.local/share/VideoCallSystem/logs/`
- 使用Qt Creator进行可视化调试
- 启用OpenCV调试输出：`export OPENCV_LOG_LEVEL=DEBUG`

## 📊 性能优化

### **系统要求**
- **CPU**: Intel i5-8400 / AMD Ryzen 5 2600 或更高
- **内存**: 8GB RAM（推荐16GB）
- **显卡**: 支持OpenGL 3.3+的独立显卡
- **网络**: 10Mbps上行带宽（多人会议）

### **优化建议**
1. **启用GPU加速** - 使用`--enable-gpu`参数
2. **调整视频质量** - 根据网络条件调整分辨率和码率
3. **关闭不必要功能** - 在低配置设备上关闭AI检测
4. **使用有线网络** - 避免WiFi带来的延迟和丢包
5. **关闭其他应用** - 释放CPU和内存资源

## 🐛 故障排除

### **常见问题**

#### 编译错误
```bash
# Qt6找不到
export CMAKE_PREFIX_PATH="/usr/lib/x86_64-linux-gnu/cmake/Qt6"

# OpenCV找不到
export OpenCV_DIR="/usr/lib/x86_64-linux-gnu/cmake/opencv4"

# 权限问题
sudo chown -R $USER:$USER build/
```

#### 运行时错误
```bash
# 缺少共享库
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"

# 显示问题
export QT_QPA_PLATFORM=xcb

# 音频问题
pulseaudio --start
```

#### 网络连接问题
- 检查防火墙设置
- 确认服务器地址和端口
- 测试网络连通性：`ping server_address`
- 检查代理设置

## 📈 路线图

### **v1.1 (计划中)**
- [ ] 移动端支持（Android/iOS）
- [ ] 更多AI检测模型
- [ ] 云端录制和存储
- [ ] 会议室管理功能
- [ ] 多语言界面支持

### **v1.2 (未来)**
- [ ] VR/AR支持
- [ ] 区块链身份验证
- [ ] 端到端加密
- [ ] 插件系统
- [ ] 企业级功能

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. **Fork项目** - 点击右上角Fork按钮
2. **创建分支** - `git checkout -b feature/new-feature`
3. **提交更改** - `git commit -am 'Add new feature'`
4. **推送分支** - `git push origin feature/new-feature`
5. **创建PR** - 在GitHub上创建Pull Request

### **代码规范**
- 使用C++17标准
- 遵循Qt编码规范
- 添加适当的注释和文档
- 编写单元测试
- 确保代码通过CI检查

## 📄 许可证

本项目采用MIT许可证 - 查看[LICENSE](LICENSE)文件了解详情。

## 🙏 致谢

- **Qt Project** - 优秀的跨平台框架
- **OpenCV** - 强大的计算机视觉库
- **WebRTC** - 实时通信技术
- **ZeroMQ** - 高性能消息传递
- **所有贡献者** - 感谢每一位贡献者的努力

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个Star！⭐**

[🏠 主页](https://github.com/gugugu5331/VideoCall-System) | 
[📖 文档](docs/) | 
[🐛 问题反馈](https://github.com/gugugu5331/VideoCall-System/issues) | 
[💬 讨论](https://github.com/gugugu5331/VideoCall-System/discussions)

</div>
