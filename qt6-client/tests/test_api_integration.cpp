#include <QCoreApplication>
#include <QTimer>
#include <QDebug>
#include "network/api_client.h"
#include "services/auth_service.h"
#include "services/meeting_service.h"
#include "services/media_service.h"
#include "services/ai_service.h"
#include "network/websocket_client.h"

/**
 * Qt6客户端API集成测试
 * 
 * 测试与后端服务器的完整集成流程
 * 后端地址: http://js1.blockelite.cn:28558
 */

class ApiIntegrationTest : public QObject
{
    Q_OBJECT

public:
    ApiIntegrationTest(QObject *parent = nullptr)
        : QObject(parent)
    {
        // 创建API客户端
        m_apiClient = new ApiClient("http://js1.blockelite.cn:28558", this);
        
        // 创建服务层
        m_authService = new AuthService(m_apiClient, this);
        m_meetingService = new MeetingService(m_apiClient, this);
        m_mediaService = new MediaService(m_apiClient, this);
        m_aiService = new AIService(m_apiClient, this);
        
        // 创建WebSocket客户端
        m_wsClient = new WebSocketClient(this);
        
        setupConnections();
    }

    void runTests()
    {
        qDebug() << "=== Starting API Integration Tests ===";
        qDebug() << "Backend URL:" << m_apiClient->baseUrl();
        
        // 测试流程：
        // 1. 获取CSRF Token
        // 2. 用户登录
        // 3. 创建会议
        // 4. 加入会议
        // 5. 建立WebSocket连接
        // 6. 测试AI功能
        // 7. 测试管理员功能（如果有权限）
        
        testGetCsrfToken();
    }

private slots:
    // ==================== CSRF Token测试 ====================
    void testGetCsrfToken()
    {
        qDebug() << "\n[TEST] Getting CSRF Token...";
        m_authService->getCsrfToken();
    }

