# 远程用户音视频 AI 分析功能实现

## 概述

本功能实现了对会议中远程用户音视频流的实时 AI 分析，包括：
- 深度伪造检测（Deepfake Detection）
- 语音识别（ASR - Automatic Speech Recognition）
- 情绪识别（Emotion Recognition）

## 核心架构

### 数据流

```
远程用户 PeerConnection
    ↓
MediaStream (接收远程流)
    ↓
RemoteStreamAnalyzer (提取音视频数据)
    ↓
AIService (调用后端API)
    ↓
AIPanelController (管理结果)
    ↓
UI显示 (VideoTile + AIPanel)
```

## 新增组件

### 1. RemoteStreamAnalyzer

**文件**: `include/webrtc/remote_stream_analyzer.h`, `src/webrtc/remote_stream_analyzer.cpp`

**功能**:
- 连接到远程用户的 MediaStream
- 定时提取视频帧（每5秒）
- 累积音频数据（每3秒）
- 视频降采样（1080p → 360p）
- 音频重采样（48kHz → 16kHz）
- 调用 AIService 进行分析

**关键方法**:
```cpp
void attachToStream(MediaStream *stream);
void startAnalysis();
void stopAnalysis();
QByteArray extractVideoFrameData(const QVideoFrame &frame);
QByteArray convertToWAV(const QByteArray &pcmData, ...);
```

### 2. 增强的 WebRTCManager

**修改**: `include/webrtc/webrtc_manager.h`, `src/webrtc/webrtc_manager.cpp`

**新增功能**:
- 为每个远程用户创建 RemoteStreamAnalyzer
- 在接收到远程流时自动启动 AI 分析
- 管理分析器生命周期

**关键方法**:
```cpp
void setAIService(AIService *aiService);
void setupAIAnalysisForRemoteStream(int remoteUserId, MediaStream *stream);
```

### 3. 增强的 AIPanelController

**修改**: `include/ui/ai_panel_controller.h`, `src/ui/ai_panel_controller.cpp`

**新增功能**:
- 按用户ID分组存储AI结果
- 提供按用户查询结果的接口
- 自动获取用户名

**新增方法**:
```cpp
Q_INVOKABLE QVariantMap getDetectionResultForUser(int userId) const;
Q_INVOKABLE QVariantMap getEmotionResultForUser(int userId) const;
Q_INVOKABLE QVariantList getAsrResultsForUser(int userId) const;
Q_INVOKABLE QString getUsernameById(int userId) const;
```

### 4. 增强的 VideoTile

**修改**: `qml/components/VideoTile.qml`

**新增功能**:
- AI结果叠加层
- 实时显示深度伪造检测结果
- 实时显示情绪识别结果
- 实时显示语音识别文本

**UI效果**:
```
┌─────────────────────────┐
│ ✅ 真实 (98%)           │
│ 😊 开心 (88%)           │
│ 💬 大家好               │
│                         │
│   [视频画面]            │
│                         │
│   用户名                │
└─────────────────────────┘
```

## API 修改

### ApiClient

**修改**: `include/network/api_client.h`, `src/network/api_client.cpp`

所有 AI 相关接口添加 `userId` 参数：

```cpp
void speechRecognition(..., int userId, ...);
void emotionDetection(..., int userId, ...);
void synthesisDetection(..., int userId, ...);
```

### AIService

**修改**: `src/services/ai_service.cpp`

更新调用 ApiClient 时传递 userId 参数。

## 性能优化

### 1. 定时批量分析
- 视频：每5秒分析一次（避免频繁请求）
- 音频：累积3秒后分析

### 2. 数据降采样
- 视频：从1080p降到360p（减少70%数据量）
- 音频：从48kHz降到16kHz（减少67%数据量）

### 3. 异步处理
- 所有数据提取和HTTP请求都是异步的
- 不阻塞主线程和渲染线程

### 4. 结果缓存
- 深度伪造/情绪：每个用户只保留最新结果
- 语音识别：每个用户最多保留20条历史记录

## 使用方法

### 1. 初始化

在 `Application::setupQmlContext()` 中自动完成：

```cpp
m_webrtcManager->setAIService(m_aiService.get());
```

### 2. 自动启动

当远程用户加入会议时，WebRTCManager 自动：
1. 接收远程流
2. 创建 RemoteStreamAnalyzer
3. 启动 AI 分析

### 3. 查询结果

在 QML 中：

```qml
VideoTile {
    userId: model.userId
    aiPanelController: root.aiPanelController
    
    // 自动显示AI结果
}
```

## 配置参数

可在 `WebRTCManager::setupAIAnalysisForRemoteStream()` 中调整：

```cpp
analyzer->setVideoAnalysisInterval(5000);      // 视频分析间隔（毫秒）
analyzer->setAudioBufferDuration(3000);        // 音频缓冲时长（毫秒）
analyzer->setVideoDownscaleSize(QSize(640, 360)); // 视频降采样尺寸
analyzer->setAudioSampleRate(16000);           // 音频采样率（Hz）

analyzer->setDeepfakeDetectionEnabled(true);   // 启用深度伪造检测
analyzer->setAsrEnabled(true);                 // 启用语音识别
analyzer->setEmotionDetectionEnabled(true);    // 启用情绪识别
```

## 日志

所有关键操作都有详细日志：

```
[INFO] Setting up AI analysis for remote user: 123
[INFO] AI analysis started for remote user: 123
[DEBUG] Analyzing video frames for user: 123 (buffer size: 15)
[DEBUG] Sent video data for deepfake detection (user: 123, size: 45678 bytes)
[INFO] Deepfake detection completed for user 123: Real (confidence: 0.98)
```

## 注意事项

1. **网络带宽**: AI分析会增加上行带宽消耗（约100-200KB/s per user）
2. **后端性能**: 确保后端AI服务有足够的GPU资源
3. **隐私**: 远程用户的音视频数据会发送到后端，需要用户同意
4. **准确性**: AI结果仅供参考，不应作为唯一判断依据

## 未来改进

1. 支持本地AI推理（使用ONNX Runtime）
2. 添加更多AI功能（人脸识别、手势识别等）
3. 优化网络传输（使用WebSocket发送数据）
4. 添加用户配置界面（允许用户自定义分析参数）

