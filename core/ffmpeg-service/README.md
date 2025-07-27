# FFmpeg 伪造检测服务

基于FFmpeg和ONNX Runtime的高性能音视频伪造检测服务，支持实时流处理和批量文件处理。

## 功能特性

- **实时音视频处理**: 支持RTSP、RTMP、HTTP等流媒体协议
- **高效压缩**: 使用FFmpeg进行音视频压缩，减少传输带宽和处理时间
- **深度学习检测**: 基于ONNX Runtime的深度学习模型推理
- **多模态检测**: 支持视频伪造、音频伪造、换脸等多种检测类型
- **高性能**: 多线程处理，支持GPU加速
- **可配置**: 丰富的配置选项，适应不同场景需求

## 系统要求

- **操作系统**: Linux, Windows, macOS
- **编译器**: GCC 7+, Clang 5+, MSVC 2017+
- **依赖库**:
  - FFmpeg 4.0+
  - ONNX Runtime 1.8+
  - OpenCV 4.0+ (可选)
  - CMake 3.16+

## 编译安装

### 1. 安装依赖

#### Ubuntu/Debian
```bash
# 安装FFmpeg
sudo apt update
sudo apt install ffmpeg libavcodec-dev libavformat-dev libavutil-dev libswscale-dev libswresample-dev

# 安装ONNX Runtime
wget https://github.com/microsoft/onnxruntime/releases/download/v1.15.1/onnxruntime-linux-x64-1.15.1.tgz
tar -xzf onnxruntime-linux-x64-1.15.1.tgz
sudo cp -r onnxruntime-linux-x64-1.15.1/include/* /usr/local/include/
sudo cp onnxruntime-linux-x64-1.15.1/lib/libonnxruntime.so* /usr/local/lib/
sudo ldconfig
```

#### Windows
```bash
# 使用vcpkg安装依赖
vcpkg install ffmpeg:x64-windows
vcpkg install onnxruntime:x64-windows
vcpkg install opencv:x64-windows
```

### 2. 编译项目

```bash
mkdir build
cd build
cmake ..
make -j$(nproc)
```

### 3. 安装

```bash
sudo make install
```

## 使用方法

### 基本用法

```bash
# 处理视频文件
./ffmpeg_detection_service -i video.mp4 -m models/detection.onnx

# 处理实时流
./ffmpeg_detection_service -i rtsp://192.168.1.100:554/stream -m models/detection.onnx

# 使用配置文件
./ffmpeg_detection_service -i input.mp4 -m models/detection.onnx -c config.json
```

### 命令行参数

- `-i, --input <url/file>`: 输入流或文件路径
- `-m, --model <path>`: 模型文件路径
- `-c, --config <file>`: 配置文件路径
- `-o, --output <file>`: 输出日志文件
- `-v, --verbose`: 详细输出
- `-h, --help`: 显示帮助信息

### 配置文件

配置文件使用JSON格式，包含以下主要部分：

#### 压缩配置
```json
{
    "compression": {
        "target_width": 640,
        "target_height": 480,
        "target_fps": 30,
        "video_bitrate": 1000000,
        "audio_bitrate": 128000,
        "video_codec": "libx264",
        "audio_codec": "aac",
        "quality": 23
    }
}
```

#### 检测配置
```json
{
    "detection": {
        "model_path": "models/detection.onnx",
        "input_width": 224,
        "input_height": 224,
        "input_channels": 3,
        "use_gpu": false,
        "gpu_device_id": 0,
        "num_threads": 4,
        "confidence_threshold": 0.5
    }
}
```

## 模型格式

服务支持ONNX格式的深度学习模型，模型应具有以下特征：

### 视频检测模型
- 输入: `[batch_size, channels, height, width]`
- 输出: `[batch_size, num_classes]`
- 支持的检测类型: 人脸伪造、Deepfake、换脸等

### 音频检测模型
- 输入: `[batch_size, 1, sequence_length]`
- 输出: `[batch_size, num_classes]`
- 支持的检测类型: 音频伪造、语音合成等

## API接口

### C++ API

```cpp
#include "ffmpeg_processor.h"

// 创建处理器
auto processor = std::make_unique<ffmpeg_detection::FFmpegProcessor>();

// 初始化
ffmpeg_detection::CompressionConfig config;
config.target_width = 640;
config.target_height = 480;

processor->initialize("models/detection.onnx", config);

// 设置回调
processor->set_result_callback([](const ffmpeg_detection::ProcessingResult& result) {
    if (result.is_fake) {
        std::cout << "检测到伪造内容!" << std::endl;
    }
});

// 开始处理
processor->start_realtime_processing("rtsp://example.com/stream");
```

## 性能优化

### 1. GPU加速
在配置文件中启用GPU：
```json
{
    "detection": {
        "use_gpu": true,
        "gpu_device_id": 0
    }
}
```

### 2. 多线程处理
调整线程数：
```json
{
    "detection": {
        "num_threads": 8
    }
}
```

### 3. 批量处理
启用批量处理：
```json
{
    "processing": {
        "batch_size": 4,
        "enable_async": true
    }
}
```

## 监控和日志

### 性能监控
服务提供详细的性能统计信息：
- 处理帧数
- 检测到伪造的帧数
- 平均处理时间
- 压缩比
- 内存使用情况

### 日志配置
```json
{
    "logging": {
        "level": "INFO",
        "output_file": "logs/ffmpeg_detection.log",
        "max_file_size_mb": 100,
        "max_files": 10
    }
}
```

## 故障排除

### 常见问题

1. **FFmpeg库找不到**
   ```bash
   export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
   ```

2. **ONNX Runtime库找不到**
   ```bash
   export LD_LIBRARY_PATH=/path/to/onnxruntime/lib:$LD_LIBRARY_PATH
   ```

3. **GPU内存不足**
   - 减少batch_size
   - 降低输入分辨率
   - 使用CPU模式

4. **处理延迟过高**
   - 启用异步处理
   - 调整压缩参数
   - 使用更快的编解码器

### 调试模式

使用详细输出模式获取更多信息：
```bash
./ffmpeg_detection_service -i input.mp4 -m model.onnx -v
```

## 开发指南

### 添加新的检测类型

1. 在`detection_engine.h`中添加新的检测类型枚举
2. 在`detection_engine.cpp`中实现相应的预处理和后处理逻辑
3. 更新配置文件支持新参数

### 自定义预处理

继承`DetectionEngine`类并重写预处理方法：
```cpp
class CustomDetectionEngine : public DetectionEngine {
protected:
    std::vector<float> preprocess_video(const std::vector<uint8_t>& frame_data,
                                       int width, int height, int channels) override {
        // 自定义预处理逻辑
    }
};
```

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request来改进项目。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交GitHub Issue
- 发送邮件至项目维护者 