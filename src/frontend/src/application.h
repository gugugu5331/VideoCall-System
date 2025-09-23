#ifndef APPLICATION_H
#define APPLICATION_H

#include <QApplication>
#include <QQmlApplicationEngine>
#include <QSettings>
#include <QTranslator>
#include <memory>

class AuthService;
class ApiService;
class MeetingService;
class DetectionService;
class WebRTCService;
class MediaService;
class MainController;
class SettingsManager;
class Logger;

/**
 * @brief 主应用程序类
 * 
 * 负责应用程序的初始化、服务管理和生命周期控制
 */
class Application : public QApplication
{
    Q_OBJECT

public:
    explicit Application(int &argc, char **argv);
    ~Application();

    /**
     * @brief 初始化应用程序
     * @return 是否初始化成功
     */
    bool initialize();

    /**
     * @brief 运行应用程序
     * @return 应用程序退出码
     */
    int run();

    // 获取服务实例
    AuthService* authService() const { return m_authService.get(); }
    ApiService* apiService() const { return m_apiService.get(); }
    MeetingService* meetingService() const { return m_meetingService.get(); }
    DetectionService* detectionService() const { return m_detectionService.get(); }
    WebRTCService* webrtcService() const { return m_webrtcService.get(); }
    MediaService* mediaService() const { return m_mediaService.get(); }
    SettingsManager* settingsManager() const { return m_settingsManager.get(); }
    Logger* logger() const { return m_logger.get(); }

    // 单例访问
    static Application* instance();

public slots:
    /**
     * @brief 退出应用程序
     */
    void quit();

    /**
     * @brief 重启应用程序
     */
    void restart();

private slots:
    /**
     * @brief 处理应用程序状态变化
     */
    void onApplicationStateChanged(Qt::ApplicationState state);

private:
    /**
     * @brief 初始化服务
     */
    void initializeServices();

    /**
     * @brief 初始化UI
     */
    void initializeUI();

    /**
     * @brief 注册QML类型
     */
    void registerQmlTypes();

    /**
     * @brief 设置应用程序属性
     */
    void setupApplicationProperties();

    /**
     * @brief 加载翻译文件
     */
    void loadTranslations();

    /**
     * @brief 应用样式表
     */
    void applyStyleSheet();

private:
    // QML引擎
    std::unique_ptr<QQmlApplicationEngine> m_engine;
    
    // 服务层
    std::unique_ptr<AuthService> m_authService;
    std::unique_ptr<ApiService> m_apiService;
    std::unique_ptr<MeetingService> m_meetingService;
    std::unique_ptr<DetectionService> m_detectionService;
    std::unique_ptr<WebRTCService> m_webrtcService;
    std::unique_ptr<MediaService> m_mediaService;
    
    // 控制器
    std::unique_ptr<MainController> m_mainController;
    
    // 工具类
    std::unique_ptr<SettingsManager> m_settingsManager;
    std::unique_ptr<Logger> m_logger;
    
    // 翻译器
    std::unique_ptr<QTranslator> m_translator;
    
    // 静态实例
    static Application* s_instance;
};

#endif // APPLICATION_H
