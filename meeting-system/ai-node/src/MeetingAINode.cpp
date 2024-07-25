/*
 * SPDX-FileCopyrightText: 2024 Meeting System
 * SPDX-License-Identifier: MIT
 */

#include "MeetingAINode.h"
#include "json.hpp"
#include <iostream>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <glog/glog.h>

using json = nlohmann::json;
using namespace MeetingAI;

MeetingAINode::MeetingAINode(const std::string& unit_name)
    : StackFlow(unit_name)
    , stop_workers_(false)
    , processed_tasks_(0)
    , failed_tasks_(0)
    , max_workers_(4)
    , max_queue_size_(1000)
    , model_base_path_("./models/")
{
    start_time_ = std::chrono::system_clock::now();
    LOG(INFO) << "MeetingAINode initialized: " << unit_name;
}

MeetingAINode::~MeetingAINode() {
    stopWorkers();
    
    // 清理模型
    for (auto& [type, model] : models_) {
        if (model) {
            model->cleanup();
        }
    }
    
    LOG(INFO) << "MeetingAINode destroyed";
}

int MeetingAINode::setup(const std::string& work_id, const std::string& object, const std::string& data) {
    LOG(INFO) << "Setting up MeetingAINode - work_id: " << work_id << ", object: " << object;
    
    try {
        json config = json::parse(data);
        
        // 解析配置
        if (config.contains("max_workers")) {
            max_workers_ = config["max_workers"];
        }
        if (config.contains("max_queue_size")) {
            max_queue_size_ = config["max_queue_size"];
        }
        if (config.contains("model_base_path")) {
            model_base_path_ = config["model_base_path"];
        }
        
        // 加载AI模型
        loadModel(AITaskType::SPEECH_RECOGNITION, model_base_path_ + "speech_recognition.model");
        loadModel(AITaskType::EMOTION_DETECTION, model_base_path_ + "emotion_detection.model");
        loadModel(AITaskType::AUDIO_DENOISING, model_base_path_ + "audio_denoising.model");
        loadModel(AITaskType::VIDEO_ENHANCEMENT, model_base_path_ + "video_enhancement.model");
        
        // 启动工作线程
        startWorkers();
        
        LOG(INFO) << "MeetingAINode setup completed successfully";
        return 0;
        
    } catch (const std::exception& e) {
        LOG(ERROR) << "Failed to setup MeetingAINode: " << e.what();
        return -1;
    }
}

int MeetingAINode::exit(const std::string& work_id, const std::string& object, const std::string& data) {
    LOG(INFO) << "Exiting MeetingAINode - work_id: " << work_id;
    
    stopWorkers();
    
    // 清理模型
    for (auto& [type, model] : models_) {
        if (model) {
            model->cleanup();
        }
    }
    models_.clear();
    
    LOG(INFO) << "MeetingAINode exit completed";
    return 0;
}

bool MeetingAINode::addTask(const AITask& task) {
    std::lock_guard<std::mutex> lock(task_queue_mutex_);
    
    if (task_queue_.size() >= max_queue_size_) {
        LOG(WARNING) << "Task queue is full, rejecting task: " << task.task_id;
        return false;
    }
    
    if (!validateTaskData(task)) {
        LOG(ERROR) << "Invalid task data: " << task.task_id;
        return false;
    }
    
    task_queue_.push(task);
    LOG(INFO) << "Task added to queue: " << task.task_id << ", type: " << taskTypeToString(task.type);
    return true;
}

void MeetingAINode::processTasks() {
    while (!stop_workers_) {
        AITask task;
        bool has_task = false;
        
        // 从队列中获取任务
        {
            std::lock_guard<std::mutex> lock(task_queue_mutex_);
            if (!task_queue_.empty()) {
                task = task_queue_.top();
                task_queue_.pop();
                has_task = true;
            }
        }
        
        if (!has_task) {
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            continue;
        }
        
        // 处理任务
        std::string result;
        bool success = false;
        
        try {
            switch (task.type) {
                case AITaskType::SPEECH_RECOGNITION:
                    result = processSpeechRecognition(task);
                    success = true;
                    break;
                case AITaskType::EMOTION_DETECTION:
                    result = processEmotionDetection(task);
                    success = true;
                    break;
                case AITaskType::AUDIO_DENOISING:
                    result = processAudioDenoising(task);
                    success = true;
                    break;
                case AITaskType::VIDEO_ENHANCEMENT:
                    result = processVideoEnhancement(task);
                    success = true;
                    break;
                default:
                    result = createErrorResponse("Unsupported task type");
                    success = false;
                    break;
            }
        } catch (const std::exception& e) {
            result = createErrorResponse(e.what());
            success = false;
        }
        
        // 发送结果
        sendTaskResult(task, result);
        
        // 更新统计
        if (success) {
            processed_tasks_++;
        } else {
            failed_tasks_++;
        }
        
        logTaskProcessing(task, success, success ? "Task completed" : "Task failed");
    }
}

