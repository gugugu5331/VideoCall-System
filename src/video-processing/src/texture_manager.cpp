#include "texture_manager.h"
#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <iostream>
#include <filesystem>

TextureManager::TextureManager()
    : initialized_(false)
{
}

TextureManager::~TextureManager() {
    cleanup();
}

bool TextureManager::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 加载默认贴纸
        loadDefaultStickers();
        
        initialized_ = true;
        std::cout << "TextureManager initialized successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Error initializing TextureManager: " << e.what() << std::endl;
        return false;
    }
}

void TextureManager::cleanup() {
    stickers_.clear();
    active_stickers_.clear();
    initialized_ = false;
}

void TextureManager::applyTextures(cv::Mat& frame, const std::vector<FaceInfo>& faces) {
    if (!initialized_ || faces.empty() || active_stickers_.empty()) {
        return;
    }

    try {
        for (const auto& face : faces) {
            for (const auto& sticker_pair : active_stickers_) {
                applyStickerToFace(frame, face, sticker_pair.second);
            }
        }
    } catch (const std::exception& e) {
        std::cerr << "Error applying textures: " << e.what() << std::endl;
    }
}

bool TextureManager::loadSticker(const std::string& path, StickerType type) {
    try {
        cv::Mat sticker_image = cv::imread(path, cv::IMREAD_UNCHANGED);
        if (sticker_image.empty()) {
            std::cerr << "Failed to load sticker: " << path << std::endl;
            return false;
        }

        StickerInfo sticker;
        sticker.image = sticker_image;
        sticker.type = type;
        sticker.path = path;
        sticker.scale = 1.0f;
        sticker.rotation = 0.0f;
        sticker.opacity = 1.0f;
        sticker.anchor_point = getDefaultAnchorPoint(type);

        stickers_[type] = sticker;
        std::cout << "Loaded sticker: " << path << " (type: " << static_cast<int>(type) << ")" << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error loading sticker: " << e.what() << std::endl;
        return false;
    }
}

void TextureManager::removeSticker(StickerType type) {
    auto it = active_stickers_.find(type);
    if (it != active_stickers_.end()) {
        active_stickers_.erase(it);
        std::cout << "Removed active sticker type: " << static_cast<int>(type) << std::endl;
    }
}

void TextureManager::activateSticker(StickerType type) {
    auto it = stickers_.find(type);
    if (it != stickers_.end()) {
        active_stickers_[type] = it->second;
        std::cout << "Activated sticker type: " << static_cast<int>(type) << std::endl;
    }
}

void TextureManager::deactivateSticker(StickerType type) {
    removeSticker(type);
}

void TextureManager::setStickerScale(StickerType type, float scale) {
    auto it = active_stickers_.find(type);
    if (it != active_stickers_.end()) {
        it->second.scale = std::clamp(scale, 0.1f, 3.0f);
    }
}

void TextureManager::setStickerRotation(StickerType type, float rotation) {
    auto it = active_stickers_.find(type);
    if (it != active_stickers_.end()) {
        it->second.rotation = rotation;
    }
}

void TextureManager::setStickerOpacity(StickerType type, float opacity) {
    auto it = active_stickers_.find(type);
    if (it != active_stickers_.end()) {
        it->second.opacity = std::clamp(opacity, 0.0f, 1.0f);
    }
}

