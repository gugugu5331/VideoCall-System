#include "application.h"
#include <QDebug>
#ifdef Q_OS_WIN
#include <windows.h>
#endif

int main(int argc, char *argv[])
{
    // Set UTF-8 encoding for console output on Windows
#ifdef Q_OS_WIN
    SetConsoleOutputCP(CP_UTF8);
    SetConsoleCP(CP_UTF8);
#endif

    try {
        Application app(argc, argv);
        return app.run();
    } catch (const std::exception &e) {
        qCritical() << "Fatal error:" << e.what();
        return 1;
    } catch (...) {
        qCritical() << "Unknown fatal error occurred";
        return 1;
    }
}

