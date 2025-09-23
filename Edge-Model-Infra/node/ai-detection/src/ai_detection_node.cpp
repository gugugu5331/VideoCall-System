/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "ai_detection_node.h"
#include "detection_utils.h"
#include <fstream>
#include <sstream>
#include <random>

using namespace StackFlows;

AIDetectionNode::AIDetectionNode(const std::string& unit_name) 
    : StackFlow(unit_name) {
    
    // Initialize detection services
    face_detector_ = std::make_unique<FaceSwapDetector>();
    voice_detector_ = std::make_unique<VoiceSynthesisDetector>();
    content_analyzer_ = std::make_unique<ContentAnalyzer>();
    
    // Register RPC actions
    register_rpc_actions();
    
    std::cout << "AIDetectionNode initialized with unit name: " << unit_name << std::endl;
}

AIDetectionNode::~AIDetectionNode() {
    std::cout << "AIDetectionNode destructor called" << std::endl;
}

int AIDetectionNode::setup(const std::string& work_id, const std::string& object, const std::string& data) {
    std::cout << "AIDetectionNode setup called - work_id: " << work_id 
              << ", object: " << object << std::endl;
    
    // Load configuration
    if (!load_configuration(data)) {
        std::cerr << "Failed to load configuration" << std::endl;
        return -1;
    }
    
    return 0;
}

int AIDetectionNode::exit(const std::string& work_id, const std::string& object, const std::string& data) {
    std::cout << "AIDetectionNode exit called - work_id: " << work_id << std::endl;
    return 0;
}

