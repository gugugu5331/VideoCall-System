#pragma once

#include <string>
#include <vector>
#include <memory>
#include <functional>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <queue>
#include <atomic>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libswscale/swscale.h>
#include <libswresample/swresample.h>
#include <libavfilter/avfilter.h>
}

namespace ffmpeg_service {

// 前向声明
class VideoProcessor;
class AudioProcessor;
class MediaCompressor;

// 媒体帧结构
struct MediaFrame {
    std::vector<uint8_t> data;
    int64_t timestamp;
    int width;
    int height;
    int channels;
    int sample_rate;
    AVPixelFormat pixel_format;
    AVSampleFormat sample_format;
    bool is_key_frame;
    
    MediaFrame() : timestamp(0), width(0), height(0), channels(0), 
                   sample_rate(0), pixel_format(AV_PIX_FMT_NONE), 
                   sample_format(AV_SAMPLE_FMT_NONE), is_key_frame(false) {}
};

// 编码参数
struct EncodingParams {
    int video_bitrate = 1000000;  // 1Mbps
    int audio_bitrate = 128000;   // 128kbps
    int video_width = 1280;
    int video_height = 720;
    int video_fps = 30;
    int audio_sample_rate = 44100;
    int audio_channels = 2;
    AVPixelFormat video_pixel_format = AV_PIX_FMT_YUV420P;
    AVSampleFormat audio_sample_format = AV_SAMPLE_FMT_FLTP;
    AVCodecID video_codec_id = AV_CODEC_ID_H264;
    AVCodecID audio_codec_id = AV_CODEC_ID_AAC;
    int gop_size = 30;
    int max_b_frames = 3;
    bool enable_hardware_acceleration = false;
};

// 处理结果
struct ProcessingResult {
    bool success;
    std::string error_message;
    std::vector<uint8_t> processed_data;
    int64_t processing_time_ms;
    float compression_ratio;
    
    ProcessingResult() : success(false), processing_time_ms(0), compression_ratio(1.0f) {}
};

// 回调函数类型
using FrameCallback = std::function<void(const MediaFrame&)>;
using ProcessingCallback = std::function<void(const ProcessingResult&)>;

// FFmpeg处理器主类
class FFmpegProcessor {
public:
    FFmpegProcessor();
    ~FFmpegProcessor();
    
    // 初始化
    bool initialize(const EncodingParams& params = EncodingParams{});
    void cleanup();
    
    // 视频处理
    bool processVideoFrame(const std::vector<uint8_t>& frame_data, 
                          int width, int height, 
                          AVPixelFormat format = AV_PIX_FMT_RGB24);
    
    // 音频处理
    bool processAudioFrame(const std::vector<uint8_t>& audio_data,
                          int sample_rate, int channels,
                          AVSampleFormat format = AV_SAMPLE_FMT_FLT);
    
    // 压缩处理
    ProcessingResult compressVideo(const std::vector<uint8_t>& video_data,
                                  const EncodingParams& params = EncodingParams{});
    
    ProcessingResult compressAudio(const std::vector<uint8_t>& audio_data,
                                  const EncodingParams& params = EncodingParams{});
    
    // 解压缩
    ProcessingResult decompressVideo(const std::vector<uint8_t>& compressed_data);
    ProcessingResult decompressAudio(const std::vector<uint8_t>& compressed_data);
    
    // 格式转换
    ProcessingResult convertVideoFormat(const std::vector<uint8_t>& video_data,
                                       AVPixelFormat target_format,
                                       int target_width, int target_height);
    
    ProcessingResult convertAudioFormat(const std::vector<uint8_t>& audio_data,
                                       AVSampleFormat target_format,
                                       int target_sample_rate, int target_channels);
    
    // 实时处理
    void startRealTimeProcessing(FrameCallback video_callback = nullptr,
                                FrameCallback audio_callback = nullptr);
    void stopRealTimeProcessing();
    
    // 设置参数
    void setEncodingParams(const EncodingParams& params);
    void setProcessingCallback(ProcessingCallback callback);
    