void TextureManager::applyStickerToFace(cv::Mat& frame, const FaceInfo& face, const StickerInfo& sticker) {
    if (sticker.image.empty() || face.landmarks.empty()) {
        return;
    }

    try {
        // 计算贴纸位置和大小
        cv::Point2f position = calculateStickerPosition(face, sticker);
        cv::Size2f size = calculateStickerSize(face, sticker);

        // 调整贴纸图像大小
        cv::Mat resized_sticker;
        cv::resize(sticker.image, resized_sticker, cv::Size(static_cast<int>(size.width), static_cast<int>(size.height)));

        // 应用旋转
        if (std::abs(sticker.rotation) > 0.01f) {
            resized_sticker = rotateImage(resized_sticker, sticker.rotation);
        }

        // 计算贴纸在帧中的位置
        cv::Rect sticker_rect(
            static_cast<int>(position.x - size.width / 2),
            static_cast<int>(position.y - size.height / 2),
            resized_sticker.cols,
            resized_sticker.rows
        );

        // 确保贴纸在帧范围内
        cv::Rect frame_rect(0, 0, frame.cols, frame.rows);
        cv::Rect valid_rect = sticker_rect & frame_rect;

        if (valid_rect.width > 0 && valid_rect.height > 0) {
            // 计算贴纸图像中对应的区域
            cv::Rect sticker_roi(
                valid_rect.x - sticker_rect.x,
                valid_rect.y - sticker_rect.y,
                valid_rect.width,
                valid_rect.height
            );

            // 应用贴纸到帧上
            blendSticker(frame(valid_rect), resized_sticker(sticker_roi), sticker.opacity);
        }

    } catch (const std::exception& e) {
        std::cerr << "Error applying sticker to face: " << e.what() << std::endl;
    }
}

cv::Point2f TextureManager::calculateStickerPosition(const FaceInfo& face, const StickerInfo& sticker) {
    cv::Point2f position;

    switch (sticker.anchor_point) {
        case AnchorPoint::FACE_CENTER:
            position = cv::Point2f(
                face.bounding_box.x + face.bounding_box.width / 2.0f,
                face.bounding_box.y + face.bounding_box.height / 2.0f
            );
            break;

        case AnchorPoint::LEFT_EYE:
            if (face.landmarks.size() > 0) {
                position = face.landmarks[0]; // 假设第一个关键点是左眼
            } else {
                position = cv::Point2f(
                    face.bounding_box.x + face.bounding_box.width * 0.3f,
                    face.bounding_box.y + face.bounding_box.height * 0.4f
                );
            }
            break;

        case AnchorPoint::RIGHT_EYE:
            if (face.landmarks.size() > 1) {
                position = face.landmarks[1]; // 假设第二个关键点是右眼
            } else {
                position = cv::Point2f(
                    face.bounding_box.x + face.bounding_box.width * 0.7f,
                    face.bounding_box.y + face.bounding_box.height * 0.4f
                );
            }
            break;

        case AnchorPoint::NOSE:
            if (face.landmarks.size() > 2) {
                position = face.landmarks[2]; // 假设第三个关键点是鼻子
            } else {
                position = cv::Point2f(
                    face.bounding_box.x + face.bounding_box.width / 2.0f,
                    face.bounding_box.y + face.bounding_box.height * 0.6f
                );
            }
            break;

        case AnchorPoint::MOUTH:
            if (face.landmarks.size() > 3) {
                // 计算嘴部中心
                cv::Point2f mouth_center = (face.landmarks[3] + face.landmarks[4]) / 2.0f;
                position = mouth_center;
            } else {
                position = cv::Point2f(
                    face.bounding_box.x + face.bounding_box.width / 2.0f,
                    face.bounding_box.y + face.bounding_box.height * 0.8f
                );
            }
            break;

        case AnchorPoint::FOREHEAD:
            position = cv::Point2f(
                face.bounding_box.x + face.bounding_box.width / 2.0f,
                face.bounding_box.y + face.bounding_box.height * 0.2f
            );
            break;

        default:
            position = cv::Point2f(
                face.bounding_box.x + face.bounding_box.width / 2.0f,
                face.bounding_box.y + face.bounding_box.height / 2.0f
            );
            break;
    }

    return position;
}

