# AI Detection Node Dockerfile
FROM ubuntu:20.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config \
    libzmq3-dev \
    libgoogle-glog-dev \
    libboost-all-dev \
    libssl-dev \
    libbsd-dev \
    libopencv-dev \
    libsndfile1-dev \
    libfftw3-dev \
    nlohmann-json3-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Create build directory and build
RUN mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc)

# Create directories for models and uploads
RUN mkdir -p /app/models /app/uploads /tmp/detection_uploads

# Copy configuration
COPY config/detection_config.json /app/config/

# Expose any necessary ports (if needed)
# EXPOSE 5000

# Set environment variables
ENV MODEL_PATH=/app/models
ENV UPLOAD_PATH=/tmp/detection_uploads

# Run the AI detection node
CMD ["./build/ai-detection"]
