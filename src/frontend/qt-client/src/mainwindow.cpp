#include "mainwindow.h"
#include "videowidget.h"
#include "webrtcmanager.h"
#include "signalingclient.h"
#include "detectionmanager.h"
#include "participantwidget.h"
#include "chatwidget.h"
#include "settingsdialog.h"

#include <QtWidgets/QApplication>
#include <QtWidgets/QInputDialog>
#include <QtCore/QStandardPaths>
#include <QtCore/QDir>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_centralWidget(nullptr)
    , m_mainSplitter(nullptr)
    , m_videoSplitter(nullptr)
    , m_videoArea(nullptr)
    , m_videoLayout(nullptr)
    , m_localVideoWidget(nullptr)
    , m_sidebarTabs(nullptr)
    , m_participantWidget(nullptr)
    , m_chatWidget(nullptr)
    , m_detectionWidget(nullptr)
    , m_detectionLog(nullptr)
    , m_controlToolBar(nullptr)
    , m_joinButton(nullptr)
    , m_leaveButton(nullptr)
    , m_cameraButton(nullptr)
    , m_microphoneButton(nullptr)
    , m_screenShareButton(nullptr)
    , m_connectionStatusLabel(nullptr)
    , m_participantCountLabel(nullptr)
    , m_detectionStatusLabel(nullptr)
    , m_webrtcManager(nullptr)
    , m_signalingClient(nullptr)
    , m_detectionManager(nullptr)
    , m_settingsDialog(nullptr)
    , m_isInMeeting(false)
    , m_isCameraEnabled(true)
    , m_isMicrophoneEnabled(true)
    , m_isScreenSharing(false)
    , m_settings(nullptr)
    , m_serverUrl("localhost")
    , m_serverPort(8080)
    , m_statusUpdateTimer(nullptr)
{
    setWindowTitle("视频会议系统 - Qt客户端");
    setMinimumSize(1200, 800);
    resize(1400, 900);
    
    // 初始化设置
    m_settings = new QSettings(this);
    loadSettings();
    
    // 设置UI
    setupUI();
    setupMenuBar();
    setupToolBar();
    setupStatusBar();
    
    // 初始化核心组件
    m_signalingClient = new SignalingClient(this);
    m_webrtcManager = new WebRTCManager(this);
    m_detectionManager = new DetectionManager(this);
    m_settingsDialog = new SettingsDialog(this);
    
    // 建立连接
    setupConnections();
    
    // 初始化状态
    updateControlButtons();
    updateConnectionStatus("未连接");
    updateParticipantCount(0);
    
    // 启动状态更新定时器
    m_statusUpdateTimer = new QTimer(this);
    connect(m_statusUpdateTimer, &QTimer::timeout, this, [this]() {
        // 定期更新状态
        if (m_detectionManager) {
            m_detectionStatusLabel->setText(
                QString("检测: %1").arg(m_detectionManager->isEnabled() ? "启用" : "禁用")
            );
        }
    });
    m_statusUpdateTimer->start(1000);
}

MainWindow::~MainWindow()
{
    saveSettings();
    
    if (m_isInMeeting) {
        leaveMeeting();
    }
}

