#include <QApplication>
#include <QMainWindow>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QPushButton>
#include <QLineEdit>
#include <QGroupBox>
#include <QMessageBox>
#include <QStyle>
#include <QPalette>
#include <QColor>
#include <QIcon>
#include <QPixmap>

class DemoWindow : public QMainWindow
{
    Q_OBJECT

public:
    DemoWindow(QWidget *parent = nullptr) : QMainWindow(parent)
    {
        setupUI();
        setupStyles();
    }

private:
    void setupUI()
    {
        setWindowTitle("音视频通话系统 - 演示版");
        setMinimumSize(800, 600);
        resize(1000, 700);

        // 创建中央窗口
        QWidget *centralWidget = new QWidget(this);
        setCentralWidget(centralWidget);

        QVBoxLayout *mainLayout = new QVBoxLayout(centralWidget);

        // Logo和标题
        QLabel *logoLabel = new QLabel("🎥");
        logoLabel->setAlignment(Qt::AlignCenter);
        logoLabel->setStyleSheet("font-size: 64px; margin: 20px;");

        QLabel *titleLabel = new QLabel("音视频通话系统");
        titleLabel->setAlignment(Qt::AlignCenter);
        titleLabel->setStyleSheet("font-size: 28px; font-weight: bold; margin: 10px; color: #4a90e2;");

        QLabel *subtitleLabel = new QLabel("基于Qt6 C++开发的高质量音视频通话系统");
        subtitleLabel->setAlignment(Qt::AlignCenter);
        subtitleLabel->setStyleSheet("font-size: 14px; margin: 10px; color: #888;");

        mainLayout->addWidget(logoLabel);
        mainLayout->addWidget(titleLabel);
        mainLayout->addWidget(subtitleLabel);

        // 功能演示区域
        QGroupBox *demoGroup = new QGroupBox("功能演示");
        QVBoxLayout *demoLayout = new QVBoxLayout(demoGroup);

        // 登录演示
        QGroupBox *loginGroup = new QGroupBox("用户登录");
        QVBoxLayout *loginLayout = new QVBoxLayout(loginGroup);

        QLineEdit *usernameEdit = new QLineEdit();
        usernameEdit->setPlaceholderText("请输入用户名");
        QLineEdit *passwordEdit = new QLineEdit();
        passwordEdit->setPlaceholderText("请输入密码");
        passwordEdit->setEchoMode(QLineEdit::Password);

        QPushButton *loginButton = new QPushButton("登录");
        connect(loginButton, &QPushButton::clicked, [this, usernameEdit, passwordEdit]() {
            QString username = usernameEdit->text().trimmed();
            QString password = passwordEdit->text();
            
            if (username.isEmpty() || password.isEmpty()) {
                QMessageBox::warning(this, "登录失败", "请输入用户名和密码");
                return;
            }
            
            QMessageBox::information(this, "登录成功", 
                QString("欢迎 %1！\n\n这是一个演示版本，展示了Qt C++音视频通话系统的界面设计。").arg(username));
        });

        loginLayout->addWidget(new QLabel("用户名:"));
        loginLayout->addWidget(usernameEdit);
        loginLayout->addWidget(new QLabel("密码:"));
        loginLayout->addWidget(passwordEdit);
        loginLayout->addWidget(loginButton);

        // 功能按钮
        QHBoxLayout *buttonLayout = new QHBoxLayout();

        QPushButton *videoCallButton = new QPushButton("🎥 开始视频通话");
        videoCallButton->setMinimumHeight(50);
        connect(videoCallButton, &QPushButton::clicked, [this]() {
            QMessageBox::information(this, "视频通话", 
                "视频通话功能演示\n\n"
                "• 支持720p/1080p高清视频\n"
                "• 实时音视频传输\n"
                "• 多摄像头支持\n"
                "• 全屏通话体验");
        });

        QPushButton *securityButton = new QPushButton("🔒 安全检测");
        securityButton->setMinimumHeight(50);
        connect(securityButton, &QPushButton::clicked, [this]() {
            QMessageBox::information(this, "安全检测", 
                "音视频鉴伪功能演示\n\n"
                "• 人脸检测和活体检测\n"
                "• 语音合成攻击检测\n"
                "• 深度伪造检测\n"
                "• 实时安全监控");
        });

        QPushButton *settingsButton = new QPushButton("⚙️ 系统设置");
        settingsButton->setMinimumHeight(50);
        connect(settingsButton, &QPushButton::clicked, [this]() {
            QMessageBox::information(this, "系统设置", 
                "系统设置功能演示\n\n"
                "• 音视频设备配置\n"
                "• 网络连接设置\n"
                "• 安全检测阈值\n"
                "• 界面主题选择");
        });

        buttonLayout->addWidget(videoCallButton);
        buttonLayout->addWidget(securityButton);
        buttonLayout->addWidget(settingsButton);

        // 技术特性
        QGroupBox *featuresGroup = new QGroupBox("技术特性");
        QVBoxLayout *featuresLayout = new QVBoxLayout(featuresGroup);

        QStringList features = {
            "🎯 基于Qt6 C++开发的跨平台应用",
            "🎥 集成WebRTC实现实时音视频通信",
            "🔒 集成OpenCV进行安全检测和鉴伪",
            "🌐 支持WebSocket实时双向通信",
            "💾 本地SQLite数据存储",
            "🎨 现代化深色主题界面设计",
            "📱 响应式布局，支持多屏幕尺寸",
            "⚡ 高性能多线程处理架构"
        };

        for (const QString &feature : features) {
            QLabel *featureLabel = new QLabel(feature);
            featureLabel->setStyleSheet("padding: 5px; font-size: 12px;");
            featuresLayout->addWidget(featureLabel);
        }

        // 添加到主布局
        demoLayout->addWidget(loginGroup);
        demoLayout->addLayout(buttonLayout);
        demoLayout->addWidget(featuresGroup);

        mainLayout->addWidget(demoGroup);

        // 状态栏
        QStatusBar *statusBar = this->statusBar();
        statusBar->showMessage("演示版本 - 基于Qt6 C++开发");
    }