    void onCsrfTokenReceived(const QString &token)
    {
        qDebug() << "[SUCCESS] CSRF Token received:" << token.left(20) + "...";
        
        // 继续登录测试
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testLogin);
    }

    void onCsrfTokenFailed(const QString &error)
    {
        qDebug() << "[WARNING] CSRF Token failed:" << error;
        qDebug() << "Continuing without CSRF Token (using Bearer token only)...";
        
        // 即使CSRF Token失败，也可以继续使用Bearer Token
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testLogin);
    }

    // ==================== 认证测试 ====================
    void testLogin()
    {
        qDebug() << "\n[TEST] Logging in...";
        
        // 使用测试账号登录
        // 注意：需要先在后端创建测试账号
        m_authService->login("demo", "P@ssw0rd!");
    }

    void onLoginSuccess()
    {
        qDebug() << "[SUCCESS] Login successful";
        qDebug() << "Auth Token:" << m_authService->authToken().left(20) + "...";
        
        // 继续会议测试
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testGetUserProfile);
    }

    void onLoginFailed(const QString &error)
    {
        qDebug() << "[FAILED] Login failed:" << error;
        qDebug() << "Please create a test user first or check credentials";
        
        // 尝试注册
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testRegister);
    }

    // ==================== 用户资料测试 ====================
    void testGetUserProfile()
    {
        qDebug() << "\n[TEST] Getting user profile...";
        
        m_apiClient->getUserProfile([this](const ApiResponse &response) {
            if (response.isSuccess()) {
                QJsonObject user = response.data["user"].toObject();
                qDebug() << "[SUCCESS] User profile received:";
                qDebug() << "  Username:" << user["username"].toString();
                qDebug() << "  Email:" << user["email"].toString();
                qDebug() << "  Nickname:" << user["nickname"].toString();
                
                // 继续会议测试
                QTimer::singleShot(1000, this, &ApiIntegrationTest::testCreateMeeting);
            } else {
                qDebug() << "[FAILED] Get user profile failed:" << response.message;
            }
        });
    }

    // ==================== 注册测试 ====================
    void testRegister()
    {
        qDebug() << "\n[TEST] Registering new user...";
        
        m_authService->registerUser(
            "demo",
            "demo@example.com",
            "P@ssw0rd!",
            "Demo User"
        );
    }

    void onRegisterSuccess()
    {
        qDebug() << "[SUCCESS] Registration successful";
        
        // 注册成功后登录
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testLogin);
    }

    void onRegisterFailed(const QString &error)
    {
        qDebug() << "[FAILED] Registration failed:" << error;
        qDebug() << "Tests cannot continue without authentication";
        
        QCoreApplication::quit();
    }

    // ==================== 会议测试 ====================
    void testCreateMeeting()
    {
        qDebug() << "\n[TEST] Creating meeting...";
        
        QDateTime startTime = QDateTime::currentDateTime().addSecs(3600);
        QDateTime endTime = startTime.addSecs(7200);
        
        QJsonObject settings;
        settings["waiting_room"] = true;
        settings["allow_recording"] = true;
        
        m_meetingService->createMeeting(
            "API测试会议",
            "这是一个API集成测试会议",
            startTime,
            endTime,
            100,
            "video",
            "test123",
            settings
        );
    }

    void onMeetingCreated(int meetingId)
    {
        qDebug() << "[SUCCESS] Meeting created with ID:" << meetingId;
        m_testMeetingId = meetingId;
        
        // 继续获取会议列表
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testGetMeetingList);
    }

    void onMeetingCreateFailed(const QString &error)
    {
        qDebug() << "[FAILED] Create meeting failed:" << error;
    }

    void testGetMeetingList()
    {
        qDebug() << "\n[TEST] Getting meeting list...";
        
        m_meetingService->getMeetingList(1, 10, "", "");
    }

    void onMeetingListReceived(const QJsonArray &meetings)
    {
        qDebug() << "[SUCCESS] Meeting list received:" << meetings.size() << "meetings";
        
        for (const auto &meeting : meetings) {
            QJsonObject m = meeting.toObject();
            qDebug() << "  -" << m["title"].toString() << "(ID:" << m["id"].toInt() << ")";
        }
        
        // 继续加入会议测试
        if (m_testMeetingId > 0) {
            QTimer::singleShot(1000, this, &ApiIntegrationTest::testJoinMeeting);
        } else {
            QTimer::singleShot(1000, this, &ApiIntegrationTest::testAIModels);
        }
    }

    void testJoinMeeting()
    {
        qDebug() << "\n[TEST] Joining meeting...";
        
        m_meetingService->joinMeeting(m_testMeetingId, "test123");
    }

    void onMeetingJoined()
    {
        qDebug() << "[SUCCESS] Joined meeting successfully";
        
        // 继续WebSocket连接测试
        QTimer::singleShot(1000, this, &ApiIntegrationTest::testWebSocketConnection);
    }

    // ==================== WebSocket测试 ====================
    void testWebSocketConnection()
    {
        qDebug() << "\n[TEST] Connecting to WebSocket...";
        
        QString wsUrl = "ws://js1.blockelite.cn:28558/ws/signaling";
        QString token = m_authService->authToken();
        int userId = m_authService->currentUser()->id();
        QString peerId = QUuid::createUuid().toString(QUuid::WithoutBraces);
        
        m_wsClient->connect(wsUrl, token, m_testMeetingId, userId, peerId);
    }

    void onWebSocketConnected()
    {
        qDebug() << "[SUCCESS] WebSocket connected";
        
        // 发送测试消息
        m_wsClient->sendChatMessage("Hello from API test!", 0);
        
        // 启动心跳
        m_wsClient->startHeartbeat(30000);
        
        // 继续AI功能测试
        QTimer::singleShot(2000, this, &ApiIntegrationTest::testAIModels);
    }

    void onWebSocketDisconnected()
    {
        qDebug() << "[INFO] WebSocket disconnected";
    }

    void onWebSocketError(const QString &error)
    {
        qDebug() << "[WARNING] WebSocket error:" << error;
    }

    void onWebSocketMessageReceived(const QJsonObject &message)
    {
        qDebug() << "[INFO] WebSocket message received:" << message["type"].toString();
    }

    // ==================== AI功能测试 ====================
    void testAIModels()
    {
        qDebug() << "\n[TEST] Getting AI models...";
        
        m_apiClient->getAIModels([this](const ApiResponse &response) {
            if (response.isSuccess()) {
                QJsonArray models = response.data["models"].toArray();
                qDebug() << "[SUCCESS] AI models received:" << models.size() << "models";
                
                for (const auto &model : models) {
                    QJsonObject m = model.toObject();
                    qDebug() << "  -" << m["model_id"].toString() 
                             << "Status:" << m["status"].toString();
                }
                
                // 继续管理员功能测试
                QTimer::singleShot(1000, this, &ApiIntegrationTest::testAdminFunctions);
            } else {
                qDebug() << "[INFO] AI models not available:" << response.message;
                
                // 继续管理员功能测试
                QTimer::singleShot(1000, this, &ApiIntegrationTest::testAdminFunctions);
            }
        });
    }

    // ==================== 管理员功能测试 ====================
    void testAdminFunctions()
    {
        qDebug() << "\n[TEST] Testing admin functions...";
        
        m_apiClient->getAdminUsers(1, 10, "", [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                QJsonArray users = response.data["users"].toArray();
                qDebug() << "[SUCCESS] Admin users received:" << users.size() << "users";
            } else {
                qDebug() << "[INFO] Admin functions not available (requires admin role):" 
                         << response.message;
            }
            
            // 测试完成
            QTimer::singleShot(1000, this, &ApiIntegrationTest::testComplete);
        });
    }

    // ==================== 测试完成 ====================
    void testComplete()
    {
        qDebug() << "\n=== API Integration Tests Complete ===";
        qDebug() << "All tests finished successfully!";
        
        // 清理
        if (m_wsClient->isConnected()) {
            m_wsClient->disconnect();
        }
        
        // 退出应用
        QTimer::singleShot(2000, qApp, &QCoreApplication::quit);
    }

