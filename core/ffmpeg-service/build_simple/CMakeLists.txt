cmake_minimum_required(VERSION 3.16)
project(ffmpeg_simple_example VERSION 1.0.0 LANGUAGES CXX)

# 设置C++标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 设置编译选项
if(MSVC)
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /W4")
    set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} /MTd")
    set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} /MT")
else()
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra -O3")
    set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -g -O0")
endif()

# 查找FFmpeg
find_package(FFMPEG REQUIRED)

# 包含目录
include_directories(${FFMPEG_INCLUDE_DIRS})

# 创建简单示例程序
add_executable(ffmpeg_simple_example ../examples/simple_example.cpp)

# 链接FFmpeg库
target_link_libraries(ffmpeg_simple_example ${FFMPEG_LIBRARIES})

# 设置库目录
target_link_directories(ffmpeg_simple_example PRIVATE ${FFMPEG_LIBRARY_DIRS})

# 设置输出目录
set_target_properties(ffmpeg_simple_example PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)

# 打印配置信息
message(STATUS "FFmpeg Simple Example Configuration:")
message(STATUS "  Version: ${PROJECT_VERSION}")
message(STATUS "  C++ Standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "  FFmpeg Libraries: ${FFMPEG_LIBRARIES}")
message(STATUS "  Build Type: ${CMAKE_BUILD_TYPE}") 