    void setupStyles()
    {
        // 设置深色主题
        QPalette darkPalette;
        darkPalette.setColor(QPalette::Window, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::WindowText, Qt::white);
        darkPalette.setColor(QPalette::Base, QColor(25, 25, 25));
        darkPalette.setColor(QPalette::AlternateBase, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::ToolTipBase, Qt::white);
        darkPalette.setColor(QPalette::ToolTipText, Qt::white);
        darkPalette.setColor(QPalette::Text, Qt::white);
        darkPalette.setColor(QPalette::Button, QColor(53, 53, 53));
        darkPalette.setColor(QPalette::ButtonText, Qt::white);
        darkPalette.setColor(QPalette::BrightText, Qt::red);
        darkPalette.setColor(QPalette::Link, QColor(42, 130, 218));
        darkPalette.setColor(QPalette::Highlight, QColor(42, 130, 218));
        darkPalette.setColor(QPalette::HighlightedText, Qt::black);
        
        qApp->setPalette(darkPalette);

        // 设置样式表
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
                font-size: 12px;
            }
            QPushButton {
                padding: 10px 20px;
                border: 1px solid #555;
                border-radius: 5px;
                background-color: #4a4a4a;
                font-size: 12px;
                font-weight: bold;
            }
            QPushButton:hover {
                background-color: #5a5a5a;
                border-color: #4a90e2;
            }
            QPushButton:pressed {
                background-color: #3a3a3a;
            }
            QStatusBar {
                background-color: #2b2b2b;
                color: #888;
            }
        )");
    }
};

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    
    app.setApplicationName("VideoCall Demo");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("VideoCall Team");
    
    DemoWindow window;
    window.show();
    
    return app.exec();
}

#include "demo_app.moc" 