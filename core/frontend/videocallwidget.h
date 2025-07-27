#ifndef VIDEOCALLWIDGET_H
#define VIDEOCALLWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QPushButton>
#include <QSlider>
#include <QComboBox>
#include <QCheckBox>
#include <QGroupBox>
#include <QFrame>
#include <QTimer>
#include <QPropertyAnimation>
#include <QGraphicsOpacityEffect>
#include <QWebEngineView>
#include <QWebChannel>
#include <QWebSocket>
#include <QMediaPlayer>
#include <QVideoWidget>
#include <QCamera>
#include <QAudioInput>
#include <QAudioOutput>
#include <QAudioDeviceInfo>
#include <QVideoFrame>
#include <QImage>
#include <QPixmap>
#include <QPainter>
#include <QPen>
#include <QBrush>
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

QT_BEGIN_NAMESPACE
class QVideoWidget;
class QAudioInput;
class QAudioOutput;
class QCamera;
class QMediaRecorder;
class QWebSocket;
class QWebEngineView;
class QWebChannel;
class QProgressBar;
class QLabel;
class QPushButton;
class QSlider;
class QComboBox;
class QCheckBox;
class QGroupBox;
class QFrame;
class QSplitter;
class QTabWidget;
class QListWidget;
class QTextEdit;
class QLineEdit;
class QSpinBox;
class QDoubleSpinBox;
class QDateTimeEdit;
class QCalendarWidget;
class QTableWidget;
class QTreeWidget;
class QTreeWidgetItem;
class QHeaderView;
class QScrollArea;
class QScrollBar;
class QToolButton;
class QMenu;
class QAction;
class QActionGroup;
class QToolBar;
class QStatusBar;
class QProgressDialog;
class QInputDialog;
class QColorDialog;
class QFontDialog;
class QFileDialog;
class QMessageBox;
class QErrorMessage;
class QWizard;
class QWizardPage;
class QDialog;
class QDialogButtonBox;
class QFormLayout;
class QBoxLayout;
class QGridLayout;
class QStackedLayout;
class QStackedWidget;
class QTabWidget;
class QToolBox;
class QGroupBox;
class QFrame;
class QSplitter;
class QScrollArea;
class QAbstractScrollArea;
class QAbstractItemView;
class QAbstractItemModel;
class QStandardItemModel;
class QStandardItem;
class QItemSelectionModel;
class QItemSelection;
class QItemSelectionRange;
class QAbstractProxyModel;
class QSortFilterProxyModel;
class QIdentityProxyModel;
class QTransposeProxyModel;
class QConcatenateTablesProxyModel;
class QAbstractTableModel;
class QAbstractListModel;
class QStringListModel;
class QDirModel;
class QFileSystemModel;
class QStandardItemModel;
class QStandardItem;
class QItemSelectionModel;
class QItemSelection;
class QItemSelectionRange;
class QAbstractProxyModel;
class QSortFilterProxyModel;
class QIdentityProxyModel;
class QTransposeProxyModel;
class QConcatenateTablesProxyModel;
class QAbstractTableModel;
class QAbstractListModel;
class QStringListModel;
class QDirModel;
class QFileSystemModel;
QT_END_NAMESPACE

class VideoCallWidget : public QWidget
{
    Q_OBJECT

public:
    explicit VideoCallWidget(QWidget *parent = nullptr);
    ~VideoCallWidget();

    // 公共方法
    void startCall(const QString &callId, const QString &remoteUser, bool isIncoming = false);
    void endCall();
    void acceptCall();
    void rejectCall();
    void muteAudio(bool muted);
    void muteVideo(bool muted);
    void switchCamera();
    void toggleFullscreen();
    void takeScreenshot();
    void startRecording();
    void stopRecording();
    void showSecurityPanel(bool show);
    
