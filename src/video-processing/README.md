# 🎥 Video Processing - OpenCV + OpenGL

一个基于OpenCV和OpenGL的高性能实时视频处理应用，支持滤镜、渲染、贴图等功能。

## ✨ 主要功能

### 🎨 滤镜效果
- **基础滤镜**: 模糊、锐化、边缘检测、浮雕
- **艺术滤镜**: 复古、卡通、素描、霓虹、热成像
- **美颜滤镜**: 磨皮、美白、瘦脸、大眼
- **几何变形**: 鱼眼、镜像、像素化

### 🖼️ 贴图系统
- **面部贴纸**: 实时面部检测和贴纸应用
- **背景替换**: 智能背景分割和替换
- **3D贴图**: 基于面部关键点的3D模型贴图
- **粒子效果**: 动态粒子系统

### 🎯 面部检测
- **实时检测**: 高性能面部检测和跟踪
- **关键点定位**: 68点面部关键点检测
- **表情识别**: 情绪分析和表情分类
- **姿态估计**: 3D面部姿态估计

### 🚀 渲染技术
- **OpenGL渲染**: 硬件加速的实时渲染
- **着色器系统**: 可编程着色器管线
- **后处理**: 多重采样抗锯齿、阴影映射
- **环境映射**: 立方体贴图和反射效果

## 🛠️ 技术栈

- **C++17**: 现代C++标准
- **OpenCV 4.x**: 计算机视觉和图像处理
- **OpenGL 3.3+**: 图形渲染和GPU计算
- **GLFW**: 窗口管理和输入处理
- **GLEW**: OpenGL扩展加载
- **GLM**: 数学库
- **CMake**: 构建系统

## 📦 安装和构建

### 系统要求

- **操作系统**: Linux (Ubuntu 20.04+), macOS (10.15+), Windows 10+
- **编译器**: GCC 9+, Clang 10+, MSVC 2019+
- **GPU**: 支持OpenGL 3.3+的显卡
- **摄像头**: USB摄像头或内置摄像头

### 依赖安装

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y \
    build-essential cmake pkg-config \
    libopencv-dev libgl1-mesa-dev libglu1-mesa-dev \
    libglfw3-dev libglew-dev libglm-dev
```

#### macOS
```bash
brew install cmake opencv glfw glew glm
```

#### Windows
使用vcpkg安装依赖：
```cmd
vcpkg install opencv glfw3 glew glm
```

### 构建项目

#### 使用构建脚本（推荐）
```bash
# 克隆项目
git clone https://github.com/gugugu5331/VideoCall-System.git
cd VideoCall-System/src/video-processing

# 设置权限
chmod +x build.sh

# 安装依赖并构建
./build.sh --all

# 或者分步执行
./build.sh --deps    # 安装依赖
./build.sh --build   # 构建项目
./build.sh --test    # 运行测试
```

#### 手动构建
```bash
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
make -j$(nproc)
```

## 🚀 使用方法

### 基本使用
```bash
# 运行应用
./build/VideoProcessing

# 指定摄像头
./build/VideoProcessing --camera 0

# 设置窗口大小
./build/VideoProcessing --width 1920 --height 1080

# 全屏模式
./build/VideoProcessing --fullscreen

# 查看所有选项
./build/VideoProcessing --help
```

### 快捷键控制

| 按键 | 功能 |
|------|------|
| `ESC` | 退出应用 |
| `SPACE` | 截图 |
| `R` | 开始/停止录制 |
| `F` | 切换全屏 |
| `U` | 显示/隐藏UI |
| `1-9` | 应用不同滤镜 |
| `0` | 移除所有滤镜 |
| `M` | 镜像模式 |
| `D` | 面部检测开关 |
| `B` | 美颜模式 |
| `C` | 卡通模式 |
| `S` | 素描模式 |

### API使用示例

```cpp
#include "video_processor.h"

using namespace VideoProcessing;