void MainWindow::setupUI()
{
    // 创建中央部件
    m_centralWidget = new QWidget;
    setCentralWidget(m_centralWidget);
    
    // 创建主分割器
    m_mainSplitter = new QSplitter(Qt::Horizontal);
    
    // 创建视频区域
    m_videoArea = new QWidget;
    m_videoLayout = new QGridLayout(m_videoArea);
    m_videoLayout->setSpacing(5);
    m_videoLayout->setContentsMargins(5, 5, 5, 5);
    
    // 创建本地视频窗口
    m_localVideoWidget = new VideoWidget("本地视频", this);
    m_localVideoWidget->setMinimumSize(320, 240);
    m_videoLayout->addWidget(m_localVideoWidget, 0, 0);
    
    // 创建视频分割器
    m_videoSplitter = new QSplitter(Qt::Vertical);
    m_videoSplitter->addWidget(m_videoArea);
    
    // 创建侧边栏标签页
    m_sidebarTabs = new QTabWidget;
    m_sidebarTabs->setMaximumWidth(350);
    m_sidebarTabs->setMinimumWidth(300);
    
    // 参与者标签页
    m_participantWidget = new ParticipantWidget(this);
    m_sidebarTabs->addTab(m_participantWidget, "参与者");
    
    // 聊天标签页
    m_chatWidget = new ChatWidget(this);
    m_sidebarTabs->addTab(m_chatWidget, "聊天");
    
    // 检测标签页
    m_detectionWidget = new QWidget;
    QVBoxLayout *detectionLayout = new QVBoxLayout(m_detectionWidget);
    
    QLabel *detectionTitle = new QLabel("AI检测日志");
    detectionTitle->setStyleSheet("font-weight: bold; font-size: 14px;");
    detectionLayout->addWidget(detectionTitle);
    
    m_detectionLog = new QTextEdit;
    m_detectionLog->setReadOnly(true);
    m_detectionLog->setMaximumBlockCount(1000);
    detectionLayout->addWidget(m_detectionLog);
    
    m_sidebarTabs->addTab(m_detectionWidget, "检测");
    
    // 添加到主分割器
    m_mainSplitter->addWidget(m_videoSplitter);
    m_mainSplitter->addWidget(m_sidebarTabs);
    m_mainSplitter->setStretchFactor(0, 3);
    m_mainSplitter->setStretchFactor(1, 1);
    
    // 设置中央布局
    QVBoxLayout *centralLayout = new QVBoxLayout(m_centralWidget);
    centralLayout->setContentsMargins(0, 0, 0, 0);
    centralLayout->addWidget(m_mainSplitter);
}

void MainWindow::setupMenuBar()
{
    m_menuBar = menuBar();
    
    // 文件菜单
    m_fileMenu = m_menuBar->addMenu("文件(&F)");
    
    m_joinAction = new QAction("加入会议(&J)", this);
    m_joinAction->setShortcut(QKeySequence::New);
    m_joinAction->setStatusTip("加入视频会议");
    connect(m_joinAction, &QAction::triggered, this, &MainWindow::joinMeeting);
    m_fileMenu->addAction(m_joinAction);
    
    m_leaveAction = new QAction("离开会议(&L)", this);
    m_leaveAction->setShortcut(QKeySequence::Close);
    m_leaveAction->setStatusTip("离开当前会议");
    m_leaveAction->setEnabled(false);
    connect(m_leaveAction, &QAction::triggered, this, &MainWindow::leaveMeeting);
    m_fileMenu->addAction(m_leaveAction);
    
    m_fileMenu->addSeparator();
    
    m_settingsAction = new QAction("设置(&S)", this);
    m_settingsAction->setShortcut(QKeySequence::Preferences);
    m_settingsAction->setStatusTip("打开设置对话框");
    connect(m_settingsAction, &QAction::triggered, this, &MainWindow::showSettings);
    m_fileMenu->addAction(m_settingsAction);
    
    m_fileMenu->addSeparator();
    
    m_exitAction = new QAction("退出(&X)", this);
    m_exitAction->setShortcut(QKeySequence::Quit);
    m_exitAction->setStatusTip("退出应用程序");
    connect(m_exitAction, &QAction::triggered, this, &QWidget::close);
    m_fileMenu->addAction(m_exitAction);
    
    // 帮助菜单
    m_helpMenu = m_menuBar->addMenu("帮助(&H)");
    
    m_aboutAction = new QAction("关于(&A)", this);
    m_aboutAction->setStatusTip("显示关于信息");
    connect(m_aboutAction, &QAction::triggered, this, &MainWindow::showAbout);
    m_helpMenu->addAction(m_aboutAction);
}

