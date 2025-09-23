#pragma once

#include <QtCore/QObject>
#include <QtCore/QString>
#include <QtCore/QStringList>
#include <QtCore/QVariant>
#include <QtCore/QVariantMap>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonDocument>
#include <QtCore/QJsonArray>
#include <QtCore/QTimer>
#include <QtCore/QThread>
#include <QtCore/QMutex>
#include <QtCore/QMutexLocker>
#include <QtCore/QWaitCondition>
#include <QtCore/QDateTime>
#include <QtCore/QUuid>
#include <QtCore/QUrl>
#include <QtCore/QDir>
#include <QtCore/QStandardPaths>
#include <QtCore/QSettings>
#include <QtCore/QDebug>

#include <QtWidgets/QApplication>
#include <QtWidgets/QMainWindow>
#include <QtWidgets/QWidget>
#include <QtWidgets/QVBoxLayout>
#include <QtWidgets/QHBoxLayout>
#include <QtWidgets/QGridLayout>
#include <QtWidgets/QPushButton>
#include <QtWidgets/QLabel>
#include <QtWidgets/QLineEdit>
#include <QtWidgets/QTextEdit>
#include <QtWidgets/QComboBox>
#include <QtWidgets/QSlider>
#include <QtWidgets/QSpinBox>
#include <QtWidgets/QCheckBox>
#include <QtWidgets/QRadioButton>
#include <QtWidgets/QGroupBox>
#include <QtWidgets/QTabWidget>
#include <QtWidgets/QSplitter>
#include <QtWidgets/QScrollArea>
#include <QtWidgets/QListWidget>
#include <QtWidgets/QTreeWidget>
#include <QtWidgets/QTableWidget>
#include <QtWidgets/QProgressBar>
#include <QtWidgets/QStatusBar>
#include <QtWidgets/QMenuBar>
#include <QtWidgets/QToolBar>
#include <QtWidgets/QDialog>
#include <QtWidgets/QMessageBox>
#include <QtWidgets/QFileDialog>
#include <QtWidgets/QColorDialog>

#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkRequest>
#include <QtNetwork/QNetworkReply>
#include <QtWebSockets/QWebSocket>

#include <QtMultimedia/QCamera>
#include <QtMultimedia/QMediaCaptureSession>
#include <QtMultimedia/QVideoSink>
#include <QtMultimedia/QAudioInput>
#include <QtMultimedia/QAudioOutput>
#include <QtMultimedia/QMediaRecorder>

#include <QtOpenGL/QOpenGLWidget>
#include <QtOpenGL/QOpenGLFunctions>
#include <QtOpenGL/QOpenGLShaderProgram>
#include <QtOpenGL/QOpenGLTexture>
#include <QtOpenGL/QOpenGLBuffer>
#include <QtOpenGL/QOpenGLVertexArrayObject>

#include <QtCharts/QChart>
#include <QtCharts/QChartView>
#include <QtCharts/QLineSeries>
#include <QtCharts/QValueAxis>

#include <QtSql/QSqlDatabase>
#include <QtSql/QSqlQuery>
#include <QtSql/QSqlError>

// OpenCV
#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <opencv2/objdetect.hpp>

// ZeroMQ (如果可用)
#ifdef HAS_ZMQ
#include <zmq.h>
#endif

// 标准库
#include <memory>
#include <vector>
#include <map>
#include <unordered_map>
#include <set>
#include <queue>
#include <stack>
#include <string>
#include <functional>
#include <algorithm>
#include <chrono>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <future>

namespace VideoCallSystem {

// 应用程序常量
namespace Constants {
    // 应用信息
    constexpr const char* APP_NAME = "VideoCall System Client";
    constexpr const char* APP_VERSION = "1.0.0";
    constexpr const char* APP_ORGANIZATION = "VideoCall System";
    constexpr const char* APP_DOMAIN = "videocall.system";

    // 网络配置
    constexpr const char* DEFAULT_SERVER_HOST = "localhost";
    constexpr int DEFAULT_SERVER_PORT = 8000;
    constexpr int DEFAULT_SIGNALING_PORT = 8080;
    constexpr int DEFAULT_AI_SERVICE_PORT = 5000;
    constexpr int DEFAULT_EDGE_INFRA_PORT = 9000;

    // 视频配置
    constexpr int DEFAULT_VIDEO_WIDTH = 640;
    constexpr int DEFAULT_VIDEO_HEIGHT = 480;
    constexpr int DEFAULT_VIDEO_FPS = 30;
    constexpr int MAX_VIDEO_WIDTH = 1920;
    constexpr int MAX_VIDEO_HEIGHT = 1080;

