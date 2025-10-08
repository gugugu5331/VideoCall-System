#ifndef AUTH_SERVICE_H
#define AUTH_SERVICE_H

#include <QObject>
#include <QString>
#include <memory>
#include "network/api_client.h"
#include "models/user.h"

class AuthService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool isAuthenticated READ isAuthenticated NOTIFY authenticationChanged)
    Q_PROPERTY(User* currentUser READ currentUser NOTIFY currentUserChanged)

public:
    explicit AuthService(ApiClient *apiClient, QObject *parent = nullptr);
    ~AuthService();

    // Authentication status
    bool isAuthenticated() const { return m_isAuthenticated; }
    User* currentUser() const { return m_currentUser.get(); }
    QString authToken() const { return m_authToken; }

    // Authentication methods
    Q_INVOKABLE void getCsrfToken();
    Q_INVOKABLE void login(const QString &username, const QString &password);
    Q_INVOKABLE void registerUser(const QString &username, const QString &email,
                                  const QString &password, const QString &fullName);
    Q_INVOKABLE void logout();
    Q_INVOKABLE void refreshToken();
    Q_INVOKABLE void requestPasswordReset(const QString &email);

    // Load/save credentials
    void loadCredentials();
    void saveCredentials();
    void clearCredentials();

signals:
    void authenticationChanged();
    void currentUserChanged();
    void csrfTokenReceived(const QString &token);
    void csrfTokenFailed(const QString &error);
    void loginSuccess();
    void loginFailed(const QString &error);
    void registerSuccess();
    void registerFailed(const QString &error);
    void logoutSuccess();
    void tokenRefreshed();
    void tokenRefreshFailed();
    void authTokenChanged();
    void isAuthenticatedChanged();
    void passwordResetSuccess();
    void passwordResetFailed(const QString &error);

private:
    void setAuthenticated(bool authenticated);
    void setCurrentUser(const QJsonObject &userData);
    void setAuthToken(const QString &token);
    void setRefreshToken(const QString &token);
    void setCsrfToken(const QString &token);

private:
    ApiClient *m_apiClient;
    std::unique_ptr<User> m_currentUser;
    QString m_authToken;
    QString m_refreshToken;
    QString m_csrfToken;
    bool m_isAuthenticated;
};

#endif // AUTH_SERVICE_H

