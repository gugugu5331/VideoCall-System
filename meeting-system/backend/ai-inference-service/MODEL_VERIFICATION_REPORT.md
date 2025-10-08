# Edge-LLM-Infra 模型验证报告

**日期**: 2025-10-08  
**目标**: 验证 Edge-LLM-Infra 是否真正加载并使用了 AI 模型

---

## 📋 执行摘要

**结论**: ❌ **Edge-LLM-Infra 使用的是虚拟模型和模拟数据，不是真实的 AI 推理**

**关键发现**:
1. ❌ 模型文件是虚拟的（仅 160-198 bytes）
2. ❌ ASR 返回硬编码的固定文本
3. ⚠️ Emotion Detection 和 Synthesis Detection 使用模型输出，但模型是虚拟的
4. ✅ ONNX Runtime 成功加载了虚拟模型

---

## 🔍 详细分析

### 1. 模型文件验证

**检查命令**:
```bash
ls -lh /work/models/
```

**结果**:
```
total 12K
-rw-r--r-- 1 root root 160 Oct  7 14:44 asr-model.onnx
-rw-r--r-- 1 root root 198 Oct  7 14:44 emotion-model.onnx
-rw-r--r-- 1 root root 176 Oct  7 14:44 synthesis-model.onnx
```

**分析**:
- ❌ **文件太小**：真实的 ONNX 模型通常是几 MB 到几百 MB
- ❌ **虚拟模型**：这些文件只包含基本的模型结构（输入/输出定义），没有真实的权重数据

**文件内容**（asr-model.onnx）:
```
dummy_asr:�
-
audio_inputtranscription_output"Identity	asr_modelZ!
audio_input


d
Pb+
transcription_output


d
�B
```

**结论**: 这些是**虚拟的 ONNX 模型**，用于测试和演示，不包含真实的 AI 权重。

---

### 2. 源代码分析

#### ASR Task（语音识别）

**文件**: `node/llm/src/asr_task.cpp`  
**方法**: `postprocess_output`（第 216-239 行）

```cpp
std::string ASRTask::postprocess_output(const std::vector<float> &output) {
    // Simple postprocessing: find argmax and convert to text
    // In a real implementation, this would involve:
    // 1. CTC decoding or attention-based decoding
    // 2. Vocabulary mapping
    // 3. Language model integration
    
    if (output.empty()) {
        return "No transcription available";
    }
    
    // Find max probability
    auto max_it = std::max_element(output.begin(), output.end());
    int max_idx = std::distance(output.begin(), max_it);
    float confidence = *max_it;
    
    // Create JSON response
    nlohmann::json result;
    result["transcription"] = "Sample transcription text";  // ❌ 硬编码！
    result["confidence"] = confidence;
    result["model"] = model_;
    
    return result.dump();
}
```

**问题**:
- ❌ **第 234 行**：硬编码了固定文本 `"Sample transcription text"`
- ❌ **不使用模型输出**：虽然计算了 `confidence`，但转录文本是固定的
- ❌ **注释说明**：代码注释明确说明这是简化实现，真实实现需要 CTC 解码、词汇映射等

**结论**: ASR 返回的是**硬编码的固定文本**，不是真实的语音识别结果。

---

#### Emotion Detection（情感检测）

**文件**: `node/llm/src/emotion_task.cpp`  
**方法**: `postprocess_output`（第 216-260 行）

```cpp
std::string EmotionTask::postprocess_output(const std::vector<float> &output) {
    // Apply softmax and find the emotion with highest probability
    
    if (output.empty()) {
        return "No emotion detected";
    }
    
    // Apply softmax
    std::vector<float> probabilities;
    float sum = 0.0f;
    for (float val : output) {
        float exp_val = std::exp(val);
        probabilities.push_back(exp_val);
        sum += exp_val;
    }
    
    for (float& prob : probabilities) {
        prob /= sum;
    }
    
    // Find emotion with highest probability
    auto max_it = std::max_element(probabilities.begin(), probabilities.end());
    int max_idx = std::distance(probabilities.begin(), max_it);
    float confidence = *max_it;
    
    std::string detected_emotion = "unknown";
    if (max_idx < emotion_labels_.size()) {
        detected_emotion = emotion_labels_[max_idx];  // ✅ 使用模型输出
    }
    
    // Create JSON response
    nlohmann::json result;
    result["emotion"] = detected_emotion;
    result["confidence"] = confidence;
    result["model"] = model_;
    
    // Add all emotion probabilities
    nlohmann::json all_emotions;
    for (size_t i = 0; i < std::min(probabilities.size(), emotion_labels_.size()); i++) {
        all_emotions[emotion_labels_[i]] = probabilities[i];
    }
    result["all_emotions"] = all_emotions;
    
    return result.dump();
}
```

