# AI Inference Node

完整的 AI 推理节点实现，支持以下功能：

## 支持的 AI 功能

1. **ASR (自动语音识别)**
   - 模型：基于 ONNX Runtime
   - 输入：音频数据
   - 输出：文本转录

2. **情绪检测**
   - 模型：基于 ONNX Runtime
   - 输入：文本或音频
   - 输出：情绪分类（anger, joy, sadness, etc.）

3. **合成检测 (Deepfake Detection)**
   - 模型：基于 ONNX Runtime
   - 输入：音频或视频
   - 输出：真伪判断

## 架构设计

- **BaseTask**: 基础任务类，所有 AI 任务继承自此类
- **ASRTask**: ASR 任务实现，负责模型加载和推理
- **EmotionTask**: 情绪检测任务实现
- **SynthesisTask**: 合成检测任务实现
- **ai_inference**: AI 推理节点类，使用 work_id 管理不同的 task 对象

## 构建步骤

### 1. 准备 Docker 环境

```bash
cd /root/meeting-system-server/meeting-system/Edge-LLM-Infra-master/docker/build
docker build -t llm:v1.0 -f base.dockerfile .
```

### 2. 下载模型

```bash
cd /root/meeting-system-server/meeting-system/Edge-LLM-Infra-master/node/llm
python3 download_models.py /work/models
```

### 3. 编译推理节点

在 Docker 容器内：

```bash
cd /work/node/llm
mkdir build && cd build
cmake .. && make -j12
```

### 4. 运行推理节点

```bash
./llm
```

## 使用示例

参考 `/work/sample/test.py` 中的请求格式：

```python
# ASR 任务
{
    "request_id": "asr_001",
    "work_id": "asr",
    "action": "setup",
    "object": "llm.setup",
    "data": {
        "model": "asr-model",
        "response_format": "llm.utf-8.stream",
        "input": "llm.utf-8.stream",
        "enoutput": True
    }
}

# 情绪检测任务
{
    "request_id": "emotion_001",
    "work_id": "emotion",
    "action": "setup",
    "object": "llm.setup",
    "data": {
        "model": "emotion-model",
        "response_format": "llm.utf-8.stream",
        "input": "llm.utf-8.stream",
        "enoutput": True
    }
}

# 合成检测任务
{
    "request_id": "synthesis_001",
    "work_id": "synthesis",
    "action": "setup",
    "object": "llm.setup",
    "data": {
        "model": "synthesis-model",
        "response_format": "llm.utf-8.stream",
        "input": "llm.utf-8.stream",
        "enoutput": True
    }
}
```

## 依赖项

- ONNX Runtime
- ZeroMQ
- nlohmann/json
- eventpp
- simdjson

## 注意事项

1. 确保模型文件位于 `/work/models/` 目录
2. 模型文件必须是 ONNX 格式
3. 根据模型名称自动选择任务类型（asr/emotion/synthesis）