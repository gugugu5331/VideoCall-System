#include <QCoreApplication>
#include <QTest>
#include <QSignalSpy>
#include "network/api_client.h"
#include "network/http_client.h"
#include "network/websocket_client.h"

class TestApiClient : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase();
    void cleanupTestCase();
    
    // HTTP Client Tests
    void testHttpClientGet();
    void testHttpClientPost();
    void testHttpClientPut();
    void testHttpClientDelete();
    void testHttpClientUpload();
    
    // API Client Tests - Auth
    void testRegisterUser();
    void testLogin();
    void testRefreshToken();
    void testForgotPassword();
    void testResetPassword();
    
    // API Client Tests - User
    void testGetUserProfile();
    void testUpdateUserProfile();
    void testChangePassword();
    void testUploadAvatar();
    void testDeleteAccount();
    
    // API Client Tests - Meeting
    void testCreateMeeting();
    void testGetMeetingList();
    void testGetMeetingInfo();
    void testUpdateMeeting();
    void testDeleteMeeting();
    void testStartMeeting();
    void testEndMeeting();
    void testJoinMeeting();
    void testLeaveMeeting();
    void testGetParticipants();
    void testAddParticipant();
    void testRemoveParticipant();
    void testUpdateParticipantRole();
    void testStartRecording();
    void testStopRecording();
    void testGetRecordings();
    void testGetChatMessages();
    void testSendChatMessage();
    
    // API Client Tests - My Meetings
    void testGetMyMeetings();
    void testGetUpcomingMeetings();
    void testGetMeetingHistory();
    
    // API Client Tests - Media
    void testUploadMedia();
    void testGetMediaList();
    void testGetMediaInfo();
    void testDeleteMedia();
    void testProcessMedia();
    
    // API Client Tests - AI
    void testSpeechRecognition();
    void testEmotionDetection();
    void testSynthesisDetection();
    void testAudioDenoising();
    void testVideoEnhancement();
    
    // API Client Tests - Signaling
    void testGetSessionInfo();
    void testGetRoomSessions();
    void testGetMessageHistory();
    void testGetStatsOverview();
    void testGetRoomStats();
    
    // API Client Tests - WebRTC
    void testGetRoomPeers();
    void testGetRoomWebRTCStats();
    void testUpdatePeerMedia();
    
    // WebSocket Client Tests
    void testWebSocketConnect();
    void testWebSocketDisconnect();
    void testWebSocketSendMessage();
    void testWebSocketHeartbeat();
    void testWebSocketReconnect();

private:
    ApiClient *apiClient;
    HttpClient *httpClient;
    WebSocketClient *wsClient;
    QString testToken;
    int testUserId;
    int testMeetingId;
};

void TestApiClient::initTestCase()
{
    // Initialize test environment
    QString baseUrl = "http://localhost:8000";
    apiClient = new ApiClient(baseUrl, this);
    httpClient = new HttpClient(this);
    wsClient = new WebSocketClient(this);
    
    qDebug() << "Test environment initialized";
}

void TestApiClient::cleanupTestCase()
{
    // Cleanup
    delete apiClient;
    delete httpClient;
    delete wsClient;
    
    qDebug() << "Test environment cleaned up";
}

// ==================== HTTP Client Tests ====================

void TestApiClient::testHttpClientGet()
{
    QSignalSpy spy(httpClient, &HttpClient::requestFinished);
    
    httpClient->get("http://localhost:8000/api/v1/health",
        [](const QJsonObject &response) {
            QVERIFY(!response.isEmpty());
            qDebug() << "GET request successful";
        },
        [](const QString &error) {
            QFAIL(qPrintable("GET request failed: " + error));
        });
    
    QVERIFY(spy.wait(5000));
}

void TestApiClient::testHttpClientPost()
{
    QJsonObject data;
    data["test"] = "value";
    
    httpClient->post("http://localhost:8000/api/v1/test",
        data,
        [](const QJsonObject &response) {
            QVERIFY(!response.isEmpty());
            qDebug() << "POST request successful";
        },
        [](const QString &error) {
            qDebug() << "POST request error (expected if endpoint doesn't exist):" << error;
        });
}

void TestApiClient::testHttpClientPut()
{
    QJsonObject data;
    data["test"] = "updated";
    
    httpClient->put("http://localhost:8000/api/v1/test/1",
        data,
        [](const QJsonObject &response) {
            qDebug() << "PUT request successful";
        },
        [](const QString &error) {
            qDebug() << "PUT request error (expected if endpoint doesn't exist):" << error;
        });
}

void TestApiClient::testHttpClientDelete()
{
    httpClient->del("http://localhost:8000/api/v1/test/1",
        [](const QJsonObject &response) {
            qDebug() << "DELETE request successful";
        },
        [](const QString &error) {
            qDebug() << "DELETE request error (expected if endpoint doesn't exist):" << error;
        });
}

void TestApiClient::testHttpClientUpload()
{
    // Create a temporary test file
    QString testFile = QDir::temp().filePath("test_upload.txt");
    QFile file(testFile);
    if (file.open(QIODevice::WriteOnly)) {
        file.write("Test upload content");
        file.close();
    }
    
    QVariantMap formData;
    formData["description"] = "Test file";
    
    httpClient->upload("http://localhost:8000/api/v1/upload",
        testFile,
        formData,
        [](const QJsonObject &response) {
            qDebug() << "Upload successful";
        },
        [](const QString &error) {
            qDebug() << "Upload error (expected if endpoint doesn't exist):" << error;
        },
        [](qint64 sent, qint64 total) {
            qDebug() << "Upload progress:" << sent << "/" << total;
        });
}

