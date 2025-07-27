#include "mainwindow.h"
#include "loginwidget.h"

#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QGridLayout>
#include <QLabel>
#include <QPushButton>
#include <QSlider>
#include <QComboBox>
#include <QCheckBox>
#include <QGroupBox>
#include <QFrame>
#include <QSplitter>
#include <QTabWidget>
#include <QListWidget>
#include <QTextEdit>
#include <QLineEdit>
#include <QProgressBar>
#include <QStatusBar>
#include <QToolBar>
#include <QMenuBar>
#include <QMenu>
#include <QAction>
#include <QSystemTrayIcon>
#include <QTimer>
#include <QSettings>
#include <QMessageBox>
#include <QApplication>
#include <QStyle>
#include <QIcon>
#include <QPixmap>
#include <QPalette>
#include <QColor>
#include <QFont>
#include <QFontMetrics>
#include <QElapsedTimer>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QBuffer>
#include <QByteArray>
#include <QDataStream>
#include <QThread>
#include <QMutex>
#include <QWaitCondition>
#include <QSemaphore>
#include <QFuture>
#include <QFutureWatcher>
#include <QtConcurrent>

// 常量定义
const QString MainWindow::SERVER_URL = "http://localhost:8000";
const QString MainWindow::APP_NAME = "VideoCall Pro";
const QString MainWindow::APP_VERSION = "1.0.0";

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_isLoggedIn(false)
    , m_isInCall(false)
    , m_isConnected(false)
    , m_callDuration(0)
    , m_networkQuality(100)
    , m_securityScore(0.0)
{
    // 初始化设置
    m_settings = new QSettings(this);
    
    // 初始化管理器
    initializeManagers();
    
    // 设置UI
    setupUI();
    setupStyles();
    
    // 加载设置
    loadSettings();
    
    // 检查系统要求
    checkSystemRequirements();
    
    // 显示登录界面
    showLogin();
}

MainWindow::~MainWindow()
{
    saveSettings();
    
    if (m_heartbeatTimer) {
        m_heartbeatTimer->stop();
    }
    if (m_statusUpdateTimer) {
        m_statusUpdateTimer->stop();
    }
    
    disconnectFromServer();
}

void MainWindow::setupUI()
{
    // 设置窗口属性
    setWindowTitle(APP_NAME + " " + APP_VERSION);
    setMinimumSize(1200, 800);
    resize(1400, 900);
    
    // 设置窗口图标
    setWindowIcon(QIcon(":/icons/app_icon.png"));
    
    // 设置菜单栏
    setupMenuBar();
    
    // 设置工具栏
    setupToolBar();
    
    // 设置状态栏
    setupStatusBar();
    
    // 设置系统托盘
    setupSystemTray();
    
    // 设置中央窗口
    setupCentralWidget();
    
    // 设置样式
    applyDarkTheme();
}

void MainWindow::setupMenuBar()
{
    m_menuBar = menuBar();
    
    // 文件菜单
    m_fileMenu = m_menuBar->addMenu("文件(&F)");
    QAction *newCallAction = m_fileMenu->addAction("新建通话(&N)");
    newCallAction->setShortcut(QKeySequence::New);
    connect(newCallAction, &QAction::triggered, this, &MainWindow::onStartCall);
    
    m_fileMenu->addSeparator();
    QAction *exitAction = m_fileMenu->addAction("退出(&X)");
    exitAction->setShortcut(QKeySequence::Quit);
    connect(exitAction, &QAction::triggered, this, &QWidget::close);
    
    // 通话菜单
    m_callMenu = m_menuBar->addMenu("通话(&C)");
    QAction *startCallAction = m_callMenu->addAction("开始通话(&S)");
    startCallAction->setShortcut(QKeySequence("Ctrl+S"));
    connect(startCallAction, &QAction::triggered, this, &MainWindow::onStartCall);
    
    QAction *endCallAction = m_callMenu->addAction("结束通话(&E)");
    endCallAction->setShortcut(QKeySequence("Ctrl+E"));
    connect(endCallAction, &QAction::triggered, this, &MainWindow::onEndCall);
    
    // 工具菜单
    m_toolsMenu = m_menuBar->addMenu("工具(&T)");
    QAction *settingsAction = m_toolsMenu->addAction("设置(&S)");
    connect(settingsAction, &QAction::triggered, [this]() {
        QMessageBox::information(this, "设置", "设置功能正在开发中...");
    });
    
    QAction *securityAction = m_toolsMenu->addAction("安全检测(&S)");
    connect(securityAction, &QAction::triggered, [this]() {
        QMessageBox::information(this, "安全检测", "安全检测功能正在开发中...");
    });
    
    // 帮助菜单
    m_helpMenu = m_menuBar->addMenu("帮助(&H)");
    QAction *aboutAction = m_helpMenu->addAction("关于(&A)");
    connect(aboutAction, &QAction::triggered, [this]() {
        QMessageBox::about(this, "关于", 
            QString("%1 %2\n\n"
                   "基于Qt6 C++开发的高质量音视频通话系统\n"
                   "支持实时音视频通话和安全检测功能\n\n"
                   "© 2025 VideoCall Team").arg(APP_NAME, APP_VERSION));
    });
}

