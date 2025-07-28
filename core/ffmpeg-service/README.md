# FFmpeg服务 + ONNX检测器

基于FFmpeg的音视频编解码和压缩模块，以及使用ONNX Runtime部署的深度学习伪造检测模型。该模块专为视频通话系统设计，提供高性能的音视频处理和实时伪造检测功能。

## 功能特性

### 🎥 FFmpeg音视频处理
- **音视频编解码**: 支持H.264/H.265视频编码和AAC音频编码
- **格式转换**: 支持多种音视频格式之间的转换
- **压缩优化**: 智能压缩算法，平衡质量和文件大小
- **实时处理**: 支持实时音视频流处理
- **硬件加速**: 支持GPU硬件加速（可选）

### 🤖 ONNX深度学习检测
- **语音伪造检测**: 基于深度学习的语音反欺骗检测
- **视频深度伪造检测**: 检测Deepfake、换脸等视频伪造
- **音频伪影检测**: 检测音频合成伪影
- **视频伪影检测**: 检测视频压缩和合成伪影
- **模型优化**: 支持模型量化和图优化

### 🔧 集成服务
- **低耦合设计**: 模块化架构，易于集成和维护
- **多语言支持**: 提供C++、Python、Go接口
- **实时检测**: 支持实时音视频流检测
- **批量处理**: 支持批量音视频文件处理
- **性能监控**: 内置性能监控和统计功能

## 系统要求

### 硬件要求
- **CPU**: Intel i5或AMD Ryzen 5以上
- **内存**: 8GB RAM（推荐16GB）
- **GPU**: NVIDIA GTX 1060或更高（可选，用于GPU加速）
- **存储**: 10GB可用空间

### 软件要求
- **操作系统**: Windows 10/11, Linux (Ubuntu 18.04+), macOS 10.15+
- **编译器**: GCC 7+, Clang 6+, MSVC 2019+
- **CMake**: 3.16+
- **依赖库**:
  - FFmpeg 4.0+
  - OpenCV 4.0+
  - ONNX Runtime 1.8+
  - vcpkg（包管理）

## 快速开始

### 🚀 一键快速开始

**Windows用户:**
```batch
# 运行快速开始脚本
quick_start.bat
```

**Linux/macOS用户:**
```bash
# 给脚本执行权限
chmod +x quick_start.sh

# 运行快速开始脚本
./quick_start.sh
```

这个脚本会自动完成以下步骤：
1. ✅ 环境准备 - 安装FFmpeg、OpenCV、ONNX Runtime等依赖
2. ✅ 项目编译 - 构建C++库和示例程序
3. ✅ 功能测试 - 验证基本功能正常
4. ✅ 项目集成 - 集成到Python AI服务、Go后端、WebRTC前端

### 📋 手动步骤

如果您想手动执行每个步骤，请按照以下说明：

#### 1. 环境准备

**Windows环境:**
```batch
# 运行环境准备脚本
setup_environment.bat
```

**Linux/macOS环境:**
```bash
# 运行环境准备脚本
./setup_environment.sh
```

#### 2. 编译项目

**Windows:**
```batch
# 运行构建脚本
build.bat
```

**Linux/macOS:**
```bash
# 运行构建脚本
./build.sh
```

#### 3. 运行测试

```bash
# 运行基本功能测试
python test_basic_functionality.py
```

#### 4. 集成到项目

```bash
# 运行项目集成脚本
python integrate_with_project.py
```

### 3. 运行示例

```bash
# 运行示例程序
./bin/ffmpeg_service_example

# 运行测试
./bin/ffmpeg_service_test
```

## 使用指南

### 基本使用

#### 1. 初始化服务

```cpp
#include "integration_service.h"

using namespace integration_service;

// 创建集成服务
IntegrationService service;

// 配置参数
IntegrationConfig config;
config.ffmpeg_params.video_bitrate = 1000000;  // 1Mbps
config.ffmpeg_params.audio_bitrate = 128000;   // 128kbps
config.video_model_config.confidence_threshold = 0.8f;
config.audio_model_config.confidence_threshold = 0.8f;

// 初始化
if (!service.initialize(config)) {
    std::cerr << "服务初始化失败!" << std::endl;
    return -1;
}
```

