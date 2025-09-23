cmake_minimum_required(VERSION 3.16)
project(VideoProcessing)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找依赖包
find_package(OpenCV REQUIRED)
find_package(OpenGL REQUIRED)
find_package(glfw3 REQUIRED)
find_package(GLEW REQUIRED)
find_package(PkgConfig REQUIRED)

# 查找 GLM (数学库)
find_path(GLM_INCLUDE_DIR glm/glm.hpp
    HINTS
    /usr/include
    /usr/local/include
    ${CMAKE_SOURCE_DIR}/third_party/glm
)

# 设置包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${OpenCV_INCLUDE_DIRS}
    ${OPENGL_INCLUDE_DIRS}
    ${GLEW_INCLUDE_DIRS}
    ${GLM_INCLUDE_DIR}
)

# 源文件
set(SOURCES
    src/main.cpp
    src/video_processor.cpp
    src/opengl_renderer.cpp
    src/filter_manager.cpp
    src/texture_manager.cpp
    src/face_detector.cpp
    src/shader_manager.cpp
    src/camera_capture.cpp
    src/effect_processor.cpp
)

# 头文件
set(HEADERS
    include/video_processor.h
    include/opengl_renderer.h
    include/filter_manager.h
    include/texture_manager.h
    include/face_detector.h
    include/shader_manager.h
    include/camera_capture.h
    include/effect_processor.h
    include/common.h
)

# 创建可执行文件
add_executable(${PROJECT_NAME} ${SOURCES} ${HEADERS})

# 链接库
target_link_libraries(${PROJECT_NAME}
    ${OpenCV_LIBS}
    ${OPENGL_LIBRARIES}
    glfw
    GLEW::GLEW
    GL
    pthread
)

# 编译选项
target_compile_options(${PROJECT_NAME} PRIVATE
    -Wall
    -Wextra
    -O3
    -march=native
)

# 设置输出目录
set_target_properties(${PROJECT_NAME} PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)

# 复制资源文件
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/shaders DESTINATION ${CMAKE_BINARY_DIR}/bin)
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/assets DESTINATION ${CMAKE_BINARY_DIR}/bin)
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/textures DESTINATION ${CMAKE_BINARY_DIR}/bin)

# 安装规则
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
)

install(DIRECTORY shaders assets textures
    DESTINATION bin
)

# 调试信息
message(STATUS "OpenCV version: ${OpenCV_VERSION}")
message(STATUS "OpenCV include dirs: ${OpenCV_INCLUDE_DIRS}")
message(STATUS "OpenCV libraries: ${OpenCV_LIBS}")
message(STATUS "OpenGL libraries: ${OPENGL_LIBRARIES}")
