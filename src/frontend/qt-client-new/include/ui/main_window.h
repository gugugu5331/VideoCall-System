#pragma once

#include "core/common.h"

namespace VideoCallSystem {

class MeetingWidget;
class ControlPanel;
class FilterPanel;
class AIDetectionPanel;
class MonitoringPanel;
class SettingsDialog;
class ParticipantWidget;
class ChatWidget;
class StatusWidget;

class WebRTCManager;
class SignalingClient;
class AIDetectionClient;
class EdgeInfraClient;
class BackendClient;

class CameraManager;
class AudioManager;
class VideoProcessor;
class FilterEngine;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

    // 初始化
    bool initialize();
    void cleanup();

    // 窗口状态
    bool isInitialized() const { return initialized_; }
    void showWindow();
    void hideWindow();

    // 会议控制
    void joinMeeting(const QString& meetingId, const QString& userName = QString());
    void leaveMeeting();
    void createMeeting(const QString& title = QString());

    // 媒体控制
    void toggleCamera();
    void toggleMicrophone();
    void toggleScreenShare();
    void startRecording();
    void stopRecording();

    // UI控制
    void showSettings();
    void showAbout();
    void switchTheme(const QString& themeName);
    void setFullScreen(bool fullScreen);

protected:
    // 事件处理
    void closeEvent(QCloseEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;
    void changeEvent(QEvent *event) override;
    void keyPressEvent(QKeyEvent *event) override;

private slots:
    // 菜单和工具栏
    void onNewMeeting();
    void onJoinMeeting();
    void onLeaveMeeting();
    void onSettings();
    void onAbout();
    void onExit();

    // 媒体控制槽
    void onCameraToggled(bool enabled);
    void onMicrophoneToggled(bool enabled);
    void onScreenShareToggled(bool enabled);
    void onRecordingToggled(bool recording);

    // 连接状态
    void onConnectionStateChanged(ConnectionState state);
    void onMeetingStateChanged(MeetingState state);
    void onParticipantJoined(const UserInfo& user);
    void onParticipantLeft(const QString& userId);

    // AI检测事件
    void onDetectionResult(const DetectionResult& result);
    void onDetectionAlert(const QString& message);

    // 滤镜事件
    void onFilterChanged(const FilterParams& params);
    void onFilterPresetSelected(const QString& presetName);

    // 聊天事件
    void onChatMessageReceived(const QString& senderId, const QString& message);
    void onChatMessageSent(const QString& message);

    // 系统事件
    void onSystemError(const QString& error);
    void onPerformanceUpdate(const QVariantMap& metrics);

    // UI事件
    void onTabChanged(int index);
    void onSplitterMoved(int pos, int index);
    void onWindowStateChanged();

private:
    // UI初始化
    void setupUI();
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupCentralWidget();
    void setupDockWidgets();

    // 连接信号槽
    void setupConnections();
    void connectNetworkManagers();
    void connectMediaManagers();
    void connectUIComponents();

    // 布局管理
    void setupMainLayout();
    void setupVideoLayout();
    void setupSidebarLayout();
    void updateLayout();

    // 状态管理
    void updateConnectionStatus();
    void updateMeetingStatus();
    void updateMediaStatus();
    void updateParticipantCount();

    // 设置管理
    void loadSettings();
    void saveSettings();
    void applySettings();

    // 主题管理
    void loadTheme(const QString& themeName);
    void applyTheme();

    // 工具函数
    void showErrorMessage(const QString& title, const QString& message);
    void showInfoMessage(const QString& title, const QString& message);
    bool confirmAction(const QString& title, const QString& message);

    // 快捷键
    void setupShortcuts();
    void registerGlobalShortcuts();

private:
    // 初始化状态
    bool initialized_;

    // 核心UI组件
    QWidget* centralWidget_;
    QSplitter* mainSplitter_;
    QSplitter* videoSplitter_;
    QTabWidget* sidebarTabs_;

    // 主要面板
    MeetingWidget* meetingWidget_;
    ControlPanel* controlPanel_;
    FilterPanel* filterPanel_;
    AIDetectionPanel* aiDetectionPanel_;
    MonitoringPanel* monitoringPanel_;
    ParticipantWidget* participantWidget_;
    ChatWidget* chatWidget_;
    StatusWidget* statusWidget_;

    // 对话框
    SettingsDialog* settingsDialog_;

    // 菜单和工具栏
    QMenuBar* menuBar_;
    QMenu* fileMenu_;
    QMenu* viewMenu_;
    QMenu* toolsMenu_;
    QMenu* helpMenu_;

    QToolBar* mainToolBar_;
    QToolBar* mediaToolBar_;

    QStatusBar* statusBar_;

    // 动作
    QAction* newMeetingAction_;
    QAction* joinMeetingAction_;
    QAction* leaveMeetingAction_;
    QAction* settingsAction_;
    QAction* exitAction_;
    QAction* aboutAction_;

    QAction* cameraAction_;
    QAction* microphoneAction_;
    QAction* screenShareAction_;
    QAction* recordingAction_;

    QAction* fullScreenAction_;
    QAction* minimizeAction_;
    QAction* maximizeAction_;

    // 网络管理器
    WebRTCManager* webrtcManager_;
    SignalingClient* signalingClient_;
    AIDetectionClient* aiDetectionClient_;
    EdgeInfraClient* edgeInfraClient_;
    BackendClient* backendClient_;

    // 媒体管理器
    CameraManager* cameraManager_;
    AudioManager* audioManager_;
    VideoProcessor* videoProcessor_;
    FilterEngine* filterEngine_;

    // 状态变量
    ConnectionState connectionState_;
    MeetingState meetingState_;
    MediaState cameraState_;
    MediaState microphoneState_;
    MediaState screenShareState_;
    bool isRecording_;
    bool isFullScreen_;

    // 会议信息
    MeetingInfo currentMeeting_;
    QList<UserInfo> participants_;

    // 性能监控
    QTimer* performanceTimer_;
    QVariantMap performanceMetrics_;

    // 设置
    QSettings* settings_;
    QString currentTheme_;

    // 快捷键
    QList<QShortcut*> shortcuts_;
};

} // namespace VideoCallSystem
