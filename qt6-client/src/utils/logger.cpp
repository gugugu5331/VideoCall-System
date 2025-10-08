#include "utils/logger.h"
#include <QDateTime>
#include <QDebug>
#include <iostream>

Logger* Logger::s_instance = nullptr;

Logger::Logger(QObject *parent)
    : QObject(parent)
    , m_logLevel(LogLevel::Info)
{
    s_instance = this;
}

Logger::~Logger()
{
    if (m_logFile.isOpen()) {
        m_logFile.close();
    }
    s_instance = nullptr;
}

Logger* Logger::instance()
{
    return s_instance;
}

void Logger::setLogLevel(LogLevel level)
{
    m_logLevel = level;
}

void Logger::setLogFile(const QString &filePath)
{
    QMutexLocker locker(&m_mutex);
    
    if (m_logFile.isOpen()) {
        m_logFile.close();
    }
    
    m_logFile.setFileName(filePath);
    if (m_logFile.open(QIODevice::WriteOnly | QIODevice::Append | QIODevice::Text)) {
        m_logStream.setDevice(&m_logFile);
    } else {
        qWarning() << "Failed to open log file:" << filePath;
    }
}

void Logger::debug(const QString &message, const QString &category)
{
    log(LogLevel::Debug, message, category);
}

void Logger::info(const QString &message, const QString &category)
{
    log(LogLevel::Info, message, category);
}

void Logger::warning(const QString &message, const QString &category)
{
    log(LogLevel::Warning, message, category);
}

void Logger::error(const QString &message, const QString &category)
{
    log(LogLevel::Error, message, category);
}

void Logger::critical(const QString &message, const QString &category)
{
    log(LogLevel::Critical, message, category);
}

void Logger::log(LogLevel level, const QString &message, const QString &category)
{
    if (level < m_logLevel) {
        return;
    }
    
    QString formattedMessage = formatMessage(level, message, category);
    
    // Emit signal
    emit logMessage(level, formattedMessage);
    
    // Write to console
    std::cout << formattedMessage.toStdString() << std::endl;
    
    // Write to file
    if (m_logFile.isOpen()) {
        QMutexLocker locker(&m_mutex);
        m_logStream << formattedMessage << "\n";
        m_logStream.flush();
    }
}

QString Logger::levelToString(LogLevel level) const
{
    switch (level) {
        case LogLevel::Debug:    return "DEBUG";
        case LogLevel::Info:     return "INFO";
        case LogLevel::Warning:  return "WARNING";
        case LogLevel::Error:    return "ERROR";
        case LogLevel::Critical: return "CRITICAL";
        default:                 return "UNKNOWN";
    }
}

QString Logger::formatMessage(LogLevel level, const QString &message, const QString &category) const
{
    QString timestamp = QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz");
    QString levelStr = levelToString(level);
    
    if (category.isEmpty()) {
        return QString("[%1] [%2] %3").arg(timestamp, levelStr, message);
    } else {
        return QString("[%1] [%2] [%3] %4").arg(timestamp, levelStr, category, message);
    }
}

