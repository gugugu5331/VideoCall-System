#include "application.h"
#include <QLoggingCategory>
#include <QDir>
#include <QStandardPaths>
#include <iostream>

// 启用Qt日志分类
Q_LOGGING_CATEGORY(app, "app")
Q_LOGGING_CATEGORY(auth, "auth")
Q_LOGGING_CATEGORY(meeting, "meeting")
Q_LOGGING_CATEGORY(detection, "detection")
Q_LOGGING_CATEGORY(webrtc, "webrtc")
Q_LOGGING_CATEGORY(media, "media")

/**
 * @brief 设置日志输出格式
 */
void setupLogging()
{
    // 设置日志格式
    qSetMessagePattern("[%{time yyyy-MM-dd hh:mm:ss.zzz}] [%{category}] [%{type}] %{message}");
    
    // 创建日志目录
    QString logDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/logs";
    QDir().mkpath(logDir);
    
    // 设置日志级别
#ifdef QT_DEBUG
    QLoggingCategory::setFilterRules("*.debug=true");
#else
    QLoggingCategory::setFilterRules("*.debug=false\n*.info=true\n*.warning=true\n*.critical=true");
#endif
}

/**
 * @brief 检查系统要求
 */
bool checkSystemRequirements()
{
    // 检查Qt版本
    if (QT_VERSION < QT_VERSION_CHECK(6, 5, 0)) {
        std::cerr << "Error: Qt 6.5.0 or higher is required" << std::endl;
        return false;
    }
    
    // 检查OpenGL支持
    // 这里可以添加更多的系统检查
    
    return true;
}

/**
 * @brief 处理未捕获的异常
 */
void handleException(const std::exception& e)
{
    qCCritical(app) << "Unhandled exception:" << e.what();
    
    // 显示错误对话框
    QMessageBox::critical(nullptr, 
                         "Fatal Error", 
                         QString("An unhandled exception occurred:\n%1\n\nThe application will now exit.")
                         .arg(e.what()));
}

int main(int argc, char *argv[])
{
    try {
        // 设置应用程序属性（必须在创建QApplication之前）
        QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
        QCoreApplication::setAttribute(Qt::AA_UseHighDpiPixmaps);
        
        // 创建应用程序实例
        Application app(argc, argv);
        
        // 设置日志
        setupLogging();
        
        // 检查系统要求
        if (!checkSystemRequirements()) {
            return -1;
        }
        
        qCInfo(app) << "Starting Video Conference Client";
        qCInfo(app) << "Version:" << app.applicationVersion();
        qCInfo(app) << "Qt Version:" << qVersion();
        
        // 运行应用程序
        int result = app.run();
        
        qCInfo(app) << "Application exited with code:" << result;
        return result;
    }
    catch (const std::exception& e) {
        handleException(e);
        return -1;
    }
    catch (...) {
        qCCritical(app) << "Unknown exception occurred";
        QMessageBox::critical(nullptr, 
                             "Fatal Error", 
                             "An unknown error occurred. The application will now exit.");
        return -1;
    }
}
