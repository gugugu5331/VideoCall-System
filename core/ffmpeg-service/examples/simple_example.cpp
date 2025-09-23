#include <iostream>
#include <string>
#include <vector>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libswscale/swscale.h>
#include <libswresample/swresample.h>
}

int main() {
    std::cout << "=== FFmpeg服务简单示例程序 ===" << std::endl;
    
    // 初始化FFmpeg
    std::cout << "初始化FFmpeg..." << std::endl;
    
    // 获取FFmpeg版本信息
    std::cout << "FFmpeg版本: " << av_version_info() << std::endl;
    
    // 列出支持的格式
    std::cout << "\n支持的输入格式:" << std::endl;
    const AVInputFormat* input_format = nullptr;
    void* opaque = nullptr;
    while ((input_format = av_demuxer_iterate(&opaque)) != nullptr) {
        std::cout << "  " << input_format->name << " - " << input_format->long_name << std::endl;
        break; // 只显示第一个作为示例
    }
    
    // 列出支持的编码器
    std::cout << "\n支持的视频编码器:" << std::endl;
    const AVCodec* codec = nullptr;
    void* codec_opaque = nullptr;
    while ((codec = av_codec_iterate(&codec_opaque)) != nullptr) {
        if (codec->type == AVMEDIA_TYPE_VIDEO && av_codec_is_encoder(codec)) {
            std::cout << "  " << codec->name << " - " << codec->long_name << std::endl;
            break; // 只显示第一个作为示例
        }
    }
    
    // 列出支持的音频编码器
    std::cout << "\n支持的音频编码器:" << std::endl;
    codec_opaque = nullptr;
    while ((codec = av_codec_iterate(&codec_opaque)) != nullptr) {
        if (codec->type == AVMEDIA_TYPE_AUDIO && av_codec_is_encoder(codec)) {
            std::cout << "  " << codec->name << " - " << codec->long_name << std::endl;
            break; // 只显示第一个作为示例
        }
    }
    
    std::cout << "\n=== FFmpeg Service Initialized Successfully! ===" << std::endl;
    std::cout << "Basic functionality test completed." << std::endl;
    
    return 0;
} 