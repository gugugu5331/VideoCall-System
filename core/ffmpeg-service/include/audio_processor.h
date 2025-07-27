#pragma once

#include <memory>
#include <string>
#include <vector>
#include <complex>
#include <functional>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libswresample/swresample.h>
}

namespace ffmpeg_detection {

// 音频格式
enum class AudioFormat {
    PCM_S16LE,      // 16-bit signed little-endian
    PCM_S32LE,      // 32-bit signed little-endian
    PCM_F32LE,      // 32-bit float little-endian
    PCM_F64LE       // 64-bit float little-endian
};

// 音频处理配置
struct AudioProcessingConfig {
    int target_sample_rate = 16000;
    int target_channels = 1;  // 单声道
    AudioFormat target_format = AudioFormat::PCM_F32LE;
    int frame_size = 1024;    // FFT窗口大小
    int hop_size = 512;       // 帧移
    float min_frequency = 80.0f;   // 最小频率
    float max_frequency = 8000.0f; // 最大频率
    bool enable_noise_reduction = true;
    bool enable_spectral_subtraction = false;
};

// 音频特征
struct AudioFeatures {
    std::vector<float> mfcc;           // MFCC特征
    std::vector<float> spectral_centroid; // 频谱质心
    std::vector<float> spectral_rolloff;  // 频谱滚降
    std::vector<float> zero_crossing_rate; // 过零率
    std::vector<float> energy;         // 能量
    std::vector<std::vector<float>> spectrogram; // 频谱图
    std::vector<float> pitch;          // 基频
    std::vector<float> formants;       // 共振峰
};

// 处理结果
struct AudioProcessingResult {
    bool success;
    std::vector<float> processed_audio;
    AudioFeatures features;
    int64_t processing_time_ms;
    std::string error_message;
};

class AudioProcessor {
public:
    AudioProcessor();
    ~AudioProcessor();

    // 初始化
    bool initialize(const AudioProcessingConfig& config);
    
    // 处理音频数据
    AudioProcessingResult process_audio(const std::vector<uint8_t>& audio_data,
                                       int sample_rate, int channels, AudioFormat format);
    
    // 处理浮点音频数据
    AudioProcessingResult process_float_audio(const std::vector<float>& audio_data,
                                             int sample_rate, int channels);
    
    // 批量处理
    std::vector<AudioProcessingResult> process_audio_batch(
        const std::vector<std::vector<uint8_t>>& audio_chunks,
        int sample_rate, int channels, AudioFormat format);
    
    // 特征提取
    AudioFeatures extract_features(const std::vector<float>& audio_data);
    
    // 音频增强
    std::vector<float> enhance_audio(const std::vector<float>& audio_data);
    
    // 噪声抑制
    std::vector<float> reduce_noise(const std::vector<float>& audio_data);
    
    // 设置回调
    using ProcessedAudioCallback = std::function<void(const AudioProcessingResult&)>;
    void set_processed_audio_callback(ProcessedAudioCallback callback);
    
    // 获取统计信息
    struct Statistics {
        int64_t audio_chunks_processed;
        double average_processing_time_ms;
        double average_compression_ratio;
        int64_t total_audio_duration_ms;
    };
    Statistics get_statistics() const;
    
    // 重置统计信息
    void reset_statistics();
    
    // 检查是否已初始化
    bool is_initialized() const;

private:
    // 内部处理函数
    bool setup_resampler();
    std::vector<float> resample_audio(const std::vector<uint8_t>& audio_data,
                                     int sample_rate, int channels, AudioFormat format);
    std::vector<float> convert_to_float(const std::vector<uint8_t>& audio_data,
                                       AudioFormat format);
    std::vector<float> apply_window(const std::vector<float>& audio_data);
    std::vector<std::complex<float>> compute_fft(const std::vector<float>& audio_data);
    std::vector<float> compute_mfcc(const std::vector<std::complex<float>>& spectrum);
    std::vector<float> compute_spectral_features(const std::vector<std::complex<float>>& spectrum);
    float compute_pitch(const std::vector<float>& audio_data);
    std::vector<float> compute_formants(const std::vector<float>& audio_data);
    void update_statistics(const AudioProcessingResult& result);
    void cleanup();

    // FFmpeg 上下文
    SwrContext* swr_ctx_;
    
    // 配置
    AudioProcessingConfig config_;
    
    // 状态
    bool is_initialized_;
    
    // 回调
    ProcessedAudioCallback processed_audio_callback_;
    
    // 统计信息
    mutable std::mutex stats_mutex_;
    Statistics stats_;
    
    // FFT 相关
    std::vector<float> window_function_;
    std::vector<std::complex<float>> fft_buffer_;
    
    // 特征提取缓存
    std::vector<float> feature_cache_;
};

} // namespace ffmpeg_detection 