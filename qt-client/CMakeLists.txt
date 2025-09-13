cmake_minimum_required(VERSION 3.16)

project(VideoConferenceClient VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find required Qt components
find_package(Qt6 REQUIRED COMPONENTS 
    Core 
    Widgets 
    Multimedia 
    MultimediaWidgets 
    Network 
    WebSockets
)

# Enable automatic MOC, UIC, and RCC
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTOUIC ON)
set(CMAKE_AUTORCC ON)

# Source files
set(SOURCES
    src/main.cpp
    src/mainwindow.cpp
    src/videowidget.cpp
    src/webrtcmanager.cpp
    src/signalingclient.cpp
    src/detectionmanager.cpp
    src/participantwidget.cpp
    src/chatwidget.cpp
    src/settingsdialog.cpp
)

# Header files
set(HEADERS
    src/mainwindow.h
    src/videowidget.h
    src/webrtcmanager.h
    src/signalingclient.h
    src/detectionmanager.h
    src/participantwidget.h
    src/chatwidget.h
    src/settingsdialog.h
)

# UI files
set(UI_FILES
    ui/mainwindow.ui
    ui/settingsdialog.ui
)

# Resource files
set(RESOURCES
    resources/resources.qrc
)

# Create executable
add_executable(VideoConferenceClient
    ${SOURCES}
    ${HEADERS}
    ${UI_FILES}
    ${RESOURCES}
)

# Link Qt libraries
target_link_libraries(VideoConferenceClient
    Qt6::Core
    Qt6::Widgets
    Qt6::Multimedia
    Qt6::MultimediaWidgets
    Qt6::Network
    Qt6::WebSockets
)

# Set target properties
set_target_properties(VideoConferenceClient PROPERTIES
    WIN32_EXECUTABLE TRUE
    MACOSX_BUNDLE TRUE
)

# Include directories
target_include_directories(VideoConferenceClient PRIVATE
    src
    ${CMAKE_CURRENT_BINARY_DIR}
)

# Compiler-specific options
if(MSVC)
    target_compile_options(VideoConferenceClient PRIVATE /W4)
else()
    target_compile_options(VideoConferenceClient PRIVATE -Wall -Wextra -Wpedantic)
endif()

# Install rules
install(TARGETS VideoConferenceClient
    BUNDLE DESTINATION .
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)
