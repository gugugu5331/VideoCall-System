/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "model_loader.h"
#include "detection_utils.h"
#include <iostream>
#include <fstream>

using namespace StackFlows;

ModelLoader::ModelLoader() : model_loaded_(false) {
}

ModelLoader::~ModelLoader() {
#ifdef USE_TENSORFLOW
    if (tf_session_) {
        tf_session_->Close();
    }
#endif
}

bool ModelLoader::load_model(const ModelInfo& model_info) {
    model_info_ = model_info;
    
    if (!DetectionUtils::file_exists(model_info.path)) {
        std::cerr << "Model file not found: " << model_info.path << std::endl;
        create_dummy_model();
        return true;
    }
    
    switch (model_info.type) {
        case ModelType::TENSORFLOW_SAVED_MODEL:
        case ModelType::TENSORFLOW_FROZEN_GRAPH:
#ifdef USE_TENSORFLOW
            return load_tensorflow_model(model_info.path);
#else
            std::cout << "TensorFlow not available, using dummy model" << std::endl;
            create_dummy_model();
            return true;
#endif
            
        case ModelType::ONNX_MODEL:
            return load_onnx_model(model_info.path);
            
        case ModelType::CUSTOM_MODEL:
        default:
            create_dummy_model();
            return true;
    }
}

bool ModelLoader::run_inference(const std::vector<float>& input_data, std::vector<float>& output_data) {
    if (!model_loaded_) {
        std::cerr << "Model not loaded" << std::endl;
        return false;
    }
    
#ifdef USE_TENSORFLOW
    if (tf_session_ && model_info_.type == ModelType::TENSORFLOW_SAVED_MODEL || 
        model_info_.type == ModelType::TENSORFLOW_FROZEN_GRAPH) {
        
        // Create input tensor
        tensorflow::Tensor input_tensor(tensorflow::DT_FLOAT, 
            tensorflow::TensorShape({1, static_cast<int>(input_data.size())}));
        
        auto input_tensor_mapped = input_tensor.tensor<float, 2>();
        for (size_t i = 0; i < input_data.size(); ++i) {
            input_tensor_mapped(0, i) = input_data[i];
        }
        
        // Run inference
        std::vector<tensorflow::Tensor> outputs;
        tensorflow::Status status = tf_session_->Run(
            {{model_info_.input_layer, input_tensor}},
            {model_info_.output_layer},
            {},
            &outputs
        );
        
        if (!status.ok()) {
            std::cerr << "TensorFlow inference failed: " << status.ToString() << std::endl;
            return false;
        }
        
        if (outputs.empty()) {
            std::cerr << "No outputs from TensorFlow model" << std::endl;
            return false;
        }
        
        // Extract output data
        auto output_tensor = outputs[0].tensor<float, 2>();
        int output_size = outputs[0].dim_size(1);
        
        output_data.clear();
        output_data.reserve(output_size);
        for (int i = 0; i < output_size; ++i) {
            output_data.push_back(output_tensor(0, i));
        }
        
        return true;
    }
#endif
    
    // Fallback: dummy inference
    output_data.clear();
    output_data.push_back(0.5f); // Dummy output
    return true;
}

#ifdef USE_TENSORFLOW
bool ModelLoader::run_inference(const std::vector<std::pair<std::string, tensorflow::Tensor>>& inputs,
                               const std::vector<std::string>& output_names,
                               std::vector<tensorflow::Tensor>* outputs) {
    if (!tf_session_) {
        std::cerr << "TensorFlow session not available" << std::endl;
        return false;
    }
    
    tensorflow::Status status = tf_session_->Run(inputs, output_names, {}, outputs);
    
    if (!status.ok()) {
        std::cerr << "TensorFlow inference failed: " << status.ToString() << std::endl;
        return false;
    }
    
    return true;
}

bool ModelLoader::load_tensorflow_model(const std::string& model_path) {
    try {
        tensorflow::SessionOptions options;
        tf_session_.reset(tensorflow::NewSession(options));
        
        if (model_info_.type == ModelType::TENSORFLOW_SAVED_MODEL) {
            // Load SavedModel
            tensorflow::SavedModelBundle bundle;
            tensorflow::Status status = tensorflow::LoadSavedModel(
                options, tensorflow::RunOptions(), model_path, {"serve"}, &bundle);
            
            if (!status.ok()) {
                std::cerr << "Failed to load SavedModel: " << status.ToString() << std::endl;
                return false;
            }
            
            tf_session_ = std::move(bundle.session);
            
        } else if (model_info_.type == ModelType::TENSORFLOW_FROZEN_GRAPH) {
            // Load frozen graph
            return load_frozen_graph(model_path);
        }
        
        model_loaded_ = true;
        std::cout << "TensorFlow model loaded successfully: " << model_path << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Exception loading TensorFlow model: " << e.what() << std::endl;
        return false;
    }
}

bool ModelLoader::load_frozen_graph(const std::string& graph_path) {
    try {
        graph_def_ = std::make_unique<tensorflow::GraphDef>();
        tensorflow::Status status = tensorflow::ReadBinaryProto(
            tensorflow::Env::Default(), graph_path, graph_def_.get());
        
        if (!status.ok()) {
            std::cerr << "Failed to read frozen graph: " << status.ToString() << std::endl;
            return false;
        }
        
        status = tf_session_->Create(*graph_def_);
        if (!status.ok()) {
            std::cerr << "Failed to create session from graph: " << status.ToString() << std::endl;
            return false;
        }
        
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Exception loading frozen graph: " << e.what() << std::endl;
        return false;
    }
}
#endif

bool ModelLoader::load_onnx_model(const std::string& model_path) {
    // ONNX Runtime integration would go here
    // For now, just use dummy model
    std::cout << "ONNX Runtime not implemented, using dummy model" << std::endl;
    create_dummy_model();
    return true;
}

void ModelLoader::create_dummy_model() {
    model_loaded_ = true;
    std::cout << "Using dummy model for inference" << std::endl;
}
