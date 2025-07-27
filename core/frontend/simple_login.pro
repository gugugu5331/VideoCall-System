QT += core gui widgets network

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

CONFIG += c++17

SOURCES += \
    simple_login.cpp

# 编译配置
CONFIG(debug, debug|release) {
    DESTDIR = debug
} else {
    DESTDIR = release
}

# Windows特定配置
win32 {
    LIBS += -lws2_32
}

# 定义
DEFINES += \
    QT_DEPRECATED_WARNINGS 