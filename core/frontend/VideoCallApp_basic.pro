QT += core gui widgets network concurrent

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

CONFIG += c++17

# 最基本版本，只包含核心功能
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

# Windows特定配置
win32 {
    LIBS += -lws2_32
}

# 定义
DEFINES += \
    QT_DEPRECATED_WARNINGS \
    VIDEO_CALL_APP_VERSION=\"1.0.0\" \
    BASIC_MODE 