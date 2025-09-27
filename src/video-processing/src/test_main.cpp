#include "video_processor.h"
#include "camera_capture.h"
#include <opencv2/opencv.hpp>
#include <iostream>
#include <memory>

int main() {
    std::cout << "=== Video Processing System Test ===" << std::endl;
    std::cout << "OpenCV Version: " << CV_VERSION << std::endl;
    
    try {
        // 初始化视频处理器
        auto video_processor = std::make_unique<VideoProcessor>();
        if (!video_processor->initialize()) {
            std::cerr << "Failed to initialize VideoProcessor" << std::endl;
            return -1;
        }
        
        // 初始化摄像头捕获
        auto camera_capture = std::make_unique<CameraCapture>();
        if (!camera_capture->initialize()) {
            std::cerr << "Failed to initialize CameraCapture" << std::endl;
            return -1;
        }
        
        std::cout << "System initialized successfully!" << std::endl;
        std::cout << "Controls:" << std::endl;
        std::cout << "  ESC - Exit" << std::endl;
        std::cout << "  SPACE - Take screenshot" << std::endl;
        std::cout << "  1-9 - Apply filters" << std::endl;
        std::cout << "  0 - Remove filters" << std::endl;
        std::cout << "  B - Beauty filter" << std::endl;
        std::cout << "  C - Cartoon filter" << std::endl;
        std::cout << "  V - Vintage filter" << std::endl;
        std::cout << "  G - Glasses sticker" << std::endl;
        std::cout << "  H - Hat sticker" << std::endl;
        std::cout << "  M - Mustache sticker" << std::endl;
        
        cv::Mat frame, processed_frame;
        int frame_count = 0;
        
        while (true) {
            // 捕获帧
            if (!camera_capture->captureFrame(frame)) {
                std::cerr << "Failed to capture frame" << std::endl;
                break;
            }
            
            if (frame.empty()) {
                continue;
            }
            
            // 处理帧
            if (video_processor->processFrame(frame, processed_frame)) {
                // 显示处理后的帧
                cv::imshow("Video Processing - Processed", processed_frame);
                
                // 显示原始帧（用于对比）
                cv::imshow("Video Processing - Original", frame);
                
                // 显示性能统计
                if (frame_count % 30 == 0) { // 每30帧显示一次
                    auto stats = video_processor->getStats();
                    std::cout << "FPS: " << stats.fps 
                             << ", Avg Processing Time: " << stats.avg_processing_time << "ms"
                             << ", Frame Count: " << stats.frame_count << std::endl;
                }
            } else {
                // 如果处理失败，显示原始帧
                cv::imshow("Video Processing - Original", frame);
            }
            
            // 处理键盘输入
            char key = cv::waitKey(1) & 0xFF;
            if (key == 27) { // ESC
                break;
            } else if (key == ' ') { // SPACE - 截图
                std::string filename = "screenshot_" + std::to_string(frame_count) + ".jpg";
                cv::imwrite(filename, processed_frame.empty() ? frame : processed_frame);
                std::cout << "Screenshot saved: " << filename << std::endl;
            } else if (key >= '0' && key <= '9') { // 数字键 - 滤镜
                FilterType filter_type = static_cast<FilterType>(key - '0');
                video_processor->setFilterType(filter_type);
                std::cout << "Applied filter: " << static_cast<int>(filter_type) << std::endl;
            } else if (key == 'b' || key == 'B') { // 美颜滤镜
                video_processor->setFilterType(FilterType::BEAUTY);
                std::cout << "Applied beauty filter" << std::endl;
            } else if (key == 'c' || key == 'C') { // 卡通滤镜
                video_processor->setFilterType(FilterType::CARTOON);
                std::cout << "Applied cartoon filter" << std::endl;
            } else if (key == 'v' || key == 'V') { // 复古滤镜
                video_processor->setFilterType(FilterType::VINTAGE);
                std::cout << "Applied vintage filter" << std::endl;
            } else if (key == 'g' || key == 'G') { // 眼镜贴纸
                video_processor->addSticker("../assets/stickers/glasses.png", StickerType::GLASSES);
                std::cout << "Added glasses sticker" << std::endl;
            } else if (key == 'h' || key == 'H') { // 帽子贴纸
                video_processor->addSticker("../assets/stickers/hat.png", StickerType::HAT);
                std::cout << "Added hat sticker" << std::endl;
            } else if (key == 'm' || key == 'M') { // 胡子贴纸
                video_processor->addSticker("../assets/stickers/mustache.png", StickerType::MUSTACHE);
                std::cout << "Added mustache sticker" << std::endl;
            } else if (key == 'r' || key == 'R') { // 重置
                video_processor->setFilterType(FilterType::NONE);
                video_processor->removeSticker(StickerType::GLASSES);
                video_processor->removeSticker(StickerType::HAT);
                video_processor->removeSticker(StickerType::MUSTACHE);
                std::cout << "Reset all effects" << std::endl;
            }
            
            frame_count++;
        }
        
        // 显示最终统计
        auto final_stats = video_processor->getStats();
        std::cout << "\n=== Final Statistics ===" << std::endl;
        std::cout << "Total frames processed: " << final_stats.frame_count << std::endl;
        std::cout << "Average FPS: " << final_stats.fps << std::endl;
        std::cout << "Average processing time: " << final_stats.avg_processing_time << "ms" << std::endl;
        
        // 清理
        video_processor->cleanup();
        camera_capture->cleanup();
        
        cv::destroyAllWindows();
        
    } catch (const std::exception& e) {
        std::cerr << "Exception: " << e.what() << std::endl;
        return -1;
    }
    
    std::cout << "Video processing system terminated successfully." << std::endl;
    return 0;
}
