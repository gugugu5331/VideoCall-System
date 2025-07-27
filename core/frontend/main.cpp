#include <QApplication>
#include <QStyleFactory>
#include <QDir>
#include <QStandardPaths>
#include <QMessageBox>
#include <QTranslator>
#include <QLocale>
#include <QSplashScreen>
#include <QPixmap>
#include <QTimer>
#include "mainwindow.h"

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    
    // 设置应用程序信息
    app.setApplicationName("VideoCall Pro");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("VideoCall Team");
    app.setOrganizationDomain("videocall.com");
    
    // 设置应用程序图标
    app.setWindowIcon(QIcon(":/icons/app_icon.png"));
    
    // 设置应用程序样式
    app.setStyle(QStyleFactory::create("Fusion"));
    
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
    app.setPalette(darkPalette);
    
    // 创建启动画面
    QPixmap splashPixmap(":/images/splash.png");
    if (splashPixmap.isNull()) {
        // 如果没有启动画面图片，创建一个简单的
        splashPixmap = QPixmap(400, 300);
        splashPixmap.fill(QColor(53, 53, 53));
    }
    
    QSplashScreen splash(splashPixmap);
    splash.show();
    
    // 显示启动信息
    splash.showMessage("正在初始化音视频通话系统...", Qt::AlignBottom | Qt::AlignCenter, Qt::white);
    app.processEvents();
    
    // 创建数据目录
    QString dataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    QDir().mkpath(dataPath);
    
    // 检查必要的组件
    splash.showMessage("检查系统组件...", Qt::AlignBottom | Qt::AlignCenter, Qt::white);
    app.processEvents();
    
    // 创建主窗口
    splash.showMessage("加载主界面...", Qt::AlignBottom | Qt::AlignCenter, Qt::white);
    app.processEvents();
    
    MainWindow window;
    
    // 延迟关闭启动画面
    QTimer::singleShot(2000, &splash, &QSplashScreen::close);
    QTimer::singleShot(2000, &window, &MainWindow::show);
    
    return app.exec();
} 