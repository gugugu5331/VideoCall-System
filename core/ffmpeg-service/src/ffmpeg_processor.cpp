#include "../include/ffmpeg_processor.h"
#include <iostream>
#include <chrono>
#include <algorithm>
#include <cstring>

extern "C" {
#include <libavutil/imgutils.h>
#include <libavutil/samplefmt.h>
#include <libavutil/opt.h>
}

namespace ffmpeg_service {

// FFmpegProcessor 实现
FFmpegProcessor::FFmpegProcessor() 
    : initialized_(false), processing_(false), should_stop_(false) {
    // 初始化FFmpeg
    av_register_all();
    avcodec_register_all();
    avformat_network_init();
}

FFmpegProcessor::~FFmpegProcessor() {
    cleanup();
    avformat_network_deinit();
}

bool FFmpegProcessor::initialize(const EncodingParams& params) {
    if (initialized_) {
        return true;
    }
    
    try {
        current_params_ = params;
        
        // 初始化各个组件
        if (!initializeCodecs()) {
            std::cerr << "Failed to initialize codecs" << std::endl;
            return false;
        }
        
        if (!initializeFilters()) {
            std::cerr << "Failed to initialize filters" << std::endl;
            return false;
        }
        
        initialized_ = true;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Initialization error: " << e.what() << std::endl;
        return false;
    }
}

void FFmpegProcessor::cleanup() {
    if (processing_) {
        stopRealTimeProcessing();
    }
    
    if (video_processor_) {
        video_processor_->cleanup();
    }
    
    if (audio_processor_) {
        audio_processor_->cleanup();
    }
    
    if (compressor_) {
        compressor_->cleanup();
    }
    
    initialized_ = false;
}

bool FFmpegProcessor::processVideoFrame(const std::vector<uint8_t>& frame_data, 
                                       int width, int height, 
                                       AVPixelFormat format) {
    if (!initialized_ || !video_processor_) {
        return false;
    }
    
    try {
        MediaFrame frame;
        frame.data = frame_data;
        frame.width = width;
        frame.height = height;
        frame.pixel_format = format;
        frame.timestamp = av_gettime();
        
        std::lock_guard<std::mutex> lock(frame_queue_mutex_);
        frame_queue_.push(frame);
        frame_cv_.notify_one();
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Video frame processing error: " << e.what() << std::endl;
        return false;
    }
}

bool FFmpegProcessor::processAudioFrame(const std::vector<uint8_t>& audio_data,
                                       int sample_rate, int channels,
                                       AVSampleFormat format) {
    if (!initialized_ || !audio_processor_) {
        return false;
    }
    
    try {
        MediaFrame frame;
        frame.data = audio_data;
        frame.sample_rate = sample_rate;
        frame.channels = channels;
        frame.sample_format = format;
        frame.timestamp = av_gettime();
        
        std::lock_guard<std::mutex> lock(frame_queue_mutex_);
        frame_queue_.push(frame);
        frame_cv_.notify_one();
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Audio frame processing error: " << e.what() << std::endl;
        return false;
    }
}

ProcessingResult FFmpegProcessor::compressVideo(const std::vector<uint8_t>& video_data,
                                               const EncodingParams& params) {
    ProcessingResult result;
    
    if (!initialized_ || !video_processor_) {
        result.error_message = "Processor not initialized";
        return result;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        result = video_processor_->compress(video_data, params);
        
        auto end_time = std::chrono::high_resolution_clock::now();
        result.processing_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            end_time - start_time).count();
        
        if (result.success && !video_data.empty()) {
            result.compression_ratio = static_cast<float>(video_data.size()) / 
                                     static_cast<float>(result.processed_data.size());
        }
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult FFmpegProcessor::compressAudio(const std::vector<uint8_t>& audio_data,
                                               const EncodingParams& params) {
    ProcessingResult result;
    
    if (!initialized_ || !audio_processor_) {
        result.error_message = "Processor not initialized";
        return result;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        result = audio_processor_->compress(audio_data, params);
        
        auto end_time = std::chrono::high_resolution_clock::now();
        result.processing_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            end_time - start_time).count();
        
        if (result.success && !audio_data.empty()) {
            result.compression_ratio = static_cast<float>(audio_data.size()) / 
                                     static_cast<float>(result.processed_data.size());
        }
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult FFmpegProcessor::decompressVideo(const std::vector<uint8_t>& compressed_data) {
    if (!initialized_ || !video_processor_) {
        ProcessingResult result;
        result.error_message = "Processor not initialized";
        return result;
    }
    
    return video_processor_->decompress(compressed_data);
}

ProcessingResult FFmpegProcessor::decompressAudio(const std::vector<uint8_t>& compressed_data) {
    if (!initialized_ || !audio_processor_) {
        ProcessingResult result;
        result.error_message = "Processor not initialized";
        return result;
    }
    
    return audio_processor_->decompress(compressed_data);
}

ProcessingResult FFmpegProcessor::convertVideoFormat(const std::vector<uint8_t>& video_data,
                                                    AVPixelFormat target_format,
                                                    int target_width, int target_height) {
    if (!initialized_ || !video_processor_) {
        ProcessingResult result;
        result.error_message = "Processor not initialized";
        return result;
    }
    
    return video_processor_->convertFormat(video_data, target_format, target_width, target_height);
}

ProcessingResult FFmpegProcessor::convertAudioFormat(const std::vector<uint8_t>& audio_data,
                                                    AVSampleFormat target_format,
                                                    int target_sample_rate, int target_channels) {
    if (!initialized_ || !audio_processor_) {
        ProcessingResult result;
        result.error_message = "Processor not initialized";
        return result;
    }
    
    return audio_processor_->convertFormat(audio_data, target_format, target_sample_rate, target_channels);
}

void FFmpegProcessor::startRealTimeProcessing(FrameCallback video_callback,
                                             FrameCallback audio_callback) {
    if (processing_) {
        return;
    }
    
    video_callback_ = video_callback;
    audio_callback_ = audio_callback;
    processing_ = true;
    should_stop_ = false;
    
    processing_thread_ = std::thread(&FFmpegProcessor::processingLoop, this);
}

void FFmpegProcessor::stopRealTimeProcessing() {
    if (!processing_) {
        return;
    }
    
    should_stop_ = true;
    frame_cv_.notify_all();
    
    if (processing_thread_.joinable()) {
        processing_thread_.join();
    }
    
    processing_ = false;
}

void FFmpegProcessor::setEncodingParams(const EncodingParams& params) {
    current_params_ = params;
    
    if (video_processor_) {
        video_processor_->cleanup();
        video_processor_->initialize(params);
    }
    
    if (audio_processor_) {
        audio_processor_->cleanup();
        audio_processor_->initialize(params);
    }
}

void FFmpegProcessor::setProcessingCallback(ProcessingCallback callback) {
    processing_callback_ = callback;
}

bool FFmpegProcessor::initializeCodecs() {
    try {
        // 初始化视频处理器
        video_processor_ = std::make_unique<VideoProcessor>();
        if (!video_processor_->initialize(current_params_)) {
            return false;
        }
        
        // 初始化音频处理器
        audio_processor_ = std::make_unique<AudioProcessor>();
        if (!audio_processor_->initialize(current_params_)) {
            return false;
        }
        
        // 初始化压缩器
        compressor_ = std::make_unique<MediaCompressor>();
        if (!compressor_->initialize(current_params_)) {
            return false;
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Codec initialization error: " << e.what() << std::endl;
        return false;
    }
}

bool FFmpegProcessor::initializeFilters() {
    // 这里可以初始化视频和音频过滤器
    // 例如：降噪、锐化、均衡器等
    return true;
}

void FFmpegProcessor::processingLoop() {
    while (!should_stop_) {
        MediaFrame frame;
        
        {
            std::unique_lock<std::mutex> lock(frame_queue_mutex_);
            frame_cv_.wait(lock, [this] { return !frame_queue_.empty() || should_stop_; });
            
            if (should_stop_) {
                break;
            }
            
            if (!frame_queue_.empty()) {
                frame = frame_queue_.front();
                frame_queue_.pop();
            }
        }
        
        // 处理帧
        if (frame.pixel_format != AV_PIX_FMT_NONE) {
            handleVideoFrame(frame);
        } else if (frame.sample_format != AV_SAMPLE_FMT_NONE) {
            handleAudioFrame(frame);
        }
    }
}

void FFmpegProcessor::handleVideoFrame(const MediaFrame& frame) {
    if (video_callback_) {
        video_callback_(frame);
    }
    
    // 这里可以添加额外的视频处理逻辑
    // 例如：压缩、格式转换等
}

void FFmpegProcessor::handleAudioFrame(const MediaFrame& frame) {
    if (audio_callback_) {
        audio_callback_(frame);
    }
    
    // 这里可以添加额外的音频处理逻辑
    // 例如：压缩、格式转换等
}

// VideoProcessor 实现
VideoProcessor::VideoProcessor() 
    : encoder_ctx_(nullptr), decoder_ctx_(nullptr), filter_ctx_(nullptr), 
      sws_ctx_(nullptr), initialized_(false) {
}

VideoProcessor::~VideoProcessor() {
    cleanup();
}

bool VideoProcessor::initialize(const EncodingParams& params) {
    if (initialized_) {
        return true;
    }
    
    try {
        params_ = params;
        
        if (!initializeCodec()) {
            return false;
        }
        
        if (!initializeFilter()) {
            return false;
        }
        
        initialized_ = true;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Video processor initialization error: " << e.what() << std::endl;
        return false;
    }
}

void VideoProcessor::cleanup() {
    cleanupCodec();
    
    if (sws_ctx_) {
        sws_freeContext(sws_ctx_);
        sws_ctx_ = nullptr;
    }
    
    if (filter_ctx_) {
        avfilter_free(filter_ctx_);
        filter_ctx_ = nullptr;
    }
    
    initialized_ = false;
}

ProcessingResult VideoProcessor::processFrame(const std::vector<uint8_t>& frame_data,
                                             int width, int height,
                                             AVPixelFormat format) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Video processor not initialized";
        return result;
    }
    
    try {
        // 创建AVFrame
        AVFrame* frame = av_frame_alloc();
        if (!frame) {
            result.error_message = "Failed to allocate frame";
            return result;
        }
        
        frame->width = width;
        frame->height = height;
        frame->format = format;
        
        if (av_frame_get_buffer(frame, 0) < 0) {
            av_frame_free(&frame);
            result.error_message = "Failed to allocate frame buffer";
            return result;
        }
        
        // 复制数据到frame
        if (av_image_fill_arrays(frame->data, frame->linesize, frame_data.data(),
                                format, width, height, 1) < 0) {
            av_frame_free(&frame);
            result.error_message = "Failed to fill frame data";
            return result;
        }
        
        // 编码帧
        AVPacket packet;
        av_init_packet(&packet);
        packet.data = nullptr;
        packet.size = 0;
        
        int ret = avcodec_send_frame(encoder_ctx_, frame);
        if (ret < 0) {
            av_frame_free(&frame);
            result.error_message = "Failed to send frame to encoder";
            return result;
        }
        
        ret = avcodec_receive_packet(encoder_ctx_, &packet);
        if (ret >= 0) {
            result.processed_data.assign(packet.data, packet.data + packet.size);
            result.success = true;
        }
        
        av_packet_unref(&packet);
        av_frame_free(&frame);
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult VideoProcessor::compress(const std::vector<uint8_t>& video_data,
                                         const EncodingParams& params) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Video processor not initialized";
        return result;
    }
    
    try {
        // 这里实现视频压缩逻辑
        // 可以使用H.264/H.265编码器进行压缩
        
        // 临时实现：简单复制数据
        result.processed_data = video_data;
        result.success = true;
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult VideoProcessor::decompress(const std::vector<uint8_t>& compressed_data) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Video processor not initialized";
        return result;
    }
    
    try {
        // 这里实现视频解压缩逻辑
        
        // 临时实现：简单复制数据
        result.processed_data = compressed_data;
        result.success = true;
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult VideoProcessor::convertFormat(const std::vector<uint8_t>& video_data,
                                              AVPixelFormat target_format,
                                              int target_width, int target_height) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Video processor not initialized";
        return result;
    }
    
    try {
        // 创建源帧
        AVFrame* src_frame = av_frame_alloc();
        AVFrame* dst_frame = av_frame_alloc();
        
        if (!src_frame || !dst_frame) {
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to allocate frames";
            return result;
        }
        
        // 设置目标帧参数
        dst_frame->width = target_width;
        dst_frame->height = target_height;
        dst_frame->format = target_format;
        
        if (av_frame_get_buffer(dst_frame, 0) < 0) {
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to allocate destination frame buffer";
            return result;
        }
        
        // 创建缩放上下文
        SwsContext* sws_ctx = sws_getContext(
            params_.video_width, params_.video_height, params_.video_pixel_format,
            target_width, target_height, target_format,
            SWS_BILINEAR, nullptr, nullptr, nullptr);
        
        if (!sws_ctx) {
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to create scaling context";
            return result;
        }
        
        // 填充源帧数据
        if (av_image_fill_arrays(src_frame->data, src_frame->linesize, video_data.data(),
                                params_.video_pixel_format, params_.video_width, 
                                params_.video_height, 1) < 0) {
            sws_freeContext(sws_ctx);
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to fill source frame data";
            return result;
        }
        
        // 执行格式转换
        if (sws_scale(sws_ctx, src_frame->data, src_frame->linesize, 0, params_.video_height,
                     dst_frame->data, dst_frame->linesize) < 0) {
            sws_freeContext(sws_ctx);
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to scale frame";
            return result;
        }
        
        // 提取转换后的数据
        int buffer_size = av_image_get_buffer_size(target_format, target_width, target_height, 1);
        result.processed_data.resize(buffer_size);
        
        if (av_image_copy_to_buffer(result.processed_data.data(), buffer_size,
                                   dst_frame->data, dst_frame->linesize,
                                   target_format, target_width, target_height, 1) < 0) {
            sws_freeContext(sws_ctx);
            av_frame_free(&src_frame);
            av_frame_free(&dst_frame);
            result.error_message = "Failed to copy frame data";
            return result;
        }
        
        result.success = true;
        
        // 清理
        sws_freeContext(sws_ctx);
        av_frame_free(&src_frame);
        av_frame_free(&dst_frame);
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

bool VideoProcessor::initializeCodec() {
    try {
        // 查找编码器
        const AVCodec* encoder = avcodec_find_encoder(params_.video_codec_id);
        if (!encoder) {
            std::cerr << "Failed to find video encoder" << std::endl;
            return false;
        }
        
        // 创建编码器上下文
        encoder_ctx_ = avcodec_alloc_context3(encoder);
        if (!encoder_ctx_) {
            std::cerr << "Failed to allocate encoder context" << std::endl;
            return false;
        }
        
        // 设置编码参数
        encoder_ctx_->width = params_.video_width;
        encoder_ctx_->height = params_.video_height;
        encoder_ctx_->time_base = {1, params_.video_fps};
        encoder_ctx_->framerate = {params_.video_fps, 1};
        encoder_ctx_->pix_fmt = params_.video_pixel_format;
        encoder_ctx_->bit_rate = params_.video_bitrate;
        encoder_ctx_->gop_size = params_.gop_size;
        encoder_ctx_->max_b_frames = params_.max_b_frames;
        
        // 打开编码器
        if (avcodec_open2(encoder_ctx_, encoder, nullptr) < 0) {
            std::cerr << "Failed to open video encoder" << std::endl;
            return false;
        }
        
        // 查找解码器
        const AVCodec* decoder = avcodec_find_decoder(params_.video_codec_id);
        if (!decoder) {
            std::cerr << "Failed to find video decoder" << std::endl;
            return false;
        }
        
        // 创建解码器上下文
        decoder_ctx_ = avcodec_alloc_context3(decoder);
        if (!decoder_ctx_) {
            std::cerr << "Failed to allocate decoder context" << std::endl;
            return false;
        }
        
        // 打开解码器
        if (avcodec_open2(decoder_ctx_, decoder, nullptr) < 0) {
            std::cerr << "Failed to open video decoder" << std::endl;
            return false;
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Video codec initialization error: " << e.what() << std::endl;
        return false;
    }
}

bool VideoProcessor::initializeFilter() {
    // 这里可以初始化视频过滤器
    // 例如：降噪、锐化等
    return true;
}

void VideoProcessor::cleanupCodec() {
    if (encoder_ctx_) {
        avcodec_free_context(&encoder_ctx_);
    }
    
    if (decoder_ctx_) {
        avcodec_free_context(&decoder_ctx_);
    }
}

// AudioProcessor 实现
AudioProcessor::AudioProcessor() 
    : encoder_ctx_(nullptr), decoder_ctx_(nullptr), swr_ctx_(nullptr), 
      initialized_(false) {
}

AudioProcessor::~AudioProcessor() {
    cleanup();
}

bool AudioProcessor::initialize(const EncodingParams& params) {
    if (initialized_) {
        return true;
    }
    
    try {
        params_ = params;
        
        if (!initializeCodec()) {
            return false;
        }
        
        if (!initializeResampler()) {
            return false;
        }
        
        initialized_ = true;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Audio processor initialization error: " << e.what() << std::endl;
        return false;
    }
}

void AudioProcessor::cleanup() {
    cleanupCodec();
    
    if (swr_ctx_) {
        swr_free(&swr_ctx_);
    }
    
    initialized_ = false;
}

ProcessingResult AudioProcessor::processFrame(const std::vector<uint8_t>& audio_data,
                                             int sample_rate, int channels,
                                             AVSampleFormat format) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Audio processor not initialized";
        return result;
    }
    
    try {
        // 创建AVFrame
        AVFrame* frame = av_frame_alloc();
        if (!frame) {
            result.error_message = "Failed to allocate frame";
            return result;
        }
        
        frame->nb_samples = audio_data.size() / (av_get_bytes_per_sample(format) * channels);
        frame->sample_rate = sample_rate;
        frame->channels = channels;
        frame->format = format;
        frame->channel_layout = av_get_default_channel_layout(channels);
        
        if (av_frame_get_buffer(frame, 0) < 0) {
            av_frame_free(&frame);
            result.error_message = "Failed to allocate frame buffer";
            return result;
        }
        
        // 复制音频数据
        memcpy(frame->data[0], audio_data.data(), audio_data.size());
        
        // 编码帧
        AVPacket packet;
        av_init_packet(&packet);
        packet.data = nullptr;
        packet.size = 0;
        
        int ret = avcodec_send_frame(encoder_ctx_, frame);
        if (ret < 0) {
            av_frame_free(&frame);
            result.error_message = "Failed to send frame to encoder";
            return result;
        }
        
        ret = avcodec_receive_packet(encoder_ctx_, &packet);
        if (ret >= 0) {
            result.processed_data.assign(packet.data, packet.data + packet.size);
            result.success = true;
        }
        
        av_packet_unref(&packet);
        av_frame_free(&frame);
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult AudioProcessor::compress(const std::vector<uint8_t>& audio_data,
                                         const EncodingParams& params) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Audio processor not initialized";
        return result;
    }
    
    try {
        // 这里实现音频压缩逻辑
        // 可以使用AAC、MP3等编码器进行压缩
        
        // 临时实现：简单复制数据
        result.processed_data = audio_data;
        result.success = true;
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult AudioProcessor::decompress(const std::vector<uint8_t>& compressed_data) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Audio processor not initialized";
        return result;
    }
    
    try {
        // 这里实现音频解压缩逻辑
        
        // 临时实现：简单复制数据
        result.processed_data = compressed_data;
        result.success = true;
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

ProcessingResult AudioProcessor::convertFormat(const std::vector<uint8_t>& audio_data,
                                              AVSampleFormat target_format,
                                              int target_sample_rate, int target_channels) {
    ProcessingResult result;
    
    if (!initialized_) {
        result.error_message = "Audio processor not initialized";
        return result;
    }
    
    try {
        // 使用重采样器进行格式转换
        if (!swr_ctx_) {
            result.error_message = "Resampler not initialized";
            return result;
        }
        
        // 计算源音频参数
        int src_samples = audio_data.size() / (av_get_bytes_per_sample(params_.audio_sample_format) * params_.audio_channels);
        
        // 计算目标音频大小
        int dst_samples = av_rescale_rnd(swr_get_delay(swr_ctx_, params_.audio_sample_rate) + src_samples,
                                       target_sample_rate, params_.audio_sample_rate, AV_ROUND_UP);
        
        // 分配目标缓冲区
        int dst_buffer_size = av_samples_get_buffer_size(nullptr, target_channels, dst_samples,
                                                        target_format, 0);
        result.processed_data.resize(dst_buffer_size);
        
        // 执行重采样
        uint8_t* dst_data = result.processed_data.data();
        const uint8_t* src_data = audio_data.data();
        
        int samples_converted = swr_convert(swr_ctx_, &dst_data, dst_samples,
                                          &src_data, src_samples);
        
        if (samples_converted < 0) {
            result.error_message = "Failed to convert audio format";
            return result;
        }
        
        // 调整缓冲区大小
        int actual_buffer_size = av_samples_get_buffer_size(nullptr, target_channels,
                                                          samples_converted, target_format, 0);
        result.processed_data.resize(actual_buffer_size);
        result.success = true;
        
    } catch (const std::exception& e) {
        result.error_message = e.what();
    }
    
    return result;
}

bool AudioProcessor::initializeCodec() {
    try {
        // 查找编码器
        const AVCodec* encoder = avcodec_find_encoder(params_.audio_codec_id);
        if (!encoder) {
            std::cerr << "Failed to find audio encoder" << std::endl;
            return false;
        }
        
        // 创建编码器上下文
        encoder_ctx_ = avcodec_alloc_context3(encoder);
        if (!encoder_ctx_) {
            std::cerr << "Failed to allocate encoder context" << std::endl;
            return false;
        }
        
        // 设置编码参数
        encoder_ctx_->sample_fmt = params_.audio_sample_format;
        encoder_ctx_->sample_rate = params_.audio_sample_rate;
        encoder_ctx_->channels = params_.audio_channels;
        encoder_ctx_->channel_layout = av_get_default_channel_layout(params_.audio_channels);
        encoder_ctx_->bit_rate = params_.audio_bitrate;
        
        // 打开编码器
        if (avcodec_open2(encoder_ctx_, encoder, nullptr) < 0) {
            std::cerr << "Failed to open audio encoder" << std::endl;
            return false;
        }
        
        // 查找解码器
        const AVCodec* decoder = avcodec_find_decoder(params_.audio_codec_id);
        if (!decoder) {
            std::cerr << "Failed to find audio decoder" << std::endl;
            return false;
        }
        
        // 创建解码器上下文
        decoder_ctx_ = avcodec_alloc_context3(decoder);
        if (!decoder_ctx_) {
            std::cerr << "Failed to allocate decoder context" << std::endl;
            return false;
        }
        
        // 打开解码器
        if (avcodec_open2(decoder_ctx_, decoder, nullptr) < 0) {
            std::cerr << "Failed to open audio decoder" << std::endl;
            return false;
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Audio codec initialization error: " << e.what() << std::endl;
        return false;
    }
}

bool AudioProcessor::initializeResampler() {
    try {
        // 创建重采样上下文
        swr_ctx_ = swr_alloc_set_opts(nullptr,
                                    av_get_default_channel_layout(params_.audio_channels),
                                    params_.audio_sample_format,
                                    params_.audio_sample_rate,
                                    av_get_default_channel_layout(params_.audio_channels),
                                    params_.audio_sample_format,
                                    params_.audio_sample_rate,
                                    0, nullptr);
        
        if (!swr_ctx_) {
            std::cerr << "Failed to allocate resampler context" << std::endl;
            return false;
        }
        
        // 初始化重采样器
        if (swr_init(swr_ctx_) < 0) {
            std::cerr << "Failed to initialize resampler" << std::endl;
            return false;
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Resampler initialization error: " << e.what() << std::endl;
        return false;
    }
}

void AudioProcessor::cleanupCodec() {
    if (encoder_ctx_) {
        avcodec_free_context(&encoder_ctx_);
    }
    
    if (decoder_ctx_) {
        avcodec_free_context(&decoder_ctx_);
    }
}

// MediaCompressor 实现
MediaCompressor::MediaCompressor() : initialized_(false) {
}

MediaCompressor::~MediaCompressor() {
    cleanup();
}

bool MediaCompressor::initialize(const EncodingParams& params) {
    if (initialized_) {
        return true;
    }
    
    try {
        params_ = params;
        
        video_processor_ = std::make_unique<VideoProcessor>();
        audio_processor_ = std::make_unique<AudioProcessor>();
        
        if (!video_processor_->initialize(params) || !audio_processor_->initialize(params)) {
            return false;
        }
        
        initialized_ = true;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Media compressor initialization error: " << e.what() << std::endl;
        return false;
    }
}

void MediaCompressor::cleanup() {
    if (video_processor_) {
        video_processor_->cleanup();
    }
    
    if (audio_processor_) {
        audio_processor_->cleanup();
    }
    
    initialized_ = false;
}

ProcessingResult MediaCompressor::compressVideo(const std::vector<uint8_t>& video_data,
                                               const EncodingParams& params) {
    if (!initialized_ || !video_processor_) {
        ProcessingResult result;
        result.error_message = "Media compressor not initialized";
        return result;
    }
    
    return video_processor_->compress(video_data, params);
}

ProcessingResult MediaCompressor::compressAudio(const std::vector<uint8_t>& audio_data,
                                               const EncodingParams& params) {
    if (!initialized_ || !audio_processor_) {
        ProcessingResult result;
        result.error_message = "Media compressor not initialized";
        return result;
    }
    
    return audio_processor_->compress(audio_data, params);
}

ProcessingResult MediaCompressor::decompressVideo(const std::vector<uint8_t>& compressed_data) {
    if (!initialized_ || !video_processor_) {
        ProcessingResult result;
        result.error_message = "Media compressor not initialized";
        return result;
    }
    
    return video_processor_->decompress(compressed_data);
}

ProcessingResult MediaCompressor::decompressAudio(const std::vector<uint8_t>& compressed_data) {
    if (!initialized_ || !audio_processor_) {
        ProcessingResult result;
        result.error_message = "Media compressor not initialized";
        return result;
    }
    
    return audio_processor_->decompress(compressed_data);
}

} // namespace ffmpeg_service 