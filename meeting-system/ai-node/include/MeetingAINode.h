/*
 * SPDX-FileCopyrightText: 2024 Meeting System
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include "StackFlow.h"
#include <memory>
#include <unordered_map>
#include <queue>
#include <mutex>
#include <thread>
#include <atomic>
#include <chrono>

namespace MeetingAI {

// AI任务类型枚举
enum class AITaskType {
    SPEECH_RECOGNITION,    // 语音识别
    EMOTION_DETECTION,     // 情绪识别
    AUDIO_DENOISING,       // 音频降噪
    VIDEO_ENHANCEMENT,     // 视频增强
    TEXT_TO_SPEECH,        // 语音合成
    FACE_DETECTION,        // 人脸检测
    GESTURE_RECOGNITION,   // 手势识别
    AUDIO_QUALITY_ANALYSIS, // 音频质量分析
    VIDEO_QUALITY_ANALYSIS  // 视频质量分析
};

// AI任务结构
struct AITask {
    std::string task_id;
    std::string request_id;
    std::string meeting_id;
    std::string user_id;
    AITaskType type;
    std::string input_data;
    std::string output_channel;
    std::chrono::system_clock::time_point timestamp;
    int priority = 5; // 1-10, 数字越小优先级越高
    int retry_count = 0;
    
    // 比较函数，用于优先队列
    bool operator<(const AITask& other) const {
        if (priority != other.priority) {
            return priority > other.priority; // 优先级高的在前
        }
        return timestamp > other.timestamp; // 时间早的在前
    }
};

// AI模型基类
class AIModel {
public:
    virtual ~AIModel() = default;
    virtual bool initialize(const std::string& model_path) = 0;
    virtual std::string process(const std::string& input_data) = 0;
    virtual void cleanup() = 0;
    virtual bool isReady() const = 0;
};

// 语音识别模型
class SpeechRecognitionModel : public AIModel {
private:
    bool ready_ = false;
    std::string model_path_;
    
public:
    bool initialize(const std::string& model_path) override;
    std::string process(const std::string& input_data) override;
    void cleanup() override;
    bool isReady() const override { return ready_; }
};

// 情绪识别模型
class EmotionDetectionModel : public AIModel {
private:
    bool ready_ = false;
    std::string model_path_;
    
public:
    bool initialize(const std::string& model_path) override;
    std::string process(const std::string& input_data) override;
    void cleanup() override;
    bool isReady() const override { return ready_; }
};

// 音频降噪模型
class AudioDenoisingModel : public AIModel {
private:
    bool ready_ = false;
    std::string model_path_;
    
public:
    bool initialize(const std::string& model_path) override;
    std::string process(const std::string& input_data) override;
    void cleanup() override;
    bool isReady() const override { return ready_; }
};

// 视频增强模型
class VideoEnhancementModel : public AIModel {
private:
    bool ready_ = false;
    std::string model_path_;
    
public:
    bool initialize(const std::string& model_path) override;
    std::string process(const std::string& input_data) override;
    void cleanup() override;
    bool isReady() const override { return ready_; }
};

// 会议AI推理节点
class MeetingAINode : public StackFlows::StackFlow {
private:
    // AI模型管理
    std::unordered_map<AITaskType, std::shared_ptr<AIModel>> models_;
    
    // 任务队列
    std::priority_queue<AITask> task_queue_;
    std::mutex task_queue_mutex_;
    
    // 工作线程
    std::vector<std::unique_ptr<std::thread>> worker_threads_;
    std::atomic<bool> stop_workers_;
    
    // 统计信息
    std::atomic<int> processed_tasks_;
    std::atomic<int> failed_tasks_;
    std::chrono::system_clock::time_point start_time_;
    
    // 配置
    int max_workers_;
    int max_queue_size_;
    std::string model_base_path_;
    
public:
    explicit MeetingAINode(const std::string& unit_name);
    virtual ~MeetingAINode();
    
    // 重写StackFlow虚函数
    int setup(const std::string& work_id, const std::string& object, const std::string& data) override;
    int exit(const std::string& work_id, const std::string& object, const std::string& data) override;
    
    // AI任务处理
    bool addTask(const AITask& task);
    void processTasks();
    
    // 具体AI任务处理方法
    std::string processSpeechRecognition(const AITask& task);
    std::string processEmotionDetection(const AITask& task);
    std::string processAudioDenoising(const AITask& task);
    std::string processVideoEnhancement(const AITask& task);
    std::string processTextToSpeech(const AITask& task);
    std::string processFaceDetection(const AITask& task);
    std::string processGestureRecognition(const AITask& task);
    std::string processAudioQualityAnalysis(const AITask& task);
    std::string processVideoQualityAnalysis(const AITask& task);
    
    // 模型管理
    bool loadModel(AITaskType type, const std::string& model_path);
    void unloadModel(AITaskType type);
    bool isModelReady(AITaskType type) const;
    
    // 工作线程管理
    void startWorkers();
    void stopWorkers();
    void workerFunction();
    
    // 统计信息
    int getProcessedTaskCount() const { return processed_tasks_.load(); }
    int getFailedTaskCount() const { return failed_tasks_.load(); }
    int getQueueSize() const;
    double getUptime() const;
    
    // 配置管理
    void setMaxWorkers(int max_workers) { max_workers_ = max_workers; }
    void setMaxQueueSize(int max_queue_size) { max_queue_size_ = max_queue_size; }
    void setModelBasePath(const std::string& path) { model_base_path_ = path; }
    
private:
    // 辅助方法
    AITaskType stringToTaskType(const std::string& type_str);
    std::string taskTypeToString(AITaskType type);
    std::string createErrorResponse(const std::string& error_message);
    std::string createSuccessResponse(const std::string& result_data);
    bool validateTaskData(const AITask& task);
    void sendTaskResult(const AITask& task, const std::string& result);
    void logTaskProcessing(const AITask& task, bool success, const std::string& message = "");
    
    // JSON处理
    std::string parseInputData(const std::string& json_data, const std::string& key);
    std::string createOutputData(const std::string& result, double confidence = 1.0);
};

// 任务工厂
class AITaskFactory {
public:
    static AITask createSpeechRecognitionTask(
        const std::string& request_id,
        const std::string& meeting_id,
        const std::string& user_id,
        const std::string& audio_data,
        const std::string& output_channel
    );
    
    static AITask createEmotionDetectionTask(
        const std::string& request_id,
        const std::string& meeting_id,
        const std::string& user_id,
        const std::string& image_data,
        const std::string& output_channel
    );
    
    static AITask createAudioDenoisingTask(
        const std::string& request_id,
        const std::string& meeting_id,
        const std::string& user_id,
        const std::string& audio_data,
        const std::string& output_channel
    );
    
    static AITask createVideoEnhancementTask(
        const std::string& request_id,
        const std::string& meeting_id,
        const std::string& user_id,
        const std::string& video_data,
        const std::string& output_channel
    );
};

// 性能监控器
class PerformanceMonitor {
private:
    std::chrono::system_clock::time_point last_report_time_;
    int last_processed_count_;
    int last_failed_count_;
    
public:
    PerformanceMonitor();
    void reportMetrics(const MeetingAINode& node);
    void logPerformanceStats(const MeetingAINode& node);
};

} // namespace MeetingAI