**分析**:
- ✅ **使用模型输出**：根据模型输出的概率选择情感
- ✅ **应用 Softmax**：正确的后处理逻辑
- ⚠️ **但模型是虚拟的**：虽然代码正确，但模型没有真实权重

**结论**: Emotion Detection 的代码实现正确，但由于模型是虚拟的，输出仍然是基于随机/固定数据。

---

#### Synthesis Detection（深度伪造检测）

**文件**: `node/llm/src/synthesis_task.cpp`  
**方法**: `postprocess_output`（第 210-238 行）

```cpp
std::string SynthesisTask::postprocess_output(const std::vector<float> &output) {
    // Binary classification: real vs synthetic
    // Apply sigmoid to get probability
    
    if (output.empty()) {
        return "No detection result available";
    }
    
    // Get the first output value (assuming binary classification)
    float raw_score = output[0];
    
    // Apply sigmoid function
    float probability = 1.0f / (1.0f + std::exp(-raw_score));
    
    // Determine if synthetic (threshold at 0.5)
    bool is_synthetic = probability > 0.5f;
    float confidence = is_synthetic ? probability : (1.0f - probability);
    
    // Create JSON response
    nlohmann::json result;
    result["is_synthetic"] = is_synthetic;
    result["is_real"] = !is_synthetic;
    result["confidence"] = confidence;
    result["probability_synthetic"] = probability;
    result["probability_real"] = 1.0f - probability;
    result["model"] = model_;
    
    return result.dump();
}
```

**分析**:
- ✅ **使用模型输出**：根据模型输出的分数判断是否为合成音频
- ✅ **应用 Sigmoid**：正确的后处理逻辑
- ⚠️ **但模型是虚拟的**：虽然代码正确，但模型没有真实权重

**结论**: Synthesis Detection 的代码实现正确，但由于模型是虚拟的，输出仍然是基于随机/固定数据。

---

## 📊 测试验证

### 测试 1: 占位符数据 vs 真实音频

| 测试项 | 输入 | 输出 | 结论 |
|--------|------|------|------|
| **ASR（占位符）** | "sample audio data" (17 bytes) | "Sample transcription text" | ❌ 固定文本 |
| **ASR（真实音频）** | 115KB MP3 文件 | "Sample transcription text" | ❌ 固定文本 |

**结论**: 无论输入什么音频数据，ASR 都返回相同的固定文本。

---

### 测试 2: 模型加载日志

**Edge-LLM-Infra 日志**:
```
[ASRTask] load_model called
[ASRTask] Config parsed successfully, model=asr-model
[ASRTask] Initializing ONNX Runtime...
[ASRTask] ONNX Runtime initialized
Loading ASR model from: /work/models/asr-model.onnx
2025-10-08 02:24:41.047231519 [W:onnxruntime:, graph.cc:108 MergeShapeInfo] Error merging shape info for output. 'transcription_output' source:{1,100,80} target:{1,100,1000}. Falling back to lenient merge.
ASR model loaded successfully
```

**分析**:
- ✅ ONNX Runtime 成功加载了模型
- ⚠️ 有警告：输出形状不匹配（`{1,100,80}` vs `{1,100,1000}`）
- ⚠️ 这说明模型结构是虚拟的，不是真实训练的模型

**结论**: ONNX Runtime 加载了虚拟模型，但模型没有真实的权重数据。

---

## 💡 问题根源

### 为什么使用虚拟模型？

**可能原因**:
1. **演示/测试目的**：Edge-LLM-Infra 是一个框架演示，不包含真实的 AI 模型
2. **模型文件太大**：真实的 AI 模型文件通常很大（几百 MB 到几 GB），不适合包含在代码仓库中
3. **许可证问题**：真实的预训练模型可能有许可证限制
4. **开发阶段**：系统仍在开发中，使用虚拟模型进行功能测试

### 虚拟模型的特征

