#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QStackedWidget>
#include <QSystemTrayIcon>
#include <QMenu>
#include <QAction>
#include <QTimer>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QSettings>
#include <QMessageBox>
#include <QCloseEvent>
#include <QShowEvent>

// 前向声明
class VideoCallWidget;
class LoginWidget;
class UserProfileWidget;
class CallHistoryWidget;
class SettingsWidget;
class SecurityDetectionWidget;
class NetworkManager;
class AudioManager;
class VideoManager;
class SecurityManager;

QT_BEGIN_NAMESPACE
class QLabel;
class QPushButton;
class QVBoxLayout;
class QHBoxLayout;
class QGridLayout;
class QTabWidget;
class QListWidget;
class QLineEdit;
class QTextEdit;
class QComboBox;
class QSlider;
class QCheckBox;
class QGroupBox;
class QFrame;
class QSplitter;
class QProgressBar;
class QStatusBar;
class QToolBar;
class QMenuBar;
QT_END_NAMESPACE

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

    // 公共方法
    void showLogin();
    void showMainInterface();
    void showVideoCall(const QString &callId, const QString &remoteUser);
    void showSecurityDetection(const QString &callId);
    void updateUserInfo(const QJsonObject &userInfo);
    void showNotification(const QString &title, const QString &message, QSystemTrayIcon::MessageIcon icon = QSystemTrayIcon::Information);

protected:
    void closeEvent(QCloseEvent *event) override;
    void showEvent(QShowEvent *event) override;

private slots:
    // UI事件处理
    void onLoginSuccess(const QJsonObject &userInfo);
    void onLogout();
    void onStartCall();
    void onEndCall();
    void onIncomingCall(const QString &callId, const QString &caller);
    void onCallEnded(const QString &callId);
    void onSecurityAlert(const QString &callId, const QString &alertType, double riskScore);
    
    // 网络事件处理
    void onNetworkConnected();
    void onNetworkDisconnected();
    void onNetworkError(const QString &error);
    
    // 系统托盘事件
    void onTrayIconActivated(QSystemTrayIcon::ActivationReason reason);
    void onShowMainWindow();
    void onQuitApplication();
    
    // 定时器事件
    void onHeartbeatTimer();
    void onStatusUpdateTimer();

private:
    // UI组件
    void setupUI();
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupSystemTray();
    void setupCentralWidget();
    void setupSidebar();
    void setupMainContent();
    
    // 样式设置
    void setupStyles();
    void applyDarkTheme();
    
    // 初始化
    void initializeManagers();
    void loadSettings();
    void saveSettings();
    void checkSystemRequirements();
    
    // 网络管理
    void connectToServer();
    void disconnectFromServer();
    void sendHeartbeat();
    
    // 状态管理
    void updateConnectionStatus();
    void updateUserStatus();
    void updateCallStatus();

private:
    // 主UI组件
    QStackedWidget *m_stackedWidget;
    QWidget *m_mainWidget;
    QSplitter *m_mainSplitter;
    QWidget *m_sidebarWidget;
    QWidget *m_contentWidget;
    
    // 各个页面
    LoginWidget *m_loginWidget;
    VideoCallWidget *m_videoCallWidget;
    UserProfileWidget *m_userProfileWidget;
    CallHistoryWidget *m_callHistoryWidget;
    SettingsWidget *m_settingsWidget;
    SecurityDetectionWidget *m_securityDetectionWidget;
    
    // 管理器
    NetworkManager *m_networkManager;
    AudioManager *m_audioManager;
    VideoManager *m_videoManager;
    SecurityManager *m_securityManager;
    
    // 系统托盘
    QSystemTrayIcon *m_trayIcon;
    QMenu *m_trayMenu;
    QAction *m_showAction;
    QAction *m_quitAction;
    
    // 菜单栏
    QMenuBar *m_menuBar;
    QMenu *m_fileMenu;
    QMenu *m_callMenu;
    QMenu *m_toolsMenu;
    QMenu *m_helpMenu;
    
    // 工具栏
    QToolBar *m_toolBar;
    QAction *m_newCallAction;
    QAction *m_endCallAction;
    QAction *m_settingsAction;
    QAction *m_securityAction;
    
    // 状态栏
    QStatusBar *m_statusBar;
    QLabel *m_connectionStatusLabel;
    QLabel *m_userStatusLabel;
    QLabel *m_callStatusLabel;
    QProgressBar *m_networkQualityBar;
    
    // 侧边栏组件
    QVBoxLayout *m_sidebarLayout;
    QPushButton *m_profileButton;
    QPushButton *m_callsButton;
    QPushButton *m_historyButton;
    QPushButton *m_settingsButton;
    QPushButton *m_securityButton;
    
    // 主内容区域
    QTabWidget *m_mainTabWidget;
    QWidget *m_dashboardTab;
    QWidget *m_callsTab;
    QWidget *m_contactsTab;
    QWidget *m_securityTab;
    
    // 数据
    QJsonObject m_currentUser;
    QString m_currentCallId;
    bool m_isLoggedIn;
    bool m_isInCall;
    bool m_isConnected;
    
    // 定时器
    QTimer *m_heartbeatTimer;
    QTimer *m_statusUpdateTimer;
    
    // 设置
    QSettings *m_settings;
    
    // 网络
    QNetworkAccessManager *m_httpManager;
    
    // 常量
    static const int HEARTBEAT_INTERVAL = 30000; // 30秒
    static const int STATUS_UPDATE_INTERVAL = 5000; // 5秒
    static const QString SERVER_URL;
    static const QString APP_NAME;
    static const QString APP_VERSION;
};

#endif // MAINWINDOW_H 