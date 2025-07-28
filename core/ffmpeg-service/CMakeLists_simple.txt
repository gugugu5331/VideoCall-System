cmake_minimum_required(VERSION 3.16)
project(ffmpeg_service VERSION 1.0.0 LANGUAGES CXX)

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

# 查找依赖包
find_package(FFMPEG REQUIRED)
find_package(OpenCV REQUIRED)
find_package(nlohmann_json REQUIRED)
find_package(spdlog REQUIRED)
find_package(fmt REQUIRED)

# 设置ONNX Runtime为可选
option(ENABLE_ONNX "Enable ONNX Runtime support" OFF)
if(ENABLE_ONNX)
    find_package(ONNXRuntime REQUIRED)
endif()

# 包含目录
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
include_directories(${FFMPEG_INCLUDE_DIRS})
include_directories(${OpenCV_INCLUDE_DIRS})
if(ENABLE_ONNX AND ONNXRuntime_FOUND)
    include_directories(${ONNXRuntime_INCLUDE_DIRS})
endif()

# 源文件
set(SOURCES
    src/ffmpeg_processor.cpp
    src/onnx_detector.cpp
    src/integration_service.cpp
)

# 头文件
set(HEADERS
    include/ffmpeg_processor.h
    include/onnx_detector.h
    include/integration_service.h
)

# 创建静态库
add_library(ffmpeg_service STATIC ${SOURCES} ${HEADERS})

# 设置库属性
set_target_properties(ffmpeg_service PROPERTIES
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
    PUBLIC_HEADER "${HEADERS}"
)

# 链接库
target_link_libraries(ffmpeg_service
    ${FFMPEG_LIBRARIES}
    ${OpenCV_LIBS}
    nlohmann_json::nlohmann_json
    spdlog::spdlog
    fmt::fmt
)

# 如果启用了ONNX Runtime，则链接
if(ENABLE_ONNX AND ONNXRuntime_FOUND)
    target_link_libraries(ffmpeg_service ${ONNXRuntime_LIBRARIES})
endif()

# 设置库目录
target_link_directories(ffmpeg_service PRIVATE
    ${FFMPEG_LIBRARY_DIRS}
)
if(ENABLE_ONNX AND ONNXRuntime_FOUND)
    target_link_directories(ffmpeg_service PRIVATE
        ${ONNXRuntime_LIBRARY_DIRS}
    )
endif()

# 编译定义
target_compile_definitions(ffmpeg_service PRIVATE
    $<$<CONFIG:Debug>:DEBUG>
    $<$<CONFIG:Release>:NDEBUG>
)

# 创建示例程序
add_executable(ffmpeg_service_example examples/main.cpp)
target_link_libraries(ffmpeg_service_example ffmpeg_service)

# 设置输出目录
set_target_properties(ffmpeg_service PROPERTIES
    ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib
    LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)

set_target_properties(ffmpeg_service_example PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)

# 打印配置信息
message(STATUS "FFmpeg Service Configuration:")
message(STATUS "  Version: ${PROJECT_VERSION}")
message(STATUS "  C++ Standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "  FFmpeg Libraries: ${FFMPEG_LIBRARIES}")
message(STATUS "  OpenCV Version: ${OpenCV_VERSION}")
if(ENABLE_ONNX AND ONNXRuntime_FOUND)
    message(STATUS "  ONNX Runtime Version: ${ONNXRuntime_VERSION}")
else()
    message(STATUS "  ONNX Runtime: Disabled")
endif()
message(STATUS "  Build Type: ${CMAKE_BUILD_TYPE}") 