    // 音频配置
    constexpr int DEFAULT_AUDIO_SAMPLE_RATE = 44100;
    constexpr int DEFAULT_AUDIO_CHANNELS = 2;
    constexpr int DEFAULT_AUDIO_BUFFER_SIZE = 1024;

    // UI配置
    constexpr int DEFAULT_WINDOW_WIDTH = 1280;
    constexpr int DEFAULT_WINDOW_HEIGHT = 720;
    constexpr int MIN_WINDOW_WIDTH = 800;
    constexpr int MIN_WINDOW_HEIGHT = 600;
}

// 枚举定义
enum class ConnectionState {
    Disconnected = 0,
    Connecting,
    Connected,
    Reconnecting,
    Error
};

enum class MeetingState {
    Idle = 0,
    Joining,
    InMeeting,
    Leaving,
    Error
};

enum class MediaState {
    Stopped = 0,
    Starting,
    Active,
    Paused,
    Error
};

enum class DetectionType {
    None = 0,
    FaceSwap,
    VoiceSynthesis,
    ContentAnalysis,
    All
};

enum class FilterType {
    None = 0,
    Blur,
    Sharpen,
    EdgeDetection,
    Emboss,
    Sepia,
    Vintage,
    Beauty,
    Cartoon,
    Sketch,
    Neon,
    Thermal,
    NightVision,
    Fisheye,
    Mirror,
    Pixelate
};

enum class LogLevel {
    Debug = 0,
    Info,
    Warning,
    Error,
    Critical
};

// 数据结构
struct UserInfo {
    QString userId;
    QString userName;
    QString email;
    QString avatar;
    bool isOnline = false;
    QDateTime lastSeen;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["userId"] = userId;
        obj["userName"] = userName;
        obj["email"] = email;
        obj["avatar"] = avatar;
        obj["isOnline"] = isOnline;
        obj["lastSeen"] = lastSeen.toString(Qt::ISODate);
        return obj;
    }
    
    static UserInfo fromJson(const QJsonObject& obj) {
        UserInfo info;
        info.userId = obj["userId"].toString();
        info.userName = obj["userName"].toString();
        info.email = obj["email"].toString();
        info.avatar = obj["avatar"].toString();
        info.isOnline = obj["isOnline"].toBool();
        info.lastSeen = QDateTime::fromString(obj["lastSeen"].toString(), Qt::ISODate);
        return info;
    }
};

struct MeetingInfo {
    QString meetingId;
    QString title;
    QString description;
    QString hostId;
    QDateTime startTime;
    QDateTime endTime;
    QStringList participants;
    bool isRecording = false;
    bool isLocked = false;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["meetingId"] = meetingId;
        obj["title"] = title;
        obj["description"] = description;
        obj["hostId"] = hostId;
        obj["startTime"] = startTime.toString(Qt::ISODate);
        obj["endTime"] = endTime.toString(Qt::ISODate);
        obj["participants"] = QJsonArray::fromStringList(participants);
        obj["isRecording"] = isRecording;
        obj["isLocked"] = isLocked;
        return obj;
    }
    
    static MeetingInfo fromJson(const QJsonObject& obj) {
        MeetingInfo info;
        info.meetingId = obj["meetingId"].toString();
        info.title = obj["title"].toString();
        info.description = obj["description"].toString();
        info.hostId = obj["hostId"].toString();
        info.startTime = QDateTime::fromString(obj["startTime"].toString(), Qt::ISODate);
        info.endTime = QDateTime::fromString(obj["endTime"].toString(), Qt::ISODate);
        
        QJsonArray participantsArray = obj["participants"].toArray();
        for (const auto& value : participantsArray) {
            info.participants.append(value.toString());
        }
        
        info.isRecording = obj["isRecording"].toBool();
        info.isLocked = obj["isLocked"].toBool();
        return info;
    }
};

struct DetectionResult {
    QString detectionId;
    DetectionType type;
    bool isFake = false;
    double confidence = 0.0;
    double riskScore = 0.0;
    QString details;
    QDateTime timestamp;
    QVariantMap metadata;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["detectionId"] = detectionId;
        obj["type"] = static_cast<int>(type);
        obj["isFake"] = isFake;
        obj["confidence"] = confidence;
        obj["riskScore"] = riskScore;
        obj["details"] = details;
        obj["timestamp"] = timestamp.toString(Qt::ISODate);
        
        QJsonObject metaObj;
        for (auto it = metadata.begin(); it != metadata.end(); ++it) {
            metaObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["metadata"] = metaObj;
        
        return obj;
    }
    
