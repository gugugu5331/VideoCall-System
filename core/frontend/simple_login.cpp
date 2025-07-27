#include <QApplication>
#include <QMainWindow>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QMessageBox>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonObject>
#include <iostream>

class SimpleLoginWindow : public QMainWindow
{
    Q_OBJECT

public:
    SimpleLoginWindow(QWidget *parent = nullptr) : QMainWindow(parent)
    {
        setWindowTitle("音视频通话系统 - 登录");
        setFixedSize(400, 300);
        
        // 创建中央部件
        QWidget *centralWidget = new QWidget(this);
        setCentralWidget(centralWidget);
        
        // 创建布局
        QVBoxLayout *mainLayout = new QVBoxLayout(centralWidget);
        
        // 标题
        QLabel *titleLabel = new QLabel("音视频通话系统", this);
        titleLabel->setAlignment(Qt::AlignCenter);
        titleLabel->setStyleSheet("font-size: 18px; font-weight: bold; margin: 20px;");
        mainLayout->addWidget(titleLabel);
        
        // 用户名输入
        QHBoxLayout *usernameLayout = new QHBoxLayout();
        QLabel *usernameLabel = new QLabel("用户名:", this);
        m_usernameEdit = new QLineEdit(this);
        m_usernameEdit->setPlaceholderText("请输入用户名");
        usernameLayout->addWidget(usernameLabel);
        usernameLayout->addWidget(m_usernameEdit);
        mainLayout->addLayout(usernameLayout);
        
        // 密码输入
        QHBoxLayout *passwordLayout = new QHBoxLayout();
        QLabel *passwordLabel = new QLabel("密码:", this);
        m_passwordEdit = new QLineEdit(this);
        m_passwordEdit->setPlaceholderText("请输入密码");
        m_passwordEdit->setEchoMode(QLineEdit::Password);
        passwordLayout->addWidget(passwordLabel);
        passwordLayout->addWidget(m_passwordEdit);
        mainLayout->addLayout(passwordLayout);
        
        // 登录按钮
        QHBoxLayout *buttonLayout = new QHBoxLayout();
        m_loginButton = new QPushButton("登录", this);
        m_registerButton = new QPushButton("注册", this);
        buttonLayout->addWidget(m_loginButton);
        buttonLayout->addWidget(m_registerButton);
        mainLayout->addLayout(buttonLayout);
        
        // 状态标签
        m_statusLabel = new QLabel("", this);
        m_statusLabel->setAlignment(Qt::AlignCenter);
        m_statusLabel->setStyleSheet("color: red;");
        mainLayout->addWidget(m_statusLabel);
        
        // 连接信号
        connect(m_loginButton, &QPushButton::clicked, this, &SimpleLoginWindow::onLoginClicked);
        connect(m_registerButton, &QPushButton::clicked, this, &SimpleLoginWindow::onRegisterClicked);
        
        // 网络管理器
        m_networkManager = new QNetworkAccessManager(this);
        
        // 设置样式
        setStyleSheet(R"(
            QMainWindow {
                background-color: #f0f0f0;
            }
            QLabel {
                font-size: 12px;
            }
            QLineEdit {
                padding: 8px;
                border: 1px solid #ccc;
                border-radius: 4px;
                font-size: 12px;
            }
            QPushButton {
                padding: 10px 20px;
                background-color: #0078d4;
                color: white;
                border: none;
                border-radius: 4px;
                font-size: 12px;
            }
            QPushButton:hover {
                background-color: #106ebe;
            }
            QPushButton:pressed {
                background-color: #005a9e;
            }
        )");
    }

private slots:
    void onLoginClicked()
    {
        QString username = m_usernameEdit->text().trimmed();
        QString password = m_passwordEdit->text().trimmed();
        
        if (username.isEmpty() || password.isEmpty()) {
            m_statusLabel->setText("请输入用户名和密码");
            return;
        }
        
        m_loginButton->setEnabled(false);
        m_statusLabel->setText("正在登录...");
        
        // 创建登录请求
        QJsonObject loginData;
        loginData["username"] = username;
        loginData["password"] = password;
        
        QJsonDocument doc(loginData);
        QByteArray data = doc.toJson();
        
        QNetworkRequest request(QUrl("http://localhost:8000/api/auth/login"));
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        
        QNetworkReply *reply = m_networkManager->post(request, data);
        connect(reply, &QNetworkReply::finished, [this, reply]() {
            reply->deleteLater();
            
            if (reply->error() == QNetworkReply::NoError) {
                QJsonDocument response = QJsonDocument::fromJson(reply->readAll());
                QJsonObject responseObj = response.object();
                
                if (responseObj["success"].toBool()) {
                    m_statusLabel->setText("登录成功！");
                    m_statusLabel->setStyleSheet("color: green;");
                    QMessageBox::information(this, "成功", "登录成功！");
                } else {
                    m_statusLabel->setText("登录失败: " + responseObj["message"].toString());
                }
            } else {
                m_statusLabel->setText("网络错误: " + reply->errorString());
            }
            
            m_loginButton->setEnabled(true);
        });
    }
    
    void onRegisterClicked()
    {
        QString username = m_usernameEdit->text().trimmed();
        QString password = m_passwordEdit->text().trimmed();
        
        if (username.isEmpty() || password.isEmpty()) {
            m_statusLabel->setText("请输入用户名和密码");
            return;
        }
        
        m_registerButton->setEnabled(false);
        m_statusLabel->setText("正在注册...");
        
        // 创建注册请求
        QJsonObject registerData;
        registerData["username"] = username;
        registerData["password"] = password;
        
        QJsonDocument doc(registerData);
        QByteArray data = doc.toJson();
        
        QNetworkRequest request(QUrl("http://localhost:8000/api/auth/register"));
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        
        QNetworkReply *reply = m_networkManager->post(request, data);
        connect(reply, &QNetworkReply::finished, [this, reply]() {
            reply->deleteLater();
            
            if (reply->error() == QNetworkReply::NoError) {
                QJsonDocument response = QJsonDocument::fromJson(reply->readAll());
                QJsonObject responseObj = response.object();
                
                if (responseObj["success"].toBool()) {
                    m_statusLabel->setText("注册成功！");
                    m_statusLabel->setStyleSheet("color: green;");
                    QMessageBox::information(this, "成功", "注册成功！");
                } else {
                    m_statusLabel->setText("注册失败: " + responseObj["message"].toString());
                }
            } else {
                m_statusLabel->setText("网络错误: " + reply->errorString());
            }
            
            m_registerButton->setEnabled(true);
        });
    }

private:
    QLineEdit *m_usernameEdit;
    QLineEdit *m_passwordEdit;
    QPushButton *m_loginButton;
    QPushButton *m_registerButton;
    QLabel *m_statusLabel;
    QNetworkAccessManager *m_networkManager;
};

int main(int argc, char *argv[])
{
    try {
        std::cout << "启动音视频通话系统登录界面..." << std::endl;
        
        QApplication app(argc, argv);
        
        SimpleLoginWindow window;
        window.show();
        
        std::cout << "登录界面已显示" << std::endl;
        
        return app.exec();
    }
    catch (const std::exception& e) {
        std::cerr << "异常: " << e.what() << std::endl;
        return 1;
    }
    catch (...) {
        std::cerr << "未知异常" << std::endl;
        return 1;
    }
}

#include "simple_login.moc" 