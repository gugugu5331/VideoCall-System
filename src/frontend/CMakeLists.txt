cmake_minimum_required(VERSION 3.16)

project(VideoConferenceClient VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找Qt6
find_package(Qt6 REQUIRED COMPONENTS Core Widgets Quick QuickControls2 Multimedia MultimediaWidgets Network WebSockets)

# 查找FFmpeg
find_package(PkgConfig REQUIRED)
pkg_check_modules(FFMPEG REQUIRED libavcodec libavformat libavutil libswscale)

# 设置Qt自动处理
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# 源文件
set(SOURCES
    src/main.cpp
    src/application.cpp
    src/models/user_model.cpp
    src/models/meeting_model.cpp
    src/models/detection_model.cpp
    src/services/api_service.cpp
    src/services/auth_service.cpp
    src/services/meeting_service.cpp
    src/services/detection_service.cpp
    src/services/webrtc_service.cpp
    src/services/media_service.cpp
    src/controllers/main_controller.cpp
    src/controllers/login_controller.cpp
    src/controllers/meeting_controller.cpp
    src/controllers/detection_controller.cpp
    src/ui/main_window.cpp
    src/ui/login_dialog.cpp
    src/ui/meeting_window.cpp
    src/ui/detection_panel.cpp
    src/ui/video_widget.cpp
    src/utils/network_manager.cpp
    src/utils/settings_manager.cpp
    src/utils/logger.cpp
)

# 头文件
set(HEADERS
    src/application.h
    src/models/user_model.h
    src/models/meeting_model.h
    src/models/detection_model.h
    src/services/api_service.h
    src/services/auth_service.h
    src/services/meeting_service.h
    src/services/detection_service.h
    src/services/webrtc_service.h
    src/services/media_service.h
    src/controllers/main_controller.h
    src/controllers/login_controller.h
    src/controllers/meeting_controller.h
    src/controllers/detection_controller.h
    src/ui/main_window.h
    src/ui/login_dialog.h
    src/ui/meeting_window.h
    src/ui/detection_panel.h
    src/ui/video_widget.h
    src/utils/network_manager.h
    src/utils/settings_manager.h
    src/utils/logger.h
)

# UI文件
set(UI_FILES
    ui/main_window.ui
    ui/login_dialog.ui
    ui/meeting_window.ui
    ui/detection_panel.ui
)

# QML文件
set(QML_FILES
    qml/main.qml
    qml/LoginView.qml
    qml/MeetingView.qml
    qml/DetectionView.qml
    qml/VideoView.qml
    qml/components/VideoTile.qml
    qml/components/DetectionAlert.qml
    qml/components/ParticipantList.qml
)

# 资源文件
set(RESOURCES
    resources/resources.qrc
)

# 创建可执行文件
qt6_add_executable(VideoConferenceClient
    ${SOURCES}
    ${HEADERS}
    ${UI_FILES}
    ${RESOURCES}
)

# 添加QML模块
qt6_add_qml_module(VideoConferenceClient
    URI VideoConference
    VERSION 1.0
    QML_FILES ${QML_FILES}
)

# 链接库
target_link_libraries(VideoConferenceClient
    Qt6::Core
    Qt6::Widgets
    Qt6::Quick
    Qt6::QuickControls2
    Qt6::Multimedia
    Qt6::MultimediaWidgets
    Qt6::Network
    Qt6::WebSockets
    ${FFMPEG_LIBRARIES}
)

# 包含目录
target_include_directories(VideoConferenceClient PRIVATE
    src
    ${FFMPEG_INCLUDE_DIRS}
)

# 编译定义
target_compile_definitions(VideoConferenceClient PRIVATE
    QT_QML_DEBUG
    ${FFMPEG_CFLAGS_OTHER}
)

# Windows特定设置
if(WIN32)
    set_target_properties(VideoConferenceClient PROPERTIES
        WIN32_EXECUTABLE TRUE
    )
    
    # 添加WebRTC库（需要预编译的WebRTC库）
    find_library(WEBRTC_LIB webrtc PATHS ${CMAKE_SOURCE_DIR}/third_party/webrtc/lib)
    if(WEBRTC_LIB)
        target_link_libraries(VideoConferenceClient ${WEBRTC_LIB})
        target_include_directories(VideoConferenceClient PRIVATE ${CMAKE_SOURCE_DIR}/third_party/webrtc/include)
    endif()
endif()

# 安装规则
install(TARGETS VideoConferenceClient
    BUNDLE DESTINATION .
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)
