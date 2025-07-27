#include <QApplication>
#include <QMainWindow>
#include <QLabel>
#include <QVBoxLayout>
#include <QWidget>
#include <QMessageBox>
#include <iostream>

int main(int argc, char *argv[])
{
    try {
        std::cout << "Starting Qt application..." << std::endl;
        
        QApplication app(argc, argv);
        
        std::cout << "QApplication created successfully" << std::endl;
        
        QMainWindow window;
        window.setWindowTitle("Qt Test Application");
        window.resize(400, 300);
        
        QWidget *centralWidget = new QWidget(&window);
        window.setCentralWidget(centralWidget);
        
        QVBoxLayout *layout = new QVBoxLayout(centralWidget);
        
        QLabel *label = new QLabel("Hello, Qt6!", centralWidget);
        label->setAlignment(Qt::AlignCenter);
        layout->addWidget(label);
        
        std::cout << "Window created successfully" << std::endl;
        
        window.show();
        
        std::cout << "Window shown successfully" << std::endl;
        
        QMessageBox::information(&window, "Test", "Qt application is working!");
        
        return app.exec();
    }
    catch (const std::exception& e) {
        std::cerr << "Exception: " << e.what() << std::endl;
        return 1;
    }
    catch (...) {
        std::cerr << "Unknown exception occurred" << std::endl;
        return 1;
    }
} 