void MainWindow::setupToolBar()
{
    m_toolBar = addToolBar("主工具栏");
    m_toolBar->setMovable(false);
    
    // 新建通话按钮
    m_newCallAction = m_toolBar->addAction(QIcon(":/icons/call.png"), "新建通话");
    m_newCallAction->setToolTip("开始新的音视频通话");
    connect(m_newCallAction, &QAction::triggered, this, &MainWindow::onStartCall);
    
    // 结束通话按钮
    m_endCallAction = m_toolBar->addAction(QIcon(":/icons/end_call.png"), "结束通话");
    m_endCallAction->setToolTip("结束当前通话");
    m_endCallAction->setEnabled(false);
    connect(m_endCallAction, &QAction::triggered, this, &MainWindow::onEndCall);
    
    m_toolBar->addSeparator();
    
    // 设置按钮
    m_settingsAction = m_toolBar->addAction(QIcon(":/icons/settings.png"), "设置");
    m_settingsAction->setToolTip("打开设置界面");
    connect(m_settingsAction, &QAction::triggered, [this]() {
        QMessageBox::information(this, "设置", "设置功能正在开发中...");
    });
    
    // 安全检测按钮
    m_securityAction = m_toolBar->addAction(QIcon(":/icons/security.png"), "安全检测");
    m_securityAction->setToolTip("查看安全检测状态");
    connect(m_securityAction, &QAction::triggered, [this]() {
        QMessageBox::information(this, "安全检测", "安全检测功能正在开发中...");
    });
}

void MainWindow::setupStatusBar()
{
    m_statusBar = statusBar();
    
    // 连接状态
    m_connectionStatusLabel = new QLabel("未连接");
    m_connectionStatusLabel->setStyleSheet("color: red;");
    m_statusBar->addWidget(m_connectionStatusLabel);
    
    m_statusBar->addSeparator();
    
    // 用户状态
    m_userStatusLabel = new QLabel("未登录");
    m_userStatusLabel->setStyleSheet("color: orange;");
    m_statusBar->addWidget(m_userStatusLabel);
    
    m_statusBar->addSeparator();
    
    // 通话状态
    m_callStatusLabel = new QLabel("空闲");
    m_callStatusLabel->setStyleSheet("color: green;");
    m_statusBar->addWidget(m_callStatusLabel);
    
    m_statusBar->addSeparator();
    
    // 网络质量
    m_networkQualityBar = new QProgressBar();
    m_networkQualityBar->setRange(0, 100);
    m_networkQualityBar->setValue(100);
    m_networkQualityBar->setMaximumWidth(100);
    m_networkQualityBar->setToolTip("网络质量");
    m_statusBar->addPermanentWidget(m_networkQualityBar);
}

void MainWindow::setupSystemTray()
{
    // 创建系统托盘图标
    m_trayIcon = new QSystemTrayIcon(this);
    m_trayIcon->setIcon(QIcon(":/icons/app_icon.png"));
    m_trayIcon->setToolTip(APP_NAME);
    
    // 创建托盘菜单
    m_trayMenu = new QMenu(this);
    
    m_showAction = m_trayMenu->addAction("显示主窗口");
    connect(m_showAction, &QAction::triggered, this, &MainWindow::onShowMainWindow);
    
    m_trayMenu->addSeparator();
    
    m_quitAction = m_trayMenu->addAction("退出");
    connect(m_quitAction, &QAction::triggered, this, &MainWindow::onQuitApplication);
    
    m_trayIcon->setContextMenu(m_trayMenu);
    
    // 连接托盘图标信号
    connect(m_trayIcon, &QSystemTrayIcon::activated, 
            this, &MainWindow::onTrayIconActivated);
    
    m_trayIcon->show();
}

void MainWindow::setupCentralWidget()
{
    // 创建堆叠窗口
    m_stackedWidget = new QStackedWidget(this);
    setCentralWidget(m_stackedWidget);
    
    // 创建各个界面
    m_loginWidget = new LoginWidget(this);
    
    // 添加到堆叠窗口
    m_stackedWidget->addWidget(m_loginWidget);
    
    // 连接信号
    connect(m_loginWidget, &LoginWidget::loginSuccess, 
            this, &MainWindow::onLoginSuccess);
}

