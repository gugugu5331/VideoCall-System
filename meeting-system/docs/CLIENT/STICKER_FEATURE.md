# 贴图功能使用文档

## 功能概述

贴图功能允许用户在视频画面上添加各种装饰性图片，包括表情包、装饰物等。贴图可以固定在画面上，也可以跟随人脸移动。

## 核心特性

### 1. 多种锚点模式

- **固定位置**：贴图固定在画面的某个位置
- **人脸中心**：贴图跟随人脸中心移动
- **左眼**：贴图跟随左眼位置
- **右眼**：贴图跟随右眼位置
- **鼻子**：贴图跟随鼻子位置
- **嘴巴**：贴图跟随嘴巴位置

### 2. 贴图属性

- **缩放**：0.1x - 5.0x
- **旋转**：0° - 360°
- **不透明度**：0% - 100%
- **位置偏移**：相对于锚点的偏移量

### 3. 预设贴图

系统内置8种预设贴图：

**表情包**：
- 😀 笑脸
- 😎 墨镜
- 😍 爱心眼
- 🤔 思考

**装饰物**：
- 👑 皇冠
- 🎩 帽子
- 🎀 蝴蝶结
- 🌟 星星

## 使用方法

### C++ API

#### 1. 启用贴图功能

```cpp
// 通过VideoEffectsController
videoEffectsController->setStickerEnabled(true);

// 或直接通过VideoEffectProcessor
videoEffectProcessor->setStickerEnabled(true);
```

#### 2. 添加预设贴图

```cpp
// 添加笑脸贴图到人脸中心
QString stickerId = videoEffectsController->addPresetSticker(
    "😀 笑脸",
    1  // 1 = 人脸中心
);
```

#### 3. 添加自定义贴图

```cpp
// 从文件添加贴图
QString stickerId = videoEffectsController->addSticker(
    "path/to/sticker.png",
    2  // 2 = 左眼
);
```

#### 4. 调整贴图属性

```cpp
// 设置缩放
videoEffectsController->setStickerScale(stickerId, 1.5f);

// 设置不透明度
videoEffectsController->setStickerOpacity(stickerId, 0.8f);
```

#### 5. 移除贴图

```cpp
// 移除单个贴图
videoEffectsController->removeSticker(stickerId);

// 清除所有贴图
videoEffectsController->clearStickers();
```

### QML API

#### 1. 使用StickerPanel组件

```qml
import QtQuick 2.15
import QtQuick.Controls 2.15

StickerPanel {
    id: stickerPanel
    controller: videoEffectsController
    anchors.fill: parent
}
```

#### 2. 直接调用API

```qml
Button {
    text: "添加笑脸"
    onClicked: {
        var stickerId = videoEffectsController.addPresetSticker("😀 笑脸", 1)
        console.log("Added sticker:", stickerId)
    }
}

Button {
    text: "清除贴图"
    onClicked: {
        videoEffectsController.clearStickers()
    }
}
```

## 技术实现

### 架构设计

```
VideoEffectProcessor
    ↓
StickerOverlay
    ↓
Sticker (多个)
```

### 数据流

```
1. 视频帧输入
    ↓
2. 人脸检测（VideoEffectProcessor）
    ↓
3. 计算贴图位置（Sticker::calculateRenderRect）
    ↓
4. Alpha混合渲染（StickerOverlay::alphaBlend）
    ↓
5. 输出合成后的视频帧
```

### 核心算法

#### 1. 人脸跟踪

```cpp
QRect Sticker::calculateRenderRect(const cv::Rect &faceRect) const
{
    switch (m_anchorType) {
        case AnchorType::Face:
            // 人脸中心
            int centerX = faceRect.x + faceRect.width / 2;
            int centerY = faceRect.y + faceRect.height / 2;
            return QRect(centerX - width / 2, centerY - height / 2, width, height);
            
        case AnchorType::LeftEye:
            // 左眼位置（人脸左上1/3处）
            int eyeX = faceRect.x + faceRect.width * 0.3;
            int eyeY = faceRect.y + faceRect.height * 0.35;
            return QRect(eyeX - width / 2, eyeY - height / 2, width, height);
        
        // ... 其他锚点
    }
}
```

#### 2. Alpha混合

