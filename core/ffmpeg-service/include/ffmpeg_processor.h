#pragma once

#include <memory>
#include <string>
#include <vector>
#include <functional>
#include <thread>
#include <atomic>
#include <mutex>
#include <condition_variable>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
#include <libswresample/swresample.h>
}

namespace ffmpeg_detection {

// 前向声明
class DetectionEngine;
class VideoCompressor;
class AudioProcessor;

// 帧数据结构
struct FrameData {
    std::vector<uint8_t> data;
    int width;
    int height;
    int channels;
    int64_t timestamp;
    bool is_keyframe;
    std::string frame_type; // "video", "audio"
};

// 处理结果
struct ProcessingResult {
    bool is_fake;
    float confidence;
    std::string detection_type;
    int64_t processing_time_ms;
    std::string details;
};

// 压缩配置
struct CompressionConfig {
    int target_width = 640;
    int target_height = 480;
    int target_fps = 30;
    int video_bitrate = 1000000; // 1Mbps
    int audio_bitrate = 128000;  // 128kbps
    std::string video_codec = "libx264";
    std::string audio_codec = "aac";
    int quality = 23; // 0-51, 越低质量越好
};

// 回调函数类型
using FrameCallback = std::function<void(const FrameData&)>;
using ResultCallback = std::function<void(const ProcessingResult&)>;

class FFmpegProcessor {
public:
    FFmpegProcessor();
    ~FFmpegProcessor();

    // 初始化
    bool initialize(const std::string& model_path, const CompressionConfig& config = {});
    
    // 处理输入流
    bool process_input_stream(const std::string& input_url);
    bool process_input_file(const std::string& input_file);
    
    // 实时处理
    bool start_realtime_processing(const std::string& input_url);
    void stop_realtime_processing();
    
    // 设置回调
    void set_frame_callback(FrameCallback callback);
    void set_result_callback(ResultCallback callback);
    
    // 获取状态
    bool is_processing() const;
    bool is_initialized() const;
    
    // 获取统计信息
    struct Statistics {
        int64_t frames_processed;
        int64_t fake_detections;
        double average_processing_time_ms;
        double compression_ratio;
    };
    Statistics get_statistics() const;

private:
    // 内部处理函数
    bool initialize_ffmpeg();
    bool initialize_codecs();
    bool process_video_frame(AVFrame* frame, int64_t timestamp);
    bool process_audio_frame(AVFrame* frame, int64_t timestamp);
    void processing_thread();
    void cleanup();

    // 成员变量
    std::unique_ptr<DetectionEngine> detection_engine_;
    std::unique_ptr<VideoCompressor> video_compressor_;
    std::unique_ptr<AudioProcessor> audio_processor_;
    
    // FFmpeg 上下文
    AVFormatContext* format_ctx_;
    AVCodecContext* video_codec_ctx_;
    AVCodecContext* audio_codec_ctx_;
    SwsContext* sws_ctx_;
    SwrContext* swr_ctx_;
    
    // 流索引
    int video_stream_idx_;
    int audio_stream_idx_;
    
    // 处理状态
    std::atomic<bool> is_processing_;
    std::atomic<bool> is_initialized_;
    std::atomic<bool> should_stop_;
    
    // 线程
    std::thread processing_thread_;
    std::mutex mutex_;
    std::condition_variable cv_;
    
    // 回调
    FrameCallback frame_callback_;
    ResultCallback result_callback_;
    
    // 统计信息
    mutable std::mutex stats_mutex_;
    Statistics stats_;
    
    // 配置
    CompressionConfig config_;
};

} // namespace ffmpeg_detection 