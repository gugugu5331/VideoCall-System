#include "detection_engine.h"
#include "utils.h"
#include <algorithm>
#include <cmath>

namespace ffmpeg_detection {

DetectionEngine::DetectionEngine()
    : is_initialized_(false)
    , total_inferences_(0)
    , total_processing_time_ms_(0.0) {
}

DetectionEngine::~DetectionEngine() {
    cleanup();
}

bool DetectionEngine::initialize(const ModelConfig& config) {
    LOG_INFO("初始化检测引擎...");
    
    config_ = config;
    
    try {
        // 创建ONNX Runtime环境
        env_ = Ort::Env(ORT_LOGGING_LEVEL_WARNING, "ffmpeg_detection");
        
        // 设置会话选项
        session_options_.SetIntraOpNumThreads(config.num_threads);
        session_options_.SetInterOpNumThreads(config.num_threads);
        
        if (config.use_gpu) {
            // 配置GPU执行提供程序
            OrtCUDAProviderOptions cuda_options;
            cuda_options.device_id = config.gpu_device_id;
            session_options_.AppendExecutionProvider_CUDA(cuda_options);
            LOG_INFO("启用GPU加速，设备ID: %d", config.gpu_device_id);
        }
        
        // 启用内存优化
        session_options_.EnableMemoryPattern();
        session_options_.EnableCpuMemArena();
        
        // 加载模型
        if (!load_model()) {
            LOG_ERROR("模型加载失败");
            return false;
        }
        
        // 设置会话
        if (!setup_session()) {
            LOG_ERROR("会话设置失败");
            return false;
        }
        
        // 预热模型
        warmup();
        
        is_initialized_ = true;
        LOG_INFO("检测引擎初始化成功");
        
        return true;
        
    } catch (const Ort::Exception& e) {
        LOG_ERROR("ONNX Runtime异常: %s", e.what());
        return false;
    } catch (const std::exception& e) {
        LOG_ERROR("初始化异常: %s", e.what());
        return false;
    }
}

DetectionResult DetectionEngine::detect_video_frame(const std::vector<uint8_t>& frame_data,
                                                   int width, int height, int channels) {
    if (!is_initialized_) {
        return {false, 0.0f, DetectionType::GENERAL_FAKE, "引擎未初始化", 0, {}};
    }
    
    Timer timer;
    timer.start();
    
    try {
        // 预处理视频帧
        auto preprocessed_data = preprocess_video(frame_data, width, height, channels);
        
        // 创建输入张量
        std::vector<const char*> input_names = {"input"};
        std::vector<const char*> output_names = {"output"};
        
        std::vector<int64_t> input_shape = {1, channels, config_.input_height, config_.input_width};
        
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);
        
        Ort::Value input_tensor = Ort::Value::CreateTensor<float>(
            memory_info, preprocessed_data.data(), preprocessed_data.size(),
            input_shape.data(), input_shape.size()
        );
        
        // 执行推理
        auto output_tensors = session_.Run(
            Ort::RunOptions{nullptr}, 
            input_names.data(), &input_tensor, 1,
            output_names.data(), output_names.size()
        );
        
        // 后处理输出
        auto output_data = output_tensors[0].GetTensorMutableData<float>();
        size_t output_size = output_tensors[0].GetTensorTypeAndShapeInfo().GetElementCount();
        
        std::vector<float> output_scores(output_data, output_data + output_size);
        auto result = postprocess_output(output_scores);
        
        timer.stop();
        result.processing_time_ms = timer.elapsed_ms();
        
        // 更新统计信息
        {
            std::lock_guard<std::mutex> lock(mutex_);
            total_inferences_++;
            total_processing_time_ms_ += result.processing_time_ms;
        }
        
        return result;
        
    } catch (const Ort::Exception& e) {
        LOG_ERROR("推理异常: %s", e.what());
        return {false, 0.0f, DetectionType::GENERAL_FAKE, e.what(), timer.elapsed_ms(), {}};
    } catch (const std::exception& e) {
        LOG_ERROR("检测异常: %s", e.what());
        return {false, 0.0f, DetectionType::GENERAL_FAKE, e.what(), timer.elapsed_ms(), {}};
    }
}