void MainWindow::initializeManagers()
{
    // 初始化定时器
    m_heartbeatTimer = new QTimer(this);
    m_heartbeatTimer->setInterval(HEARTBEAT_INTERVAL);
    connect(m_heartbeatTimer, &QTimer::timeout, this, &MainWindow::onHeartbeatTimer);
    
    m_statusUpdateTimer = new QTimer(this);
    m_statusUpdateTimer->setInterval(STATUS_UPDATE_INTERVAL);
    connect(m_statusUpdateTimer, &QTimer::timeout, this, &MainWindow::onStatusUpdateTimer);
}

void MainWindow::setupStyles()
{
    // 应用深色主题
    applyDarkTheme();
}

void MainWindow::applyDarkTheme()
{
    QPalette darkPalette;
    darkPalette.setColor(QPalette::Window, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::WindowText, Qt::white);
    darkPalette.setColor(QPalette::Base, QColor(25, 25, 25));
    darkPalette.setColor(QPalette::AlternateBase, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::ToolTipBase, Qt::white);
    darkPalette.setColor(QPalette::ToolTipText, Qt::white);
    darkPalette.setColor(QPalette::Text, Qt::white);
    darkPalette.setColor(QPalette::Button, QColor(53, 53, 53));
    darkPalette.setColor(QPalette::ButtonText, Qt::white);
    darkPalette.setColor(QPalette::BrightText, Qt::red);
    darkPalette.setColor(QPalette::Link, QColor(42, 130, 218));
    darkPalette.setColor(QPalette::Highlight, QColor(42, 130, 218));
    darkPalette.setColor(QPalette::HighlightedText, Qt::black);
    
    qApp->setPalette(darkPalette);
}

void MainWindow::loadSettings()
{
    // 加载窗口位置和大小
    restoreGeometry(m_settings->value("geometry").toByteArray());
    restoreState(m_settings->value("windowState").toByteArray());
    
    // 加载用户设置
    m_currentUser = m_settings->value("currentUser").toJsonObject();
    m_isLoggedIn = !m_currentUser.isEmpty();
}

void MainWindow::saveSettings()
{
    // 保存窗口位置和大小
    m_settings->setValue("geometry", saveGeometry());
    m_settings->setValue("windowState", saveState());
    
    // 保存用户设置
    m_settings->setValue("currentUser", m_currentUser);
}

void MainWindow::checkSystemRequirements()
{
    // 检查Qt版本
    QString qtVersion = qVersion();
    qDebug() << "Qt版本:" << qtVersion;
    
    // 检查网络连接
    QNetworkAccessManager *nam = new QNetworkAccessManager(this);
    QNetworkReply *reply = nam->get(QNetworkRequest(QUrl("http://localhost:8000/health")));
    
    connect(reply, &QNetworkReply::finished, [this, reply, nam]() {
        if (reply->error() == QNetworkReply::NoError) {
            qDebug() << "后端服务连接正常";
        } else {
            qDebug() << "后端服务连接失败:" << reply->errorString();
        }
        reply->deleteLater();
        nam->deleteLater();
    });
}

void MainWindow::showLogin()
{
    m_stackedWidget->setCurrentWidget(m_loginWidget);
    m_isLoggedIn = false;
    updateUserStatus();
}

void MainWindow::showMainInterface()
{
    // 这里可以显示主界面，暂时显示登录界面
    m_stackedWidget->setCurrentWidget(m_loginWidget);
}

void MainWindow::showVideoCall(const QString &callId, const QString &remoteUser)
{
    // 暂时显示登录界面，视频通话功能待实现
    m_stackedWidget->setCurrentWidget(m_loginWidget);
    m_isInCall = true;
    updateCallStatus();
}

void MainWindow::showSecurityDetection(const QString &callId)
{
    // 暂时显示登录界面，安全检测功能待实现
    m_stackedWidget->setCurrentWidget(m_loginWidget);
}

void MainWindow::updateUserInfo(const QJsonObject &userInfo)
{
    m_currentUser = userInfo;
    saveSettings();
    updateUserStatus();
}

void MainWindow::showNotification(const QString &title, const QString &message, QSystemTrayIcon::MessageIcon icon)
{
    if (m_trayIcon && m_trayIcon->isVisible()) {
        m_trayIcon->showMessage(title, message, icon, 5000);
    }
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    if (m_trayIcon && m_trayIcon->isVisible()) {
        hide();
        event->ignore();
    } else {
        event->accept();
    }
}

void MainWindow::showEvent(QShowEvent *event)
{
    QMainWindow::showEvent(event);
    
    if (m_isLoggedIn && !m_isConnected) {
        connectToServer();
    }
}