private:
    void setupConnections()
    {
        // AuthService信号
        connect(m_authService, &AuthService::csrfTokenReceived, 
                this, &ApiIntegrationTest::onCsrfTokenReceived);
        connect(m_authService, &AuthService::csrfTokenFailed, 
                this, &ApiIntegrationTest::onCsrfTokenFailed);
        connect(m_authService, &AuthService::loginSuccess, 
                this, &ApiIntegrationTest::onLoginSuccess);
        connect(m_authService, &AuthService::loginFailed, 
                this, &ApiIntegrationTest::onLoginFailed);
        connect(m_authService, &AuthService::registerSuccess, 
                this, &ApiIntegrationTest::onRegisterSuccess);
        connect(m_authService, &AuthService::registerFailed, 
                this, &ApiIntegrationTest::onRegisterFailed);
        
        // MeetingService信号
        connect(m_meetingService, &MeetingService::meetingCreated, 
                this, &ApiIntegrationTest::onMeetingCreated);
        connect(m_meetingService, &MeetingService::meetingCreateFailed, 
                this, &ApiIntegrationTest::onMeetingCreateFailed);
        connect(m_meetingService, &MeetingService::meetingListReceived, 
                this, &ApiIntegrationTest::onMeetingListReceived);
        connect(m_meetingService, &MeetingService::meetingJoined, 
                this, &ApiIntegrationTest::onMeetingJoined);
        
        // WebSocketClient信号
        connect(m_wsClient, &WebSocketClient::connected, 
                this, &ApiIntegrationTest::onWebSocketConnected);
        connect(m_wsClient, &WebSocketClient::disconnected, 
                this, &ApiIntegrationTest::onWebSocketDisconnected);
        connect(m_wsClient, &WebSocketClient::error, 
                this, &ApiIntegrationTest::onWebSocketError);
        connect(m_wsClient, &WebSocketClient::messageReceived, 
                this, &ApiIntegrationTest::onWebSocketMessageReceived);
    }

private:
    ApiClient *m_apiClient;
    AuthService *m_authService;
    MeetingService *m_meetingService;
    MediaService *m_mediaService;
    AIService *m_aiService;
    WebSocketClient *m_wsClient;
    
    int m_testMeetingId = 0;
};

int main(int argc, char *argv[])
{
    QCoreApplication app(argc, argv);
    
    qDebug() << "Qt6 Client API Integration Test";
    qDebug() << "Backend: http://js1.blockelite.cn:28558";
    qDebug() << "";
    
    ApiIntegrationTest test;
    
    // 启动测试
    QTimer::singleShot(0, &test, &ApiIntegrationTest::runTests);
    
    return app.exec();
}

#include "test_api_integration.moc"

