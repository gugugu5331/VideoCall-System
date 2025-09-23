#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QtWidgets/QMainWindow>
#include <QtWidgets/QVBoxLayout>
#include <QtWidgets/QHBoxLayout>
#include <QtWidgets/QGridLayout>
#include <QtWidgets/QPushButton>
#include <QtWidgets/QLabel>
#include <QtWidgets/QLineEdit>
#include <QtWidgets/QTextEdit>
#include <QtWidgets/QSplitter>
#include <QtWidgets/QTabWidget>
#include <QtWidgets/QListWidget>
#include <QtWidgets/QStatusBar>
#include <QtWidgets/QMenuBar>
#include <QtWidgets/QToolBar>
#include <QtWidgets/QAction>
#include <QtWidgets/QDialog>
#include <QtWidgets/QMessageBox>
#include <QtCore/QTimer>
#include <QtCore/QSettings>

QT_BEGIN_NAMESPACE
class QAction;
class QMenu;
QT_END_NAMESPACE

class VideoWidget;
class WebRTCManager;
class SignalingClient;
class DetectionManager;
class ParticipantWidget;
class ChatWidget;
class SettingsDialog;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

protected:
    void closeEvent(QCloseEvent *event) override;

private slots:
    // 会议控制
    void joinMeeting();
    void leaveMeeting();
    void toggleCamera();
    void toggleMicrophone();
    void toggleScreenShare();
    
    // 用户界面
    void showSettings();
    void showAbout();
    void updateConnectionStatus(const QString &status);
    void updateParticipantCount(int count);
    
    // WebRTC事件
    void onUserJoined(const QString &userId, const QString &userName);
    void onUserLeft(const QString &userId);
    void onLocalStreamReady();
    void onRemoteStreamReceived(const QString &userId, QObject *stream);
    
    // 检测事件
    void onDetectionResult(const QString &type, bool isFake, double confidence);
    void onDetectionAlert(const QString &message);
    
    // 聊天事件
    void onChatMessageReceived(const QString &sender, const QString &message);
    void sendChatMessage();

private:
    void setupUI();
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupConnections();
    void loadSettings();
    void saveSettings();
    void updateControlButtons();
    void showJoinDialog();
    
    // UI组件
    QWidget *m_centralWidget;
    QSplitter *m_mainSplitter;
    QSplitter *m_videoSplitter;
    
    // 视频区域
    QWidget *m_videoArea;
    QGridLayout *m_videoLayout;
    VideoWidget *m_localVideoWidget;
    QMap<QString, VideoWidget*> m_remoteVideoWidgets;
    
    // 侧边栏
    QTabWidget *m_sidebarTabs;
    ParticipantWidget *m_participantWidget;
    ChatWidget *m_chatWidget;
    QWidget *m_detectionWidget;
    QTextEdit *m_detectionLog;
    
    // 控制按钮
    QToolBar *m_controlToolBar;
    QPushButton *m_joinButton;
    QPushButton *m_leaveButton;
    QPushButton *m_cameraButton;
    QPushButton *m_microphoneButton;
    QPushButton *m_screenShareButton;
    
    // 状态栏
    QLabel *m_connectionStatusLabel;
    QLabel *m_participantCountLabel;
    QLabel *m_detectionStatusLabel;
    
    // 菜单和动作
    QMenuBar *m_menuBar;
    QMenu *m_fileMenu;
    QMenu *m_viewMenu;
    QMenu *m_toolsMenu;
    QMenu *m_helpMenu;
    
    QAction *m_joinAction;
    QAction *m_leaveAction;
    QAction *m_settingsAction;
    QAction *m_exitAction;
    QAction *m_aboutAction;
    
    // 核心组件
    WebRTCManager *m_webrtcManager;
    SignalingClient *m_signalingClient;
    DetectionManager *m_detectionManager;
    SettingsDialog *m_settingsDialog;
    
    // 会议状态
    bool m_isInMeeting;
    bool m_isCameraEnabled;
    bool m_isMicrophoneEnabled;
    bool m_isScreenSharing;
    QString m_currentMeetingId;
    QString m_currentUserName;
    
    // 设置
    QSettings *m_settings;
    QString m_serverUrl;
    int m_serverPort;
    
    // 定时器
    QTimer *m_statusUpdateTimer;
};

#endif // MAINWINDOW_H
