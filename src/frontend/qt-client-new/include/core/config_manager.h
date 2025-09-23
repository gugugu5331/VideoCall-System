#pragma once

#include "common.h"

namespace VideoCallSystem {

class ConfigManager : public QObject
{
    Q_OBJECT

public:
    explicit ConfigManager(QObject* parent = nullptr);
    ~ConfigManager();

    // 初始化和保存
    bool initialize();
    bool loadConfig(const QString& configPath = QString());
    bool saveConfig(const QString& configPath = QString());
    void resetToDefaults();

    // 服务器配置
    QString serverHost() const;
    void setServerHost(const QString& host);
    
    int serverPort() const;
    void setServerPort(int port);
    
    QString serverUrl() const;
    void setServerUrl(const QString& url);

    // 信令服务器配置
    QString signalingHost() const;
    void setSignalingHost(const QString& host);
    
    int signalingPort() const;
    void setSignalingPort(int port);
    
    QString signalingUrl() const;

    // AI检测服务配置
    QString aiServiceHost() const;
    void setAIServiceHost(const QString& host);
    
    int aiServicePort() const;
    void setAIServicePort(int port);
    
    QString aiServiceUrl() const;

    // Edge-Model-Infra配置
    QString edgeInfraHost() const;
    void setEdgeInfraHost(const QString& host);
    
    int edgeInfraPort() const;
    void setEdgeInfraPort(int port);
    
    QString edgeInfraUrl() const;

    // 用户配置
    QString userId() const;
    void setUserId(const QString& userId);
    
    QString userName() const;
    void setUserName(const QString& userName);
    
    QString userEmail() const;
    void setUserEmail(const QString& email);
    
    QString userAvatar() const;
    void setUserAvatar(const QString& avatar);

    // 视频配置
    int videoWidth() const;
    void setVideoWidth(int width);
    
    int videoHeight() const;
    void setVideoHeight(int height);
    
    int videoFps() const;
    void setVideoFps(int fps);
    
    QString cameraDevice() const;
    void setCameraDevice(const QString& device);
    
    bool cameraEnabled() const;
    void setCameraEnabled(bool enabled);

    // 音频配置
    int audioSampleRate() const;
    void setAudioSampleRate(int sampleRate);
    
    int audioChannels() const;
    void setAudioChannels(int channels);
    
    int audioBufferSize() const;
    void setAudioBufferSize(int bufferSize);
    
    QString audioInputDevice() const;
    void setAudioInputDevice(const QString& device);
    
    QString audioOutputDevice() const;
    void setAudioOutputDevice(const QString& device);
    
    bool microphoneEnabled() const;
    void setMicrophoneEnabled(bool enabled);
    
    bool speakerEnabled() const;
    void setSpeakerEnabled(bool enabled);
    
    float microphoneVolume() const;
    void setMicrophoneVolume(float volume);
    
    float speakerVolume() const;
    void setSpeakerVolume(float volume);

    // UI配置
    int windowWidth() const;
    void setWindowWidth(int width);
    
    int windowHeight() const;
    void setWindowHeight(int height);
    
    bool windowMaximized() const;
    void setWindowMaximized(bool maximized);
    
    QString theme() const;
    void setTheme(const QString& theme);
    
    QString language() const;
    void setLanguage(const QString& language);
    
    bool showStatusBar() const;
    void setShowStatusBar(bool show);
    
    bool showToolBar() const;
    void setShowToolBar(bool show);

    // 滤镜配置
    FilterParams defaultFilterParams() const;
    void setDefaultFilterParams(const FilterParams& params);
    
    QList<FilterParams> savedFilterPresets() const;
    void setSavedFilterPresets(const QList<FilterParams>& presets);
    
    void addFilterPreset(const FilterParams& params, const QString& name);
    void removeFilterPreset(const QString& name);

    // AI检测配置
    bool faceSwapDetectionEnabled() const;
    void setFaceSwapDetectionEnabled(bool enabled);
    
    bool voiceSynthesisDetectionEnabled() const;
    void setVoiceSynthesisDetectionEnabled(bool enabled);
    
    bool contentAnalysisEnabled() const;
    void setContentAnalysisEnabled(bool enabled);
    
    float detectionThreshold() const;
    void setDetectionThreshold(float threshold);
    
    int detectionInterval() const;
    void setDetectionInterval(int interval);

    // 录制配置
    QString recordingPath() const;
    void setRecordingPath(const QString& path);
    
    QString recordingFormat() const;
    void setRecordingFormat(const QString& format);
    
    int recordingQuality() const;
    void setRecordingQuality(int quality);
    
    bool autoStartRecording() const;
    void setAutoStartRecording(bool autoStart);

    // 网络配置
    int connectionTimeout() const;
    void setConnectionTimeout(int timeout);
    
    int reconnectInterval() const;
    void setReconnectInterval(int interval);
    
    int maxReconnectAttempts() const;
    void setMaxReconnectAttempts(int attempts);
    
    bool useProxy() const;
    void setUseProxy(bool use);
    
    QString proxyHost() const;
    void setProxyHost(const QString& host);
    
    int proxyPort() const;
    void setProxyPort(int port);
    
    QString proxyUser() const;
    void setProxyUser(const QString& user);
    
    QString proxyPassword() const;
    void setProxyPassword(const QString& password);

    // 日志配置
    LogLevel logLevel() const;
    void setLogLevel(LogLevel level);
    
    QString logPath() const;
    void setLogPath(const QString& path);
    
    int maxLogFiles() const;
    void setMaxLogFiles(int maxFiles);
    
    int maxLogFileSize() const;
    void setMaxLogFileSize(int maxSize);

    // 高级配置
    bool enableHardwareAcceleration() const;
    void setEnableHardwareAcceleration(bool enable);
    
    bool enableGPUProcessing() const;
    void setEnableGPUProcessing(bool enable);
    
    int maxConcurrentConnections() const;
    void setMaxConcurrentConnections(int max);
    
    bool enableTelemetry() const;
    void setEnableTelemetry(bool enable);
    
    bool enableCrashReporting() const;
    void setEnableCrashReporting(bool enable);

    // 配置验证
    bool validateConfig() const;
    QStringList getConfigErrors() const;

    // 配置导入导出
    bool exportConfig(const QString& filePath) const;
    bool importConfig(const QString& filePath);

    // 获取所有配置
    QJsonObject getAllConfig() const;
    void setAllConfig(const QJsonObject& config);

public slots:
    void reloadConfig();
    void saveConfigAsync();

signals:
    void configChanged();
    void configLoaded();
    void configSaved();
    void configError(const QString& error);

private:
    // 内部函数
    void setupDefaults();
    void validateAndFixConfig();
    QString getDefaultConfigPath() const;
    QJsonObject createDefaultConfig() const;
    
    // 配置文件处理
    bool loadFromFile(const QString& filePath);
    bool saveToFile(const QString& filePath) const;
    
    // 配置迁移
    void migrateConfig(const QJsonObject& oldConfig);
    int getConfigVersion(const QJsonObject& config) const;

private:
    QJsonObject config_;
    QString configPath_;
    bool initialized_;
    mutable QMutex configMutex_;
    
    // 默认值
    static const QJsonObject defaultConfig_;
    static const int currentConfigVersion_;
};

} // namespace VideoCallSystem