    // 状态查询
    bool isInCall() const { return m_isInCall; }
    bool isAudioMuted() const { return m_isAudioMuted; }
    bool isVideoMuted() const { return m_isVideoMuted; }
    bool isRecording() const { return m_isRecording; }
    bool isFullscreen() const { return m_isFullscreen; }
    QString getCurrentCallId() const { return m_currentCallId; }
    QString getRemoteUser() const { return m_remoteUser; }

signals:
    void callEnded(const QString &callId);
    void callAccepted(const QString &callId);
    void callRejected(const QString &callId);
    void audioMutedChanged(bool muted);
    void videoMutedChanged(bool muted);
    void cameraSwitched();
    void fullscreenToggled(bool fullscreen);
    void screenshotTaken(const QString &filePath);
    void recordingStarted(const QString &filePath);
    void recordingStopped(const QString &filePath);
    void securityAlert(const QString &alertType, double riskScore);
    void networkQualityChanged(int quality);
    void callDurationChanged(int seconds);

protected:
    void resizeEvent(QResizeEvent *event) override;
    void keyPressEvent(QKeyEvent *event) override;
    void mouseDoubleClickEvent(QMouseEvent *event) override;
    void closeEvent(QCloseEvent *event) override;

private slots:
    // UI事件处理
    void onEndCallClicked();
    void onAcceptCallClicked();
    void onRejectCallClicked();
    void onMuteAudioClicked();
    void onMuteVideoClicked();
    void onSwitchCameraClicked();
    void onFullscreenClicked();
    void onScreenshotClicked();
    void onRecordingClicked();
    void onSecurityClicked();
    void onSettingsClicked();
    void onChatClicked();
    void onParticipantsClicked();
    
    // 音视频处理
    void onLocalVideoFrame(const QVideoFrame &frame);
    void onRemoteVideoFrame(const QVideoFrame &frame);
    void onAudioLevelChanged(int level);
    void onVideoQualityChanged(int quality);
    void onNetworkQualityChanged(int quality);
    
    // 安全检测
    void onSecurityDetectionResult(const QString &type, double riskScore, const QString &details);
    void onFaceDetectionResult(const QRect &faceRect, double confidence);
    void onVoiceDetectionResult(bool isSpoofed, double confidence);
    void onVideoDetectionResult(bool isDeepfake, double confidence);
    
    // 网络事件
    void onWebSocketConnected();
    void onWebSocketDisconnected();
    void onWebSocketError(const QString &error);
    void onWebSocketMessageReceived(const QString &message);
    
    // 定时器事件
    void onCallDurationTimer();
    void onQualityCheckTimer();
    void onSecurityCheckTimer();
    
    // 录制事件
    void onRecordingStateChanged(QMediaRecorder::State state);
    void onRecordingError(QMediaRecorder::Error error, const QString &errorString);

private:
    // UI设置
    void setupUI();
    void setupVideoLayout();
    void setupControlPanel();
    void setupSecurityPanel();
    void setupChatPanel();
    void setupParticipantsPanel();
    void setupSettingsPanel();
    
    // 音视频管理
    void initializeAudio();
    void initializeVideo();
    void initializeWebRTC();
    void setupCamera();
    void setupMicrophone();
    void setupSpeakers();
    
    // 安全检测
    void initializeSecurityDetection();
    void setupFaceDetection();
    void setupVoiceDetection();
    void setupVideoDetection();
    void processSecurityFrame(const QImage &frame);
    
    // 网络管理
    void connectToSignalingServer();
    void disconnectFromSignalingServer();
    void sendSignalingMessage(const QJsonObject &message);
    void handleSignalingMessage(const QJsonObject &message);
    
    // 录制管理
    void setupRecording();
    void startVideoRecording();
    void startAudioRecording();
    void stopVideoRecording();
    void stopAudioRecording();
    
    // 工具方法
    void updateCallDuration();
    void updateNetworkQuality();
    void updateSecurityStatus();
    void updateRecordingStatus();
    void updateUIState();
    void applyVideoEffects();
    void applyAudioEffects();
    QString formatDuration(int seconds);
    QString formatFileSize(qint64 bytes);
    void saveScreenshot(const QPixmap &pixmap);
    void showNotification(const QString &title, const QString &message);

private:
    // 主UI组件
    QVBoxLayout *m_mainLayout;
    QHBoxLayout *m_videoLayout;
    QHBoxLayout *m_controlLayout;
    QVBoxLayout *m_securityLayout;
    QVBoxLayout *m_chatLayout;
    QVBoxLayout *m_participantsLayout;
    QVBoxLayout *m_settingsLayout;
    
    // 视频显示区域
    QWidget *m_videoContainer;
    QVideoWidget *m_localVideoWidget;
    QVideoWidget *m_remoteVideoWidget;
    QLabel *m_localVideoLabel;
    QLabel *m_remoteVideoLabel;
    QLabel *m_callStatusLabel;
    QLabel *m_callDurationLabel;
    QLabel *m_networkQualityLabel;
    QLabel *m_securityStatusLabel;
    
