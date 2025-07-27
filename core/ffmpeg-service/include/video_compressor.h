#pragma once

#include <memory>
#include <string>
#include <vector>
#include <functional>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libswscale/swscale.h>
}

namespace ffmpeg_detection {

// 压缩质量级别
enum class CompressionQuality {
    LOW,        // 高压缩，低质量
    MEDIUM,     // 平衡压缩和质量
    HIGH,       // 低压缩，高质量
    CUSTOM      // 自定义参数
};

// 压缩配置
struct VideoCompressionConfig {
    int target_width = 640;
    int target_height = 480;
    int target_fps = 30;
    int bitrate = 1000000;  // 1Mbps
    std::string codec = "libx264";
    int quality = 23;       // 0-51, 越低质量越好
    int gop_size = 30;      // GOP大小
    int max_b_frames = 2;   // B帧数量
    bool enable_fast_decode = true;
    CompressionQuality quality_level = CompressionQuality::MEDIUM;
};

// 压缩结果
struct CompressionResult {
    bool success;
    std::vector<uint8_t> compressed_data;
    int original_size;
    int compressed_size;
    double compression_ratio;
    int64_t processing_time_ms;
    std::string error_message;
};

// 帧信息
struct FrameInfo {
    int width;
    int height;
    int channels;
    int64_t timestamp;
    bool is_keyframe;
    std::string pixel_format;
};

class VideoCompressor {
public:
    VideoCompressor();
    ~VideoCompressor();

    // 初始化
    bool initialize(const VideoCompressionConfig& config);
    
    // 压缩单个帧
    CompressionResult compress_frame(const std::vector<uint8_t>& frame_data,
                                    const FrameInfo& frame_info);
    
    // 批量压缩
    std::vector<CompressionResult> compress_frames(
        const std::vector<std::vector<uint8_t>>& frames,
        const std::vector<FrameInfo>& frame_infos);
    
    // 流式压缩
    bool start_stream_compression();
    CompressionResult compress_stream_frame(const std::vector<uint8_t>& frame_data,
                                           const FrameInfo& frame_info);
    std::vector<uint8_t> finish_stream_compression();
    
    // 设置回调
    using CompressedFrameCallback = std::function<void(const CompressionResult&)>;
    void set_compressed_frame_callback(CompressedFrameCallback callback);
    
    // 获取统计信息
    struct Statistics {
        int64_t frames_compressed;
        double average_compression_ratio;
        double average_processing_time_ms;
        int64_t total_bytes_saved;
    };
    Statistics get_statistics() const;
    
    // 重置统计信息
    void reset_statistics();
    
    // 检查是否已初始化
    bool is_initialized() const;

private:
    // 内部处理函数
    bool setup_codec();
    bool setup_scaler();
    std::vector<uint8_t> scale_frame(const std::vector<uint8_t>& frame_data,
                                    const FrameInfo& frame_info);
    CompressionResult encode_frame(AVFrame* frame);
    void update_statistics(const CompressionResult& result);
    void cleanup();

    // FFmpeg 上下文
    AVCodecContext* codec_ctx_;
    SwsContext* sws_ctx_;
    AVFrame* frame_;
    AVFrame* scaled_frame_;
    AVPacket* packet_;
    
    // 配置
    VideoCompressionConfig config_;
    
    // 状态
    bool is_initialized_;
    bool is_streaming_;
    
    // 回调
    CompressedFrameCallback compressed_frame_callback_;
    
    // 统计信息
    mutable std::mutex stats_mutex_;
    Statistics stats_;
    
    // 流式压缩缓冲区
    std::vector<uint8_t> stream_buffer_;
};

} // namespace ffmpeg_detection 