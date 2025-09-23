#include <QtWidgets/QApplication>
#include <QtCore/QDir>
#include <QtCore/QStandardPaths>
#include <QtCore/QLoggingCategory>
#include <QtCore/QCommandLineParser>
#include <QtCore/QCommandLineOption>
#include <QtWidgets/QMessageBox>
#include <QtWidgets/QStyleFactory>
#include <QtCore/QTranslator>
#include <QtCore/QLibraryInfo>

#include "mainwindow.h"

// 日志分类
Q_LOGGING_CATEGORY(app, "videoconference.app")
Q_LOGGING_CATEGORY(webrtc, "videoconference.webrtc")
Q_LOGGING_CATEGORY(signaling, "videoconference.signaling")
Q_LOGGING_CATEGORY(detection, "videoconference.detection")

void setupLogging()
{
    // 设置日志格式
    qSetMessagePattern("[%{time yyyy-MM-dd hh:mm:ss.zzz}] [%{category}] [%{type}] %{message}");
    
    // 启用所有日志类别
    QLoggingCategory::setFilterRules("videoconference.*=true");
}

void setupApplication(QApplication &app)
{
    // 设置应用程序信息
    app.setApplicationName("VideoConferenceClient");
    app.setApplicationDisplayName("视频会议系统");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("VideoConference");
    app.setOrganizationDomain("videoconference.com");
    
    // 设置应用程序图标
    app.setWindowIcon(QIcon(":/icons/app.png"));
    
    // 设置样式
    app.setStyle(QStyleFactory::create("Fusion"));
    
    // 设置深色主题
    QPalette darkPalette;
    darkPalette.setColor(QPalette::Window, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::WindowText, Qt::white);
    darkPalette.setColor(QPalette::Base, QColor(25, 25, 25));
    darkPalette.setColor(QPalette::AlternateBase, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::ToolTipBase, Qt::white);
    darkPalette.setColor(QPalette::ToolTipText, Qt::white);
    darkPalette.setColor(QPalette::Text, Qt::white);
    darkPalette.setColor(QPalette::Button, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::ButtonText, Qt::white);
    darkPalette.setColor(QPalette::BrightText, Qt::red);
    darkPalette.setColor(QPalette::Link, QColor(42, 130, 218));
    darkPalette.setColor(QPalette::Highlight, QColor(42, 130, 218));
    darkPalette.setColor(QPalette::HighlightedText, Qt::black);
    app.setPalette(darkPalette);
    
    // 设置样式表
    QString styleSheet = R"(
        QMainWindow {
            background-color: #2d2d2d;
        }
        QToolBar {
            background-color: #3d3d3d;
            border: none;
            spacing: 3px;
            padding: 5px;
        }
        QToolBar::separator {
            background-color: #555;
            width: 1px;
            margin: 5px;
        }
        QStatusBar {
            background-color: #3d3d3d;
            border-top: 1px solid #555;
        }
        QTabWidget::pane {
            border: 1px solid #555;
            background-color: #2d2d2d;
        }
        QTabBar::tab {
            background-color: #3d3d3d;
            color: white;
            padding: 8px 16px;
            margin-right: 2px;
        }
        QTabBar::tab:selected {
            background-color: #4facfe;
        }
        QTabBar::tab:hover {
            background-color: #555;
        }
        QSplitter::handle {
            background-color: #555;
        }
        QSplitter::handle:horizontal {
            width: 3px;
        }
        QSplitter::handle:vertical {
            height: 3px;
        }
    )";
    app.setStyleSheet(styleSheet);
}

bool checkSystemRequirements()
{
    // 检查Qt版本
    if (QT_VERSION < QT_VERSION_CHECK(6, 0, 0)) {
        QMessageBox::critical(nullptr, "系统要求", 
                             "此应用程序需要Qt 6.0或更高版本。");
        return false;
    }
    
    // 检查多媒体支持
    // 这里可以添加更多的系统要求检查
    
    return true;
}

void setupTranslation(QApplication &app)
{
    // 加载Qt翻译
    QTranslator *qtTranslator = new QTranslator(&app);
    if (qtTranslator->load("qt_" + QLocale::system().name(),
                          QLibraryInfo::path(QLibraryInfo::TranslationsPath))) {
        app.installTranslator(qtTranslator);
    }
    
    // 加载应用程序翻译
    QTranslator *appTranslator = new QTranslator(&app);
    if (appTranslator->load("videoconference_" + QLocale::system().name(),
                           ":/translations")) {
        app.installTranslator(appTranslator);
    }
}

int main(int argc, char *argv[])
{
    // 启用高DPI支持
    QApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication::setAttribute(Qt::AA_UseHighDpiPixmaps);
    
    QApplication app(argc, argv);
    
    // 设置日志
    setupLogging();
    
    qCInfo(app) << "启动视频会议客户端...";
    qCInfo(app) << "Qt版本:" << QT_VERSION_STR;
    qCInfo(app) << "应用程序版本:" << app.applicationVersion();
    
    // 解析命令行参数
    QCommandLineParser parser;
    parser.setApplicationDescription("视频会议系统Qt客户端");
    parser.addHelpOption();
    parser.addVersionOption();
    
    QCommandLineOption serverOption(QStringList() << "s" << "server",
                                   "服务器地址 (默认: localhost)",
                                   "address", "localhost");
    parser.addOption(serverOption);
    
    QCommandLineOption portOption(QStringList() << "p" << "port",
                                 "服务器端口 (默认: 8080)",
                                 "port", "8080");
    parser.addOption(portOption);
    
    QCommandLineOption debugOption(QStringList() << "d" << "debug",
                                  "启用调试模式");
    parser.addOption(debugOption);
    
    parser.process(app);
    
    // 检查系统要求
    if (!checkSystemRequirements()) {
        return 1;
    }
    
    // 设置应用程序
    setupApplication(app);
    setupTranslation(app);
    
    // 创建主窗口
    MainWindow window;
    
    // 应用命令行参数
    if (parser.isSet(serverOption)) {
        // 设置服务器地址
        qCInfo(app) << "使用服务器地址:" << parser.value(serverOption);
    }
    
    if (parser.isSet(portOption)) {
        // 设置服务器端口
        qCInfo(app) << "使用服务器端口:" << parser.value(portOption);
    }
    
    if (parser.isSet(debugOption)) {
        // 启用调试模式
        QLoggingCategory::setFilterRules("*.debug=true");
        qCInfo(app) << "调试模式已启用";
    }
    
    // 显示主窗口
    window.show();
    
    qCInfo(app) << "应用程序启动完成";
    
    // 运行应用程序
    int result = app.exec();
    
    qCInfo(app) << "应用程序退出，返回码:" << result;
    
    return result;
}