DetectionResult DetectionEngine::detect_audio_frame(const std::vector<float>& audio_data,
                                                   int sample_rate, int channels) {
    if (!is_initialized_) {
        return {false, 0.0f, DetectionType::AUDIO_FORGERY, "引擎未初始化", 0, {}};
    }
    
    Timer timer;
    timer.start();
    
    try {
        // 预处理音频数据
        auto preprocessed_data = preprocess_audio(audio_data, sample_rate, channels);
        
        // 创建输入张量
        std::vector<const char*> input_names = {"input"};
        std::vector<const char*> output_names = {"output"};
        
        std::vector<int64_t> input_shape = {1, 1, static_cast<int64_t>(preprocessed_data.size())};
        
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);
        
        Ort::Value input_tensor = Ort::Value::CreateTensor<float>(
            memory_info, preprocessed_data.data(), preprocessed_data.size(),
            input_shape.data(), input_shape.size()
        );
        
        // 执行推理
        auto output_tensors = session_.Run(
            Ort::RunOptions{nullptr}, 
            input_names.data(), &input_tensor, 1,
            output_names.data(), output_names.size()
        );
        
        // 后处理输出
        auto output_data = output_tensors[0].GetTensorMutableData<float>();
        size_t output_size = output_tensors[0].GetTensorTypeAndShapeInfo().GetElementCount();
        
        std::vector<float> output_scores(output_data, output_data + output_size);
        auto result = postprocess_output(output_scores);
        result.type = DetectionType::AUDIO_FORGERY;
        
        timer.stop();
        result.processing_time_ms = timer.elapsed_ms();
        
        // 更新统计信息
        {
            std::lock_guard<std::mutex> lock(mutex_);
            total_inferences_++;
            total_processing_time_ms_ += result.processing_time_ms;
        }
        
        return result;
        
    } catch (const Ort::Exception& e) {
        LOG_ERROR("音频推理异常: %s", e.what());
        return {false, 0.0f, DetectionType::AUDIO_FORGERY, e.what(), timer.elapsed_ms(), {}};
    } catch (const std::exception& e) {
        LOG_ERROR("音频检测异常: %s", e.what());
        return {false, 0.0f, DetectionType::AUDIO_FORGERY, e.what(), timer.elapsed_ms(), {}};
    }
}

DetectionResult DetectionEngine::detect_combined(const std::vector<uint8_t>& video_data,
                                                const std::vector<float>& audio_data,
                                                int video_width, int video_height,
                                                int audio_sample_rate) {
    if (!is_initialized_) {
        return {false, 0.0f, DetectionType::GENERAL_FAKE, "引擎未初始化", 0, {}};
    }
    
    Timer timer;
    timer.start();
    
    try {
        // 预处理视频和音频
        auto preprocessed_video = preprocess_video(video_data, video_width, video_height, 3);
        auto preprocessed_audio = preprocess_audio(audio_data, audio_sample_rate, 1);
        
        // 创建多模态输入张量
        std::vector<const char*> input_names = {"video_input", "audio_input"};
        std::vector<const char*> output_names = {"output"};
        
        std::vector<int64_t> video_shape = {1, 3, config_.input_height, config_.input_width};
        std::vector<int64_t> audio_shape = {1, 1, static_cast<int64_t>(preprocessed_audio.size())};
        
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);
        
        Ort::Value video_tensor = Ort::Value::CreateTensor<float>(
            memory_info, preprocessed_video.data(), preprocessed_video.size(),
            video_shape.data(), video_shape.size()
        );
        
        Ort::Value audio_tensor = Ort::Value::CreateTensor<float>(
            memory_info, preprocessed_audio.data(), preprocessed_audio.size(),
            audio_shape.data(), audio_shape.size()
        );
        
        std::vector<Ort::Value> input_tensors = {std::move(video_tensor), std::move(audio_tensor)};
        
        // 执行推理
        auto output_tensors = session_.Run(
            Ort::RunOptions{nullptr}, 
            input_names.data(), input_tensors.data(), input_tensors.size(),
            output_names.data(), output_names.size()
        );
        
        // 后处理输出
        auto output_data = output_tensors[0].GetTensorMutableData<float>();
        size_t output_size = output_tensors[0].GetTensorTypeAndShapeInfo().GetElementCount();
        
        std::vector<float> output_scores(output_data, output_data + output_size);
        auto result = postprocess_output(output_scores);
        result.type = DetectionType::GENERAL_FAKE;
        
        timer.stop();
        result.processing_time_ms = timer.elapsed_ms();
        
        // 更新统计信息
        {
            std::lock_guard<std::mutex> lock(mutex_);
            total_inferences_++;
            total_processing_time_ms_ += result.processing_time_ms;
        }
        
        return result;
        
    } catch (const Ort::Exception& e) {
        LOG_ERROR("多模态推理异常: %s", e.what());
        return {false, 0.0f, DetectionType::GENERAL_FAKE, e.what(), timer.elapsed_ms(), {}};
    } catch (const std::exception& e) {
        LOG_ERROR("多模态检测异常: %s", e.what());
        return {false, 0.0f, DetectionType::GENERAL_FAKE, e.what(), timer.elapsed_ms(), {}};
    }
}

