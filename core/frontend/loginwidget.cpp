#include "loginwidget.h"
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QCheckBox>
#include <QGroupBox>
#include <QFrame>
#include <QJsonObject>
#include <QJsonDocument>
#include <QMessageBox>

LoginWidget::LoginWidget(QWidget *parent)
    : QWidget(parent)
{
    setupUI();
    setupStyles();
}

LoginWidget::~LoginWidget()
{
}

void LoginWidget::setupUI()
{
    m_mainLayout = new QVBoxLayout(this);
    
    // Logo和标题
    m_logoLabel = new QLabel("🎥");
    m_logoLabel->setAlignment(Qt::AlignCenter);
    m_logoLabel->setStyleSheet("font-size: 48px; margin: 20px;");
    
    m_titleLabel = new QLabel("音视频通话系统");
    m_titleLabel->setAlignment(Qt::AlignCenter);
    m_titleLabel->setStyleSheet("font-size: 24px; font-weight: bold; margin: 10px;");
    
    m_mainLayout->addWidget(m_logoLabel);
    m_mainLayout->addWidget(m_titleLabel);
    
    // 登录组
    m_loginGroup = new QGroupBox("用户登录");
    QVBoxLayout *loginLayout = new QVBoxLayout(m_loginGroup);
    
    // 用户名
    QLabel *usernameLabel = new QLabel("用户名:");
    m_usernameEdit = new QLineEdit();
    m_usernameEdit->setPlaceholderText("请输入用户名");
    
    // 密码
    QLabel *passwordLabel = new QLabel("密码:");
    m_passwordEdit = new QLineEdit();
    m_passwordEdit->setPlaceholderText("请输入密码");
    m_passwordEdit->setEchoMode(QLineEdit::Password);
    
    // 记住我
    m_rememberMeCheckBox = new QCheckBox("记住我");
    
    // 按钮
    QHBoxLayout *buttonLayout = new QHBoxLayout();
    m_loginButton = new QPushButton("登录");
    m_registerButton = new QPushButton("注册");
    m_forgotPasswordButton = new QPushButton("忘记密码");
    
    buttonLayout->addWidget(m_loginButton);
    buttonLayout->addWidget(m_registerButton);
    buttonLayout->addWidget(m_forgotPasswordButton);
    
    // 添加到登录组
    loginLayout->addWidget(usernameLabel);
    loginLayout->addWidget(m_usernameEdit);
    loginLayout->addWidget(passwordLabel);
    loginLayout->addWidget(m_passwordEdit);
    loginLayout->addWidget(m_rememberMeCheckBox);
    loginLayout->addLayout(buttonLayout);
    
    m_mainLayout->addWidget(m_loginGroup);
    
    // 连接信号
    connect(m_loginButton, &QPushButton::clicked, this, &LoginWidget::onLoginClicked);
    connect(m_registerButton, &QPushButton::clicked, this, &LoginWidget::onRegisterClicked);
    connect(m_forgotPasswordButton, &QPushButton::clicked, this, &LoginWidget::onForgotPasswordClicked);
    
    // 回车键登录
    connect(m_passwordEdit, &QLineEdit::returnPressed, this, &LoginWidget::onLoginClicked);
}

void LoginWidget::setupStyles()
{
    setStyleSheet(R"(
        QWidget {
            background-color: #2b2b2b;
            color: white;
        }
        QGroupBox {
            font-weight: bold;
            border: 2px solid #555;
            border-radius: 5px;
            margin-top: 10px;
            padding-top: 10px;
        }
        QLineEdit {
            padding: 8px;
            border: 1px solid #555;
            border-radius: 3px;
            background-color: #3b3b3b;
        }
        QPushButton {
            padding: 8px 16px;
            border: 1px solid #555;
            border-radius: 3px;
            background-color: #4a4a4a;
        }
        QPushButton:hover {
            background-color: #5a5a5a;
        }
        QPushButton:pressed {
            background-color: #3a3a3a;
        }
    )");
}

void LoginWidget::onLoginClicked()
{
    QString username = m_usernameEdit->text().trimmed();
    QString password = m_passwordEdit->text();
    
    if (username.isEmpty() || password.isEmpty()) {
        QMessageBox::warning(this, "登录失败", "请输入用户名和密码");
        return;
    }
    
    // 模拟登录成功
    QJsonObject userInfo;
    userInfo["username"] = username;
    userInfo["email"] = username + "@example.com";
    userInfo["token"] = "mock_token_" + QString::number(QDateTime::currentMSecsSinceEpoch());
    userInfo["id"] = 1;
    
    emit loginSuccess(userInfo);
}

void LoginWidget::onRegisterClicked()
{
    QMessageBox::information(this, "注册", "注册功能正在开发中...");
}

void LoginWidget::onForgotPasswordClicked()
{
    QMessageBox::information(this, "忘记密码", "密码重置功能正在开发中...");
} 