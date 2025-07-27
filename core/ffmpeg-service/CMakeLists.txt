cmake_minimum_required(VERSION 3.16)
project(ffmpeg_detection_service)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找必要的包
find_package(PkgConfig REQUIRED)
find_package(Threads REQUIRED)

# FFmpeg
pkg_check_modules(FFMPEG REQUIRED
    libavcodec
    libavformat
    libavutil
    libswscale
    libswresample
)

# ONNX Runtime
find_path(ONNXRUNTIME_INCLUDE_DIRS
    NAMES onnxruntime_cxx_api.h
    PATHS /usr/include /usr/local/include /opt/onnxruntime/include
    PATH_SUFFIXES onnxruntime
)

find_library(ONNXRUNTIME_LIBRARIES
    NAMES onnxruntime
    PATHS /usr/lib /usr/local/lib /opt/onnxruntime/lib
)

# OpenCV (可选，用于图像处理)
find_package(OpenCV QUIET)

# 设置源文件
set(SOURCES
    src/main.cpp
    src/ffmpeg_processor.cpp
    src/detection_engine.cpp
    src/video_compressor.cpp
    src/audio_processor.cpp
    src/onnx_inference.cpp
    src/utils.cpp
)

# 创建可执行文件
add_executable(${PROJECT_NAME} ${SOURCES})

# 包含目录
target_include_directories(${PROJECT_NAME} PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${FFMPEG_INCLUDE_DIRS}
    ${ONNXRUNTIME_INCLUDE_DIRS}
)

if(OpenCV_FOUND)
    target_include_directories(${PROJECT_NAME} PRIVATE ${OpenCV_INCLUDE_DIRS})
endif()

# 链接库
target_link_libraries(${PROJECT_NAME}
    ${FFMPEG_LIBRARIES}
    ${ONNXRUNTIME_LIBRARIES}
    Threads::Threads
)

if(OpenCV_FOUND)
    target_link_libraries(${PROJECT_NAME} ${OpenCV_LIBS})
endif()

# 编译选项
target_compile_options(${PROJECT_NAME} PRIVATE
    ${FFMPEG_CFLAGS_OTHER}
    -O3
    -march=native
    -ffast-math
)

# 安装
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
)

# 复制模型文件
install(DIRECTORY models/
    DESTINATION share/${PROJECT_NAME}/
    FILES_MATCHING PATTERN "*.onnx"
) 