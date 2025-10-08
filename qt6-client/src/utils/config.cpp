#include "utils/config.h"
#include <QFile>
#include <QJsonDocument>
#include <QJsonObject>
#include <QDebug>

Config::Config(QObject *parent)
    : QObject(parent)
{
}

Config::~Config()
{
}

bool Config::load(const QString &filePath)
{
    m_filePath = filePath;
    
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        qWarning() << "Failed to open config file:" << filePath;
        return false;
    }

    QByteArray data = file.readAll();
    file.close();

    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(data, &error);
    
    if (error.error != QJsonParseError::NoError) {
        qWarning() << "Failed to parse config file:" << error.errorString();
        return false;
    }

    m_config = doc.object();
    emit configChanged();
    
    return true;
}

bool Config::save(const QString &filePath)
{
    QString path = filePath.isEmpty() ? m_filePath : filePath;
    
    QFile file(path);
    if (!file.open(QIODevice::WriteOnly)) {
        qWarning() << "Failed to open config file for writing:" << path;
        return false;
    }

    QJsonDocument doc(m_config);
    file.write(doc.toJson(QJsonDocument::Indented));
    file.close();
    
    return true;
}

QString Config::appName() const
{
    return value("app.name", "Meeting System").toString();
}

QString Config::appVersion() const
{
    return value("app.version", "1.0.0").toString();
}

QString Config::apiBaseUrl() const
{
    return value("server.api_url", "http://localhost:8080/api").toString();
}

QString Config::wsUrl() const
{
    return value("server.websocket_url", "ws://localhost:8080/ws").toString();
}

int Config::apiTimeout() const
{
    return value("api.timeout", 30000).toInt();
}

QVariantMap Config::webrtcConfig() const
{
    return value("webrtc").toMap();
}

QVariantMap Config::uiConfig() const
{
    return value("ui").toMap();
}

QVariantMap Config::aiConfig() const
{
    return value("ai").toMap();
}

QVariant Config::value(const QString &key, const QVariant &defaultValue) const
{
    QVariant val = getNestedValue(m_config, key);
    return val.isNull() ? defaultValue : val;
}

void Config::setValue(const QString &key, const QVariant &value)
{
    setNestedValue(m_config, key, QJsonValue::fromVariant(value));
    emit configChanged();
}

QVariant Config::getNestedValue(const QJsonObject &obj, const QString &path) const
{
    QStringList keys = path.split('.');
    QJsonValue current = obj;
    
    for (const QString &key : keys) {
        if (!current.isObject()) {
            return QJsonValue::Undefined;
        }
        current = current.toObject()[key];
    }
    
    return current.toVariant();
}

void Config::setNestedValue(QJsonObject &obj, const QString &path, const QVariant &value)
{
    QStringList keys = path.split('.');
    QJsonObject *current = &obj;
    
    for (int i = 0; i < keys.size() - 1; ++i) {
        const QString &key = keys[i];
        if (!current->contains(key) || !(*current)[key].isObject()) {
            current->insert(key, QJsonObject());
        }
        QJsonObject nested = (*current)[key].toObject();
        current->insert(key, nested);
        current = &nested;
    }
    
    current->insert(keys.last(), QJsonValue::fromVariant(value));
}

