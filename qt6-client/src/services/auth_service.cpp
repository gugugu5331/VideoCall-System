#include "services/auth_service.h"
#include "utils/logger.h"
#include <QSettings>

AuthService::AuthService(ApiClient *apiClient, QObject *parent)
    : QObject(parent)
    , m_apiClient(apiClient)
    , m_isAuthenticated(false)
{
    m_currentUser = std::make_unique<User>();
    loadCredentials();
}

AuthService::~AuthService()
{
}

void AuthService::getCsrfToken()
{
    LOG_INFO("Fetching CSRF token");

    m_apiClient->getCsrfToken([this](const ApiResponse &response) {
        if (response.isSuccess()) {
            QString token = response.data["csrf_token"].toString();
            setCsrfToken(token);
            LOG_INFO("CSRF token received");
            emit csrfTokenReceived(token);
        } else {
            LOG_ERROR("Failed to get CSRF token: " + response.message);
            emit csrfTokenFailed(response.message);
        }
    });
}

void AuthService::login(const QString &username, const QString &password)
{
    LOG_INFO("Attempting login for user: " + username);
    
    m_apiClient->login(username, password, [this, username, password](const ApiResponse &response) {
        if (response.isSuccess()) {
            QString token = response.data["token"].toString();
            QString refreshToken = response.data["refresh_token"].toString();
            
            setAuthToken(token);
            setRefreshToken(refreshToken);
            
            // Update current user
            QJsonObject userData = response.data["user"].toObject();
            m_currentUser->fromJson(userData);
            
            setAuthenticated(true);

            // Save credentials if remember me
            saveCredentials();
            
            LOG_INFO("Login successful");
            emit loginSuccess();
        } else {
            LOG_ERROR("Login failed: " + response.message);
            emit loginFailed(response.message);
        }
    });
}

void AuthService::registerUser(const QString &username, const QString &email,
                              const QString &password, const QString &fullName)
{
    LOG_INFO("Attempting registration for user: " + username);
    
    m_apiClient->registerUser(username, email, password, fullName,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                LOG_INFO("Registration successful");
                emit registerSuccess();
            } else {
                LOG_ERROR("Registration failed: " + response.message);
                emit registerFailed(response.message);
            }
        });
}

void AuthService::logout()
{
    LOG_INFO("Logging out");

    setAuthToken("");
    setRefreshToken("");
    setAuthenticated(false);

    m_currentUser = std::make_unique<User>();

    emit logoutSuccess();
}

void AuthService::refreshToken()
{
    if (m_refreshToken.isEmpty()) {
        LOG_WARNING("No refresh token available");
        return;
    }

    LOG_INFO("Refreshing auth token");

    m_apiClient->refreshToken(m_refreshToken, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            QString newToken = response.data["token"].toString();
            setAuthToken(newToken);
            LOG_INFO("Token refreshed successfully");
            emit tokenRefreshed();
        } else {
            LOG_ERROR("Token refresh failed: " + response.message);
            emit tokenRefreshFailed();
            // Token refresh failed, need to re-login
            logout();
        }
    });
}

void AuthService::setAuthToken(const QString &token)
{
    if (m_authToken != token) {
        m_authToken = token;
        m_apiClient->setAuthToken(token);
        emit authTokenChanged();
    }
}

void AuthService::setRefreshToken(const QString &token)
{
    m_refreshToken = token;
}

void AuthService::setCsrfToken(const QString &token)
{
    m_csrfToken = token;
    m_apiClient->setCsrfToken(token);
}

void AuthService::setAuthenticated(bool authenticated)
{
    if (m_isAuthenticated != authenticated) {
        m_isAuthenticated = authenticated;
        emit authenticationChanged();
        emit isAuthenticatedChanged();
    }
}

void AuthService::setCurrentUser(const QJsonObject &userData)
{
    m_currentUser->fromJson(userData);
    emit currentUserChanged();
}

void AuthService::saveCredentials()
{
    QSettings settings("MeetingSystem", "Client");
    settings.setValue("auth_token", m_authToken);
    settings.setValue("refresh_token", m_refreshToken);
}

void AuthService::loadCredentials()
{
    QSettings settings("MeetingSystem", "Client");
    QString authToken = settings.value("auth_token").toString();
    QString refreshTokenStr = settings.value("refresh_token").toString();

    if (!authToken.isEmpty() && !refreshTokenStr.isEmpty()) {
        m_authToken = authToken;
        m_refreshToken = refreshTokenStr;
        setAuthenticated(true);
        // Try to refresh token to verify it's still valid
        this->refreshToken();
    }
}

void AuthService::clearCredentials()
{
    QSettings settings("MeetingSystem", "Client");
    settings.remove("auth_token");
    settings.remove("refresh_token");
}

void AuthService::requestPasswordReset(const QString &email)
{
    LOG_INFO("Requesting password reset for: " + email);

    // TODO: Implement API call when backend is ready
    // For now, just emit success
    emit passwordResetSuccess();

    /* Future implementation:
    m_apiClient->requestPasswordReset(email, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Password reset email sent");
            emit passwordResetSuccess();
        } else {
            LOG_ERROR("Password reset failed: " + response.message);
            emit passwordResetFailed(response.message);
        }
    });
    */
}

