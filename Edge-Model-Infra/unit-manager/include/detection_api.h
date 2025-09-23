/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include <string>
#include <memory>
#include <unordered_map>
#include <mutex>
#include "pzmq.hpp"
#include "json.hpp"

using namespace StackFlows;

class DetectionAPI {
public:
    DetectionAPI();
    ~DetectionAPI();

    // Initialize the detection API
    bool initialize();

    // Handle detection requests
    std::string handle_detect_request(const std::string& json_request);
    std::string handle_status_request(const std::string& task_id);
    std::string handle_setup_request(const std::string& json_request);

    // File upload handling
    std::string handle_file_upload(const std::string& file_data, const std::string& file_type, const std::string& filename);

private:
    // RPC communication with AI detection node
    std::unique_ptr<pzmq> detection_client_;
    
    // Task management
    struct DetectionTask {
        std::string task_id;
        std::string status;
        std::string result;
        std::chrono::system_clock::time_point created_at;
    };
    
    std::unordered_map<std::string, DetectionTask> tasks_;
    std::mutex tasks_mutex_;
    
    // Helper methods
    std::string generate_task_id();
    std::string save_uploaded_file(const std::string& file_data, const std::string& filename);
    std::string create_response(const std::string& status, const nlohmann::json& data = {});
    std::string create_error_response(const std::string& error_message, int error_code = -1);
    
    // RPC call helpers
    std::string call_detection_rpc(const std::string& action, const std::string& data);
    
    // Configuration
    std::string upload_dir_;
    bool initialized_;
};

// Global detection API instance
extern std::unique_ptr<DetectionAPI> g_detection_api;
