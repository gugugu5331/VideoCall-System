cmake_minimum_required(VERSION 3.16)

project(MeetingSystemClient VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# Qt6 packages
find_package(Qt6 REQUIRED COMPONENTS
    Core
    Gui
    Widgets
    Quick
    QuickControls2
    Network
    WebSockets
    Multimedia
    MultimediaWidgets
    Sql
    OpenGL
    OpenGLWidgets
)

# OpenCV package
find_package(OpenCV REQUIRED)
if(OpenCV_FOUND)
    message(STATUS "OpenCV found: ${OpenCV_VERSION}")
    include_directories(${OpenCV_INCLUDE_DIRS})
else()
    message(FATAL_ERROR "OpenCV not found. Please install OpenCV 4.x")
endif()

# WebRTC package (新增) - 暂时注释，项目使用Qt Multimedia模拟WebRTC
# list(APPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")
# find_package(WebRTC REQUIRED)
# if(WebRTC_FOUND)
#     message(STATUS "WebRTC found: ${WebRTC_LIBRARIES}")
#     include_directories(${WebRTC_INCLUDE_DIRS})
# else()
#     message(FATAL_ERROR "WebRTC not found. Please set WEBRTC_ROOT environment variable")
# endif()

# Include directories
include_directories(
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_SOURCE_DIR}/src
)

# Source files
set(SOURCES
    src/main.cpp
    src/application.cpp
    
    # Network layer
    src/network/http_client.cpp
    src/network/websocket_client.cpp
    src/network/api_client.cpp
    
    # WebRTC layer
    src/webrtc/webrtc_manager.cpp
    src/webrtc/peer_connection.cpp
    src/webrtc/media_stream.cpp
    src/webrtc/remote_stream_analyzer.cpp

    # Video Effects
    src/video_effects/video_effect_processor.cpp
    src/video_effects/gl_video_renderer.cpp
    src/video_effects/sticker_overlay.cpp

    # Services
    src/services/auth_service.cpp
    src/services/meeting_service.cpp
    src/services/media_service.cpp
    src/services/ai_service.cpp
    
    # Models
    src/models/user.cpp
    src/models/meeting.cpp
    src/models/participant.cpp
    src/models/message.cpp
    
    # UI Controllers
    src/ui/login_controller.cpp
    src/ui/main_window_controller.cpp
    src/ui/meeting_room_controller.cpp
    src/ui/ai_panel_controller.cpp
    src/ui/video_effects_controller.cpp
    
    # Utils
    src/utils/logger.cpp
    src/utils/config.cpp
    src/utils/json_helper.cpp
)

# Header files
set(HEADERS
    include/application.h
    
    # Network layer
    include/network/http_client.h
    include/network/websocket_client.h
    include/network/api_client.h
    
    # WebRTC layer
    include/webrtc/webrtc_manager.h
    include/webrtc/peer_connection.h
    include/webrtc/media_stream.h
    include/webrtc/remote_stream_analyzer.h

    # Video Effects
    include/video_effects/video_effect_processor.h
    include/video_effects/gl_video_renderer.h
    include/video_effects/sticker_overlay.h

    # Services
    include/services/auth_service.h
    include/services/meeting_service.h
    include/services/media_service.h
    include/services/ai_service.h
    
    # Models
    include/models/user.h
    include/models/meeting.h
    include/models/participant.h
    include/models/message.h
    
    # UI Controllers
    include/ui/login_controller.h
    include/ui/main_window_controller.h
    include/ui/meeting_room_controller.h
    include/ui/ai_panel_controller.h
    include/ui/video_effects_controller.h
    
    # Utils
    include/utils/logger.h
    include/utils/config.h
    include/utils/json_helper.h
)

# QML files
set(QML_FILES
    qml/main.qml
    qml/LoginPage.qml
    qml/MainWindow.qml
    qml/MeetingRoom.qml
    qml/AIPanel.qml
    qml/VideoEffectsPanel.qml
    qml/components/VideoTile.qml
    qml/components/MeetingToolBar.qml
    qml/components/ParticipantList.qml
    qml/components/ChatPanel.qml
)

# Resources
qt6_add_resources(RESOURCES
    resources/resources.qrc
)

# Create executable
add_executable(${PROJECT_NAME}
    ${SOURCES}
    ${HEADERS}
    ${RESOURCES}
)

# Link Qt libraries
target_link_libraries(${PROJECT_NAME} PRIVATE
    Qt6::Core
    Qt6::Gui
    Qt6::Widgets
    Qt6::Quick
    Qt6::QuickControls2
    Qt6::Network
    Qt6::WebSockets
    Qt6::Multimedia
    Qt6::MultimediaWidgets
    Qt6::Sql
    Qt6::OpenGL
    Qt6::OpenGLWidgets
    ${OpenCV_LIBS}
    # WebRTC::WebRTC  # 新增WebRTC链接 - 暂时注释，使用Qt Multimedia
)

# Set output directory
set_target_properties(${PROJECT_NAME} PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin
)

# Install
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
)

# Copy config file
configure_file(
    ${CMAKE_SOURCE_DIR}/config/config.json
    ${CMAKE_BINARY_DIR}/bin/config.json
    COPYONLY
)

