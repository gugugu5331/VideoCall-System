# Video Processing Dockerfile
# 基于Ubuntu 22.04构建OpenCV + OpenGL视频处理应用

FROM ubuntu:22.04

# 设置环境变量
ENV DEBIAN_FRONTEND=noninteractive
ENV DISPLAY=:0
ENV QT_X11_NO_MITSHM=1

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    # 基础工具
    build-essential \
    cmake \
    git \
    wget \
    curl \
    pkg-config \
    # C++编译器和工具
    g++ \
    clang \
    make \
    ninja-build \
    # OpenCV依赖
    libopencv-dev \
    libopencv-contrib-dev \
    # OpenGL依赖
    libgl1-mesa-dev \
    libglu1-mesa-dev \
    libglfw3-dev \
    libglew-dev \
    libglm-dev \
    # X11和显示相关
    xorg-dev \
    libx11-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxcursor-dev \
    libxi-dev \
    # 音频支持
    libasound2-dev \
    libpulse-dev \
    # 视频编解码
    libavcodec-dev \
    libavformat-dev \
    libavutil-dev \
    libswscale-dev \
    # 图像格式支持
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    libwebp-dev \
    # 其他工具
    vim \
    htop \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制源代码
COPY . .

# 设置构建脚本权限
RUN chmod +x build.sh

# 构建应用
RUN ./build.sh --build

# 创建运行用户
RUN useradd -m -s /bin/bash videouser && \
    usermod -a -G audio,video videouser

# 设置权限
RUN chown -R videouser:videouser /app

# 切换到普通用户
USER videouser

# 暴露端口（如果需要Web界面）
EXPOSE 8080

# 设置入口点
ENTRYPOINT ["./build/VideoProcessing"]

# 默认参数
CMD ["--help"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep VideoProcessing || exit 1

# 标签信息
LABEL maintainer="VideoCall-System"
LABEL description="OpenCV + OpenGL Video Processing Application"
LABEL version="1.0.0"