void MainWindow::setupToolBar()
{
    m_controlToolBar = addToolBar("控制");
    m_controlToolBar->setMovable(false);
    
    // 加入/离开按钮
    m_joinButton = new QPushButton("加入会议");
    m_joinButton->setStyleSheet("QPushButton { background-color: #4CAF50; color: white; padding: 8px 16px; border: none; border-radius: 4px; } QPushButton:hover { background-color: #45a049; }");
    connect(m_joinButton, &QPushButton::clicked, this, &MainWindow::joinMeeting);
    m_controlToolBar->addWidget(m_joinButton);
    
    m_leaveButton = new QPushButton("离开会议");
    m_leaveButton->setStyleSheet("QPushButton { background-color: #f44336; color: white; padding: 8px 16px; border: none; border-radius: 4px; } QPushButton:hover { background-color: #da190b; }");
    m_leaveButton->setEnabled(false);
    connect(m_leaveButton, &QPushButton::clicked, this, &MainWindow::leaveMeeting);
    m_controlToolBar->addWidget(m_leaveButton);
    
    m_controlToolBar->addSeparator();
    
    // 媒体控制按钮
    m_cameraButton = new QPushButton("📹 摄像头");
    m_cameraButton->setCheckable(true);
    m_cameraButton->setChecked(true);
    m_cameraButton->setEnabled(false);
    connect(m_cameraButton, &QPushButton::clicked, this, &MainWindow::toggleCamera);
    m_controlToolBar->addWidget(m_cameraButton);
    
    m_microphoneButton = new QPushButton("🎤 麦克风");
    m_microphoneButton->setCheckable(true);
    m_microphoneButton->setChecked(true);
    m_microphoneButton->setEnabled(false);
    connect(m_microphoneButton, &QPushButton::clicked, this, &MainWindow::toggleMicrophone);
    m_controlToolBar->addWidget(m_microphoneButton);
    
    m_screenShareButton = new QPushButton("🖥️ 屏幕共享");
    m_screenShareButton->setCheckable(true);
    m_screenShareButton->setEnabled(false);
    connect(m_screenShareButton, &QPushButton::clicked, this, &MainWindow::toggleScreenShare);
    m_controlToolBar->addWidget(m_screenShareButton);
}

void MainWindow::setupStatusBar()
{
    // 连接状态
    m_connectionStatusLabel = new QLabel("未连接");
    statusBar()->addWidget(m_connectionStatusLabel);
    
    statusBar()->addPermanentWidget(new QLabel("|"));
    
    // 参与者数量
    m_participantCountLabel = new QLabel("参与者: 0");
    statusBar()->addPermanentWidget(m_participantCountLabel);
    
    statusBar()->addPermanentWidget(new QLabel("|"));
    
    // 检测状态
    m_detectionStatusLabel = new QLabel("检测: 禁用");
    statusBar()->addPermanentWidget(m_detectionStatusLabel);
}

void MainWindow::setupConnections()
{
    // 信令客户端连接
    connect(m_signalingClient, &SignalingClient::connected, this, [this]() {
        updateConnectionStatus("已连接");
    });
    
    connect(m_signalingClient, &SignalingClient::disconnected, this, [this]() {
        updateConnectionStatus("连接断开");
    });
    
    connect(m_signalingClient, &SignalingClient::userJoined, this, &MainWindow::onUserJoined);
    connect(m_signalingClient, &SignalingClient::userLeft, this, &MainWindow::onUserLeft);
    
    // WebRTC管理器连接
    connect(m_webrtcManager, &WebRTCManager::localStreamReady, this, &MainWindow::onLocalStreamReady);
    connect(m_webrtcManager, &WebRTCManager::remoteStreamReceived, this, &MainWindow::onRemoteStreamReceived);
    
    // 检测管理器连接
    connect(m_detectionManager, &DetectionManager::detectionResult, this, &MainWindow::onDetectionResult);
    connect(m_detectionManager, &DetectionManager::detectionAlert, this, &MainWindow::onDetectionAlert);
    
    // 聊天连接
    connect(m_chatWidget, &ChatWidget::messageSent, this, &MainWindow::sendChatMessage);
    connect(m_signalingClient, &SignalingClient::chatMessageReceived, this, &MainWindow::onChatMessageReceived);
}

void MainWindow::loadSettings()
{
    m_serverUrl = m_settings->value("server/url", "localhost").toString();
    m_serverPort = m_settings->value("server/port", 8080).toInt();
    
    // 恢复窗口几何
    restoreGeometry(m_settings->value("window/geometry").toByteArray());
    restoreState(m_settings->value("window/state").toByteArray());
}

void MainWindow::saveSettings()
{
    m_settings->setValue("server/url", m_serverUrl);
    m_settings->setValue("server/port", m_serverPort);
    
    // 保存窗口几何
    m_settings->setValue("window/geometry", saveGeometry());
    m_settings->setValue("window/state", saveState());
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    if (m_isInMeeting) {
        int ret = QMessageBox::question(this, "确认退出", 
                                       "您正在会议中，确定要退出吗？",
                                       QMessageBox::Yes | QMessageBox::No,
                                       QMessageBox::No);
        if (ret == QMessageBox::No) {
            event->ignore();
            return;
        }
        leaveMeeting();
    }
    
    saveSettings();
    event->accept();
}

// 会议控制槽函数
void MainWindow::joinMeeting()
{
    showJoinDialog();
}

