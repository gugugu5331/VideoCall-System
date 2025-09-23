cmake_minimum_required(VERSION 3.20)
project(VideoCallSystemClient VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 设置Qt6路径（如果需要）
if(WIN32)
    set(CMAKE_PREFIX_PATH "C:/Qt/6.5.0/msvc2019_64" ${CMAKE_PREFIX_PATH})
elseif(APPLE)
    set(CMAKE_PREFIX_PATH "/usr/local/Qt-6.5.0" ${CMAKE_PREFIX_PATH})
endif()

# 查找Qt6组件
find_package(Qt6 REQUIRED COMPONENTS
    Core
    Widgets
    Gui
    Network
    Multimedia
    MultimediaWidgets
    OpenGL
    OpenGLWidgets
    WebSockets
    WebEngine
    WebEngineWidgets
    Concurrent
    Charts
    Sql
)

# 查找其他依赖
find_package(OpenCV REQUIRED)
find_package(OpenGL REQUIRED)
find_package(PkgConfig REQUIRED)

# 查找ZeroMQ (用于Edge-Model-Infra通信)
find_path(ZMQ_INCLUDE_DIR zmq.h)
find_library(ZMQ_LIBRARY NAMES zmq libzmq)

# 查找Protocol Buffers
find_package(Protobuf REQUIRED)

# 查找FFmpeg
pkg_check_modules(FFMPEG REQUIRED
    libavcodec
    libavformat
    libavutil
    libswscale
    libswresample
)

# WebRTC (如果可用)
find_path(WEBRTC_INCLUDE_DIR
    NAMES api/peer_connection_interface.h
    PATHS
    /usr/local/include/webrtc
    /opt/webrtc/include
    ${CMAKE_SOURCE_DIR}/third_party/webrtc/include
)

# 设置包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${OpenCV_INCLUDE_DIRS}
    ${OPENGL_INCLUDE_DIRS}
    ${ZMQ_INCLUDE_DIR}
    ${FFMPEG_INCLUDE_DIRS}
)

if(WEBRTC_INCLUDE_DIR)
    include_directories(${WEBRTC_INCLUDE_DIR})
    add_definitions(-DHAS_WEBRTC)
endif()

if(ZMQ_LIBRARY)
    add_definitions(-DHAS_ZMQ)
endif()

# 源文件
set(SOURCES
    src/main.cpp
    src/application.cpp
    src/mainwindow.cpp
    src/videowidget.cpp
    src/cameramanager.cpp
    src/filtercontroller.cpp
    src/effectsmanager.cpp
    src/facedetectionmanager.cpp
    src/webrtcmanager.cpp
    src/signalingclient.cpp
    src/participantmanager.cpp
    src/chatmanager.cpp
    src/settingsmanager.cpp
    src/recordingmanager.cpp
    src/screensharemanager.cpp
    src/audiomanager.cpp
    src/networkmanager.cpp
    src/performancemonitor.cpp
    src/pluginmanager.cpp
    src/thememanager.cpp
    
    # UI组件
    src/ui/controlpanel.cpp
    src/ui/filterpanel.cpp
    src/ui/effectspanel.cpp
    src/ui/participantspanel.cpp
    src/ui/chatpanel.cpp
    src/ui/settingsdialog.cpp
    src/ui/aboutdialog.cpp
    src/ui/joindialog.cpp
    src/ui/recordingdialog.cpp
    src/ui/customwidgets.cpp
    
    # 视频处理
    src/video/videoprocessor.cpp
    src/video/openglrenderer.cpp
    src/video/filtermanager.cpp
    src/video/texturemanager.cpp
    src/video/shadermanager.cpp
    
    # 网络
    src/network/websocketclient.cpp
    src/network/httpclient.cpp
    src/network/p2pmanager.cpp
    
    # 工具类
    src/utils/logger.cpp
    src/utils/config.cpp
    src/utils/fileutils.cpp
    src/utils/imageutils.cpp
    src/utils/mathutils.cpp
)

# 头文件
set(HEADERS
    include/application.h
    include/mainwindow.h
    include/videowidget.h
    include/cameramanager.h
    include/filtercontroller.h
    include/effectsmanager.h
    include/facedetectionmanager.h
    include/webrtcmanager.h
    include/signalingclient.h
    include/participantmanager.h
    include/chatmanager.h
    include/settingsmanager.h
    include/recordingmanager.h
    include/screensharemanager.h
    include/audiomanager.h
    include/networkmanager.h
    include/performancemonitor.h
    include/pluginmanager.h
    include/thememanager.h
    include/common.h
    include/constants.h
    
    # UI组件头文件
    include/ui/controlpanel.h
    include/ui/filterpanel.h
    include/ui/effectspanel.h
    include/ui/participantspanel.h
    include/ui/chatpanel.h
    include/ui/settingsdialog.h
    include/ui/aboutdialog.h
    include/ui/joindialog.h
    include/ui/recordingdialog.h
    include/ui/customwidgets.h
    
    # 视频处理头文件
    include/video/videoprocessor.h
    include/video/openglrenderer.h
    include/video/filtermanager.h
    include/video/texturemanager.h
    include/video/shadermanager.h
    
    # 网络头文件
    include/network/websocketclient.h
    include/network/httpclient.h
    include/network/p2pmanager.h
    
    # 工具类头文件
    include/utils/logger.h
    include/utils/config.h
    include/utils/fileutils.h
    include/utils/imageutils.h
    include/utils/mathutils.h
)

