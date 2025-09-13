#include "application.h"
#include "services/auth_service.h"
#include "services/api_service.h"
#include "services/meeting_service.h"
#include "services/detection_service.h"
#include "services/webrtc_service.h"
#include "services/media_service.h"
#include "controllers/main_controller.h"
#include "utils/settings_manager.h"
#include "utils/logger.h"
#include "models/user_model.h"
#include "models/meeting_model.h"
#include "models/detection_model.h"

#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <QQuickStyle>
#include <QDir>
#include <QStandardPaths>
#include <QLoggingCategory>

Application* Application::s_instance = nullptr;

Application::Application(int &argc, char **argv)
    : QApplication(argc, argv)
    , m_engine(std::make_unique<QQmlApplicationEngine>())
{
    s_instance = this;
    setupApplicationProperties();
}

Application::~Application()
{
    s_instance = nullptr;
}

bool Application::initialize()
{
    try {
        // 初始化日志系统
        m_logger = std::make_unique<Logger>();
        m_logger->initialize();

        // 初始化设置管理器
        m_settingsManager = std::make_unique<SettingsManager>();

        // 加载翻译
        loadTranslations();

        // 应用样式
        applyStyleSheet();

        // 初始化服务
        initializeServices();

        // 注册QML类型
        registerQmlTypes();

        // 初始化UI
        initializeUI();

        // 连接应用程序状态变化信号
        connect(this, &QApplication::applicationStateChanged,
                this, &Application::onApplicationStateChanged);

        m_logger->info("Application initialized successfully");
        return true;
    }
    catch (const std::exception& e) {
        if (m_logger) {
            m_logger->error(QString("Failed to initialize application: %1").arg(e.what()));
        }
        return false;
    }
}

int Application::run()
{
    if (!initialize()) {
        return -1;
    }

    return exec();
}

Application* Application::instance()
{
    return s_instance;
}

void Application::quit()
{
    if (m_logger) {
        m_logger->info("Application quitting...");
    }
    
    // 清理资源
    if (m_webrtcService) {
        m_webrtcService->cleanup();
    }
    
    if (m_mediaService) {
        m_mediaService->cleanup();
    }

    QApplication::quit();
}

void Application::restart()
{
    if (m_logger) {
        m_logger->info("Application restarting...");
    }
    
    quit();
    
    // 重启应用程序
    QProcess::startDetached(applicationFilePath(), arguments());
}

void Application::onApplicationStateChanged(Qt::ApplicationState state)
{
    switch (state) {
    case Qt::ApplicationActive:
        if (m_logger) m_logger->debug("Application became active");
        break;
    case Qt::ApplicationInactive:
        if (m_logger) m_logger->debug("Application became inactive");
        break;
    case Qt::ApplicationSuspended:
        if (m_logger) m_logger->debug("Application suspended");
        // 暂停媒体服务
        if (m_mediaService) {
            m_mediaService->pause();
        }
        break;
    case Qt::ApplicationHidden:
        if (m_logger) m_logger->debug("Application hidden");
        break;
    }
}

void Application::initializeServices()
{
    // 创建API服务
    m_apiService = std::make_unique<ApiService>();
    m_apiService->setBaseUrl(m_settingsManager->value("api/base_url", "http://localhost:8080").toString());

    // 创建认证服务
    m_authService = std::make_unique<AuthService>(m_apiService.get());

    // 创建会议服务
    m_meetingService = std::make_unique<MeetingService>(m_apiService.get());

    // 创建检测服务
    m_detectionService = std::make_unique<DetectionService>(m_apiService.get());

    // 创建WebRTC服务
    m_webrtcService = std::make_unique<WebRTCService>();

    // 创建媒体服务
    m_mediaService = std::make_unique<MediaService>();

    // 创建主控制器
    m_mainController = std::make_unique<MainController>(
        m_authService.get(),
        m_meetingService.get(),
        m_detectionService.get(),
        m_webrtcService.get(),
        m_mediaService.get()
    );
}

void Application::initializeUI()
{
    // 设置QML上下文属性
    QQmlContext* context = m_engine->rootContext();
    context->setContextProperty("app", this);
    context->setContextProperty("mainController", m_mainController.get());
    context->setContextProperty("authService", m_authService.get());
    context->setContextProperty("meetingService", m_meetingService.get());
    context->setContextProperty("detectionService", m_detectionService.get());
    context->setContextProperty("settingsManager", m_settingsManager.get());

    // 加载主QML文件
    const QUrl url(QStringLiteral("qrc:/qml/main.qml"));
    QObject::connect(m_engine.get(), &QQmlApplicationEngine::objectCreated,
                     this, [url](QObject *obj, const QUrl &objUrl) {
        if (!obj && url == objUrl) {
            QCoreApplication::exit(-1);
        }
    }, Qt::QueuedConnection);

    m_engine->load(url);
}

void Application::registerQmlTypes()
{
    // 注册C++模型到QML
    qmlRegisterType<UserModel>("VideoConference", 1, 0, "UserModel");
    qmlRegisterType<MeetingModel>("VideoConference", 1, 0, "MeetingModel");
    qmlRegisterType<DetectionModel>("VideoConference", 1, 0, "DetectionModel");

    // 注册枚举类型
    qmlRegisterUncreatableMetaObject(
        UserModel::staticMetaObject,
        "VideoConference",
        1, 0,
        "UserStatus",
        "Error: only enums"
    );
}

void Application::setupApplicationProperties()
{
    setApplicationName("Video Conference Client");
    setApplicationVersion("1.0.0");
    setOrganizationName("Video Conference System");
    setOrganizationDomain("videoconference.com");

    // 设置应用程序图标
    setWindowIcon(QIcon(":/icons/app_icon.png"));

    // 设置Qt Quick样式
    QQuickStyle::setStyle("Material");
}

void Application::loadTranslations()
{
    m_translator = std::make_unique<QTranslator>();
    
    QString locale = m_settingsManager->value("ui/language", "en").toString();
    QString translationFile = QString(":/translations/app_%1.qm").arg(locale);
    
    if (m_translator->load(translationFile)) {
        installTranslator(m_translator.get());
        if (m_logger) {
            m_logger->info(QString("Loaded translation: %1").arg(locale));
        }
    }
}

void Application::applyStyleSheet()
{
    QString theme = m_settingsManager->value("ui/theme", "dark").toString();
    QString styleFile = QString(":/styles/%1.qss").arg(theme);
    
    QFile file(styleFile);
    if (file.open(QFile::ReadOnly)) {
        QString styleSheet = QLatin1String(file.readAll());
        setStyleSheet(styleSheet);
        if (m_logger) {
            m_logger->info(QString("Applied theme: %1").arg(theme));
        }
    }
}
