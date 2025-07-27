#ifndef LOGINWIDGET_H
#define LOGINWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QCheckBox>
#include <QGroupBox>
#include <QFrame>

class LoginWidget : public QWidget
{
    Q_OBJECT

public:
    explicit LoginWidget(QWidget *parent = nullptr);
    ~LoginWidget();

signals:
    void loginSuccess(const QJsonObject &userInfo);
    void loginFailed(const QString &error);

private slots:
    void onLoginClicked();
    void onRegisterClicked();
    void onForgotPasswordClicked();

private:
    void setupUI();
    void setupStyles();

private:
    QVBoxLayout *m_mainLayout;
    QLabel *m_logoLabel;
    QLabel *m_titleLabel;
    QGroupBox *m_loginGroup;
    QLineEdit *m_usernameEdit;
    QLineEdit *m_passwordEdit;
    QCheckBox *m_rememberMeCheckBox;
    QPushButton *m_loginButton;
    QPushButton *m_registerButton;
    QPushButton *m_forgotPasswordButton;
};

#endif // LOGINWIDGET_H 