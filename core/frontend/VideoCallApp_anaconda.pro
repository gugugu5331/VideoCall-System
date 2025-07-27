QT += core gui widgets

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

CONFIG += c++14

# Anaconda Qt兼容版本
SOURCES += \
    main.cpp \
    mainwindow.cpp \
    loginwidget.cpp

HEADERS += \
    mainwindow.h \
    loginwidget.h

FORMS += \
    mainwindow.ui \
    loginwidget.ui

# 资源文件
RESOURCES += \
    resources.qrc

# 编译配置
CONFIG(debug, debug|release) {
    DESTDIR = debug
} else {
    DESTDIR = release
}

# 包含路径
INCLUDEPATH += \
    include \
    src

# Windows特定配置
win32 {
    LIBS += -lws2_32 -liphlpapi
}

# 定义
DEFINES += \
    QT_DEPRECATED_WARNINGS \
    VIDEO_CALL_APP_VERSION=\"1.0.0\" \
    SIMPLE_MODE

# Anaconda Qt库配置
# 注释掉OpenCV依赖，使用简化模式
# LIBS += \
#     -lopencv_core \
#     -lopencv_imgproc \
#     -lopencv_videoio \
#     -lopencv_face \
#     -lopencv_dnn 