#include "utils/json_helper.h"
#include <QJsonParseError>

QJsonObject JsonHelper::parseObject(const QString &jsonString, bool *ok)
{
    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(jsonString.toUtf8(), &error);
    
    if (ok) {
        *ok = (error.error == QJsonParseError::NoError && doc.isObject());
    }
    
    return doc.object();
}

QJsonArray JsonHelper::parseArray(const QString &jsonString, bool *ok)
{
    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(jsonString.toUtf8(), &error);
    
    if (ok) {
        *ok = (error.error == QJsonParseError::NoError && doc.isArray());
    }
    
    return doc.array();
}

QString JsonHelper::stringify(const QJsonObject &obj, bool compact)
{
    QJsonDocument doc(obj);
    return QString::fromUtf8(doc.toJson(compact ? QJsonDocument::Compact : QJsonDocument::Indented));
}

QString JsonHelper::stringify(const QJsonArray &arr, bool compact)
{
    QJsonDocument doc(arr);
    return QString::fromUtf8(doc.toJson(compact ? QJsonDocument::Compact : QJsonDocument::Indented));
}

QJsonObject JsonHelper::fromVariantMap(const QVariantMap &map)
{
    return QJsonObject::fromVariantMap(map);
}

QVariantMap JsonHelper::toVariantMap(const QJsonObject &obj)
{
    return obj.toVariantMap();
}

QJsonValue JsonHelper::getValue(const QJsonObject &obj, const QString &path, const QJsonValue &defaultValue)
{
    QStringList keys = path.split('.');
    QJsonValue current = obj;
    
    for (const QString &key : keys) {
        if (!current.isObject()) {
            return defaultValue;
        }
        current = current.toObject()[key];
        if (current.isUndefined() || current.isNull()) {
            return defaultValue;
        }
    }
    
    return current;
}

void JsonHelper::setValue(QJsonObject &obj, const QString &path, const QJsonValue &value)
{
    QStringList keys = path.split('.');
    if (keys.isEmpty()) {
        return;
    }
    
    QJsonObject *current = &obj;
    
    for (int i = 0; i < keys.size() - 1; ++i) {
        const QString &key = keys[i];
        if (!current->contains(key) || !(*current)[key].isObject()) {
            current->insert(key, QJsonObject());
        }
        QJsonValue val = (*current)[key];
        QJsonObject nested = val.toObject();
        current->insert(key, nested);
        current = &nested;
    }
    
    current->insert(keys.last(), value);
}

bool JsonHelper::hasKey(const QJsonObject &obj, const QString &key)
{
    return obj.contains(key);
}

QJsonObject JsonHelper::merge(const QJsonObject &obj1, const QJsonObject &obj2)
{
    QJsonObject result = obj1;
    
    for (auto it = obj2.begin(); it != obj2.end(); ++it) {
        if (result.contains(it.key()) && result[it.key()].isObject() && it.value().isObject()) {
            // Recursively merge nested objects
            result[it.key()] = merge(result[it.key()].toObject(), it.value().toObject());
        } else {
            result[it.key()] = it.value();
        }
    }
    
    return result;
}