#### 2. 视频检测

```cpp
// 准备视频数据
std::vector<uint8_t> video_data = loadVideoData("test.mp4");
int width = 1280, height = 720, fps = 30;

// 执行检测
auto result = service.detectVideo(video_data, width, height, fps);

// 处理结果
if (result.is_fake) {
    std::cout << "检测到伪造视频!" << std::endl;
    std::cout << "置信度: " << result.overall_confidence << std::endl;
    std::cout << "风险评分: " << result.overall_risk_score << std::endl;
}
```

#### 3. 音频检测

```cpp
// 准备音频数据
std::vector<uint8_t> audio_data = loadAudioData("test.wav");
int sample_rate = 44100, channels = 2;

// 执行检测
auto result = service.detectAudio(audio_data, sample_rate, channels);

// 处理结果
if (result.is_fake) {
    std::cout << "检测到伪造音频!" << std::endl;
}
```

#### 4. 混合检测

```cpp
// 同时检测音视频
auto result = service.detectHybrid(video_data, audio_data, 
                                  width, height, fps, 
                                  sample_rate, channels);
```

### 高级功能

#### 1. 实时检测

```cpp
// 设置回调函数
auto callback = [](const IntegratedDetectionResult& result) {
    if (result.is_fake) {
        std::cout << "实时检测到伪造内容!" << std::endl;
    }
};

// 启动实时检测
service.startRealTimeDetection(IntegratedDetectionType::REAL_TIME_VIDEO, callback);

// 处理实时数据
while (running) {
    auto frame_data = getNextFrame();
    service.processVideoFrame(frame_data, width, height);
}

// 停止检测
service.stopRealTimeDetection();
```

#### 2. 批量处理

```cpp
// 准备批量数据
std::vector<std::vector<uint8_t>> video_batch;
for (const auto& file : video_files) {
    video_batch.push_back(loadVideoData(file));
}

// 进度回调
auto progress_callback = [](int progress, const std::string& status) {
    std::cout << "进度: " << progress << "% - " << status << std::endl;
};

// 批量检测
auto results = service.batchDetectVideo(video_batch, progress_callback);
```

#### 3. 性能监控

```cpp
// 启用性能监控
service.enablePerformanceMonitoring(true);

// 执行检测操作
for (int i = 0; i < 100; ++i) {
    service.detectVideo(test_data, width, height, fps);
}

// 获取性能统计
std::unordered_map<std::string, double> stats;
service.getPerformanceStats(stats);

for (const auto& stat : stats) {
    std::cout << stat.first << ": " << stat.second << std::endl;
}
```

### 与现有项目集成

#### 1. 与Python AI服务集成

```cpp
#include "project_integration.h"

using namespace project_integration;

// 创建Python AI服务集成
PythonAIServiceIntegration python_integration;

// 初始化
if (!python_integration.initialize("config.json")) {
    return -1;
}

// 创建检测请求
DetectionRequest request;
request.detection_id = "test_001";
request.detection_type = "video_deepfake";
request.video_data = video_data;
request.metadata["user_id"] = "user123";

// 执行检测
auto response = python_integration.detect(request);
```

#### 2. 与Go后端集成

```cpp
// 创建Go后端集成
GoBackendIntegration go_integration;

// 初始化
if (!go_integration.initialize("go_config.json")) {
    return -1;
}

// 执行检测
auto result = go_integration.detectVideo(video_data, width, height);

// 获取性能统计
auto stats = go_integration.getPerformanceStats();
```

#### 3. 与WebRTC集成

```cpp
// 创建WebRTC集成
WebRTCIntegration webrtc_integration;

// 初始化
if (!webrtc_integration.initialize("webrtc_config.json")) {
    return -1;
}

// 配置流检测
StreamConfig stream_config;
stream_config.detection_interval_ms = 1000;
stream_config.enable_video_detection = true;
stream_config.enable_audio_detection = true;
webrtc_integration.setStreamConfig(stream_config);

// 启动流检测
webrtc_integration.startStreamDetection([](const IntegratedDetectionResult& result) {
    // 处理检测结果
});
```

