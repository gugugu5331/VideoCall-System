#ifndef LOGIN_CONTROLLER_H
#define LOGIN_CONTROLLER_H

#include <QObject>

class LoginController : public QObject
{
    Q_OBJECT

public:
    explicit LoginController(QObject *parent = nullptr);
    ~LoginController();

    Q_INVOKABLE void login(const QString &username, const QString &password);
    Q_INVOKABLE void registerUser(const QString &username, const QString &email, 
                                  const QString &password, const QString &fullName);

signals:
    void loginSuccess();
    void loginFailed(const QString &error);
    void registerSuccess();
    void registerFailed(const QString &error);
};

#endif // LOGIN_CONTROLLER_H