std::vector<DetectionResult> DetectionEngine::detect_batch(const std::vector<std::vector<uint8_t>>& frames) {
    std::vector<DetectionResult> results;
    results.reserve(frames.size());
    
    for (const auto& frame : frames) {
        // 假设所有帧都是相同尺寸
        auto result = detect_video_frame(frame, config_.input_width, config_.input_height, config_.input_channels);
        results.push_back(result);
    }
    
    return results;
}

std::string DetectionEngine::get_model_info() const {
    if (!is_initialized_) {
        return "模型未初始化";
    }
    
    std::string info = "模型信息:\n";
    info += "  路径: " + config_.model_path + "\n";
    info += "  输入尺寸: " + std::to_string(config_.input_width) + "x" + std::to_string(config_.input_height) + "\n";
    info += "  输入通道: " + std::to_string(config_.input_channels) + "\n";
    info += "  使用GPU: " + std::string(config_.use_gpu ? "是" : "否") + "\n";
    info += "  线程数: " + std::to_string(config_.num_threads) + "\n";
    info += "  置信度阈值: " + std::to_string(config_.confidence_threshold) + "\n";
    info += "  总推理次数: " + std::to_string(total_inferences_) + "\n";
    info += "  平均处理时间: " + std::to_string(total_inferences_ > 0 ? total_processing_time_ms_ / total_inferences_ : 0) + "ms\n";
    
    return info;
}

bool DetectionEngine::is_initialized() const {
    return is_initialized_;
}

void DetectionEngine::warmup() {
    LOG_INFO("预热模型...");
    
    // 创建随机输入数据进行预热
    std::vector<uint8_t> dummy_frame(config_.input_width * config_.input_height * config_.input_channels, 128);
    
    for (int i = 0; i < 10; i++) {
        detect_video_frame(dummy_frame, config_.input_width, config_.input_height, config_.input_channels);
    }
    
    LOG_INFO("模型预热完成");
}

bool DetectionEngine::load_model() {
    if (!FileUtils::file_exists(config_.model_path)) {
        LOG_ERROR("模型文件不存在: %s", config_.model_path.c_str());
        return false;
    }
    
    try {
        session_ = Ort::Session(env_, config_.model_path.c_str(), session_options_);
        LOG_INFO("模型加载成功: %s", config_.model_path.c_str());
        return true;
    } catch (const Ort::Exception& e) {
        LOG_ERROR("模型加载失败: %s", e.what());
        return false;
    }
}

bool DetectionEngine::setup_session() {
    try {
        // 获取输入输出信息
        Ort::AllocatorWithDefaultOptions allocator;
        
        // 获取输入名称
        size_t num_input_nodes = session_.GetInputCount();
        input_names_.reserve(num_input_nodes);
        for (size_t i = 0; i < num_input_nodes; i++) {
            char* input_name = session_.GetInputName(i, allocator);
            input_names_.push_back(input_name);
            allocator.Free(input_name);
        }
        
        // 获取输出名称
        size_t num_output_nodes = session_.GetOutputCount();
        output_names_.reserve(num_output_nodes);
        for (size_t i = 0; i < num_output_nodes; i++) {
            char* output_name = session_.GetOutputName(i, allocator);
            output_names_.push_back(output_name);
            allocator.Free(output_name);
        }
        
        // 获取输入输出形状
        input_shapes_.reserve(num_input_nodes);
        output_shapes_.reserve(num_output_nodes);
        
        for (size_t i = 0; i < num_input_nodes; i++) {
            auto type_info = session_.GetInputTypeInfo(i);
            auto tensor_info = type_info.GetTensorTypeAndShapeInfo();
            input_shapes_.push_back(tensor_info.GetShape());
        }
        
        for (size_t i = 0; i < num_output_nodes; i++) {
            auto type_info = session_.GetOutputTypeInfo(i);
            auto tensor_info = type_info.GetTensorTypeAndShapeInfo();
            output_shapes_.push_back(tensor_info.GetShape());
        }
        
        LOG_INFO("会话设置完成 - 输入: %zu, 输出: %zu", num_input_nodes, num_output_nodes);
        return true;
        
    } catch (const Ort::Exception& e) {
        LOG_ERROR("会话设置失败: %s", e.what());
        return false;
    }
}