## 配置说明

### 配置文件结构

```json
{
  "ffmpeg_params": {
    "video_bitrate": 1000000,
    "audio_bitrate": 128000,
    "video_width": 1280,
    "video_height": 720,
    "video_fps": 30,
    "audio_sample_rate": 44100,
    "audio_channels": 2,
    "video_codec_id": "h264",
    "audio_codec_id": "aac",
    "enable_hardware_acceleration": false
  },
  "video_model_config": {
    "model_path": "models/video_deepfake.onnx",
    "confidence_threshold": 0.8,
    "risk_threshold": 0.7,
    "enable_gpu": false,
    "num_threads": 4,
    "enable_optimization": true
  },
  "audio_model_config": {
    "model_path": "models/voice_spoofing.onnx",
    "confidence_threshold": 0.8,
    "risk_threshold": 0.7,
    "enable_gpu": false,
    "num_threads": 4,
    "enable_optimization": true
  },
  "integration_config": {
    "video_weight": 0.6,
    "audio_weight": 0.4,
    "confidence_threshold": 0.8,
    "risk_threshold": 0.7,
    "enable_compression": true,
    "enable_real_time": true,
    "enable_feature_cache": true,
    "max_batch_size": 10,
    "processing_threads": 4
  }
}
```

### 环境变量

```bash
# FFmpeg配置
export FFMPEG_HOME=/usr/local/ffmpeg
export FFMPEG_LIBS=/usr/local/ffmpeg/lib

# ONNX Runtime配置
export ONNXRUNTIME_HOME=/usr/local/onnxruntime
export ONNXRUNTIME_LIBS=/usr/local/onnxruntime/lib

# OpenCV配置
export OPENCV_HOME=/usr/local/opencv
export OPENCV_LIBS=/usr/local/opencv/lib

# 模型路径
export MODEL_PATH=/path/to/models
export CACHE_PATH=/path/to/cache
```

## 性能优化

### 1. 编译优化

```bash
# 启用优化编译
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-O3 -march=native"

# 启用多线程编译
make -j$(nproc)
```

### 2. 运行时优化

```cpp
// 启用GPU加速
config.video_model_config.enable_gpu = true;
config.audio_model_config.enable_gpu = true;

// 优化线程数
config.processing_threads = std::thread::hardware_concurrency();

// 启用模型优化
config.video_model_config.enable_optimization = true;
config.audio_model_config.enable_optimization = true;
```

### 3. 内存优化

```cpp
// 启用特征缓存
config.enable_feature_cache = true;
config.cache_size = 1000;
config.cache_ttl_seconds = 3600;

// 批量处理优化
config.max_batch_size = 16;  // 根据内存大小调整
```

## 故障排除

### 常见问题

#### 1. 编译错误

**问题**: 找不到FFmpeg库
```bash
# 解决方案
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH
pkg-config --libs libavcodec libavformat libavutil
```

**问题**: ONNX Runtime链接错误
```bash
# 解决方案
export LD_LIBRARY_PATH=/usr/local/onnxruntime/lib:$LD_LIBRARY_PATH
```

#### 2. 运行时错误

**问题**: 模型加载失败
```cpp
// 检查模型文件路径
std::cout << "模型路径: " << config.video_model_config.model_path << std::endl;
// 确保模型文件存在且可读
```

**问题**: 内存不足
```cpp
// 减少批量大小
config.max_batch_size = 4;

// 减少缓存大小
config.cache_size = 100;
```

#### 3. 性能问题

**问题**: 检测速度慢
```cpp
// 启用GPU加速
config.video_model_config.enable_gpu = true;

// 增加线程数
config.processing_threads = 8;

// 启用模型优化
config.video_model_config.enable_optimization = true;
```

### 调试模式

