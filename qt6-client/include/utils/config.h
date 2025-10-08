#ifndef CONFIG_H
#define CONFIG_H

#include <QObject>
#include <QString>
#include <QVariantMap>
#include <QJsonObject>

class Config : public QObject
{
    Q_OBJECT

public:
    explicit Config(QObject *parent = nullptr);
    ~Config();

    bool load(const QString &filePath);
    bool save(const QString &filePath);

    // Getters
    QString appName() const;
    QString appVersion() const;
    QString apiBaseUrl() const;
    QString wsUrl() const;
    int apiTimeout() const;
    QVariantMap webrtcConfig() const;
    QVariantMap uiConfig() const;
    QVariantMap aiConfig() const;

    // Generic getter
    QVariant value(const QString &key, const QVariant &defaultValue = QVariant()) const;

    // Setters
    void setValue(const QString &key, const QVariant &value);

signals:
    void configChanged();

private:
    QVariant getNestedValue(const QJsonObject &obj, const QString &path) const;
    void setNestedValue(QJsonObject &obj, const QString &path, const QVariant &value);

private:
    QJsonObject m_config;
    QString m_filePath;
};

#endif // CONFIG_H

