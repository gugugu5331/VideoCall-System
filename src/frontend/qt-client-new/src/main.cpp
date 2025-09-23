#include "core/application.h"
#include "core/common.h"

#include <QCommandLineParser>
#include <QDir>
#include <QStandardPaths>
#include <QStyleFactory>
#include <QLoggingCategory>

using namespace VideoCallSystem;

// 日志分类
Q_LOGGING_CATEGORY(lcMain, "videocall.main")

// 命令行参数结构
struct CommandLineOptions {
    QString configFile;
    QString logLevel = "info";
    QString theme = "system";
    QString language = "en";
    bool enableGPU = true;
    bool enableCrashReporting = true;
    bool enableTelemetry = false;
    bool debugMode = false;
    bool safeMode = false;
    QString serverUrl;
    int serverPort = 0;
    bool showHelp = false;
    bool showVersion = false;
};

// 解析命令行参数
CommandLineOptions parseCommandLine(QCoreApplication& app) {
    CommandLineOptions options;
    
    QCommandLineParser parser;
    parser.setApplicationDescription("VideoCall System Client - Advanced video conferencing with AI detection");
    parser.addHelpOption();
    parser.addVersionOption();
    
    // 配置选项
    QCommandLineOption configOption(QStringList() << "c" << "config",
        "Configuration file path", "file");
    parser.addOption(configOption);
    
    QCommandLineOption logLevelOption(QStringList() << "l" << "log-level",
        "Log level (debug, info, warning, error, critical)", "level", "info");
    parser.addOption(logLevelOption);
    
    QCommandLineOption themeOption(QStringList() << "t" << "theme",
        "UI theme (light, dark, system)", "theme", "system");
    parser.addOption(themeOption);
    
    QCommandLineOption languageOption(QStringList() << "lang" << "language",
        "UI language (en, zh, ja, ko)", "language", "en");
    parser.addOption(languageOption);
    
    // 功能选项
    QCommandLineOption noGpuOption("no-gpu",
        "Disable GPU acceleration");
    parser.addOption(noGpuOption);
    
    QCommandLineOption noCrashReportingOption("no-crash-reporting",
        "Disable crash reporting");
    parser.addOption(noCrashReportingOption);
    
    QCommandLineOption enableTelemetryOption("enable-telemetry",
        "Enable telemetry data collection");
    parser.addOption(enableTelemetryOption);
    
    QCommandLineOption debugOption(QStringList() << "d" << "debug",
        "Enable debug mode");
    parser.addOption(debugOption);
    
    QCommandLineOption safeModeOption("safe-mode",
        "Start in safe mode (minimal features)");
    parser.addOption(safeModeOption);
    
    // 服务器选项
    QCommandLineOption serverUrlOption(QStringList() << "s" << "server",
        "Server URL", "url");
    parser.addOption(serverUrlOption);
    
    QCommandLineOption serverPortOption(QStringList() << "p" << "port",
        "Server port", "port");
    parser.addOption(serverPortOption);
    
    // 解析参数
    parser.process(app);
    
    // 提取选项值
    if (parser.isSet(configOption)) {
        options.configFile = parser.value(configOption);
    }
    
    if (parser.isSet(logLevelOption)) {
        options.logLevel = parser.value(logLevelOption);
    }
    
    if (parser.isSet(themeOption)) {
        options.theme = parser.value(themeOption);
    }
    
    if (parser.isSet(languageOption)) {
        options.language = parser.value(languageOption);
    }
    
    options.enableGPU = !parser.isSet(noGpuOption);
    options.enableCrashReporting = !parser.isSet(noCrashReportingOption);
    options.enableTelemetry = parser.isSet(enableTelemetryOption);
    options.debugMode = parser.isSet(debugOption);
    options.safeMode = parser.isSet(safeModeOption);
    
    if (parser.isSet(serverUrlOption)) {
        options.serverUrl = parser.value(serverUrlOption);
    }
    
    if (parser.isSet(serverPortOption)) {
        bool ok;
        int port = parser.value(serverPortOption).toInt(&ok);
        if (ok && port > 0 && port <= 65535) {
            options.serverPort = port;
        }
    }
    
    return options;
}

// 设置日志级别
void setupLogging(const QString& logLevel) {
    QLoggingCategory::setFilterRules("*=false");
    
    if (logLevel == "debug") {
        QLoggingCategory::setFilterRules("*.debug=true\n*.info=true\n*.warning=true\n*.critical=true");
    } else if (logLevel == "info") {
        QLoggingCategory::setFilterRules("*.info=true\n*.warning=true\n*.critical=true");
    } else if (logLevel == "warning") {
        QLoggingCategory::setFilterRules("*.warning=true\n*.critical=true");
    } else if (logLevel == "error") {
        QLoggingCategory::setFilterRules("*.critical=true");
    } else if (logLevel == "critical") {
        QLoggingCategory::setFilterRules("*.critical=true");
    }
}