```cpp
void StickerOverlay::alphaBlend(cv::Mat &target, const cv::Mat &overlay, 
                                const cv::Mat &mask, const QRect &rect)
{
    // 转换为浮点数
    cv::Mat targetFloat, overlayFloat, maskFloat;
    targetROI.convertTo(targetFloat, CV_32F);
    overlayROI.convertTo(overlayFloat, CV_32F);
    maskROI.convertTo(maskFloat, CV_32F, 1.0 / 255.0);
    
    // 逐通道混合
    for (int i = 0; i < 3; i++) {
        cv::Mat blended = overlayChannel.mul(maskFloat) + 
                         targetChannel.mul(cv::Scalar(1.0) - maskFloat);
        resultChannels.push_back(blended);
    }
    
    // 合并通道
    cv::merge(resultChannels, result);
    result.convertTo(targetROI, CV_8U);
}
```

## 性能优化

### 1. 贴图缓存

- 贴图图像加载后缓存在内存中
- Alpha通道预先提取，避免重复计算

### 2. 边界检查

```cpp
// 完全在画面外的贴图不渲染
if (renderRect.x() >= target.cols || renderRect.y() >= target.rows) {
    return;
}
```

### 3. ROI优化

```cpp
// 只处理贴图覆盖的区域
cv::Mat targetROI = target(cv::Rect(x1, y1, x2 - x1, y2 - y1));
cv::Mat overlayROI = overlay(cv::Rect(ox1, oy1, ox2 - ox1, oy2 - oy1));
```

### 性能指标

| 贴图数量 | 处理时间 | 帧率影响 |
|---------|---------|---------|
| 1个 | ~2ms | 几乎无 |
| 3个 | ~5ms | <5fps |
| 5个 | ~8ms | ~10fps |

## 贴图图片要求

### 格式要求

- **支持格式**：PNG（推荐）、JPG、JPEG、BMP
- **推荐格式**：PNG（支持透明通道）
- **颜色模式**：RGB或RGBA

### 尺寸建议

- **最小尺寸**：64x64 像素
- **推荐尺寸**：128x128 - 512x512 像素
- **最大尺寸**：1024x1024 像素

### 透明度

- PNG格式的Alpha通道会被自动识别
- JPG格式会自动添加全不透明的Alpha通道
- 建议使用PNG格式以获得最佳效果

## 常见问题

### Q1: 贴图不显示？

**A:** 检查以下几点：
1. 是否启用了贴图功能：`setStickerEnabled(true)`
2. 贴图文件路径是否正确
3. 贴图图片格式是否支持
4. 如果使用人脸跟踪，是否检测到人脸

### Q2: 贴图位置不准确？

**A:** 人脸跟踪的精度取决于人脸检测算法。可以：
1. 确保光线充足
2. 人脸正对摄像头
3. 使用位置偏移微调：`setPosition(QPoint(offsetX, offsetY))`

### Q3: 贴图边缘有锯齿？

**A:** 这是Alpha混合的正常现象。可以：
1. 使用高分辨率的贴图图片
2. 确保贴图PNG有平滑的Alpha通道
3. 调整不透明度使边缘更柔和

### Q4: 性能影响大？

**A:** 优化建议：
1. 限制同时显示的贴图数量（建议≤3个）
2. 使用适当尺寸的贴图（不要过大）
3. 降低视频分辨率（如720p）

## 扩展开发

### 添加新的预设贴图

1. 准备贴图图片（PNG格式，带透明通道）
2. 添加到资源文件或指定目录
3. 在`initializePresets()`中注册：

```cpp
void StickerOverlay::initializePresets()
{
    m_presetStickers.insert(
        QString::fromUtf8("🎃 南瓜"),
        ":/stickers/pumpkin.png"
    );
}
```

### 自定义锚点类型

1. 在`Sticker::AnchorType`枚举中添加新类型
2. 在`calculateRenderRect()`中实现位置计算逻辑
3. 更新QML界面的锚点选择列表

### 添加动画效果

```cpp
// 在Sticker类中添加动画属性
class Sticker {
    // ...
    float m_animationPhase;  // 0.0 - 1.0
    
    void updateAnimation(float deltaTime) {
        m_animationPhase += deltaTime;
        if (m_animationPhase > 1.0f) {
            m_animationPhase -= 1.0f;
        }
        
        // 根据动画相位调整属性
        m_rotation = m_animationPhase * 360.0f;  // 旋转动画
        m_scale = 1.0f + 0.2f * sin(m_animationPhase * 2 * M_PI);  // 缩放动画
    }
};
```

## 总结

贴图功能为视频会议增添了趣味性和个性化。通过人脸跟踪技术，贴图可以智能地跟随用户的面部移动，提供了丰富的互动体验。

**核心优势**：
- ✅ 简单易用的API
- ✅ 多种锚点模式
- ✅ 实时人脸跟踪
- ✅ 高性能Alpha混合
- ✅ 灵活的属性调整

**适用场景**：
- 视频会议娱乐
- 直播互动
- 在线教育
- 虚拟活动