cv::Size2f TextureManager::calculateStickerSize(const FaceInfo& face, const StickerInfo& sticker) {
    float base_size = std::min(face.bounding_box.width, face.bounding_box.height) * 0.3f;
    
    // 根据贴纸类型调整基础大小
    switch (sticker.type) {
        case StickerType::GLASSES:
            base_size *= 0.8f;
            break;
        case StickerType::HAT:
            base_size *= 1.2f;
            break;
        case StickerType::MUSTACHE:
            base_size *= 0.4f;
            break;
        case StickerType::EARS:
            base_size *= 0.6f;
            break;
        default:
            break;
    }

    base_size *= sticker.scale;

    // 保持贴纸的宽高比
    float aspect_ratio = static_cast<float>(sticker.image.cols) / static_cast<float>(sticker.image.rows);
    
    return cv::Size2f(base_size * aspect_ratio, base_size);
}

cv::Mat TextureManager::rotateImage(const cv::Mat& image, float angle_degrees) {
    cv::Point2f center(image.cols / 2.0f, image.rows / 2.0f);
    cv::Mat rotation_matrix = cv::getRotationMatrix2D(center, angle_degrees, 1.0);
    
    cv::Mat rotated;
    cv::warpAffine(image, rotated, rotation_matrix, image.size());
    
    return rotated;
}

void TextureManager::blendSticker(cv::Mat& background, const cv::Mat& sticker, float opacity) {
    if (sticker.channels() == 4) {
        // 带Alpha通道的贴纸
        std::vector<cv::Mat> sticker_channels;
        cv::split(sticker, sticker_channels);
        
        cv::Mat alpha = sticker_channels[3] / 255.0f * opacity;
        cv::Mat sticker_rgb;
        cv::merge(std::vector<cv::Mat>{sticker_channels[0], sticker_channels[1], sticker_channels[2]}, sticker_rgb);
        
        for (int c = 0; c < 3; c++) {
            background.col(c) = background.col(c).mul(1.0f - alpha) + sticker_rgb.col(c).mul(alpha);
        }
    } else {
        // 不带Alpha通道的贴纸
        cv::addWeighted(background, 1.0f - opacity, sticker, opacity, 0, background);
    }
}

AnchorPoint TextureManager::getDefaultAnchorPoint(StickerType type) {
    switch (type) {
        case StickerType::GLASSES:
            return AnchorPoint::NOSE;
        case StickerType::HAT:
            return AnchorPoint::FOREHEAD;
        case StickerType::MUSTACHE:
            return AnchorPoint::MOUTH;
        case StickerType::EARS:
            return AnchorPoint::FACE_CENTER;
        case StickerType::CROWN:
            return AnchorPoint::FOREHEAD;
        case StickerType::MASK:
            return AnchorPoint::FACE_CENTER;
        default:
            return AnchorPoint::FACE_CENTER;
    }
}

void TextureManager::loadDefaultStickers() {
    // 尝试加载默认贴纸
    std::vector<std::pair<std::string, StickerType>> default_stickers = {
        {"../assets/stickers/glasses.png", StickerType::GLASSES},
        {"../assets/stickers/hat.png", StickerType::HAT},
        {"../assets/stickers/mustache.png", StickerType::MUSTACHE},
        {"../assets/stickers/ears.png", StickerType::EARS},
        {"../assets/stickers/crown.png", StickerType::CROWN},
        {"../assets/stickers/mask.png", StickerType::MASK}
    };

    for (const auto& sticker_info : default_stickers) {
        if (std::filesystem::exists(sticker_info.first)) {
            loadSticker(sticker_info.first, sticker_info.second);
        }
    }
}

std::vector<std::string> TextureManager::getAvailableStickers() const {
    std::vector<std::string> sticker_names;
    for (const auto& sticker_pair : stickers_) {
        sticker_names.push_back(getStickerTypeName(sticker_pair.first));
    }
    return sticker_names;
}

std::string TextureManager::getStickerTypeName(StickerType type) const {
    switch (type) {
        case StickerType::GLASSES: return "Glasses";
        case StickerType::HAT: return "Hat";
        case StickerType::MUSTACHE: return "Mustache";
        case StickerType::EARS: return "Ears";
        case StickerType::CROWN: return "Crown";
        case StickerType::MASK: return "Mask";
        default: return "Unknown";
    }
}
