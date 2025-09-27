# 视频处理系统测试指南

## 概述

本视频处理系统实现了基于OpenCV和OpenGL的实时滤镜、贴图和人脸检测功能，专为视频会议应用设计。

## 功能特性

### ✅ 已实现功能

#### 🎨 **实时滤镜系统**
- **基础滤镜**: 模糊、锐化、边缘检测
- **艺术滤镜**: 复古、卡通、素描、霓虹
- **美颜滤镜**: 磨皮、美白、瘦脸
- **色彩滤镜**: 灰度、棕褐色、热成像

#### 🎭 **贴图和贴纸系统**
- **面部贴纸**: 眼镜、帽子、胡子、耳朵
- **动态贴纸**: 支持动画和旋转效果
- **智能定位**: 基于面部关键点的精确贴纸定位
- **透明度控制**: 可调节贴纸透明度和混合模式

#### 👤 **人脸检测和跟踪**
- **实时检测**: 基于Haar级联分类器和DNN模型
- **关键点检测**: 68点面部关键点定位
- **多人脸支持**: 同时检测和处理多个人脸
- **跟踪优化**: 智能跟踪减少检测频率

#### ⚡ **性能优化**
- **GPU加速**: OpenGL硬件加速渲染
- **多线程处理**: 异步处理提高性能
- **自适应质量**: 根据性能自动调整处理质量
- **内存优化**: 高效的内存管理和缓存

## 编译和测试

### 方法1: 完整系统编译

#### Linux/macOS
```bash
cd src/video-processing

# 安装依赖
./build.sh --deps

# 编译项目
./build.sh --build

# 运行测试
./build/bin/VideoProcessingTest
```

#### Windows (MinGW/MSYS2)
```bash
cd src/video-processing

# 创建构建目录
mkdir build && cd build

# 配置CMake
cmake -G "MinGW Makefiles" ..

# 编译
mingw32-make

# 运行测试
./bin/VideoProcessingTest.exe
```

### 方法2: 简化版本测试

如果完整系统编译遇到问题，可以使用简化版本进行功能测试：

#### Linux/macOS
```bash
cd src/video-processing

# 编译简化版本
./compile_simple.sh

# 运行测试
./simple_video_test
```

#### Windows
```cmd
cd src\video-processing

# 编译简化版本
compile_simple.bat

# 运行测试
simple_video_test.exe
```

## 功能测试

### 🎮 **控制说明**

#### 基础控制
- `ESC` - 退出程序
- `SPACE` - 截图保存
- `R` - 开始/停止录制

#### 滤镜控制
- `1` - 模糊滤镜
- `2` - 锐化滤镜
- `3` - 边缘检测
- `4` - 复古滤镜
- `5` - 灰度滤镜
- `6` - 美颜滤镜
- `7` - 卡通滤镜
- `0` - 移除所有滤镜

#### 贴纸控制
- `G` - 眼镜贴纸
- `H` - 帽子贴纸
- `M` - 胡子贴纸
- `C` - 皇冠贴纸

#### 高级控制
- `F` - 切换人脸检测显示
- `B` - 美颜模式开关
- `+/-` - 调整滤镜强度
- `D` - 切换检测模式

### 📊 **性能测试**

#### 测试指标
- **帧率**: 目标60 FPS @ 1080p
- **延迟**: <50ms端到端处理延迟
- **CPU使用率**: <30%（启用GPU加速）
- **内存使用**: <500MB

#### 测试场景
1. **基础功能测试**
   - 摄像头捕获和显示
   - 基础滤镜应用
   - 截图功能

2. **人脸检测测试**
   - 单人脸检测准确性
   - 多人脸同时检测
   - 检测稳定性和跟踪

3. **滤镜性能测试**
   - 各种滤镜的处理速度
   - 滤镜切换的流畅性
   - 滤镜强度调节

4. **贴纸功能测试**
   - 贴纸定位准确性
   - 动态贴纸效果
   - 多贴纸同时应用

## 故障排除

### 常见问题

#### 1. 编译错误
```
错误: 找不到OpenCV库
```
**解决方案**:
- Ubuntu/Debian: `sudo apt-get install libopencv-dev`
- CentOS/RHEL: `sudo yum install opencv-devel`
- macOS: `brew install opencv`
- Windows: 下载并安装OpenCV，配置环境变量

#### 2. 摄像头问题
```
错误: 无法打开摄像头
```
**解决方案**:
- 检查摄像头是否被其他程序占用
- 尝试不同的摄像头ID (0, 1, 2...)
- 检查摄像头权限设置

#### 3. OpenGL错误
```
错误: OpenGL初始化失败
```
**解决方案**:
- 更新显卡驱动
- 检查OpenGL支持版本
- 尝试软件渲染模式

#### 4. 人脸检测不工作
```
警告: 无法加载人脸检测模型
```
**解决方案**:
- 下载Haar级联分类器文件
- 检查模型文件路径
- 使用DNN模型替代

### 性能优化建议

#### 1. 硬件要求
- **最低配置**: Intel i5 / AMD Ryzen 5, 8GB RAM, 集成显卡
- **推荐配置**: Intel i7 / AMD Ryzen 7, 16GB RAM, 独立显卡
- **最佳配置**: Intel i9 / AMD Ryzen 9, 32GB RAM, RTX 3060+

#### 2. 软件优化
- 启用GPU加速
- 调整处理分辨率
- 优化检测频率
- 使用多线程处理

## 集成到视频会议系统

### API接口

```cpp
// 初始化视频处理器
VideoProcessor processor;
processor.initialize();

// 设置滤镜
processor.setFilterType(FilterType::BEAUTY);
processor.setFilterIntensity(0.8f);

// 添加贴纸
processor.addSticker("glasses.png", StickerType::GLASSES);

// 处理视频帧
cv::Mat input_frame, output_frame;
processor.processFrame(input_frame, output_frame);
```

### WebRTC集成

```cpp
// 在WebRTC视频帧回调中使用
void OnFrame(webrtc::VideoFrame& frame) {
    cv::Mat cv_frame = ConvertToMat(frame);
    cv::Mat processed_frame;
    
    video_processor_->processFrame(cv_frame, processed_frame);
    
    webrtc::VideoFrame processed_webrtc_frame = ConvertToWebRTC(processed_frame);
    // 发送处理后的帧
}
```

## 下一步开发

### 🔄 **待实现功能**
- [ ] 更多艺术滤镜效果
- [ ] 3D面部建模和贴图
- [ ] 背景替换和虚拟背景
- [ ] 手势识别和控制
- [ ] 语音驱动的面部动画
- [ ] 实时美颜参数调节
- [ ] 自定义滤镜编辑器

### 🚀 **性能优化**
- [ ] CUDA加速支持
- [ ] 移动端优化
- [ ] 网络传输优化
- [ ] 电池续航优化

### 🔧 **工程改进**
- [ ] 单元测试覆盖
- [ ] 性能基准测试
- [ ] 内存泄漏检测
- [ ] 跨平台兼容性测试

---

## 总结

本视频处理系统成功实现了：
- ✅ 实时视频滤镜处理
- ✅ 人脸检测和贴纸应用
- ✅ 高性能GPU加速渲染
- ✅ 模块化架构设计
- ✅ 跨平台兼容性

系统已准备好集成到视频会议应用中，为用户提供丰富的视频特效功能！
