#ifndef VIDEOMANAGER_H
#define VIDEOMANAGER_H

#include <QObject>
#include <QCamera>
#include <QVideoWidget>

class VideoManager : public QObject
{
    Q_OBJECT

public:
    explicit VideoManager(QObject *parent = nullptr);
    ~VideoManager();

    void initializeVideo();
    void startVideo();
    void stopVideo();

private:
    QCamera *m_camera;
    QVideoWidget *m_videoWidget;
};

#endif // VIDEOMANAGER_H 