// 设置应用程序属性
void setupApplicationAttributes() {
    // 启用高DPI支持
    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QCoreApplication::setAttribute(Qt::AA_UseHighDpiPixmaps);
    
    // 设置OpenGL属性
    QCoreApplication::setAttribute(Qt::AA_UseDesktopOpenGL);
    QCoreApplication::setAttribute(Qt::AA_ShareOpenGLContexts);
    
    // 其他属性
    QCoreApplication::setAttribute(Qt::AA_DontCreateNativeWidgetSiblings);
    QCoreApplication::setAttribute(Qt::AA_SynthesizeMouseForUnhandledTouchEvents);
}

// 检查系统要求
bool checkSystemRequirements() {
    // 检查Qt版本
    if (QT_VERSION < QT_VERSION_CHECK(6, 0, 0)) {
        qCCritical(lcMain) << "Qt 6.0 or higher is required";
        return false;
    }
    
    // 检查OpenGL支持
    // 这里可以添加更多的系统检查
    
    return true;
}

// 创建应用程序目录
void createApplicationDirectories() {
    QStringList dirs = {
        QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation),
        QStandardPaths::writableLocation(QStandardPaths::AppDataLocation),
        QStandardPaths::writableLocation(QStandardPaths::CacheLocation),
        QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/logs",
        QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/recordings",
        QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/screenshots",
        QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/cache"
    };
    
    for (const QString& dir : dirs) {
        QDir().mkpath(dir);
    }
}

// 应用命令行选项
void applyCommandLineOptions(Application* app, const CommandLineOptions& options) {
    auto config = app->configManager();
    
    // 应用服务器配置
    if (!options.serverUrl.isEmpty()) {
        config->setServerUrl(options.serverUrl);
    }
    
    if (options.serverPort > 0) {
        config->setServerPort(options.serverPort);
    }
    
    // 应用UI配置
    if (options.theme != "system") {
        config->setTheme(options.theme);
    }
    
    if (options.language != "en") {
        config->setLanguage(options.language);
    }
    
    // 应用功能配置
    config->setEnableHardwareAcceleration(options.enableGPU);
    config->setEnableCrashReporting(options.enableCrashReporting);
    config->setEnableTelemetry(options.enableTelemetry);
    
    // 保存配置
    config->saveConfig();
}

// 显示启动画面
void showSplashScreen() {
    // 这里可以显示启动画面
    // QSplashScreen* splash = new QSplashScreen(pixmap);
    // splash->show();
    // QCoreApplication::processEvents();
}

// 主函数
int main(int argc, char *argv[])
{
    // 设置应用程序属性
    setupApplicationAttributes();
    
    // 创建应用程序实例
    Application app(argc, argv);
    
    // 设置应用程序信息
    app.setApplicationName(Constants::APP_NAME);
    app.setApplicationVersion(Constants::APP_VERSION);
    app.setOrganizationName(Constants::APP_ORGANIZATION);
    app.setOrganizationDomain(Constants::APP_DOMAIN);
    
    // 解析命令行参数
    CommandLineOptions options = parseCommandLine(app);
    
    // 设置日志
    setupLogging(options.logLevel);
    
    qCInfo(lcMain) << "Starting" << Constants::APP_NAME << "version" << Constants::APP_VERSION;
    qCInfo(lcMain) << "Qt version:" << qVersion();
    qCInfo(lcMain) << "Build info:" << app.buildInfo();
    
    // 检查系统要求
    if (!checkSystemRequirements()) {
        qCCritical(lcMain) << "System requirements not met";
        return -1;
    }
    
    // 创建应用程序目录
    createApplicationDirectories();
    
    // 显示启动画面
    if (!options.safeMode) {
        showSplashScreen();
    }
    
    try {
        // 初始化应用程序
        qCInfo(lcMain) << "Initializing application...";
        if (!app.initialize()) {
            qCCritical(lcMain) << "Failed to initialize application";
            return -1;
        }
        
        // 应用命令行选项
        applyCommandLineOptions(&app, options);
        
        // 启用崩溃报告
        if (options.enableCrashReporting) {
            app.enableCrashReporting(true);
        }
        
        // 启动性能监控
        if (options.debugMode) {
            app.startPerformanceMonitoring();
        }
        
        // 显示主窗口
        qCInfo(lcMain) << "Showing main window...";
        app.showMainWindow();
        
        qCInfo(lcMain) << "Application started successfully";
        
        // 运行事件循环
        int result = app.exec();
        
        qCInfo(lcMain) << "Application finished with code:" << result;
        return result;
        
    } catch (const std::exception& e) {
        qCCritical(lcMain) << "Unhandled exception:" << e.what();
        app.handleCriticalError(QString("Unhandled exception: %1").arg(e.what()));
        return -1;
    } catch (...) {
        qCCritical(lcMain) << "Unknown exception occurred";
        app.handleCriticalError("Unknown exception occurred");
        return -1;
    }
}