int main() {
    VideoProcessor processor;
    
    // 初始化
    processor.Initialize(1280, 720);
    
    // 启动摄像头
    processor.StartCamera(0);
    
    // 设置滤镜
    processor.SetFilter(FilterType::BEAUTY);
    
    // 加载贴纸
    processor.LoadSticker("heart", "assets/heart.png");
    processor.SetActiveSticker("heart");
    
    // 启用面部检测
    processor.EnableFaceDetection(true);
    
    // 运行主循环
    processor.Run();
    
    return 0;
}
```

## 🐳 Docker部署

### 构建镜像
```bash
docker build -t video-processing .
```

### 运行容器
```bash
# 基本运行
docker run --rm -it \
    --device /dev/video0 \
    -e DISPLAY=$DISPLAY \
    -v /tmp/.X11-unix:/tmp/.X11-unix \
    video-processing

# 使用Docker Compose
docker-compose up -d
```

### 访问服务
- 应用界面: http://localhost:80
- 监控面板: http://localhost:3000 (Grafana)
- 指标数据: http://localhost:9090 (Prometheus)

## 📁 项目结构

```
src/video-processing/
├── include/                 # 头文件
│   ├── common.h            # 通用定义
│   ├── video_processor.h   # 主处理器
│   ├── camera_capture.h    # 摄像头捕获
│   ├── opengl_renderer.h   # OpenGL渲染器
│   ├── filter_manager.h    # 滤镜管理器
│   ├── face_detector.h     # 面部检测器
│   ├── texture_manager.h   # 纹理管理器
│   └── shader_manager.h    # 着色器管理器
├── src/                    # 源文件
├── shaders/                # 着色器文件
│   ├── basic.vert         # 基础顶点着色器
│   └── basic.frag         # 基础片段着色器
├── assets/                 # 资源文件
├── textures/              # 纹理文件
├── config/                # 配置文件
├── CMakeLists.txt         # CMake配置
├── build.sh              # 构建脚本
├── Dockerfile            # Docker配置
├── docker-compose.yml    # Docker Compose配置
└── README.md             # 说明文档
```

## 🎯 性能优化

### GPU加速
- 使用OpenGL进行硬件加速渲染
- 着色器并行处理图像效果
- 纹理内存优化

### CPU优化
- 多线程处理管线
- SIMD指令优化
- 内存池管理

### 实时性能
- 帧率控制和VSync
- 延迟优化
- 缓存策略

## 🔧 配置选项

### 滤镜参数
```cpp
EffectParams params;
params.intensity = 0.8f;      // 效果强度
params.brightness = 0.1f;     // 亮度调整
params.contrast = 1.2f;       // 对比度
params.saturation = 1.1f;     // 饱和度
params.hue = 0.0f;           // 色相偏移
```

### 渲染设置
```cpp
Settings settings;
settings.target_fps = 60;     // 目标帧率
settings.msaa_samples = 4;    // 抗锯齿采样
settings.vsync = true;        // 垂直同步
settings.fullscreen = false;  // 全屏模式
```

## 🐛 故障排除

### 常见问题

1. **摄像头无法打开**
   - 检查设备权限: `ls -l /dev/video*`
   - 确认摄像头未被其他程序占用

2. **OpenGL错误**
   - 更新显卡驱动
   - 检查OpenGL版本: `glxinfo | grep OpenGL`

3. **编译错误**
   - 确认所有依赖已安装
   - 检查CMake版本 >= 3.16

4. **性能问题**
   - 降低分辨率或帧率
   - 关闭不必要的滤镜效果
   - 检查GPU使用率

### 调试模式
```bash
# 编译调试版本
cmake -DCMAKE_BUILD_TYPE=Debug ..
make

# 使用GDB调试
gdb ./VideoProcessing
```

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支: `git checkout -b feature/new-filter`
3. 提交更改: `git commit -am 'Add new filter'`
4. 推送分支: `git push origin feature/new-filter`
5. 创建Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [OpenCV](https://opencv.org/) - 计算机视觉库
- [OpenGL](https://www.opengl.org/) - 图形API
- [GLFW](https://www.glfw.org/) - 窗口管理
- [GLM](https://glm.g-truc.net/) - 数学库

## 📞 联系方式

- 项目主页: https://github.com/gugugu5331/VideoCall-System
- 问题反馈: https://github.com/gugugu5331/VideoCall-System/issues

---

⭐ 如果这个项目对您有帮助，请给我们一个星标！
