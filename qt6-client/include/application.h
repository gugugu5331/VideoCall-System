#ifndef APPLICATION_H
#define APPLICATION_H

#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <memory>

// Forward declarations
class AuthService;
class MeetingService;
class MediaService;
class AIService;
class WebRTCManager;
class VideoEffectsController;
class Config;
class Logger;

class Application : public QObject
{
    Q_OBJECT

public:
    explicit Application(int &argc, char **argv);
    ~Application();

    int run();

    // Singleton access
    static Application* instance();

    // Service getters
    AuthService* authService() const { return m_authService.get(); }
    MeetingService* meetingService() const { return m_meetingService.get(); }
    MediaService* mediaService() const { return m_mediaService.get(); }
    AIService* aiService() const { return m_aiService.get(); }
    WebRTCManager* webrtcManager() const { return m_webrtcManager.get(); }
    VideoEffectsController* videoEffectsController() const { return m_videoEffectsController.get(); }
    Config* config() const { return m_config.get(); }
    Logger* logger() const { return m_logger.get(); }

signals:
    void initialized();
    void error(const QString &message);

private:
    void initializeServices();
    void registerQmlTypes();
    void setupQmlContext();

private:
    static Application* s_instance;

    std::unique_ptr<QGuiApplication> m_app;
    std::unique_ptr<QQmlApplicationEngine> m_engine;

    // Core services
    std::unique_ptr<Config> m_config;
    std::unique_ptr<Logger> m_logger;

    // Business services
    std::unique_ptr<AuthService> m_authService;
    std::unique_ptr<MeetingService> m_meetingService;
    std::unique_ptr<MediaService> m_mediaService;
    std::unique_ptr<AIService> m_aiService;

    // WebRTC
    std::unique_ptr<WebRTCManager> m_webrtcManager;

    // Video Effects
    std::unique_ptr<VideoEffectsController> m_videoEffectsController;
};

#endif // APPLICATION_H

