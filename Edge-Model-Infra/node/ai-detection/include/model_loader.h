/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include <string>
#include <memory>
#include <vector>

#ifdef USE_TENSORFLOW
#include <tensorflow/cc/client/client_session.h>
#include <tensorflow/cc/ops/standard_ops.h>
#include <tensorflow/core/framework/tensor.h>
#include <tensorflow/core/public/session.h>
#endif

namespace StackFlows {

enum class ModelType {
    TENSORFLOW_SAVED_MODEL,
    TENSORFLOW_FROZEN_GRAPH,
    ONNX_MODEL,
    CUSTOM_MODEL
};

struct ModelInfo {
    ModelType type;
    std::string path;
    std::string input_layer;
    std::string output_layer;
    std::vector<int> input_shape;
    std::vector<int> output_shape;
};

class ModelLoader {
public:
    ModelLoader();
    virtual ~ModelLoader();

    // Load model from file
    bool load_model(const ModelInfo& model_info);

    // Check if model is loaded
    bool is_loaded() const { return model_loaded_; }

    // Get model type
    ModelType get_model_type() const { return model_info_.type; }

#ifdef USE_TENSORFLOW
    // Get TensorFlow session (if using TensorFlow)
    tensorflow::Session* get_tf_session() const { return tf_session_.get(); }
    
    // Run inference with TensorFlow
    bool run_inference(const std::vector<std::pair<std::string, tensorflow::Tensor>>& inputs,
                      const std::vector<std::string>& output_names,
                      std::vector<tensorflow::Tensor>* outputs);
#endif

    // Generic inference interface (for future ONNX support)
    bool run_inference(const std::vector<float>& input_data, std::vector<float>& output_data);

private:
#ifdef USE_TENSORFLOW
    bool load_tensorflow_model(const std::string& model_path);
    bool load_frozen_graph(const std::string& graph_path);
#endif
    
    bool load_onnx_model(const std::string& model_path);
    void create_dummy_model();

private:
    bool model_loaded_;
    ModelInfo model_info_;
    
#ifdef USE_TENSORFLOW
    std::unique_ptr<tensorflow::Session> tf_session_;
    std::unique_ptr<tensorflow::GraphDef> graph_def_;
#endif
    
    // For ONNX models (future implementation)
    // std::unique_ptr<Ort::Session> onnx_session_;
};

} // namespace StackFlows