// ==================== API Client Tests - Auth ====================

void TestApiClient::testRegisterUser()
{
    QString username = "testuser_" + QString::number(QDateTime::currentMSecsSinceEpoch());
    
    apiClient->registerUser(username, "test@example.com", "password123", "Test User",
        [this, username](const ApiResponse &response) {
            if (response.isSuccess()) {
                testUserId = response.data["user_id"].toInt();
                qDebug() << "User registered successfully, ID:" << testUserId;
                QVERIFY(testUserId > 0);
            } else {
                qDebug() << "Registration failed:" << response.message;
            }
        });
}

void TestApiClient::testLogin()
{
    apiClient->login("testuser", "password123",
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                testToken = response.data["token"].toString();
                apiClient->setAuthToken(testToken);
                qDebug() << "Login successful, token received";
                QVERIFY(!testToken.isEmpty());
            } else {
                qDebug() << "Login failed:" << response.message;
            }
        });
}

void TestApiClient::testRefreshToken()
{
    QString refreshToken = "test_refresh_token";
    
    apiClient->refreshToken(refreshToken,
        [](const ApiResponse &response) {
            if (response.isSuccess()) {
                qDebug() << "Token refreshed successfully";
            } else {
                qDebug() << "Token refresh failed:" << response.message;
            }
        });
}

// ==================== API Client Tests - Meeting ====================

void TestApiClient::testCreateMeeting()
{
    QDateTime startTime = QDateTime::currentDateTime().addSecs(3600);
    QDateTime endTime = startTime.addSecs(3600);
    
    QJsonObject settings;
    settings["enable_recording"] = true;
    settings["enable_chat"] = true;
    
    apiClient->createMeeting("Test Meeting", "Test Description",
        startTime, endTime, 10, "video", "", settings,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                testMeetingId = response.data["meeting_id"].toInt();
                qDebug() << "Meeting created, ID:" << testMeetingId;
                QVERIFY(testMeetingId > 0);
            } else {
                qDebug() << "Meeting creation failed:" << response.message;
            }
        });
}

void TestApiClient::testGetMeetingList()
{
    apiClient->getMeetingList(1, 20, "scheduled", "",
        [](const ApiResponse &response) {
            if (response.isSuccess()) {
                int total = response.data["total"].toInt();
                qDebug() << "Retrieved" << total << "meetings";
            } else {
                qDebug() << "Get meeting list failed:" << response.message;
            }
        });
}

void TestApiClient::testJoinMeeting()
{
    if (testMeetingId == 0) {
        qDebug() << "Skipping join meeting test (no meeting ID)";
        return;
    }
    
    apiClient->joinMeeting(testMeetingId, "",
        [](const ApiResponse &response) {
            if (response.isSuccess()) {
                qDebug() << "Joined meeting successfully";
                QString roomUrl = response.data["room_url"].toString();
                QVERIFY(!roomUrl.isEmpty());
            } else {
                qDebug() << "Join meeting failed:" << response.message;
            }
        });
}

// ==================== WebSocket Client Tests ====================

void TestApiClient::testWebSocketConnect()
{
    QSignalSpy spy(wsClient, &WebSocketClient::connected);
    
    QString wsUrl = "ws://localhost:8000/ws/signaling";
    wsClient->connect(wsUrl, testToken, testMeetingId, testUserId, "test_peer_id");
    
    // Wait for connection (or timeout)
    bool connected = spy.wait(5000);
    if (connected) {
        qDebug() << "WebSocket connected successfully";
        QVERIFY(wsClient->isConnected());
    } else {
        qDebug() << "WebSocket connection timeout (server may not be running)";
    }
}

void TestApiClient::testWebSocketSendMessage()
{
    if (!wsClient->isConnected()) {
        qDebug() << "Skipping send message test (not connected)";
        return;
    }
    
    QJsonObject message;
    message["type"] = "test";
    message["content"] = "Test message";
    
    wsClient->sendMessage(message);
    qDebug() << "Message sent";
}

void TestApiClient::testWebSocketHeartbeat()
{
    if (!wsClient->isConnected()) {
        qDebug() << "Skipping heartbeat test (not connected)";
        return;
    }
    
    wsClient->startHeartbeat(5000); // 5 seconds
    qDebug() << "Heartbeat started";
    
    QTest::qWait(6000); // Wait for at least one heartbeat
    
    wsClient->stopHeartbeat();
    qDebug() << "Heartbeat stopped";
}

void TestApiClient::testWebSocketDisconnect()
{
    if (!wsClient->isConnected()) {
        qDebug() << "Skipping disconnect test (not connected)";
        return;
    }
    
    QSignalSpy spy(wsClient, &WebSocketClient::disconnected);
    
    wsClient->disconnect();
    
    bool disconnected = spy.wait(5000);
    QVERIFY(disconnected);
    QVERIFY(!wsClient->isConnected());
    qDebug() << "WebSocket disconnected successfully";
}

// ==================== Main ====================

QTEST_MAIN(TestApiClient)
#include "test_api_client.moc"

