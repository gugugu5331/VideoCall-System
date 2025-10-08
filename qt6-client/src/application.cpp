#include "application.h"
#include "utils/config.h"
#include "utils/logger.h"
#include "network/api_client.h"
#include "network/websocket_client.h"
#include "services/auth_service.h"
#include "services/meeting_service.h"
#include "services/media_service.h"
#include "services/ai_service.h"
#include "webrtc/webrtc_manager.h"
#include "ui/login_controller.h"
#include "ui/main_window_controller.h"
#include "ui/meeting_room_controller.h"
#include "ui/ai_panel_controller.h"
#include "ui/video_effects_controller.h"

#include <QQmlContext>
#include <QDebug>
#include <QQuickStyle>

Application* Application::s_instance = nullptr;

Application::Application(int &argc, char **argv)
    : QObject(nullptr)
{
    s_instance = this;

    // Set Qt Quick style to Basic (supports customization)
    QQuickStyle::setStyle("Basic");

    // Create QGuiApplication
    m_app = std::make_unique<QGuiApplication>(argc, argv);
    m_app->setOrganizationName("Meeting System");
    m_app->setOrganizationDomain("meeting.com");
    m_app->setApplicationName("智能会议系统");
    m_app->setApplicationVersion("1.0.0");

    // Initialize core services
    m_config = std::make_unique<Config>();
    m_logger = std::make_unique<Logger>();

    // Load configuration
    QString configPath = "config.json";
    if (!m_config->load(configPath)) {
        qWarning() << "Failed to load config from" << configPath << ", using defaults";
    }

    // Setup logger
    m_logger->setLogLevel(LogLevel::Info);
    m_logger->setLogFile("meeting-client.log");

    LOG_INFO("Application starting...");

    // Initialize services
    initializeServices();

    // Create QML engine
    m_engine = std::make_unique<QQmlApplicationEngine>();

    // Register QML types
    registerQmlTypes();

    // Setup QML context
    setupQmlContext();

    // Load main QML file
    const QUrl url(QStringLiteral("qrc:/qml/main.qml"));
    QObject::connect(m_engine.get(), &QQmlApplicationEngine::objectCreated,
                    m_app.get(), [url](QObject *obj, const QUrl &objUrl) {
        if (!obj && url == objUrl)
            QCoreApplication::exit(-1);
    }, Qt::QueuedConnection);

    m_engine->load(url);

    LOG_INFO("Application initialized successfully");
}

Application::~Application()
{
    LOG_INFO("Application shutting down...");
    s_instance = nullptr;
}

int Application::run()
{
    return m_app->exec();
}

Application* Application::instance()
{
    return s_instance;
}

void Application::initializeServices()
{
    // Create API client
    QString apiBaseUrl = m_config->apiBaseUrl();
    auto apiClient = new ApiClient(apiBaseUrl, this);

    // Create WebSocket client
    auto wsClient = new WebSocketClient(this);

    // Create business services
    m_authService = std::make_unique<AuthService>(apiClient);
    m_meetingService = std::make_unique<MeetingService>(apiClient, wsClient);
    m_mediaService = std::make_unique<MediaService>(apiClient);
    m_aiService = std::make_unique<AIService>(apiClient);

    // Create WebRTC manager
    m_webrtcManager = std::make_unique<WebRTCManager>(wsClient);

    // 设置AIService到WebRTCManager
    m_webrtcManager->setAIService(m_aiService.get());

    // Create Video Effects Controller
    m_videoEffectsController = std::make_unique<VideoEffectsController>();

    // Connect signals
    connect(m_authService.get(), &AuthService::loginSuccess, this, [this]() {
        LOG_INFO("User logged in successfully");
    });

    connect(m_authService.get(), &AuthService::loginFailed, this, [this](const QString &error) {
        LOG_ERROR("Login failed: " + error);
    });

    LOG_INFO("Services initialized");
}

void Application::registerQmlTypes()
{
    // Register custom types for QML
    qmlRegisterType<LoginController>("MeetingSystem", 1, 0, "LoginController");
    qmlRegisterType<MainWindowController>("MeetingSystem", 1, 0, "MainWindowController");
    qmlRegisterType<MeetingRoomController>("MeetingSystem", 1, 0, "MeetingRoomController");
    qmlRegisterType<AIPanelController>("MeetingSystem", 1, 0, "AIPanelController");
    qmlRegisterType<VideoEffectsController>("MeetingSystem", 1, 0, "VideoEffectsController");

    LOG_INFO("QML types registered");
}

void Application::setupQmlContext()
{
    QQmlContext *context = m_engine->rootContext();

    // Expose services to QML
    context->setContextProperty("authService", m_authService.get());
    context->setContextProperty("meetingService", m_meetingService.get());
    context->setContextProperty("mediaService", m_mediaService.get());
    context->setContextProperty("aiService", m_aiService.get());
    context->setContextProperty("webrtcManager", m_webrtcManager.get());
    context->setContextProperty("videoEffectsController", m_videoEffectsController.get());
    context->setContextProperty("config", m_config.get());

    LOG_INFO("QML context setup complete");
}