std::vector<float> DetectionEngine::preprocess_video(const std::vector<uint8_t>& frame_data,
                                                    int width, int height, int channels) {
    // 调整图像尺寸到模型输入尺寸
    std::vector<uint8_t> resized_data;
    if (width != config_.input_width || height != config_.input_height) {
        // 这里应该使用图像缩放库，简化处理
        resized_data = frame_data; // 实际应该进行缩放
    } else {
        resized_data = frame_data;
    }
    
    // 转换为浮点数并归一化
    std::vector<float> normalized_data;
    normalized_data.reserve(config_.input_width * config_.input_height * config_.input_channels);
    
    for (int c = 0; c < config_.input_channels; c++) {
        for (int h = 0; h < config_.input_height; h++) {
            for (int w = 0; w < config_.input_width; w++) {
                int idx = (h * config_.input_width + w) * config_.input_channels + c;
                float pixel_value = static_cast<float>(resized_data[idx]) / 255.0f;
                
                // 应用均值和标准差归一化
                pixel_value = (pixel_value - config_.mean[c]) / config_.std[c];
                
                normalized_data.push_back(pixel_value);
            }
        }
    }
    
    return normalized_data;
}

std::vector<float> DetectionEngine::preprocess_audio(const std::vector<float>& audio_data,
                                                    int sample_rate, int channels) {
    // 重采样到目标采样率
    std::vector<float> resampled_data;
    if (sample_rate != 16000) {
        // 这里应该使用音频重采样库，简化处理
        resampled_data = audio_data; // 实际应该进行重采样
    } else {
        resampled_data = audio_data;
    }
    
    // 转换为单声道
    std::vector<float> mono_data;
    if (channels > 1) {
        mono_data.reserve(resampled_data.size() / channels);
        for (size_t i = 0; i < resampled_data.size(); i += channels) {
            float sum = 0.0f;
            for (int c = 0; c < channels; c++) {
                sum += resampled_data[i + c];
            }
            mono_data.push_back(sum / channels);
        }
    } else {
        mono_data = resampled_data;
    }
    
    // 应用窗口函数和归一化
    for (size_t i = 0; i < mono_data.size(); i++) {
        // 应用汉宁窗
        float window = 0.5f * (1.0f - cos(2.0f * M_PI * i / (mono_data.size() - 1)));
        mono_data[i] *= window;
        
        // 归一化到[-1, 1]范围
        mono_data[i] = std::max(-1.0f, std::min(1.0f, mono_data[i]));
    }
    
    return mono_data;
}

DetectionResult DetectionEngine::postprocess_output(const std::vector<float>& output) {
    DetectionResult result;
    
    if (output.empty()) {
        result.is_fake = false;
        result.confidence = 0.0f;
        result.type = DetectionType::GENERAL_FAKE;
        result.details = "输出为空";
        result.raw_scores = output;
        return result;
    }
    
    // 应用softmax
    auto softmax_scores = MathUtils::softmax(output);
    
    // 找到最大概率的类别
    auto max_it = std::max_element(softmax_scores.begin(), softmax_scores.end());
    int max_class = std::distance(softmax_scores.begin(), max_it);
    float max_confidence = *max_it;
    
    // 判断是否为伪造
    result.is_fake = max_confidence > config_.confidence_threshold;
    result.confidence = max_confidence;
    result.raw_scores = softmax_scores;
    
    // 根据类别确定检测类型
    switch (max_class) {
        case 0:
            result.type = DetectionType::FACE_FORGERY;
            result.details = "人脸伪造";
            break;
        case 1:
            result.type = DetectionType::DEEPFAKE;
            result.details = "Deepfake";
            break;
        case 2:
            result.type = DetectionType::FACE_SWAP;
            result.details = "换脸";
            break;
        case 3:
            result.type = DetectionType::AUDIO_FORGERY;
            result.details = "音频伪造";
            break;
        case 4:
            result.type = DetectionType::LIP_SYNC;
            result.details = "唇同步";
            break;
        default:
            result.type = DetectionType::GENERAL_FAKE;
            result.details = "通用伪造";
            break;
    }
    
    return result;
}

std::string DetectionEngine::detection_type_to_string(DetectionType type) {
    switch (type) {
        case DetectionType::FACE_FORGERY: return "人脸伪造";
        case DetectionType::DEEPFAKE: return "Deepfake";
        case DetectionType::FACE_SWAP: return "换脸";
        case DetectionType::AUDIO_FORGERY: return "音频伪造";
        case DetectionType::LIP_SYNC: return "唇同步";
        case DetectionType::GENERAL_FAKE: return "通用伪造";
        default: return "未知";
    }
}

void DetectionEngine::cleanup() {
    if (is_initialized_) {
        session_.release();
        is_initialized_ = false;
    }
}

} // namespace ffmpeg_detection 