    // 状态查询
    bool isInitialized() const { return initialized_; }
    bool isProcessing() const { return processing_; }
    EncodingParams getCurrentParams() const { return current_params_; }

private:
    // 内部处理函数
    bool initializeCodecs();
    bool initializeFilters();
    void processingLoop();
    void handleVideoFrame(const MediaFrame& frame);
    void handleAudioFrame(const MediaFrame& frame);
    
    // 成员变量
    std::unique_ptr<VideoProcessor> video_processor_;
    std::unique_ptr<AudioProcessor> audio_processor_;
    std::unique_ptr<MediaCompressor> compressor_;
    
    EncodingParams current_params_;
    ProcessingCallback processing_callback_;
    
    std::thread processing_thread_;
    std::mutex frame_queue_mutex_;
    std::condition_variable frame_cv_;
    std::queue<MediaFrame> frame_queue_;
    
    std::atomic<bool> initialized_;
    std::atomic<bool> processing_;
    std::atomic<bool> should_stop_;
    
    FrameCallback video_callback_;
    FrameCallback audio_callback_;
};

// 视频处理器
class VideoProcessor {
public:
    VideoProcessor();
    ~VideoProcessor();
    
    bool initialize(const EncodingParams& params);
    void cleanup();
    
    ProcessingResult processFrame(const std::vector<uint8_t>& frame_data,
                                 int width, int height,
                                 AVPixelFormat format);
    
    ProcessingResult compress(const std::vector<uint8_t>& video_data,
                             const EncodingParams& params);
    
    ProcessingResult decompress(const std::vector<uint8_t>& compressed_data);
    
    ProcessingResult convertFormat(const std::vector<uint8_t>& video_data,
                                  AVPixelFormat target_format,
                                  int target_width, int target_height);

private:
    bool initializeCodec();
    bool initializeFilter();
    void cleanupCodec();
    
    AVCodecContext* encoder_ctx_;
    AVCodecContext* decoder_ctx_;
    AVFilterContext* filter_ctx_;
    SwsContext* sws_ctx_;
    
    EncodingParams params_;
    bool initialized_;
};

// 音频处理器
class AudioProcessor {
public:
    AudioProcessor();
    ~AudioProcessor();
    
    bool initialize(const EncodingParams& params);
    void cleanup();
    
    ProcessingResult processFrame(const std::vector<uint8_t>& audio_data,
                                 int sample_rate, int channels,
                                 AVSampleFormat format);
    
    ProcessingResult compress(const std::vector<uint8_t>& audio_data,
                             const EncodingParams& params);
    
    ProcessingResult decompress(const std::vector<uint8_t>& compressed_data);
    
    ProcessingResult convertFormat(const std::vector<uint8_t>& audio_data,
                                  AVSampleFormat target_format,
                                  int target_sample_rate, int target_channels);

private:
    bool initializeCodec();
    bool initializeResampler();
    void cleanupCodec();
    
    AVCodecContext* encoder_ctx_;
    AVCodecContext* decoder_ctx_;
    SwrContext* swr_ctx_;
    
    EncodingParams params_;
    bool initialized_;
};

// 媒体压缩器
class MediaCompressor {
public:
    MediaCompressor();
    ~MediaCompressor();
    
    bool initialize(const EncodingParams& params);
    void cleanup();
    
    ProcessingResult compressVideo(const std::vector<uint8_t>& video_data,
                                  const EncodingParams& params);
    
    ProcessingResult compressAudio(const std::vector<uint8_t>& audio_data,
                                  const EncodingParams& params);
    
    ProcessingResult decompressVideo(const std::vector<uint8_t>& compressed_data);
    ProcessingResult decompressAudio(const std::vector<uint8_t>& compressed_data);

private:
    std::unique_ptr<VideoProcessor> video_processor_;
    std::unique_ptr<AudioProcessor> audio_processor_;
    EncodingParams params_;
    bool initialized_;
};

} // namespace ffmpeg_service 