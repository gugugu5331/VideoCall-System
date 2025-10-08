#include "ui/main_window_controller.h"
#include "application.h"
#include "services/meeting_service.h"
#include "utils/logger.h"
#include <QDateTime>

MainWindowController::MainWindowController(QObject *parent)
    : QObject(parent)
{
}

MainWindowController::~MainWindowController()
{
}

void MainWindowController::createQuickMeeting()
{
    LOG_INFO("Creating quick meeting");
    
    MeetingService *meetingService = Application::instance()->meetingService();
    
    QString title = "快速会议 - " + QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm");
    QString description = "快速创建的会议";
    QDateTime startTime = QDateTime::currentDateTime();
    QDateTime endTime = startTime.addSecs(3600); // 1 hour duration
    int maxParticipants = 10;
    QString meetingType = "video";
    QString password = "";
    QJsonObject settings;

    QObject::connect(meetingService, &MeetingService::meetingCreated, this, [this](Meeting *meeting) {
        emit meetingCreated(meeting->meetingId());
        emit meetingJoined();
    }, Qt::SingleShotConnection);

    QObject::connect(meetingService, &MeetingService::meetingError, this, [this](const QString &errorMsg) {
        emit error(errorMsg);
    }, Qt::SingleShotConnection);

    meetingService->createMeeting(title, description, startTime, endTime, maxParticipants, meetingType, password, settings);
}

void MainWindowController::joinMeeting(int meetingId, const QString &password)
{
    LOG_INFO(QString("Joining meeting: %1").arg(meetingId));
    
    MeetingService *meetingService = Application::instance()->meetingService();
    
    connect(meetingService, &MeetingService::meetingJoined, this, [this]() {
        emit meetingJoined();
    }, Qt::SingleShotConnection);
    
    connect(meetingService, &MeetingService::meetingError, this, [this](const QString &errorMsg) {
        emit error(errorMsg);
    }, Qt::SingleShotConnection);
    
    meetingService->joinMeeting(meetingId, password);
}

void MainWindowController::scheduleMeeting(const QString &title, const QString &description,
                                          const QDateTime &startTime, int duration)
{
    LOG_INFO("Scheduling meeting: " + title);

    MeetingService *meetingService = Application::instance()->meetingService();

    QDateTime endTime = startTime.addSecs(duration * 60); // duration in minutes
    int maxParticipants = 10;
    QString meetingType = "video";
    QString password = "";
    QJsonObject settings;

    QObject::connect(meetingService, &MeetingService::meetingCreated, this, [this](Meeting *meeting) {
        emit meetingCreated(meeting->meetingId());
    }, Qt::SingleShotConnection);

    QObject::connect(meetingService, &MeetingService::meetingError, this, [this](const QString &errorMsg) {
        emit error(errorMsg);
    }, Qt::SingleShotConnection);

    meetingService->createMeeting(title, description, startTime, endTime, maxParticipants, meetingType, password, settings);
}

void MainWindowController::getMeetingList()
{
    LOG_INFO("Fetching meeting list");
    
    MeetingService *meetingService = Application::instance()->meetingService();
    
    connect(meetingService, &MeetingService::meetingListUpdated, this, [this]() {
        emit meetingListUpdated();
    }, Qt::SingleShotConnection);
    
    connect(meetingService, &MeetingService::meetingError, this, [this](const QString &errorMsg) {
        emit error(errorMsg);
    }, Qt::SingleShotConnection);
    
    meetingService->getMeetingList();
}