```bash
# 启用调试编译
cmake .. -DCMAKE_BUILD_TYPE=Debug

# 启用详细日志
export FFMPEG_LOG_LEVEL=debug
export ONNXRUNTIME_LOG_LEVEL=debug
```

## API参考

### 核心类

#### IntegrationService
主要的集成服务类，提供统一的检测接口。

**主要方法**:
- `initialize(config)`: 初始化服务
- `detectVideo(data, width, height, fps)`: 视频检测
- `detectAudio(data, sample_rate, channels)`: 音频检测
- `detectHybrid(video_data, audio_data, ...)`: 混合检测
- `startRealTimeDetection(type, callback)`: 启动实时检测
- `batchDetectVideo(batch, callback)`: 批量视频检测

#### FFmpegProcessor
FFmpeg音视频处理器，负责编解码和格式转换。

**主要方法**:
- `compressVideo(data, params)`: 压缩视频
- `compressAudio(data, params)`: 压缩音频
- `convertVideoFormat(data, format, width, height)`: 视频格式转换
- `convertAudioFormat(data, format, sample_rate, channels)`: 音频格式转换

#### ONNXDetector
ONNX深度学习检测器，负责模型推理。

**主要方法**:
- `detectVoiceSpoofing(data, sample_rate, channels)`: 语音伪造检测
- `detectVideoDeepfake(data, width, height, fps)`: 视频深度伪造检测
- `detectFaceSwap(data, width, height, fps)`: 换脸检测
- `batchDetect(batch, type)`: 批量检测

### 数据结构

#### IntegratedDetectionResult
集成检测结果结构。

```cpp
struct IntegratedDetectionResult {
    bool is_fake;                    // 是否为伪造
    float overall_confidence;        // 整体置信度
    float overall_risk_score;        // 整体风险评分
    DetectionResult video_result;    // 视频检测结果
    DetectionResult audio_result;    // 音频检测结果
    ProcessingResult compression_result; // 压缩结果
    int64_t total_processing_time_ms;   // 总处理时间
    float compression_ratio;         // 压缩比
    int64_t frame_count;             // 帧数
    std::unordered_map<std::string, float> detailed_metrics; // 详细指标
    std::string detection_summary;   // 检测摘要
};
```

#### IntegrationConfig
集成服务配置结构。

```cpp
struct IntegrationConfig {
    ffmpeg_service::EncodingParams ffmpeg_params;      // FFmpeg参数
    onnx_detector::ModelConfig video_model_config;     // 视频模型配置
    onnx_detector::ModelConfig audio_model_config;     // 音频模型配置
    onnx_detector::PreprocessingParams preprocessing_params; // 预处理参数
    float video_weight;            // 视频权重
    float audio_weight;            // 音频权重
    float confidence_threshold;    // 置信度阈值
    float risk_threshold;          // 风险阈值
    int max_batch_size;            // 最大批量大小
    int processing_threads;        // 处理线程数
    bool enable_compression;       // 启用压缩
    bool enable_real_time;         // 启用实时处理
    bool enable_feature_cache;     // 启用特征缓存
    size_t cache_size;             // 缓存大小
    int cache_ttl_seconds;         // 缓存TTL
};
```

## 贡献指南

### 开发环境设置

1. 克隆项目
2. 安装依赖
3. 配置开发环境
4. 运行测试

### 代码规范

- 使用C++17标准
- 遵循Google C++ Style Guide
- 添加适当的注释和文档
- 编写单元测试

### 提交规范

- 使用清晰的提交信息
- 包含测试用例
- 更新相关文档

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 联系方式

- 项目维护者: [Your Name]
- 邮箱: [your.email@example.com]
- 项目地址: [GitHub Repository URL]

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持基本的音视频处理和检测功能
- 提供C++ API接口

### v1.1.0 (计划中)
- 添加GPU加速支持
- 优化模型推理性能
- 增加更多检测算法

---

**注意**: 本模块专为视频通话系统设计，请根据具体需求调整配置参数。如有问题，请参考故障排除部分或联系维护者。 