1. **文件很小**：160-198 bytes（真实模型通常是 MB 级别）
2. **只有结构**：包含输入/输出定义，但没有权重数据
3. **Identity 操作**：虚拟模型通常使用 Identity 操作（直接传递输入）
4. **固定输出**：由于没有真实权重，输出是随机或固定的

---

## 🎯 修复建议

### 选项 1: 使用真实的 AI 模型（推荐）

**步骤**:
1. **获取真实的 ONNX 模型**：
   - ASR: 使用 Whisper、DeepSpeech 等模型
   - Emotion Detection: 使用情感分类模型
   - Synthesis Detection: 使用深度伪造检测模型

2. **转换为 ONNX 格式**：
   ```bash
   # 示例：将 PyTorch 模型转换为 ONNX
   python -m torch.onnx.export model.pth model.onnx
   ```

3. **替换虚拟模型**：
   ```bash
   cp real-asr-model.onnx /work/models/asr-model.onnx
   cp real-emotion-model.onnx /work/models/emotion-model.onnx
   cp real-synthesis-model.onnx /work/models/synthesis-model.onnx
   ```

4. **重新启动 Edge-LLM-Infra**：
   ```bash
   pkill -9 llm
   cd /root/meeting-system-server/meeting-system/Edge-LLM-Infra-master/node/llm/build
   ./llm > llm.log 2>&1 &
   ```

---

### 选项 2: 修改 ASR 代码使用模型输出

**目标**: 即使使用虚拟模型，也让 ASR 返回不同的结果

**修改文件**: `node/llm/src/asr_task.cpp`

**修改前**（第 234 行）:
```cpp
result["transcription"] = "Sample transcription text";  // 固定文本
```

**修改后**:
```cpp
// 使用模型输出生成不同的文本（虽然仍然是模拟的）
std::string transcription = "Transcription_" + std::to_string(max_idx) + "_conf_" + std::to_string(confidence);
result["transcription"] = transcription;
```

**效果**:
- 不同的音频输入会产生不同的 `max_idx` 和 `confidence`
- 返回的文本会不同（虽然仍然不是真实的转录）
- 至少可以验证系统是否真正处理了音频数据

---

### 选项 3: 保持现状（不推荐）

**说明**: 如果只是为了演示系统架构，可以保持使用虚拟模型

**优点**:
- 不需要大文件
- 系统运行快速
- 适合功能测试

**缺点**:
- ❌ 不是真实的 AI 推理
- ❌ 无法验证音频处理逻辑
- ❌ 用户会认为系统有问题

---

## 📈 真实模型 vs 虚拟模型对比

| 特征 | 虚拟模型（当前） | 真实模型（推荐） |
|------|----------------|----------------|
| **文件大小** | 160-198 bytes | 几百 MB 到几 GB |
| **权重数据** | ❌ 无 | ✅ 有 |
| **推理结果** | ❌ 固定/随机 | ✅ 真实的 AI 推理 |
| **ASR 输出** | "Sample transcription text" | 真实的语音转录 |
| **Emotion 输出** | 基于随机数据 | 真实的情感分析 |
| **Synthesis 输出** | 基于随机数据 | 真实的深度伪造检测 |
| **加载时间** | 快速（<1s） | 较慢（几秒到几十秒） |
| **内存占用** | 很小 | 较大（几百 MB 到几 GB） |
| **适用场景** | 演示、测试 | 生产环境 |

---

## 🎉 结论

### 当前状态
- ❌ **Edge-LLM-Infra 使用虚拟模型**，不是真实的 AI 推理
- ❌ **ASR 返回固定文本**，不处理音频内容
- ⚠️ **Emotion 和 Synthesis Detection 代码正确**，但模型是虚拟的
- ✅ **系统架构正确**，流式传输、数据格式都没问题

### 建议
1. **如果需要真实的 AI 推理**：替换为真实的 ONNX 模型
2. **如果只是演示系统**：可以保持现状，但需要向用户说明
3. **如果需要验证数据处理**：修改 ASR 代码使用模型输出

### 下一步
- 决定是否需要真实的 AI 模型
- 如果需要，获取或训练真实的 ONNX 模型
- 如果不需要，向用户说明当前是演示版本

---

**报告生成时间**: 2025-10-08 02:35:00  
**验证状态**: ✅ **完成 - 确认使用虚拟模型**  
**建议**: 替换为真实的 ONNX 模型以获得真实的 AI 推理结果