std::string MeetingAINode::processSpeechRecognition(const AITask& task) {
    auto model = models_.find(AITaskType::SPEECH_RECOGNITION);
    if (model == models_.end() || !model->second->isReady()) {
        return createErrorResponse("Speech recognition model not available");
    }
    
    try {
        json input = json::parse(task.input_data);
        std::string audio_data = input["audio_data"];
        
        // 调用模型进行语音识别
        std::string recognition_result = model->second->process(audio_data);
        
        // 创建输出结果
        json output;
        output["text"] = recognition_result;
        output["confidence"] = 0.95; // 模拟置信度
        output["language"] = "zh-CN";
        output["timestamp"] = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
        
        return createSuccessResponse(output.dump());
        
    } catch (const std::exception& e) {
        return createErrorResponse("Speech recognition processing failed: " + std::string(e.what()));
    }
}

std::string MeetingAINode::processEmotionDetection(const AITask& task) {
    auto model = models_.find(AITaskType::EMOTION_DETECTION);
    if (model == models_.end() || !model->second->isReady()) {
        return createErrorResponse("Emotion detection model not available");
    }
    
    try {
        json input = json::parse(task.input_data);
        std::string image_data = input["image_data"];
        
        // 调用模型进行情绪识别
        std::string emotion_result = model->second->process(image_data);
        
        // 创建输出结果
        json output;
        output["emotion"] = emotion_result;
        output["confidence"] = 0.88;
        output["emotions"] = json::array({
            {{"emotion", "happy"}, {"confidence", 0.88}},
            {{"emotion", "neutral"}, {"confidence", 0.12}}
        });
        output["timestamp"] = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
        
        return createSuccessResponse(output.dump());
        
    } catch (const std::exception& e) {
        return createErrorResponse("Emotion detection processing failed: " + std::string(e.what()));
    }
}

std::string MeetingAINode::processAudioDenoising(const AITask& task) {
    auto model = models_.find(AITaskType::AUDIO_DENOISING);
    if (model == models_.end() || !model->second->isReady()) {
        return createErrorResponse("Audio denoising model not available");
    }
    
    try {
        json input = json::parse(task.input_data);
        std::string audio_data = input["audio_data"];
        
        // 调用模型进行音频降噪
        std::string denoised_audio = model->second->process(audio_data);
        
        // 创建输出结果
        json output;
        output["denoised_audio"] = denoised_audio;
        output["noise_reduction_db"] = 15.5;
        output["quality_score"] = 0.92;
        output["timestamp"] = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
        
        return createSuccessResponse(output.dump());
        
    } catch (const std::exception& e) {
        return createErrorResponse("Audio denoising processing failed: " + std::string(e.what()));
    }
}

std::string MeetingAINode::processVideoEnhancement(const AITask& task) {
    auto model = models_.find(AITaskType::VIDEO_ENHANCEMENT);
    if (model == models_.end() || !model->second->isReady()) {
        return createErrorResponse("Video enhancement model not available");
    }
    
    try {
        json input = json::parse(task.input_data);
        std::string video_data = input["video_data"];
        
        // 调用模型进行视频增强
        std::string enhanced_video = model->second->process(video_data);
        
        // 创建输出结果
        json output;
        output["enhanced_video"] = enhanced_video;
        output["enhancement_type"] = "super_resolution";
        output["quality_improvement"] = 0.85;
        output["timestamp"] = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
        
        return createSuccessResponse(output.dump());
        
    } catch (const std::exception& e) {
        return createErrorResponse("Video enhancement processing failed: " + std::string(e.what()));
    }
}

bool MeetingAINode::loadModel(AITaskType type, const std::string& model_path) {
    std::shared_ptr<AIModel> model;
    
    switch (type) {
        case AITaskType::SPEECH_RECOGNITION:
            model = std::make_shared<SpeechRecognitionModel>();
            break;
        case AITaskType::EMOTION_DETECTION:
            model = std::make_shared<EmotionDetectionModel>();
            break;
        case AITaskType::AUDIO_DENOISING:
            model = std::make_shared<AudioDenoisingModel>();
            break;
        case AITaskType::VIDEO_ENHANCEMENT:
            model = std::make_shared<VideoEnhancementModel>();
            break;
        default:
            LOG(ERROR) << "Unsupported model type: " << static_cast<int>(type);
            return false;
    }
    
    if (model->initialize(model_path)) {
        models_[type] = model;
        LOG(INFO) << "Model loaded successfully: " << taskTypeToString(type);
        return true;
    } else {
        LOG(ERROR) << "Failed to load model: " << taskTypeToString(type);
        return false;
    }
}

void MeetingAINode::startWorkers() {
    stop_workers_ = false;
    
    for (int i = 0; i < max_workers_; ++i) {
        worker_threads_.emplace_back(
            std::make_unique<std::thread>(&MeetingAINode::workerFunction, this)
        );
    }
    
    LOG(INFO) << "Started " << max_workers_ << " worker threads";
}

void MeetingAINode::stopWorkers() {
    stop_workers_ = true;
    
    for (auto& thread : worker_threads_) {
        if (thread && thread->joinable()) {
            thread->join();
        }
    }
    
    worker_threads_.clear();
    LOG(INFO) << "All worker threads stopped";
}

void MeetingAINode::workerFunction() {
    LOG(INFO) << "Worker thread started";
    processTasks();
    LOG(INFO) << "Worker thread stopped";
}
