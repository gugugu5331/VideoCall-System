# FFmpeg服务使用指南

## 目录

1. [环境准备](#环境准备)
2. [编译项目](#编译项目)
3. [运行示例](#运行示例)
4. [集成到项目](#集成到项目)
5. [API使用](#api使用)
6. [配置管理](#配置管理)
7. [性能优化](#性能优化)
8. [故障排除](#故障排除)

## 环境准备

### 系统要求

- **操作系统**: Windows 10+, Linux (Ubuntu 18.04+), macOS 10.15+
- **编译器**: 
  - Windows: Visual Studio 2019+ 或 MinGW-w64
  - Linux: GCC 7+ 或 Clang 8+
  - macOS: Xcode 11+ 或 Clang 8+
- **CMake**: 3.16+
- **Python**: 3.7+ (用于测试和集成)
- **vcpkg**: 最新版本

### 依赖安装

#### Windows环境

```batch
# 1. 安装vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
bootstrap-vcpkg.bat

# 2. 安装依赖包
vcpkg install ffmpeg:x64-windows
vcpkg install opencv4:x64-windows
vcpkg install onnxruntime:x64-windows
vcpkg install nlohmann-json:x64-windows
vcpkg install spdlog:x64-windows
vcpkg install fmt:x64-windows

# 3. 集成到Visual Studio (可选)
vcpkg integrate install
```

#### Linux/macOS环境

```bash
# 1. 安装系统依赖
sudo apt update
sudo apt install build-essential cmake git python3 python3-pip

# 2. 安装vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
./bootstrap-vcpkg.sh

# 3. 安装依赖包
./vcpkg install ffmpeg
./vcpkg install opencv4
./vcpkg install onnxruntime
./vcpkg install nlohmann-json
./vcpkg install spdlog
./vcpkg install fmt
```

### 自动化环境准备

使用提供的脚本可以自动完成环境准备：

**Windows:**
```batch
setup_environment.bat
```

**Linux/macOS:**
```bash
chmod +x setup_environment.sh
./setup_environment.sh
```

## 编译项目

### 使用构建脚本

**Windows:**
```batch
build.bat
```

**Linux/macOS:**
```bash
chmod +x build.sh
./build.sh
```

### 手动编译

```bash
# 1. 创建构建目录
mkdir build && cd build

# 2. 配置CMake
cmake .. -DCMAKE_TOOLCHAIN_FILE=/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake

# 3. 编译
make -j$(nproc)  # Linux/macOS
# 或
cmake --build . --config Release  # Windows
```

### 编译选项

```bash
# Debug版本
cmake .. -DCMAKE_BUILD_TYPE=Debug

# Release版本（推荐）
cmake .. -DCMAKE_BUILD_TYPE=Release

# 启用GPU加速
cmake .. -DENABLE_GPU=ON

# 自定义安装路径
cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local
```

## 运行示例

### 基本功能测试

```bash
# 运行Python测试脚本
python test_basic_functionality.py
```

### 运行C++示例程序

```bash
# Windows
build\bin\ffmpeg_service_example.exe

# Linux/macOS
./build/bin/ffmpeg_service_example
```

### 示例程序功能

示例程序演示了以下功能：

1. **单次检测**
   - 视频伪造检测
   - 音频伪造检测
   - 混合检测

2. **批量处理**
   - 批量视频检测
   - 批量音频检测

3. **实时检测**
   - 实时视频流检测
   - 实时音频流检测

4. **性能监控**
   - 处理时间统计
   - 内存使用监控
   - 吞吐量计算

## 集成到项目

### 自动化集成

使用提供的脚本可以自动完成项目集成：

```bash
python integrate_with_project.py
```

这个脚本会：
- 集成到Python AI服务
- 集成到Go后端
- 集成到WebRTC前端
- 创建配置文件

### Python AI服务集成

```python
from app.services.ffmpeg_integration import get_ffmpeg_service

# 获取服务实例
ffmpeg_service = get_ffmpeg_service()

# 检测视频伪造
result = ffmpeg_service.detect_video_forgery(video_data, config={
    "detection_type": "video_deepfake",
    "confidence_threshold": 0.8
})

# 检测音频伪造
result = ffmpeg_service.detect_audio_forgery(audio_data, config={
    "detection_type": "voice_spoofing",
    "confidence_threshold": 0.8
})

# 压缩媒体
result = ffmpeg_service.compress_media(media_data, "video", config={
    "compression_level": "medium",
    "quality": 0.8
})
```

### Go后端集成

```go
package main

import (
    "net/http"
    "encoding/json"
)

func main() {
    // 创建FFmpeg处理器
    config := FFmpegServiceConfig{
        ServicePath: "./ffmpeg_service_example",
        Timeout: 60,
    }
    handler := NewFFmpegHandler(config)
    
    // 注册路由
    http.HandleFunc("/api/ffmpeg/detect", handler.DetectForgery)
    http.HandleFunc("/api/ffmpeg/compress", handler.CompressMedia)
    
    // 启动服务器
    http.ListenAndServe(":8080", nil)
}
```

### WebRTC前端集成

```javascript
// 初始化FFmpeg服务
await window.ffmpegService.initialize();

// 实时检测
const stopDetection = window.ffmpegService.startRealTimeDetection(
    mediaStream,
    (result) => {
        console.log('检测结果:', result);
        if (result.video.is_forgery || result.audio.is_forgery) {
            alert('检测到伪造内容！');
        }
    },
    {
        interval: 5000,  // 5秒检测一次
        detection: {
            confidence_threshold: 0.8
        }
    }
);

// 压缩视频流
const compressedStream = await window.ffmpegService.compressVideoStream(
    videoStream,
    {
        bitrate: 1000000,  // 1Mbps
        framerate: 30,
        width: 1280,
        height: 720
    }
);
```

## API使用

### C++ API

#### 基本使用

```cpp
#include "integration_service.h"

using namespace integration_service;

int main() {
    // 1. 配置服务
    IntegrationConfig config;
    config.ffmpeg.video_codec = "libx264";
    config.ffmpeg.audio_codec = "aac";
    config.onnx.model_path = "models/detection.onnx";
    config.processing.max_threads = 4;
    
    // 2. 初始化服务
    IntegrationService service;
    if (!service.initialize(config)) {
        std::cerr << "服务初始化失败!" << std::endl;
        return -1;
    }
    
    // 3. 执行检测
    std::vector<uint8_t> video_data = loadVideoData("test.mp4");
    auto result = service.detectVideo(video_data, DetectionType::VIDEO_DEEPFAKE);
    
    if (result.success) {
        std::cout << "检测结果: " << result.confidence << std::endl;
    }
    
    // 4. 清理资源
    service.cleanup();
    return 0;
}
```

#### 实时检测

```cpp
// 设置回调函数
auto detectionCallback = [](const DetectionResult& result) {
    std::cout << "实时检测结果: " << result.confidence << std::endl;
};

// 启动实时检测
service.startRealTimeDetection(detectionCallback);

// 输入视频帧
std::vector<uint8_t> frame_data = getVideoFrame();
service.processVideoFrame(frame_data, 1920, 1080, AV_PIX_FMT_YUV420P);

// 停止实时检测
service.stopRealTimeDetection();
```

#### 批量处理

```cpp
// 批量视频检测
std::vector<std::string> video_files = {"video1.mp4", "video2.mp4", "video3.mp4"};
auto results = service.detectVideoBatch(video_files, DetectionType::VIDEO_DEEPFAKE);

for (const auto& result : results) {
    std::cout << result.filename << ": " << result.confidence << std::endl;
}
```

### Python API

#### 基本使用

```python
from app.services.ffmpeg_integration import get_ffmpeg_service

# 获取服务实例
service = get_ffmpeg_service()

# 检测视频伪造
with open('test_video.mp4', 'rb') as f:
    video_data = f.read()

result = service.detect_video_forgery(video_data, {
    'detection_type': 'video_deepfake',
    'confidence_threshold': 0.8,
    'processing_mode': 'fast'
})

if result['success']:
    print(f"检测结果: {result['result']}")
else:
    print(f"检测失败: {result['error']}")
```

#### 批量处理

```python
import os

video_files = ['video1.mp4', 'video2.mp4', 'video3.mp4']
results = []

for video_file in video_files:
    with open(video_file, 'rb') as f:
        video_data = f.read()
    
    result = service.detect_video_forgery(video_data)
    results.append({
        'file': video_file,
        'result': result
    })

# 分析结果
for result in results:
    if result['result']['success']:
        print(f"{result['file']}: {result['result']['result']['confidence']}")
```

### Go API

#### 基本使用

```go
package main

import (
    "encoding/json"
    "net/http"
)

func detectVideoHandler(w http.ResponseWriter, r *http.Request) {
    // 解析请求
    var req FFmpegDetectionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 创建处理器
    config := FFmpegServiceConfig{
        ServicePath: "./ffmpeg_service_example",
        Timeout: 60,
    }
    handler := NewFFmpegHandler(config)
    
    // 执行检测
    handler.DetectForgery(w, r)
}

func main() {
    http.HandleFunc("/api/detect", detectVideoHandler)
    http.ListenAndServe(":8080", nil)
}
```

## 配置管理

### 配置文件结构

```json
{
    "ffmpeg_service": {
        "enabled": true,
        "service_path": "./ffmpeg_service_example",
        "timeout": 60,
        "max_concurrent_requests": 10
    },
    "integration": {
        "python_ai_service": {
            "enabled": true,
            "endpoint": "/api/ffmpeg"
        },
        "go_backend": {
            "enabled": true,
            "endpoint": "/api/ffmpeg"
        },
        "webrtc_frontend": {
            "enabled": true,
            "real_time_detection": true
        }
    },
    "detection": {
        "video_deepfake": {
            "enabled": true,
            "confidence_threshold": 0.8,
            "processing_mode": "fast"
        },
        "voice_spoofing": {
            "enabled": true,
            "confidence_threshold": 0.8,
            "processing_mode": "fast"
        }
    },
    "compression": {
        "video": {
            "codec": "h264",
            "bitrate": 1000000,
            "quality": 0.8
        },
        "audio": {
            "codec": "aac",
            "bitrate": 128000,
            "quality": 0.8
        }
    }
}
```

### 环境变量

```bash
# FFmpeg服务配置
export FFMPEG_SERVICE_PATH="./ffmpeg_service_example"
export FFMPEG_SERVICE_TIMEOUT=60
export FFMPEG_MAX_CONCURRENT_REQUESTS=10

# 检测配置
export DETECTION_CONFIDENCE_THRESHOLD=0.8
export DETECTION_PROCESSING_MODE=fast

# 压缩配置
export VIDEO_BITRATE=1000000
export AUDIO_BITRATE=128000
export COMPRESSION_QUALITY=0.8
```

### 运行时配置

```cpp
// C++ 运行时配置
IntegrationConfig config;
config.loadFromFile("config.json");
config.loadFromEnvironment();  // 从环境变量加载

// 动态修改配置
config.detection.video_deepfake.confidence_threshold = 0.9;
service.updateConfig(config);
```

```python
# Python 运行时配置
import os

# 从环境变量加载配置
service_path = os.getenv('FFMPEG_SERVICE_PATH', './ffmpeg_service_example')
timeout = int(os.getenv('FFMPEG_SERVICE_TIMEOUT', '60'))

# 创建服务实例
service = FFmpegServiceIntegration(service_path)
```

## 性能优化

### 编译优化

```bash
# 启用优化编译
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-O3 -march=native"

# 启用并行编译
make -j$(nproc)  # Linux/macOS
cmake --build . --config Release --parallel  # Windows
```

### 运行时优化

```cpp
// 1. 调整线程数
config.processing.max_threads = std::thread::hardware_concurrency();

// 2. 启用GPU加速
config.onnx.device = "gpu";
config.onnx.gpu_id = 0;

// 3. 优化内存使用
config.processing.buffer_size = 1024 * 1024;  // 1MB缓冲区

// 4. 启用批处理
config.processing.batch_size = 4;
```

### 模型优化

```cpp
// 1. 模型量化
ModelOptimizer optimizer;
optimizer.quantizeModel("model.onnx", "model_quantized.onnx");

// 2. 图优化
optimizer.optimizeGraph("model.onnx", "model_optimized.onnx");

// 3. 模型剪枝
optimizer.pruneModel("model.onnx", "model_pruned.onnx", 0.3);
```

### 监控和调优

```cpp
// 启用性能监控
PerformanceMonitor monitor;
monitor.start();

// 执行检测
auto result = service.detectVideo(video_data);

// 获取性能统计
auto stats = monitor.getStats();
std::cout << "处理时间: " << stats.processing_time_ms << "ms" << std::endl;
std::cout << "内存使用: " << stats.memory_usage_mb << "MB" << std::endl;
std::cout << "吞吐量: " << stats.throughput_fps << "fps" << std::endl;
```

## 故障排除

### 常见问题

#### 1. 编译错误

**问题**: `find_package(FFMPEG REQUIRED)` 失败
**解决方案**:
```bash
# 确保vcpkg正确安装
vcpkg install ffmpeg:x64-windows

# 检查CMake工具链文件路径
cmake .. -DCMAKE_TOOLCHAIN_FILE=/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake
```

**问题**: ONNX Runtime链接错误
**解决方案**:
```bash
# 重新安装ONNX Runtime
vcpkg remove onnxruntime:x64-windows
vcpkg install onnxruntime:x64-windows
```

#### 2. 运行时错误

**问题**: 找不到FFmpeg库
**解决方案**:
```bash
# 检查库文件路径
ls /path/to/vcpkg/installed/x64-windows/lib/

# 设置环境变量
export LD_LIBRARY_PATH=/path/to/vcpkg/installed/x64-windows/lib:$LD_LIBRARY_PATH
```

**问题**: 模型文件不存在
**解决方案**:
```bash
# 检查模型文件路径
ls models/

# 下载预训练模型
wget https://example.com/models/detection.onnx -O models/detection.onnx
```

#### 3. 性能问题

**问题**: 检测速度慢
**解决方案**:
```cpp
// 1. 启用GPU加速
config.onnx.device = "gpu";

// 2. 减少输入分辨率
config.preprocessing.target_width = 640;
config.preprocessing.target_height = 480;

// 3. 使用更快的模型
config.onnx.model_path = "models/fast_detection.onnx";
```

**问题**: 内存使用过高
**解决方案**:
```cpp
// 1. 减少批处理大小
config.processing.batch_size = 1;

// 2. 减少缓冲区大小
config.processing.buffer_size = 512 * 1024;  // 512KB

// 3. 启用内存池
config.processing.enable_memory_pool = true;
```

### 调试技巧

#### 启用详细日志

```cpp
// C++
#include <spdlog/spdlog.h>

spdlog::set_level(spdlog::level::debug);
spdlog::debug("调试信息: {}", variable);
```

```python
# Python
import logging

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)
logger.debug("调试信息: %s", variable)
```

#### 性能分析

```cpp
// 使用性能监控
PerformanceMonitor monitor;
monitor.start();

// 执行操作
auto result = service.detectVideo(video_data);

// 获取详细统计
auto stats = monitor.getDetailedStats();
for (const auto& stat : stats) {
    std::cout << stat.name << ": " << stat.value << " " << stat.unit << std::endl;
}
```

#### 内存检查

```cpp
// 启用内存检查
#ifdef _DEBUG
    _CrtSetDbgFlag(_CRTDBG_ALLOC_MEM_DF | _CRTDBG_LEAK_CHECK_DF);
#endif
```

### 获取帮助

1. **查看日志**: 检查 `logs/` 目录下的日志文件
2. **运行测试**: 使用 `test_basic_functionality.py` 诊断问题
3. **检查配置**: 验证 `integration_config.json` 配置正确
4. **查看文档**: 参考 `README.md` 和 `IMPLEMENTATION_SUMMARY.md`

### 联系支持

如果遇到无法解决的问题，请：

1. 收集错误日志和系统信息
2. 运行诊断脚本
3. 查看项目文档
4. 提交Issue到项目仓库 