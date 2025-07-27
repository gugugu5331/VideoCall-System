#include "ffmpeg_processor.h"
#include "detection_engine.h"
#include "video_compressor.h"
#include "audio_processor.h"
#include "utils.h"

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/frame.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
#include <libswresample/swresample.h>
}

namespace ffmpeg_detection {

FFmpegProcessor::FFmpegProcessor()
    : format_ctx_(nullptr)
    , video_codec_ctx_(nullptr)
    , audio_codec_ctx_(nullptr)
    , sws_ctx_(nullptr)
    , swr_ctx_(nullptr)
    , video_stream_idx_(-1)
    , audio_stream_idx_(-1)
    , is_processing_(false)
    , is_initialized_(false)
    , should_stop_(false) {
    
    // 初始化FFmpeg
    av_register_all();
    avformat_network_init();
    
    // 初始化统计信息
    stats_ = {0, 0, 0.0, 0.0};
}

FFmpegProcessor::~FFmpegProcessor() {
    stop_realtime_processing();
    cleanup();
}

bool FFmpegProcessor::initialize(const std::string& model_path, const CompressionConfig& config) {
    LOG_INFO("初始化FFmpeg处理器...");
    
    config_ = config;
    
    // 创建检测引擎
    detection_engine_ = std::make_unique<DetectionEngine>();
    ModelConfig model_config;
    model_config.model_path = model_path;
    model_config.input_width = config.target_width;
    model_config.input_height = config.target_height;
    model_config.use_gpu = false; // 可以根据需要启用GPU
    
    if (!detection_engine_->initialize(model_config)) {
        LOG_ERROR("检测引擎初始化失败");
        return false;
    }
    
    // 创建视频压缩器
    video_compressor_ = std::make_unique<VideoCompressor>();
    VideoCompressionConfig video_config;
    video_config.target_width = config.target_width;
    video_config.target_height = config.target_height;
    video_config.bitrate = config.video_bitrate;
    video_config.codec = config.video_codec;
    video_config.quality = config.quality;
    
    if (!video_compressor_->initialize(video_config)) {
        LOG_ERROR("视频压缩器初始化失败");
        return false;
    }
    
    // 创建音频处理器
    audio_processor_ = std::make_unique<AudioProcessor>();
    AudioProcessingConfig audio_config;
    audio_config.target_sample_rate = 16000;
    audio_config.target_channels = 1;
    audio_config.enable_noise_reduction = true;
    
    if (!audio_processor_->initialize(audio_config)) {
        LOG_ERROR("音频处理器初始化失败");
        return false;
    }
    
    is_initialized_ = true;
    LOG_INFO("FFmpeg处理器初始化成功");
    
    return true;
}

bool FFmpegProcessor::process_input_stream(const std::string& input_url) {
    if (!is_initialized_) {
        LOG_ERROR("处理器未初始化");
        return false;
    }
    
    LOG_INFO("开始处理输入流: %s", input_url.c_str());
    
    // 打开输入流
    if (avformat_open_input(&format_ctx_, input_url.c_str(), nullptr, nullptr) != 0) {
        LOG_ERROR("无法打开输入流: %s", input_url.c_str());
        return false;
    }
    
    // 查找流信息
    if (avformat_find_stream_info(format_ctx_, nullptr) < 0) {
        LOG_ERROR("无法查找流信息");
        return false;
    }
    
    // 初始化编解码器
    if (!initialize_codecs()) {
        LOG_ERROR("编解码器初始化失败");
        return false;
    }
    
    // 开始处理
    return process_stream();
}

bool FFmpegProcessor::process_input_file(const std::string& input_file) {
    return process_input_stream(input_file);
}

bool FFmpegProcessor::start_realtime_processing(const std::string& input_url) {
    if (is_processing_) {
        LOG_WARNING("已经在处理中");
        return false;
    }
    
    should_stop_ = false;
    is_processing_ = true;
    
    // 启动处理线程
    processing_thread_ = std::thread([this, input_url]() {
        this->processing_thread();
    });
    
    return true;
}

void FFmpegProcessor::stop_realtime_processing() {
    if (!is_processing_) {
        return;
    }
    
    should_stop_ = true;
    cv_.notify_all();
    
    if (processing_thread_.joinable()) {
        processing_thread_.join();
    }
    
    is_processing_ = false;
}

void FFmpegProcessor::set_frame_callback(FrameCallback callback) {
    frame_callback_ = callback;
}

void FFmpegProcessor::set_result_callback(ResultCallback callback) {
    result_callback_ = callback;
}

bool FFmpegProcessor::is_processing() const {
    return is_processing_;
}

bool FFmpegProcessor::is_initialized() const {
    return is_initialized_;
}

FFmpegProcessor::Statistics FFmpegProcessor::get_statistics() const {
    std::lock_guard<std::mutex> lock(stats_mutex_);
    return stats_;
}

bool FFmpegProcessor::initialize_ffmpeg() {
    // FFmpeg已经在构造函数中初始化
    return true;
}

bool FFmpegProcessor::initialize_codecs() {
    // 查找视频流
    video_stream_idx_ = av_find_best_stream(format_ctx_, AVMEDIA_TYPE_VIDEO, -1, -1, nullptr, 0);
    if (video_stream_idx_ >= 0) {
        AVStream* video_stream = format_ctx_->streams[video_stream_idx_];
        const AVCodec* video_codec = avcodec_find_decoder(video_stream->codecpar->codec_id);
        
        if (!video_codec) {
            LOG_ERROR("不支持的视频编解码器");
            return false;
        }
        
        video_codec_ctx_ = avcodec_alloc_context3(video_codec);
        if (!video_codec_ctx_) {
            LOG_ERROR("无法分配视频编解码器上下文");
            return false;
        }
        
        if (avcodec_parameters_to_context(video_codec_ctx_, video_stream->codecpar) < 0) {
            LOG_ERROR("无法复制视频编解码器参数");
            return false;
        }
        
        if (avcodec_open2(video_codec_ctx_, video_codec, nullptr) < 0) {
            LOG_ERROR("无法打开视频编解码器");
            return false;
        }
        
        LOG_INFO("视频流初始化成功: %dx%d, %d fps", 
                video_codec_ctx_->width, video_codec_ctx_->height, 
                video_stream->r_frame_rate.num / video_stream->r_frame_rate.den);
    }
    
    // 查找音频流
    audio_stream_idx_ = av_find_best_stream(format_ctx_, AVMEDIA_TYPE_AUDIO, -1, -1, nullptr, 0);
    if (audio_stream_idx_ >= 0) {
        AVStream* audio_stream = format_ctx_->streams[audio_stream_idx_];
        const AVCodec* audio_codec = avcodec_find_decoder(audio_stream->codecpar->codec_id);
        
        if (!audio_codec) {
            LOG_ERROR("不支持的音频编解码器");
            return false;
        }
        
        audio_codec_ctx_ = avcodec_alloc_context3(audio_codec);
        if (!audio_codec_ctx_) {
            LOG_ERROR("无法分配音频编解码器上下文");
            return false;
        }
        
        if (avcodec_parameters_to_context(audio_codec_ctx_, audio_stream->codecpar) < 0) {
            LOG_ERROR("无法复制音频编解码器参数");
            return false;
        }
        
        if (avcodec_open2(audio_codec_ctx_, audio_codec, nullptr) < 0) {
            LOG_ERROR("无法打开音频编解码器");
            return false;
        }
        
        LOG_INFO("音频流初始化成功: %d Hz, %d 声道", 
                audio_codec_ctx_->sample_rate, audio_codec_ctx_->channels);
    }
    
    return true;
}

bool FFmpegProcessor::process_video_frame(AVFrame* frame, int64_t timestamp) {
    if (!frame || !video_compressor_ || !detection_engine_) {
        return false;
    }
    
    Timer timer;
    timer.start();
    
    // 将AVFrame转换为字节数据
    std::vector<uint8_t> frame_data;
    int frame_size = frame->linesize[0] * frame->height * 3; // RGB
    frame_data.resize(frame_size);
    
    // 这里需要根据实际的像素格式进行转换
    // 简化处理，假设是RGB格式
    memcpy(frame_data.data(), frame->data[0], frame_size);
    
    // 压缩视频帧
    FrameInfo frame_info;
    frame_info.width = frame->width;
    frame_info.height = frame->height;
    frame_info.channels = 3;
    frame_info.timestamp = timestamp;
    frame_info.is_keyframe = frame->key_frame;
    frame_info.pixel_format = "RGB";
    
    auto compression_result = video_compressor_->compress_frame(frame_data, frame_info);
    
    if (compression_result.success) {
        // 检测伪造
        auto detection_result = detection_engine_->detect_video_frame(
            compression_result.compressed_data,
            config_.target_width,
            config_.target_height,
            3
        );
        
        timer.stop();
        
        // 更新统计信息
        {
            std::lock_guard<std::mutex> lock(stats_mutex_);
            stats_.frames_processed++;
            if (detection_result.is_fake) {
                stats_.fake_detections++;
            }
            stats_.average_processing_time_ms = 
                (stats_.average_processing_time_ms * (stats_.frames_processed - 1) + timer.elapsed_ms()) / stats_.frames_processed;
            stats_.compression_ratio = compression_result.compression_ratio;
        }
        
        // 调用回调
        if (frame_callback_) {
            FrameData frame_data_callback;
            frame_data_callback.data = compression_result.compressed_data;
            frame_data_callback.width = config_.target_width;
            frame_data_callback.height = config_.target_height;
            frame_data_callback.channels = 3;
            frame_data_callback.timestamp = timestamp;
            frame_data_callback.is_keyframe = frame->key_frame;
            frame_data_callback.frame_type = "video";
            
            frame_callback_(frame_data_callback);
        }
        
        if (result_callback_) {
            ProcessingResult result;
            result.is_fake = detection_result.is_fake;
            result.confidence = detection_result.confidence;
            result.detection_type = detection_result.details;
            result.processing_time_ms = timer.elapsed_ms();
            result.details = "视频帧检测完成";
            
            result_callback_(result);
        }
        
        return true;
    }
    
    return false;
}

bool FFmpegProcessor::process_audio_frame(AVFrame* frame, int64_t timestamp) {
    if (!frame || !audio_processor_ || !detection_engine_) {
        return false;
    }
    
    Timer timer;
    timer.start();
    
    // 将音频数据转换为浮点数组
    std::vector<float> audio_data;
    int samples = frame->nb_samples * frame->channels;
    audio_data.resize(samples);
    
    // 根据音频格式转换
    if (frame->format == AV_SAMPLE_FMT_FLT) {
        memcpy(audio_data.data(), frame->data[0], samples * sizeof(float));
    } else if (frame->format == AV_SAMPLE_FMT_S16) {
        int16_t* samples_int16 = (int16_t*)frame->data[0];
        for (int i = 0; i < samples; i++) {
            audio_data[i] = samples_int16[i] / 32768.0f;
        }
    } else {
        LOG_WARNING("不支持的音频格式: %d", frame->format);
        return false;
    }
    
    // 处理音频
    auto audio_result = audio_processor_->process_float_audio(
        audio_data, frame->sample_rate, frame->channels
    );
    
    if (audio_result.success) {
        // 检测音频伪造
        auto detection_result = detection_engine_->detect_audio_frame(
            audio_result.processed_audio,
            audio_processor_->get_config().target_sample_rate,
            audio_processor_->get_config().target_channels
        );
        
        timer.stop();
        
        // 调用回调
        if (result_callback_) {
            ProcessingResult result;
            result.is_fake = detection_result.is_fake;
            result.confidence = detection_result.confidence;
            result.detection_type = detection_result.details;
            result.processing_time_ms = timer.elapsed_ms();
            result.details = "音频帧检测完成";
            
            result_callback_(result);
        }
        
        return true;
    }
    
    return false;
}

void FFmpegProcessor::processing_thread() {
    LOG_INFO("处理线程启动");
    
    // 打开输入流
    if (avformat_open_input(&format_ctx_, input_url_.c_str(), nullptr, nullptr) != 0) {
        LOG_ERROR("无法打开输入流: %s", input_url_.c_str());
        return;
    }
    
    // 查找流信息
    if (avformat_find_stream_info(format_ctx_, nullptr) < 0) {
        LOG_ERROR("无法查找流信息");
        return;
    }
    
    // 初始化编解码器
    if (!initialize_codecs()) {
        LOG_ERROR("编解码器初始化失败");
        return;
    }
    
    // 分配帧和包
    AVFrame* frame = av_frame_alloc();
    AVPacket* packet = av_packet_alloc();
    
    if (!frame || !packet) {
        LOG_ERROR("无法分配帧或包");
        return;
    }
    
    // 主处理循环
    while (!should_stop_) {
        int ret = av_read_frame(format_ctx_, packet);
        if (ret < 0) {
            if (ret == AVERROR_EOF) {
                LOG_INFO("到达文件末尾");
                break;
            }
            LOG_WARNING("读取帧失败: %d", ret);
            continue;
        }
        
        // 处理视频包
        if (packet->stream_index == video_stream_idx_ && video_codec_ctx_) {
            ret = avcodec_send_packet(video_codec_ctx_, packet);
            if (ret < 0) {
                LOG_WARNING("发送视频包失败: %d", ret);
                av_packet_unref(packet);
                continue;
            }
            
            while (ret >= 0 && !should_stop_) {
                ret = avcodec_receive_frame(video_codec_ctx_, frame);
                if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
                    break;
                } else if (ret < 0) {
                    LOG_WARNING("接收视频帧失败: %d", ret);
                    break;
                }
                
                process_video_frame(frame, frame->pts);
                av_frame_unref(frame);
            }
        }
        
        // 处理音频包
        if (packet->stream_index == audio_stream_idx_ && audio_codec_ctx_) {
            ret = avcodec_send_packet(audio_codec_ctx_, packet);
            if (ret < 0) {
                LOG_WARNING("发送音频包失败: %d", ret);
                av_packet_unref(packet);
                continue;
            }
            
            while (ret >= 0 && !should_stop_) {
                ret = avcodec_receive_frame(audio_codec_ctx_, frame);
                if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) {
                    break;
                } else if (ret < 0) {
                    LOG_WARNING("接收音频帧失败: %d", ret);
                    break;
                }
                
                process_audio_frame(frame, frame->pts);
                av_frame_unref(frame);
            }
        }
        
        av_packet_unref(packet);
    }
    
    // 清理
    av_frame_free(&frame);
    av_packet_free(&packet);
    
    LOG_INFO("处理线程结束");
}

void FFmpegProcessor::cleanup() {
    if (video_codec_ctx_) {
        avcodec_free_context(&video_codec_ctx_);
    }
    
    if (audio_codec_ctx_) {
        avcodec_free_context(&audio_codec_ctx_);
    }
    
    if (format_ctx_) {
        avformat_close_input(&format_ctx_);
    }
    
    if (sws_ctx_) {
        sws_freeContext(sws_ctx_);
        sws_ctx_ = nullptr;
    }
    
    if (swr_ctx_) {
        swr_free(&swr_ctx_);
    }
}

} // namespace ffmpeg_detection 