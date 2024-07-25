cmake_minimum_required(VERSION 3.10)
project(meeting-ai-node)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 设置构建类型
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# 编译选项
set(CMAKE_CXX_FLAGS_DEBUG "-g -O0 -Wall -Wextra")
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -DNDEBUG")

# 包含目录
include_directories(include)
include_directories(../Edge-LLM-Infra/hybrid-comm/include)
include_directories(../Edge-LLM-Infra/hybrid-comm/include/libzmq)
include_directories(../Edge-LLM-Infra/infra-controller/include)
include_directories(../Edge-LLM-Infra/utils)

# 查找依赖包
find_package(PkgConfig REQUIRED)
find_package(Threads REQUIRED)

# 查找glog
find_package(glog REQUIRED)

# 查找eventpp
find_package(eventpp REQUIRED)

# 查找nlohmann_json
find_package(nlohmann_json REQUIRED)

# 查找OpenCV (用于图像处理)
find_package(OpenCV REQUIRED)

# 查找FFmpeg (用于音视频处理)
pkg_check_modules(FFMPEG REQUIRED 
    libavcodec 
    libavformat 
    libavutil 
    libswscale 
    libswresample
)

# 源文件
file(GLOB_RECURSE SOURCES 
    "src/*.cpp"
    "../Edge-LLM-Infra/hybrid-comm/src/*.cpp"
    "../Edge-LLM-Infra/infra-controller/src/*.cpp"
)

# 排除不需要的文件
list(FILTER SOURCES EXCLUDE REGEX ".*main\\.cpp$")

# 创建库
add_library(meeting-ai-lib STATIC ${SOURCES})

# 链接库
target_link_libraries(meeting-ai-lib
    glog::glog
    eventpp::eventpp
    nlohmann_json::nlohmann_json
    ${OpenCV_LIBS}
    ${FFMPEG_LIBRARIES}
    zmq
    pthread
    utilities
)

# 包含目录
target_include_directories(meeting-ai-lib PUBLIC
    include
    ../Edge-LLM-Infra/hybrid-comm/include
    ../Edge-LLM-Infra/infra-controller/include
    ../Edge-LLM-Infra/utils
    ${OpenCV_INCLUDE_DIRS}
    ${FFMPEG_INCLUDE_DIRS}
)

# 创建可执行文件
add_executable(meeting-ai-node src/main.cpp)

# 链接库到可执行文件
target_link_libraries(meeting-ai-node
    meeting-ai-lib
    glog::glog
    eventpp::eventpp
    nlohmann_json::nlohmann_json
    ${OpenCV_LIBS}
    ${FFMPEG_LIBRARIES}
    zmq
    pthread
    utilities
)

# 安装目标
install(TARGETS meeting-ai-node
    RUNTIME DESTINATION bin
)

install(TARGETS meeting-ai-lib
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
)

install(DIRECTORY include/
    DESTINATION include/meeting-ai
    FILES_MATCHING PATTERN "*.h" PATTERN "*.hpp"
)

# 创建配置文件目录
install(DIRECTORY DESTINATION etc/meeting-ai)

# 复制配置文件
install(FILES config/ai_node_config.json
    DESTINATION etc/meeting-ai
)

# 创建模型目录
install(DIRECTORY DESTINATION share/meeting-ai/models)

# 创建日志目录
install(DIRECTORY DESTINATION var/log/meeting-ai)

# 编译选项
target_compile_definitions(meeting-ai-node PRIVATE
    $<$<CONFIG:Debug>:DEBUG>
    $<$<CONFIG:Release>:NDEBUG>
)

# 设置RPATH
set_target_properties(meeting-ai-node PROPERTIES
    INSTALL_RPATH_USE_LINK_PATH TRUE
)

# 打印配置信息
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "C++ standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "OpenCV version: ${OpenCV_VERSION}")
message(STATUS "Install prefix: ${CMAKE_INSTALL_PREFIX}")

# 添加自定义目标用于清理
add_custom_target(clean-all
    COMMAND ${CMAKE_BUILD_TOOL} clean
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/CMakeFiles
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/CMakeCache.txt
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/cmake_install.cmake
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/Makefile
)

# 添加测试目标
enable_testing()

# 如果有测试文件，添加测试
if(EXISTS ${CMAKE_CURRENT_SOURCE_DIR}/tests)
    add_subdirectory(tests)
endif()

# 生成编译数据库（用于IDE支持）
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