void MainWindow::leaveMeeting()
{
    if (!m_isInMeeting) return;
    
    // 离开会议逻辑
    m_webrtcManager->leaveMeeting();
    m_signalingClient->leaveMeeting();
    m_detectionManager->stopDetection();
    
    // 清理远程视频
    for (auto it = m_remoteVideoWidgets.begin(); it != m_remoteVideoWidgets.end(); ++it) {
        m_videoLayout->removeWidget(it.value());
        it.value()->deleteLater();
    }
    m_remoteVideoWidgets.clear();
    
    // 更新状态
    m_isInMeeting = false;
    m_currentMeetingId.clear();
    updateControlButtons();
    updateConnectionStatus("未连接");
    updateParticipantCount(0);
    
    // 清理UI
    m_participantWidget->clearParticipants();
    m_chatWidget->clearMessages();
    m_detectionLog->clear();
}

void MainWindow::showJoinDialog()
{
    bool ok;
    QString userName = QInputDialog::getText(this, "加入会议", 
                                           "请输入您的姓名:", 
                                           QLineEdit::Normal, 
                                           m_settings->value("user/name", "用户").toString(), 
                                           &ok);
    if (!ok || userName.isEmpty()) return;
    
    QString meetingId = QInputDialog::getText(this, "加入会议", 
                                            "请输入会议ID:", 
                                            QLineEdit::Normal, 
                                            "demo-meeting", 
                                            &ok);
    if (!ok || meetingId.isEmpty()) return;
    
    // 保存用户名
    m_settings->setValue("user/name", userName);
    
    // 开始加入会议
    m_currentUserName = userName;
    m_currentMeetingId = meetingId;
    
    updateConnectionStatus("连接中...");
    
    // 连接信令服务器
    QString serverUrl = QString("ws://%1:%2/signaling").arg(m_serverUrl).arg(m_serverPort);
    m_signalingClient->connectToServer(serverUrl);
    
    // 初始化WebRTC
    m_webrtcManager->initialize();
    
    // 启动检测
    m_detectionManager->startDetection();
    
    // 加入会议
    m_signalingClient->joinMeeting(meetingId, userName);
    
    m_isInMeeting = true;
    updateControlButtons();
}

void MainWindow::toggleCamera()
{
    m_isCameraEnabled = !m_isCameraEnabled;
    m_webrtcManager->toggleCamera(m_isCameraEnabled);
    updateControlButtons();
}

void MainWindow::toggleMicrophone()
{
    m_isMicrophoneEnabled = !m_isMicrophoneEnabled;
    m_webrtcManager->toggleMicrophone(m_isMicrophoneEnabled);
    updateControlButtons();
}

void MainWindow::toggleScreenShare()
{
    m_isScreenSharing = !m_isScreenSharing;
    m_webrtcManager->toggleScreenShare(m_isScreenSharing);
    updateControlButtons();
}

void MainWindow::showSettings()
{
    if (m_settingsDialog->exec() == QDialog::Accepted) {
        // 应用新设置
        loadSettings();
    }
}

void MainWindow::showAbout()
{
    QMessageBox::about(this, "关于",
                      "视频会议系统 Qt客户端\n\n"
                      "版本: 1.0.0\n"
                      "基于Qt6和WebRTC技术\n"
                      "支持多人视频会议和AI检测功能");
}

void MainWindow::updateConnectionStatus(const QString &status)
{
    m_connectionStatusLabel->setText(QString("状态: %1").arg(status));
}

void MainWindow::updateParticipantCount(int count)
{
    m_participantCountLabel->setText(QString("参与者: %1").arg(count));
}

void MainWindow::updateControlButtons()
{
    // 更新按钮状态
    m_joinButton->setEnabled(!m_isInMeeting);
    m_leaveButton->setEnabled(m_isInMeeting);
    m_cameraButton->setEnabled(m_isInMeeting);
    m_microphoneButton->setEnabled(m_isInMeeting);
    m_screenShareButton->setEnabled(m_isInMeeting);

    m_joinAction->setEnabled(!m_isInMeeting);
    m_leaveAction->setEnabled(m_isInMeeting);

    // 更新按钮文本和样式
    if (m_isInMeeting) {
        m_cameraButton->setText(m_isCameraEnabled ? "📹 摄像头" : "📹 摄像头(关)");
        m_cameraButton->setStyleSheet(m_isCameraEnabled ?
            "QPushButton { background-color: #4CAF50; color: white; }" :
            "QPushButton { background-color: #f44336; color: white; }");

        m_microphoneButton->setText(m_isMicrophoneEnabled ? "🎤 麦克风" : "🎤 麦克风(关)");
        m_microphoneButton->setStyleSheet(m_isMicrophoneEnabled ?
            "QPushButton { background-color: #4CAF50; color: white; }" :
            "QPushButton { background-color: #f44336; color: white; }");

        m_screenShareButton->setText(m_isScreenSharing ? "🖥️ 停止共享" : "🖥️ 屏幕共享");
        m_screenShareButton->setStyleSheet(m_isScreenSharing ?
            "QPushButton { background-color: #ff9800; color: white; }" :
            "QPushButton { background-color: #2196F3; color: white; }");
    }
}

