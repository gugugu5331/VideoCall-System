#include "ui/login_controller.h"
#include "application.h"
#include "services/auth_service.h"
#include "utils/logger.h"

LoginController::LoginController(QObject *parent)
    : QObject(parent)
{
}

LoginController::~LoginController()
{
}

void LoginController::login(const QString &username, const QString &password)
{
    if (username.isEmpty() || password.isEmpty()) {
        emit loginFailed("用户名和密码不能为空");
        return;
    }
    
    LOG_INFO("Login attempt for user: " + username);
    
    AuthService *authService = Application::instance()->authService();
    
    connect(authService, &AuthService::loginSuccess, this, [this]() {
        emit loginSuccess();
    }, Qt::SingleShotConnection);
    
    connect(authService, &AuthService::loginFailed, this, [this](const QString &error) {
        emit loginFailed(error);
    }, Qt::SingleShotConnection);
    
    authService->login(username, password);
}

void LoginController::registerUser(const QString &username, const QString &email,
                                  const QString &password, const QString &fullName)
{
    if (username.isEmpty() || email.isEmpty() || password.isEmpty()) {
        emit registerFailed("所有字段都必须填写");
        return;
    }
    
    LOG_INFO("Registration attempt for user: " + username);
    
    AuthService *authService = Application::instance()->authService();
    
    connect(authService, &AuthService::registerSuccess, this, [this]() {
        emit registerSuccess();
    }, Qt::SingleShotConnection);
    
    connect(authService, &AuthService::registerFailed, this, [this](const QString &error) {
        emit registerFailed(error);
    }, Qt::SingleShotConnection);
    
    authService->registerUser(username, email, password, fullName);
}

