/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include "StackFlow.h"
#include "channel.h"
#include "face_swap_detector.h"
#include "voice_synthesis_detector.h"
#include "content_analyzer.h"
#include <memory>
#include <unordered_map>
#include <string>

namespace StackFlows {

class AIDetectionNode : public StackFlow {
public:
    AIDetectionNode(const std::string& unit_name);
    virtual ~AIDetectionNode();

    // Override StackFlow methods
    virtual int setup(const std::string& work_id, const std::string& object, const std::string& data) override;
    virtual int exit(const std::string& work_id, const std::string& object, const std::string& data) override;

    // Detection task management
    struct DetectionTask {
        std::string task_id;
        std::string file_path;
        std::string file_type;
        std::string status;
        std::string result;
        std::chrono::system_clock::time_point created_at;
    };

private:
    // Detection services
    std::unique_ptr<FaceSwapDetector> face_detector_;
    std::unique_ptr<VoiceSynthesisDetector> voice_detector_;
    std::unique_ptr<ContentAnalyzer> content_analyzer_;

    // Task management
    std::unordered_map<std::string, DetectionTask> detection_tasks_;
    std::mutex tasks_mutex_;

    // RPC action handlers
    std::string rpc_setup_face_detector(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_setup_voice_detector(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_detect_image(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_detect_audio(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_detect_video(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_analyze_content(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);
    std::string rpc_get_detection_status(pzmq* zmq_obj, const std::shared_ptr<pzmq_data>& data);

    // Helper methods
    void register_rpc_actions();
    std::string generate_task_id();
    void update_task_status(const std::string& task_id, const std::string& status, const std::string& result = "");
    bool load_configuration(const std::string& config_data);
};

} // namespace StackFlows
