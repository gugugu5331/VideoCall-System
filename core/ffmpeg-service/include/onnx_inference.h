#pragma once

#include <memory>
#include <string>
#include <vector>
#include <unordered_map>
#include <chrono>
#include <mutex>

// ONNX Runtime
#include <onnxruntime_cxx_api.h>

namespace ffmpeg_detection {

// 推理设备
enum class InferenceDevice {
    CPU,
    CUDA,
    DirectML,
    OpenVINO
};

// 模型优化级别
enum class OptimizationLevel {
    NONE,
    BASIC,
    EXTENDED,
    ALL
};

// 推理配置
struct InferenceConfig {
    std::string model_path;
    InferenceDevice device = InferenceDevice::CPU;
    OptimizationLevel optimization = OptimizationLevel::BASIC;
    int num_threads = 4;
    int gpu_device_id = 0;
    bool enable_memory_pattern = true;
    bool enable_cpu_mem_arena = true;
    bool enable_graph_optimization = true;
    int execution_mode = 0; // 0: sequential, 1: parallel
    float confidence_threshold = 0.5f;
};

// 推理结果
struct InferenceResult {
    bool success;
    std::vector<float> output_scores;
    std::vector<std::string> output_labels;
    int64_t inference_time_ms;
    int64_t preprocessing_time_ms;
    int64_t postprocessing_time_ms;
    std::string error_message;
};

// 模型信息
struct ModelInfo {
    std::string name;
    std::string version;
    std::vector<std::string> input_names;
    std::vector<std::string> output_names;
    std::vector<std::vector<int64_t>> input_shapes;
    std::vector<std::vector<int64_t>> output_shapes;
    std::vector<std::string> input_types;
    std::vector<std::string> output_types;
};

class ONNXInference {
public:
    ONNXInference();
    ~ONNXInference();

    // 初始化
    bool initialize(const InferenceConfig& config);
    
    // 推理函数
    InferenceResult infer(const std::vector<std::vector<float>>& inputs);
    InferenceResult infer_single_input(const std::vector<float>& input);
    InferenceResult infer_batch(const std::vector<std::vector<std::vector<float>>>& batch_inputs);
    
    // 异步推理
    std::future<InferenceResult> infer_async(const std::vector<std::vector<float>>& inputs);
    
    // 预热模型
    void warmup(int num_runs = 10);
    
    // 获取模型信息
    ModelInfo get_model_info() const;
    bool is_initialized() const;
    
    // 性能分析
    struct PerformanceStats {
        int64_t total_inferences;
        double average_inference_time_ms;
        double average_preprocessing_time_ms;
        double average_postprocessing_time_ms;
        double throughput_fps;
        int64_t peak_memory_usage_mb;
    };
    PerformanceStats get_performance_stats() const;
    
    // 重置统计信息
    void reset_performance_stats();
    
    // 模型优化
    bool optimize_model(const std::string& output_path);
    bool quantize_model(const std::string& output_path, int bits = 8);

private:
    // 内部处理函数
    bool load_model();
    bool setup_session();
    bool setup_providers();
    std::vector<Ort::Value> create_input_tensors(const std::vector<std::vector<float>>& inputs);
    std::vector<float> extract_output_scores(const std::vector<Ort::Value>& outputs);
    std::vector<std::string> get_output_labels(const std::vector<float>& scores);
    void update_performance_stats(const InferenceResult& result);
    void cleanup();

    // ONNX Runtime 相关
    Ort::Env env_;
    Ort::Session session_;
    Ort::SessionOptions session_options_;
    
    // 配置
    InferenceConfig config_;
    ModelInfo model_info_;
    
    // 状态
    bool is_initialized_;
    mutable std::mutex mutex_;
    
    // 性能统计
    mutable std::mutex stats_mutex_;
    PerformanceStats performance_stats_;
    
    // 输入输出名称
    std::vector<const char*> input_names_;
    std::vector<const char*> output_names_;
    
    // 内存分配器
    Ort::MemoryInfo memory_info_;
    
    // 标签映射
    std::unordered_map<int, std::string> label_mapping_;
};

} // namespace ffmpeg_detection 