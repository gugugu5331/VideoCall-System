/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "detection_api.h"
#include <fstream>
#include <sstream>
#include <random>
#include <iomanip>
#include <filesystem>
#include <base64.h>

// Global instance
std::unique_ptr<DetectionAPI> g_detection_api = nullptr;

DetectionAPI::DetectionAPI() : initialized_(false), upload_dir_("/tmp/detection_uploads") {
    // Create upload directory if it doesn't exist
    std::filesystem::create_directories(upload_dir_);
}

DetectionAPI::~DetectionAPI() {
    if (detection_client_) {
        detection_client_.reset();
    }
}

bool DetectionAPI::initialize() {
    try {
        // Initialize RPC client to communicate with AI detection node
        detection_client_ = std::make_unique<pzmq>("ai-detection");
        
        initialized_ = true;
        std::cout << "DetectionAPI initialized successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Failed to initialize DetectionAPI: " << e.what() << std::endl;
        return false;
    }
}

std::string DetectionAPI::handle_detect_request(const std::string& json_request) {
    if (!initialized_) {
        return create_error_response("Detection API not initialized", -1);
    }
    
    try {
        nlohmann::json request = nlohmann::json::parse(json_request);
        
        std::string file_path = request.value("file_path", "");
        std::string file_type = request.value("file_type", "");
        std::string detection_type = request.value("detection_type", "auto");
        
        if (file_path.empty()) {
            return create_error_response("Missing file_path parameter", -2);
        }
        
        // Determine detection action based on file type
        std::string action;
        if (file_type == "image" || detection_type == "image") {
            action = "detect_image";
        } else if (file_type == "audio" || detection_type == "audio") {
            action = "detect_audio";
        } else if (file_type == "video" || detection_type == "video") {
            action = "detect_video";
        } else {
            // Auto-detect based on file extension
            std::string ext = file_path.substr(file_path.find_last_of('.'));
            std::transform(ext.begin(), ext.end(), ext.begin(), ::tolower);
            
            if (ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".bmp") {
                action = "detect_image";
            } else if (ext == ".wav" || ext == ".mp3" || ext == ".flac" || ext == ".ogg") {
                action = "detect_audio";
            } else if (ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".mkv") {
                action = "detect_video";
            } else {
                return create_error_response("Unsupported file type", -3);
            }
        }
        
        // Call detection RPC
        std::string rpc_response = call_detection_rpc(action, file_path);
        
        // Parse response and extract task ID
        nlohmann::json response_json = nlohmann::json::parse(rpc_response);
        
        return create_response("accepted", response_json);
        
    } catch (const std::exception& e) {
        return create_error_response("Error processing detection request: " + std::string(e.what()), -4);
    }
}

std::string DetectionAPI::handle_status_request(const std::string& task_id) {
    if (!initialized_) {
        return create_error_response("Detection API not initialized", -1);
    }
    
    try {
        // Call status RPC
        std::string rpc_response = call_detection_rpc("get_detection_status", task_id);
        
        nlohmann::json response_json = nlohmann::json::parse(rpc_response);
        
        return create_response("success", response_json);
        
    } catch (const std::exception& e) {
        return create_error_response("Error getting task status: " + std::string(e.what()), -4);
    }
}

std::string DetectionAPI::handle_setup_request(const std::string& json_request) {
    if (!initialized_) {
        return create_error_response("Detection API not initialized", -1);
    }
    
    try {
        nlohmann::json request = nlohmann::json::parse(json_request);
        
        std::string detector_type = request.value("detector_type", "");
        std::string model_path = request.value("model_path", "");
        
        if (detector_type.empty() || model_path.empty()) {
            return create_error_response("Missing detector_type or model_path parameter", -2);
        }
        
        std::string action;
        if (detector_type == "face_swap") {
            action = "setup_face_detector";
        } else if (detector_type == "voice_synthesis") {
            action = "setup_voice_detector";
        } else {
            return create_error_response("Unsupported detector type", -3);
        }
        
        // Call setup RPC
        std::string rpc_response = call_detection_rpc(action, model_path);
        
        nlohmann::json response_json = nlohmann::json::parse(rpc_response);
        
        return create_response("success", response_json);
        
    } catch (const std::exception& e) {
        return create_error_response("Error setting up detector: " + std::string(e.what()), -4);
    }
}

std::string DetectionAPI::handle_file_upload(const std::string& file_data, const std::string& file_type, const std::string& filename) {
    try {
        // Save uploaded file
        std::string file_path = save_uploaded_file(file_data, filename);
        
        if (file_path.empty()) {
            return create_error_response("Failed to save uploaded file", -5);
        }
        
        // Create detection request
        nlohmann::json detect_request;
        detect_request["file_path"] = file_path;
        detect_request["file_type"] = file_type;
        
        return handle_detect_request(detect_request.dump());
        
    } catch (const std::exception& e) {
        return create_error_response("Error handling file upload: " + std::string(e.what()), -4);
    }
}

std::string DetectionAPI::generate_task_id() {
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<> dis(0, 15);
    
    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < 8; i++) {
        ss << dis(gen);
    }
    ss << "-";
    for (int i = 0; i < 4; i++) {
        ss << dis(gen);
    }
    ss << "-4";
    for (int i = 0; i < 3; i++) {
        ss << dis(gen);
    }
    ss << "-";
    ss << dis(gen);
    for (int i = 0; i < 3; i++) {
        ss << dis(gen);
    }
    ss << "-";
    for (int i = 0; i < 12; i++) {
        ss << dis(gen);
    }
    return ss.str();
}

std::string DetectionAPI::save_uploaded_file(const std::string& file_data, const std::string& filename) {
    try {
        // Generate unique filename
        std::string unique_filename = generate_task_id() + "_" + filename;
        std::string file_path = upload_dir_ + "/" + unique_filename;
        
        // Decode base64 data (assuming file_data is base64 encoded)
        // For simplicity, we'll assume it's already binary data
        std::ofstream file(file_path, std::ios::binary);
        if (!file.is_open()) {
            std::cerr << "Failed to open file for writing: " << file_path << std::endl;
            return "";
        }
        
        file.write(file_data.c_str(), file_data.size());
        file.close();
        
        return file_path;
        
    } catch (const std::exception& e) {
        std::cerr << "Error saving file: " << e.what() << std::endl;
        return "";
    }
}

std::string DetectionAPI::create_response(const std::string& status, const nlohmann::json& data) {
    nlohmann::json response;
    response["status"] = status;
    response["timestamp"] = std::chrono::duration_cast<std::chrono::seconds>(
        std::chrono::system_clock::now().time_since_epoch()).count();
    
    if (!data.empty()) {
        response["data"] = data;
    }
    
    return response.dump();
}

std::string DetectionAPI::create_error_response(const std::string& error_message, int error_code) {
    nlohmann::json response;
    response["status"] = "error";
    response["error"] = {
        {"code", error_code},
        {"message", error_message}
    };
    response["timestamp"] = std::chrono::duration_cast<std::chrono::seconds>(
        std::chrono::system_clock::now().time_since_epoch()).count();
    
    return response.dump();
}

std::string DetectionAPI::call_detection_rpc(const std::string& action, const std::string& data) {
    if (!detection_client_) {
        throw std::runtime_error("Detection client not initialized");
    }
    
    std::string response;
    
    int result = detection_client_->call_rpc_action(action, data, 
        [&response](pzmq* self, const std::shared_ptr<pzmq_data>& msg) {
            response = msg->string();
        });
    
    if (result != 0) {
        throw std::runtime_error("RPC call failed with code: " + std::to_string(result));
    }
    
    return response;
}
