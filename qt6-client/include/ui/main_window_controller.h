#ifndef MAIN_WINDOW_CONTROLLER_H
#define MAIN_WINDOW_CONTROLLER_H

#include <QObject>

class MainWindowController : public QObject
{
    Q_OBJECT

public:
    explicit MainWindowController(QObject *parent = nullptr);
    ~MainWindowController();

    Q_INVOKABLE void createQuickMeeting();
    Q_INVOKABLE void joinMeeting(int meetingId, const QString &password);
    Q_INVOKABLE void scheduleMeeting(const QString &title, const QString &description,
                                     const QDateTime &startTime, int duration);
    Q_INVOKABLE void getMeetingList();

signals:
    void meetingCreated(int meetingId);
    void meetingJoined();
    void meetingListUpdated();
    void error(const QString &message);
};

#endif // MAIN_WINDOW_CONTROLLER_H