// WebRTC事件处理
void MainWindow::onUserJoined(const QString &userId, const QString &userName)
{
    m_participantWidget->addParticipant(userId, userName);
    updateParticipantCount(m_participantWidget->getParticipantCount());

    m_detectionLog->append(QString("[%1] 用户加入: %2")
                          .arg(QTime::currentTime().toString())
                          .arg(userName));
}

void MainWindow::onUserLeft(const QString &userId)
{
    QString userName = m_participantWidget->getParticipantName(userId);
    m_participantWidget->removeParticipant(userId);
    updateParticipantCount(m_participantWidget->getParticipantCount());

    // 移除远程视频
    if (m_remoteVideoWidgets.contains(userId)) {
        m_videoLayout->removeWidget(m_remoteVideoWidgets[userId]);
        m_remoteVideoWidgets[userId]->deleteLater();
        m_remoteVideoWidgets.remove(userId);
    }

    m_detectionLog->append(QString("[%1] 用户离开: %2")
                          .arg(QTime::currentTime().toString())
                          .arg(userName));
}

void MainWindow::onLocalStreamReady()
{
    m_localVideoWidget->setVideoSource(m_webrtcManager->getLocalVideoSource());
    m_detectionLog->append(QString("[%1] 本地视频流就绪")
                          .arg(QTime::currentTime().toString()));
}

void MainWindow::onRemoteStreamReceived(const QString &userId, QObject *stream)
{
    // 创建远程视频窗口
    if (!m_remoteVideoWidgets.contains(userId)) {
        QString userName = m_participantWidget->getParticipantName(userId);
        VideoWidget *remoteWidget = new VideoWidget(userName, this);
        remoteWidget->setMinimumSize(320, 240);

        // 计算网格位置
        int count = m_remoteVideoWidgets.size() + 1; // +1 for local video
        int cols = qCeil(qSqrt(count + 1));
        int row = count / cols;
        int col = count % cols;

        m_videoLayout->addWidget(remoteWidget, row, col);
        m_remoteVideoWidgets[userId] = remoteWidget;
    }

    // 设置视频源
    m_remoteVideoWidgets[userId]->setVideoSource(stream);

    m_detectionLog->append(QString("[%1] 收到远程视频流: %2")
                          .arg(QTime::currentTime().toString())
                          .arg(m_participantWidget->getParticipantName(userId)));
}

// 检测事件处理
void MainWindow::onDetectionResult(const QString &type, bool isFake, double confidence)
{
    QString status = isFake ? "可疑" : "正常";
    QString message = QString("[%1] %2检测: %3 (置信度: %4%)")
                     .arg(QTime::currentTime().toString())
                     .arg(type)
                     .arg(status)
                     .arg(qRound(confidence * 100));

    m_detectionLog->append(message);

    if (isFake && confidence > 0.7) {
        m_detectionLog->append(QString("<font color='red'>⚠️ 高风险检测结果！</font>"));
    }
}

void MainWindow::onDetectionAlert(const QString &message)
{
    m_detectionLog->append(QString("<font color='red'>[%1] 告警: %2</font>")
                          .arg(QTime::currentTime().toString())
                          .arg(message));

    // 显示系统通知
    QMessageBox::warning(this, "检测告警", message);
}

// 聊天事件处理
void MainWindow::onChatMessageReceived(const QString &sender, const QString &message)
{
    m_chatWidget->addMessage(sender, message, false);
}

void MainWindow::sendChatMessage()
{
    QString message = m_chatWidget->getCurrentMessage();
    if (!message.isEmpty()) {
        m_signalingClient->sendChatMessage(message);
        m_chatWidget->addMessage(m_currentUserName, message, true);
        m_chatWidget->clearCurrentMessage();
    }
}
