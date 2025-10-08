#ifndef JSON_HELPER_H
#define JSON_HELPER_H

#include <QJsonObject>
#include <QJsonArray>
#include <QJsonDocument>
#include <QString>
#include <QVariantMap>

class JsonHelper
{
public:
    // Parse JSON string to QJsonObject
    static QJsonObject parseObject(const QString &jsonString, bool *ok = nullptr);
    
    // Parse JSON string to QJsonArray
    static QJsonArray parseArray(const QString &jsonString, bool *ok = nullptr);
    
    // Convert QJsonObject to QString
    static QString stringify(const QJsonObject &obj, bool compact = false);
    
    // Convert QJsonArray to QString
    static QString stringify(const QJsonArray &arr, bool compact = false);
    
    // Convert QVariantMap to QJsonObject
    static QJsonObject fromVariantMap(const QVariantMap &map);
    
    // Convert QJsonObject to QVariantMap
    static QVariantMap toVariantMap(const QJsonObject &obj);
    
    // Get nested value from JSON object
    static QJsonValue getValue(const QJsonObject &obj, const QString &path, const QJsonValue &defaultValue = QJsonValue());
    
    // Set nested value in JSON object
    static void setValue(QJsonObject &obj, const QString &path, const QJsonValue &value);
    
    // Check if JSON object has key
    static bool hasKey(const QJsonObject &obj, const QString &key);
    
    // Merge two JSON objects
    static QJsonObject merge(const QJsonObject &obj1, const QJsonObject &obj2);
};

#endif // JSON_HELPER_H

