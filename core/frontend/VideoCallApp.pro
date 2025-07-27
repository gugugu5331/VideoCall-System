QT += core gui widgets multimedia multimediawidgets network webenginewidgets
greaterThan(QT_MAJOR_VERSION, 4): QT += widgets
CONFIG += c++17
# You can make your code fail to compile if it uses deprecated APIs.
# In order to do so, uncomment the following line.
#DEFINES += QT_DISABLE_DEPRECATED_BEFORE=0x060000    # disables all the APIs deprecated before Qt 6.0.0
SOURCES += \
    main.cpp \
    mainwindow.cpp \
    videocallwidget.cpp \
    loginwidget.cpp \
    userprofilewidget.cpp \
    callhistorywidget.cpp \
    settingswidget.cpp \
    securitydetectionwidget.cpp \
    networkmanager.cpp \
    audiomanager.cpp \
    videomanager.cpp \
    securitymanager.cpp
HEADERS += \
    mainwindow.h \
    videocallwidget.h \
    loginwidget.h \
    userprofilewidget.h \
    callhistorywidget.h \
    settingswidget.h \
    securitydetectionwidget.h \
    networkmanager.h \
    audiomanager.h \
    videomanager.h \
    securitymanager.h
FORMS += \
    mainwindow.ui \
    videocallwidget.ui \
    loginwidget.ui \
    userprofilewidget.ui \
    callhistorywidget.ui \
    settingswidget.ui \
    securitydetectionwidget.ui
# Default rules for deployment.
qnx: target.path = /tmp/$${TARGET}/bin
else: unix: target.path = /opt/$${TARGET}/bin
 INSTALLS += target
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
# 库文件
LIBS += \
    -lopencv_core \
    -lopencv_imgproc \
    -lopencv_videoio \
    -lopencv_face \
    -lopencv_dnn
# Windows特定配置
win32 {
    LIBS += -lws2_32 -liphlpapi
    RC_ICONS = resources/icon.ico
}
# 定义
DEFINES += \
    QT_DEPRECATED_WARNINGS \
    VIDEO_CALL_APP_VERSION=\"1.0.0\" 
 
# OpenCV 配置 
INCLUDEPATH += C:\opencv\build\x64\vc16\bin\..\include 
LIBS += -LC:\opencv\build\x64\vc16\bin -lopencv_world 
