#ifndef LOGGER_H
#define LOGGER_H

#include <QObject>
#include <QString>
#include <QFile>
#include <QTextStream>
#include <QMutex>

enum class LogLevel {
    Debug,
    Info,
    Warning,
    Error,
    Critical
};

class Logger : public QObject
{
    Q_OBJECT

public:
    explicit Logger(QObject *parent = nullptr);
    ~Logger();

    static Logger* instance();

    void setLogLevel(LogLevel level);
    void setLogFile(const QString &filePath);

    void debug(const QString &message, const QString &category = QString());
    void info(const QString &message, const QString &category = QString());
    void warning(const QString &message, const QString &category = QString());
    void error(const QString &message, const QString &category = QString());
    void critical(const QString &message, const QString &category = QString());

signals:
    void logMessage(LogLevel level, const QString &message);

private:
    void log(LogLevel level, const QString &message, const QString &category);
    QString levelToString(LogLevel level) const;
    QString formatMessage(LogLevel level, const QString &message, const QString &category) const;

private:
    static Logger* s_instance;
    LogLevel m_logLevel;
    QFile m_logFile;
    QTextStream m_logStream;
    QMutex m_mutex;
};

// Convenience macros
#define LOG_DEBUG(msg) Logger::instance()->debug(msg, __FUNCTION__)
#define LOG_INFO(msg) Logger::instance()->info(msg, __FUNCTION__)
#define LOG_WARNING(msg) Logger::instance()->warning(msg, __FUNCTION__)
#define LOG_ERROR(msg) Logger::instance()->error(msg, __FUNCTION__)
#define LOG_CRITICAL(msg) Logger::instance()->critical(msg, __FUNCTION__)

#endif // LOGGER_H

