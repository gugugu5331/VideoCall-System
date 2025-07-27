QT += core gui widgets

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

CONFIG += c++17

SOURCES += \
    demo_app.cpp

# Default rules for deployment.
qnx: target.path = /tmp/$${TARGET}/bin
else: unix:!android: target.path = /opt/$${TARGET}/bin
!isEmpty(target.path): INSTALLS += target

# 编译配置
CONFIG(debug, debug|release) {
    DESTDIR = debug
} else {
    DESTDIR = release
}

# Windows特定配置
win32 {
    LIBS += -lws2_32 -liphlpapi
}

# 定义
DEFINES += \
    QT_DEPRECATED_WARNINGS \
    DEMO_APP_VERSION=\"1.0.0\" 