# 资源文件
set(RESOURCES
    resources/resources.qrc
)

# UI文件
set(UI_FILES
    ui/mainwindow.ui
    ui/settingsdialog.ui
    ui/aboutdialog.ui
    ui/joindialog.ui
    ui/recordingdialog.ui
)

# Qt预处理
qt6_add_executable(${PROJECT_NAME} ${SOURCES} ${HEADERS})
qt6_add_resources(${PROJECT_NAME} "resources" FILES ${RESOURCES})

if(UI_FILES)
    qt6_add_resources(${PROJECT_NAME} "ui_files" FILES ${UI_FILES})
endif()

# 链接库
target_link_libraries(${PROJECT_NAME}
    Qt6::Core
    Qt6::Widgets
    Qt6::Gui
    Qt6::Network
    Qt6::Multimedia
    Qt6::MultimediaWidgets
    Qt6::OpenGL
    Qt6::OpenGLWidgets
    Qt6::WebSockets
    Qt6::Concurrent
    Qt6::Charts
    ${OpenCV_LIBS}
    ${OPENGL_LIBRARIES}
)

# WebRTC链接（如果可用）
if(WEBRTC_INCLUDE_DIR)
    find_library(WEBRTC_LIBRARY
        NAMES webrtc
        PATHS
        /usr/local/lib
        /opt/webrtc/lib
        ${CMAKE_SOURCE_DIR}/third_party/webrtc/lib
    )
    if(WEBRTC_LIBRARY)
        target_link_libraries(${PROJECT_NAME} ${WEBRTC_LIBRARY})
    endif()
endif()

# 编译选项
target_compile_options(${PROJECT_NAME} PRIVATE
    $<$<CXX_COMPILER_ID:GNU>:-Wall -Wextra -O3>
    $<$<CXX_COMPILER_ID:Clang>:-Wall -Wextra -O3>
    $<$<CXX_COMPILER_ID:MSVC>:/W4 /O2>
)

# 预处理器定义
target_compile_definitions(${PROJECT_NAME} PRIVATE
    QT_NO_DEBUG_OUTPUT
    QT_NO_WARNING_OUTPUT
    $<$<CONFIG:Debug>:DEBUG_BUILD>
    $<$<CONFIG:Release>:RELEASE_BUILD>
)

# 设置输出目录
set_target_properties(${PROJECT_NAME} PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
    LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib
    ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib
)

# 复制资源文件
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/shaders DESTINATION ${CMAKE_BINARY_DIR}/bin)
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/assets DESTINATION ${CMAKE_BINARY_DIR}/bin)
file(COPY ${CMAKE_CURRENT_SOURCE_DIR}/config DESTINATION ${CMAKE_BINARY_DIR}/bin)

# 安装规则
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

install(DIRECTORY shaders assets config
    DESTINATION bin
)

# 平台特定设置
if(WIN32)
    set_target_properties(${PROJECT_NAME} PROPERTIES
        WIN32_EXECUTABLE TRUE
    )
    # 复制Qt DLL（发布时）
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        find_program(WINDEPLOYQT_EXECUTABLE windeployqt HINTS ${Qt6_DIR}/../../../bin)
        if(WINDEPLOYQT_EXECUTABLE)
            add_custom_command(TARGET ${PROJECT_NAME} POST_BUILD
                COMMAND ${WINDEPLOYQT_EXECUTABLE} $<TARGET_FILE:${PROJECT_NAME}>
                COMMENT "Deploying Qt libraries"
            )
        endif()
    endif()
elseif(APPLE)
    set_target_properties(${PROJECT_NAME} PROPERTIES
        MACOSX_BUNDLE TRUE
        MACOSX_BUNDLE_INFO_PLIST ${CMAKE_CURRENT_SOURCE_DIR}/Info.plist
    )
endif()

# 调试信息
message(STATUS "Qt6 version: ${Qt6_VERSION}")
message(STATUS "OpenCV version: ${OpenCV_VERSION}")
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
if(WEBRTC_INCLUDE_DIR)
    message(STATUS "WebRTC found: ${WEBRTC_INCLUDE_DIR}")
else()
    message(STATUS "WebRTC not found - some features will be disabled")
endif()
