/*
 * Base Task Interface for AI Inference
 */
#pragma once

#include <string>
#include <vector>
#include <functional>
#include <memory>
#include "json.hpp"

namespace AIInference {

typedef std::function<void(const std::string &data, bool finish)> task_callback_t;

// Base task class that all AI tasks inherit from
class BaseTask {
public:
    std::string model_;
    std::string response_format_;
    std::vector<std::string> inputs_;
    task_callback_t out_callback_;
    bool enoutput_;
    bool enstream_;
    std::string work_id_;

    BaseTask(const std::string &workid) : work_id_(workid), enoutput_(false), enstream_(false) {}

    virtual ~BaseTask() {
        // Don't call pure virtual function from destructor
        // Derived classes should call stop() in their own destructors
    }

    void set_output(task_callback_t out_callback) {
        out_callback_ = out_callback;
    }

    bool parse_config(const nlohmann::json &config_body) {
        try {
            model_ = config_body.at("model");
            response_format_ = config_body.at("response_format");
            enoutput_ = config_body.at("enoutput");
            if (config_body.contains("input")) {
                if (config_body["input"].is_string()) {
                    inputs_.push_back(config_body["input"].get<std::string>());
                } else if (config_body["input"].is_array()) {
                    for (auto _in : config_body["input"]) {
                        inputs_.push_back(_in.get<std::string>());
                    }
                }
            }
        } catch (...) {
            return true;
        }
        enstream_ = (response_format_.find("stream") != std::string::npos);
        return false;
    }

    // Pure virtual functions that must be implemented by derived classes
    virtual int load_model(const nlohmann::json &config_body) = 0;
    virtual void inference(const std::string &msg) = 0;
    virtual void start() = 0;
    virtual void stop() = 0;
};

} // namespace AIInference