    static DetectionResult fromJson(const QJsonObject& obj) {
        DetectionResult result;
        result.detectionId = obj["detectionId"].toString();
        result.type = static_cast<DetectionType>(obj["type"].toInt());
        result.isFake = obj["isFake"].toBool();
        result.confidence = obj["confidence"].toDouble();
        result.riskScore = obj["riskScore"].toDouble();
        result.details = obj["details"].toString();
        result.timestamp = QDateTime::fromString(obj["timestamp"].toString(), Qt::ISODate);
        
        QJsonObject metaObj = obj["metadata"].toObject();
        for (auto it = metaObj.begin(); it != metaObj.end(); ++it) {
            result.metadata[it.key()] = it.value().toVariant();
        }
        
        return result;
    }
};

struct FilterParams {
    FilterType type = FilterType::None;
    float intensity = 1.0f;
    float brightness = 0.0f;
    float contrast = 1.0f;
    float saturation = 1.0f;
    float hue = 0.0f;
    float gamma = 1.0f;
    QColor colorBalance = QColor(255, 255, 255);
    bool enabled = true;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["type"] = static_cast<int>(type);
        obj["intensity"] = intensity;
        obj["brightness"] = brightness;
        obj["contrast"] = contrast;
        obj["saturation"] = saturation;
        obj["hue"] = hue;
        obj["gamma"] = gamma;
        obj["colorBalance"] = colorBalance.name();
        obj["enabled"] = enabled;
        return obj;
    }
    
    static FilterParams fromJson(const QJsonObject& obj) {
        FilterParams params;
        params.type = static_cast<FilterType>(obj["type"].toInt());
        params.intensity = obj["intensity"].toDouble();
        params.brightness = obj["brightness"].toDouble();
        params.contrast = obj["contrast"].toDouble();
        params.saturation = obj["saturation"].toDouble();
        params.hue = obj["hue"].toDouble();
        params.gamma = obj["gamma"].toDouble();
        params.colorBalance = QColor(obj["colorBalance"].toString());
        params.enabled = obj["enabled"].toBool();
        return params;
    }
};

// 工具函数
inline QString stateToString(ConnectionState state) {
    switch (state) {
        case ConnectionState::Disconnected: return "Disconnected";
        case ConnectionState::Connecting: return "Connecting";
        case ConnectionState::Connected: return "Connected";
        case ConnectionState::Reconnecting: return "Reconnecting";
        case ConnectionState::Error: return "Error";
        default: return "Unknown";
    }
}

inline QString stateToString(MeetingState state) {
    switch (state) {
        case MeetingState::Idle: return "Idle";
        case MeetingState::Joining: return "Joining";
        case MeetingState::InMeeting: return "In Meeting";
        case MeetingState::Leaving: return "Leaving";
        case MeetingState::Error: return "Error";
        default: return "Unknown";
    }
}

inline QString filterTypeToString(FilterType type) {
    switch (type) {
        case FilterType::None: return "None";
        case FilterType::Blur: return "Blur";
        case FilterType::Sharpen: return "Sharpen";
        case FilterType::EdgeDetection: return "Edge Detection";
        case FilterType::Emboss: return "Emboss";
        case FilterType::Sepia: return "Sepia";
        case FilterType::Vintage: return "Vintage";
        case FilterType::Beauty: return "Beauty";
        case FilterType::Cartoon: return "Cartoon";
        case FilterType::Sketch: return "Sketch";
        case FilterType::Neon: return "Neon";
        case FilterType::Thermal: return "Thermal";
        case FilterType::NightVision: return "Night Vision";
        case FilterType::Fisheye: return "Fisheye";
        case FilterType::Mirror: return "Mirror";
        case FilterType::Pixelate: return "Pixelate";
        default: return "Unknown";
    }
}

// 错误处理宏
#define SAFE_CALL(func) \
    try { \
        func; \
    } catch (const std::exception& e) { \
        qCritical() << "Exception in" << __FUNCTION__ << ":" << e.what(); \
    } catch (...) { \
        qCritical() << "Unknown exception in" << __FUNCTION__; \
    }

#define LOG_DEBUG(msg) qDebug() << "[DEBUG]" << __FUNCTION__ << ":" << msg
#define LOG_INFO(msg) qInfo() << "[INFO]" << __FUNCTION__ << ":" << msg
#define LOG_WARNING(msg) qWarning() << "[WARNING]" << __FUNCTION__ << ":" << msg
#define LOG_ERROR(msg) qCritical() << "[ERROR]" << __FUNCTION__ << ":" << msg

} // namespace VideoCallSystem