void AIDetectionNode::register_rpc_actions() {
    if (!rpc_ctx_) {
        rpc_ctx_ = std::make_unique<pzmq>(unit_name_);
    }
    
    // Register detection RPC actions
    rpc_ctx_->register_rpc_action("setup_face_detector", 
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_setup_face_detector(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("setup_voice_detector",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_setup_voice_detector(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("detect_image",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_detect_image(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("detect_audio",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_detect_audio(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("detect_video",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_detect_video(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("analyze_content",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_analyze_content(zmq_obj, data);
        });
    
    rpc_ctx_->register_rpc_action("get_detection_status",
        [this](pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
            return this->rpc_get_detection_status(zmq_obj, data);
        });
    
    std::cout << "RPC actions registered successfully" << std::endl;
}

std::string AIDetectionNode::rpc_setup_face_detector(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string model_path = data->get_param(0);
        
        if (face_detector_->initialize(model_path)) {
            return DetectionUtils::create_detection_response(true, 1.0f, "Face detector initialized successfully");
        } else {
            return DetectionUtils::create_error_response("Failed to initialize face detector");
        }
    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in setup_face_detector: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::rpc_setup_voice_detector(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string model_path = data->get_param(0);
        
        if (voice_detector_->initialize(model_path)) {
            return DetectionUtils::create_detection_response(true, 1.0f, "Voice detector initialized successfully");
        } else {
            return DetectionUtils::create_error_response("Failed to initialize voice detector");
        }
    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in setup_voice_detector: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::rpc_detect_image(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string image_path = data->get_param(0);
        std::string task_id = generate_task_id();
        
        // Create detection task
        DetectionTask task;
        task.task_id = task_id;
        task.file_path = image_path;
        task.file_type = "image";
        task.status = "processing";
        task.created_at = std::chrono::system_clock::now();
        
        {
            std::lock_guard<std::mutex> lock(tasks_mutex_);
            detection_tasks_[task_id] = task;
        }
        
        // Load and process image
        cv::Mat image = cv::imread(image_path);
        if (image.empty()) {
            update_task_status(task_id, "failed", "Failed to load image");
            return DetectionUtils::create_error_response("Failed to load image");
        }
        
        // Perform detection
        DetectionResult result = face_detector_->detect_image(image);
        
        // Update task with result
        std::string result_json = DetectionUtils::create_detection_response(
            result.is_fake, result.confidence, result.details);
        update_task_status(task_id, "completed", result_json);
        
        return DetectionUtils::create_task_status_response(task_id, "completed", result_json);
        
    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in detect_image: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::generate_task_id() {
    return DetectionUtils::generate_uuid();
}

void AIDetectionNode::update_task_status(const std::string& task_id, const std::string& status, const std::string& result) {
    std::lock_guard<std::mutex> lock(tasks_mutex_);
    auto it = detection_tasks_.find(task_id);
    if (it != detection_tasks_.end()) {
        it->second.status = status;
        it->second.result = result;
    }
}

std::string AIDetectionNode::rpc_detect_audio(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string audio_path = data->get_param(0);
        std::string task_id = generate_task_id();

        // Create detection task
        DetectionTask task;
        task.task_id = task_id;
        task.file_path = audio_path;
        task.file_type = "audio";
        task.status = "processing";
        task.created_at = std::chrono::system_clock::now();

        {
            std::lock_guard<std::mutex> lock(tasks_mutex_);
            detection_tasks_[task_id] = task;
        }

        // Perform detection
        AudioDetectionResult result = voice_detector_->detect_audio(audio_path);

        // Update task with result
        std::string result_json = DetectionUtils::create_detection_response(
            result.is_fake, result.confidence, result.details);
        update_task_status(task_id, "completed", result_json);

        return DetectionUtils::create_task_status_response(task_id, "completed", result_json);

    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in detect_audio: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::rpc_detect_video(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string video_path = data->get_param(0);
        std::string task_id = generate_task_id();

        // Create detection task
        DetectionTask task;
        task.task_id = task_id;
        task.file_path = video_path;
        task.file_type = "video";
        task.status = "processing";
        task.created_at = std::chrono::system_clock::now();

        {
            std::lock_guard<std::mutex> lock(tasks_mutex_);
            detection_tasks_[task_id] = task;
        }

        // Perform detection
        DetectionResult result = face_detector_->detect_video(video_path);

        // Update task with result
        std::string result_json = DetectionUtils::create_detection_response(
            result.is_fake, result.confidence, result.details);
        update_task_status(task_id, "completed", result_json);

        return DetectionUtils::create_task_status_response(task_id, "completed", result_json);

    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in detect_video: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::rpc_analyze_content(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string video_path = data->get_param(0);
        std::string task_id = generate_task_id();

        // Create analysis task
        DetectionTask task;
        task.task_id = task_id;
        task.file_path = video_path;
        task.file_type = "content_analysis";
        task.status = "processing";
        task.created_at = std::chrono::system_clock::now();

        {
            std::lock_guard<std::mutex> lock(tasks_mutex_);
            detection_tasks_[task_id] = task;
        }

        // Perform content analysis
        ContentAnalysisResult result = content_analyzer_->analyze_video(video_path);

        // Create result JSON
        nlohmann::json result_json;
        result_json["summary"] = result.summary;
        result_json["emotions_count"] = result.emotions.size();
        result_json["motion_segments"] = result.motion_data.size();
        result_json["voice_activity_segments"] = result.voice_activity.size();
        result_json["scene_changes"] = result.scene_changes.size();

        std::string result_str = result_json.dump();
        update_task_status(task_id, "completed", result_str);

        return DetectionUtils::create_task_status_response(task_id, "completed", result_str);

    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in analyze_content: " + std::string(e.what()));
    }
}

std::string AIDetectionNode::rpc_get_detection_status(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data) {
    try {
        std::string task_id = data->get_param(0);

        std::lock_guard<std::mutex> lock(tasks_mutex_);
        auto it = detection_tasks_.find(task_id);
        if (it != detection_tasks_.end()) {
            return DetectionUtils::create_task_status_response(
                task_id, it->second.status, it->second.result);
        } else {
            return DetectionUtils::create_error_response("Task not found");
        }

    } catch (const std::exception& e) {
        return DetectionUtils::create_error_response("Exception in get_detection_status: " + std::string(e.what()));
    }
}

bool AIDetectionNode::load_configuration(const std::string& config_data) {
    try {
        // Parse JSON configuration
        nlohmann::json config = nlohmann::json::parse(config_data);

        // Load model configurations if provided
        if (config.contains("models")) {
            auto models = config["models"];

            if (models.contains("face_swap_detector")) {
                std::string model_path = models["face_swap_detector"]["model_path"];
                face_detector_->initialize(model_path);
            }

            if (models.contains("voice_synthesis_detector")) {
                std::string model_path = models["voice_synthesis_detector"]["model_path"];
                voice_detector_->initialize(model_path);
            }
        }

        return true;
    } catch (const std::exception& e) {
        std::cerr << "Failed to parse configuration: " << e.what() << std::endl;
        return false;
    }
}
