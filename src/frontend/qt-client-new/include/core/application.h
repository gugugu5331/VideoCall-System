#pragma once

#include "common.h"

namespace VideoCallSystem {

class ConfigManager;
class Logger;
class ServiceManager;
class MainWindow;

class Application : public QApplication
{
    Q_OBJECT

public:
    explicit Application(int &argc, char **argv);
    ~Application();

    // 单例访问
    static Application* instance();

    // 初始化和清理
    bool initialize();
    void cleanup();

    // 配置管理
    ConfigManager* configManager() const { return configManager_; }
    Logger* logger() const { return logger_; }
    ServiceManager* serviceManager() const { return serviceManager_; }

    // 主窗口管理
    MainWindow* mainWindow() const { return mainWindow_; }
    void showMainWindow();
    void hideMainWindow();

    // 应用程序状态
    bool isInitialized() const { return initialized_; }
    QString version() const { return APP_VERSION; }
    QString buildInfo() const;

    // 系统信息
    QString systemInfo() const;
    QString qtVersion() const;
    QString openCVVersion() const;

    // 错误处理
    void handleCriticalError(const QString& error);
    void showErrorMessage(const QString& title, const QString& message);
    void showInfoMessage(const QString& title, const QString& message);

    // 应用程序设置
    void setTheme(const QString& themeName);
    QString currentTheme() const { return currentTheme_; }
    QStringList availableThemes() const;

    // 语言设置
    void setLanguage(const QString& languageCode);
    QString currentLanguage() const { return currentLanguage_; }
    QStringList availableLanguages() const;

    // 自动更新
    void checkForUpdates();
    bool isUpdateAvailable() const { return updateAvailable_; }
    QString latestVersion() const { return latestVersion_; }

    // 崩溃报告
    void enableCrashReporting(bool enable);
    bool isCrashReportingEnabled() const { return crashReportingEnabled_; }

    // 性能监控
    void startPerformanceMonitoring();
    void stopPerformanceMonitoring();
    QVariantMap getPerformanceMetrics() const;

public slots:
    // 应用程序控制
    void restart();
    void quit();
    void aboutQt();
    void aboutApplication();

    // 主题切换
    void switchToLightTheme();
    void switchToDarkTheme();
    void switchToSystemTheme();

    // 语言切换
    void switchLanguage(const QString& languageCode);

    // 更新处理
    void downloadUpdate();
    void installUpdate();

signals:
    // 应用程序状态信号
    void initialized();
    void aboutToQuit();
    void themeChanged(const QString& themeName);
    void languageChanged(const QString& languageCode);

    // 错误信号
    void criticalError(const QString& error);
    void warningMessage(const QString& message);
    void infoMessage(const QString& message);

    // 更新信号
    void updateAvailable(const QString& version);
    void updateDownloaded();
    void updateInstalled();

    // 性能信号
    void performanceMetricsUpdated(const QVariantMap& metrics);

protected:
    // 事件处理
    bool event(QEvent* event) override;
    bool notify(QObject* receiver, QEvent* event) override;

private slots:
    // 内部槽函数
    void onConfigChanged();
    void onServiceStatusChanged();
    void onUpdateCheckFinished();
    void onPerformanceTimer();

private:
    // 初始化函数
    bool initializeCore();
    bool initializeServices();
    bool initializeUI();
    bool initializeTheme();
    bool initializeLanguage();

    // 清理函数
    void cleanupServices();
    void cleanupUI();
    void cleanupCore();

    // 配置函数
    void loadSettings();
    void saveSettings();
    void applyTheme(const QString& themeName);
    void loadTranslations(const QString& languageCode);

    // 工具函数
    QString getConfigPath() const;
    QString getLogPath() const;
    QString getCachePath() const;
    QString getThemePath(const QString& themeName) const;
    QString getTranslationPath(const QString& languageCode) const;

    // 错误处理
    void setupCrashHandler();
    void handleException(const std::exception& e);

    // 性能监控
    void collectPerformanceMetrics();
    void updateMemoryUsage();
    void updateCpuUsage();

private:
    // 核心组件
    ConfigManager* configManager_;
    Logger* logger_;
    ServiceManager* serviceManager_;
    MainWindow* mainWindow_;

    // 应用程序状态
    bool initialized_;
    QString currentTheme_;
    QString currentLanguage_;

    // 更新相关
    bool updateAvailable_;
    QString latestVersion_;
    QNetworkAccessManager* updateManager_;

    // 崩溃报告
    bool crashReportingEnabled_;

    // 性能监控
    QTimer* performanceTimer_;
    QVariantMap performanceMetrics_;
    bool performanceMonitoringEnabled_;

    // 翻译器
    QTranslator* appTranslator_;
    QTranslator* qtTranslator_;

    // 设置
    QSettings* settings_;

    // 静态实例
    static Application* instance_;
};

} // namespace VideoCallSystem