    // 控制面板
    QWidget *m_controlPanel;
    QPushButton *m_endCallButton;
    QPushButton *m_acceptCallButton;
    QPushButton *m_rejectCallButton;
    QPushButton *m_muteAudioButton;
    QPushButton *m_muteVideoButton;
    QPushButton *m_switchCameraButton;
    QPushButton *m_fullscreenButton;
    QPushButton *m_screenshotButton;
    QPushButton *m_recordingButton;
    QPushButton *m_securityButton;
    QPushButton *m_settingsButton;
    QPushButton *m_chatButton;
    QPushButton *m_participantsButton;
    
    // 音视频控制
    QSlider *m_audioVolumeSlider;
    QSlider *m_videoBrightnessSlider;
    QSlider *m_videoContrastSlider;
    QComboBox *m_cameraComboBox;
    QComboBox *m_microphoneComboBox;
    QComboBox *m_speakerComboBox;
    QCheckBox *m_echoCancellationCheckBox;
    QCheckBox *m_noiseSuppressionCheckBox;
    QCheckBox *m_autoGainControlCheckBox;
    
    // 安全检测面板
    QWidget *m_securityPanel;
    QLabel *m_faceDetectionLabel;
    QLabel *m_voiceDetectionLabel;
    QLabel *m_videoDetectionLabel;
    QProgressBar *m_faceRiskBar;
    QProgressBar *m_voiceRiskBar;
    QProgressBar *m_videoRiskBar;
    QLabel *m_securityScoreLabel;
    QTextEdit *m_securityDetailsText;
    
    // 聊天面板
    QWidget *m_chatPanel;
    QTextEdit *m_chatHistoryText;
    QLineEdit *m_chatInputLine;
    QPushButton *m_sendMessageButton;
    QListWidget *m_emojiListWidget;
    
    // 参与者面板
    QWidget *m_participantsPanel;
    QListWidget *m_participantsListWidget;
    QPushButton *m_addParticipantButton;
    QPushButton *m_removeParticipantButton;
    QPushButton *m_muteAllButton;
    
    // 设置面板
    QWidget *m_settingsPanel;
    QTabWidget *m_settingsTabWidget;
    QWidget *m_audioSettingsTab;
    QWidget *m_videoSettingsTab;
    QWidget *m_networkSettingsTab;
    QWidget *m_securitySettingsTab;
    
    // 音视频设备
    QCamera *m_camera;
    QAudioInput *m_audioInput;
    QAudioOutput *m_audioOutput;
    QMediaRecorder *m_videoRecorder;
    QMediaRecorder *m_audioRecorder;
    
    // WebRTC相关
    QWebEngineView *m_webRTCView;
    QWebChannel *m_webChannel;
    QWebSocket *m_signalingSocket;
    
    // 网络管理
    QNetworkAccessManager *m_networkManager;
    QNetworkReply *m_currentReply;
    
    // 定时器
    QTimer *m_callDurationTimer;
    QTimer *m_qualityCheckTimer;
    QTimer *m_securityCheckTimer;
    
    // 线程
    QThread *m_videoProcessingThread;
    QThread *m_audioProcessingThread;
    QThread *m_securityProcessingThread;
    QMutex m_videoMutex;
    QMutex m_audioMutex;
    QMutex m_securityMutex;
    
    // 数据
    QString m_currentCallId;
    QString m_remoteUser;
    bool m_isInCall;
    bool m_isIncomingCall;
    bool m_isAudioMuted;
    bool m_isVideoMuted;
    bool m_isRecording;
    bool m_isFullscreen;
    bool m_isSecurityEnabled;
    int m_callDuration;
    int m_networkQuality;
    double m_securityScore;
    
    // 录制相关
    QString m_recordingFilePath;
    QDateTime m_recordingStartTime;
    qint64 m_recordingFileSize;
    
    // 安全检测结果
    double m_faceRiskScore;
    double m_voiceRiskScore;
    double m_videoRiskScore;
    QString m_lastSecurityAlert;
    QDateTime m_lastSecurityCheck;
    
    // 样式
    QString m_currentTheme;
    QPalette m_customPalette;
    
    // 常量
    static const int CALL_DURATION_UPDATE_INTERVAL = 1000; // 1秒
    static const int QUALITY_CHECK_INTERVAL = 5000; // 5秒
    static const int SECURITY_CHECK_INTERVAL = 10000; // 10秒
    static const int MAX_RECORDING_DURATION = 3600000; // 1小时
    static const qint64 MAX_RECORDING_FILE_SIZE = 1073741824; // 1GB
};

#endif // VIDEOCALLWIDGET_H 