// 槽函数实现
void MainWindow::onLoginSuccess(const QJsonObject &userInfo)
{
    m_currentUser = userInfo;
    m_isLoggedIn = true;
    updateUserInfo(userInfo);
    showMainInterface();
    
    // 连接到服务器
    connectToServer();
    
    showNotification("登录成功", "欢迎使用音视频通话系统！");
}

void MainWindow::onLogout()
{
    m_isLoggedIn = false;
    m_currentUser = QJsonObject();
    disconnectFromServer();
    showLogin();
    saveSettings();
}

void MainWindow::onStartCall()
{
    // 这里可以显示拨号界面或联系人列表
    // 暂时直接开始一个测试通话
    QString callId = "test_call_" + QString::number(QDateTime::currentMSecsSinceEpoch());
    showVideoCall(callId, "测试用户");
}

void MainWindow::onEndCall()
{
    if (m_isInCall) {
        // 暂时直接结束通话，视频通话功能待实现
        m_isInCall = false;
        updateCallStatus();
    }
}

void MainWindow::onIncomingCall(const QString &callId, const QString &caller)
{
    showNotification("来电", QString("来自 %1 的来电").arg(caller), QSystemTrayIcon::Information);
    
    // 这里可以显示来电界面
    QMessageBox::StandardButton reply = QMessageBox::question(this, "来电", 
        QString("来自 %1 的来电\n是否接听？").arg(caller),
        QMessageBox::Yes | QMessageBox::No);
    
    if (reply == QMessageBox::Yes) {
        showVideoCall(callId, caller);
    }
}

void MainWindow::onCallEnded(const QString &callId)
{
    m_isInCall = false;
    updateCallStatus();
    showNotification("通话结束", "通话已结束");
}

void MainWindow::onSecurityAlert(const QString &callId, const QString &alertType, double riskScore)
{
    QString message = QString("安全警报: %1 (风险评分: %2)").arg(alertType).arg(riskScore);
    showNotification("安全警报", message, QSystemTrayIcon::Warning);
    
    // 更新安全状态
    m_securityScore = riskScore;
    // updateSecurityStatus(); // 暂时注释掉
}

void MainWindow::onNetworkConnected()
{
    m_isConnected = true;
    updateConnectionStatus();
    m_heartbeatTimer->start();
}

void MainWindow::onNetworkDisconnected()
{
    m_isConnected = false;
    updateConnectionStatus();
    m_heartbeatTimer->stop();
}

void MainWindow::onNetworkError(const QString &error)
{
    showNotification("网络错误", error, QSystemTrayIcon::Critical);
    updateConnectionStatus();
}

void MainWindow::onTrayIconActivated(QSystemTrayIcon::ActivationReason reason)
{
    if (reason == QSystemTrayIcon::DoubleClick) {
        onShowMainWindow();
    }
}

void MainWindow::onShowMainWindow()
{
    show();
    raise();
    activateWindow();
}

void MainWindow::onQuitApplication()
{
    qApp->quit();
}

void MainWindow::onHeartbeatTimer()
{
    // 心跳功能待实现
}

void MainWindow::onStatusUpdateTimer()
{
    updateConnectionStatus();
    updateUserStatus();
    updateCallStatus();
}

void MainWindow::connectToServer()
{
    // 网络连接功能待实现
    m_isConnected = true;
    updateConnectionStatus();
}

void MainWindow::disconnectFromServer()
{
    // 网络断开功能待实现
    m_isConnected = false;
    updateConnectionStatus();
    m_heartbeatTimer->stop();
}

void MainWindow::sendHeartbeat()
{
    // 心跳功能待实现
}

void MainWindow::updateConnectionStatus()
{
    if (m_connectionStatusLabel) {
        if (m_isConnected) {
            m_connectionStatusLabel->setText("已连接");
            m_connectionStatusLabel->setStyleSheet("color: green;");
        } else {
            m_connectionStatusLabel->setText("未连接");
            m_connectionStatusLabel->setStyleSheet("color: red;");
        }
    }
}

void MainWindow::updateUserStatus()
{
    if (m_userStatusLabel) {
        if (m_isLoggedIn) {
            QString username = m_currentUser["username"].toString();
            m_userStatusLabel->setText(QString("用户: %1").arg(username));
            m_userStatusLabel->setStyleSheet("color: green;");
        } else {
            m_userStatusLabel->setText("未登录");
            m_userStatusLabel->setStyleSheet("color: orange;");
        }
    }
}

void MainWindow::updateCallStatus()
{
    if (m_callStatusLabel) {
        if (m_isInCall) {
            m_callStatusLabel->setText("通话中");
            m_callStatusLabel->setStyleSheet("color: red;");
            m_endCallAction->setEnabled(true);
        } else {
            m_callStatusLabel->setText("空闲");
            m_callStatusLabel->setStyleSheet("color: green;");
            m_endCallAction->setEnabled(false);
        }
    }
} 