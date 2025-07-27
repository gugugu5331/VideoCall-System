#ifndef AUDIOMANAGER_H
#define AUDIOMANAGER_H

#include <QObject>
#include <QAudioInput>
#include <QAudioOutput>
#include <QAudioDeviceInfo>

class AudioManager : public QObject
{
    Q_OBJECT

public:
    explicit AudioManager(QObject *parent = nullptr);
    ~AudioManager();

    void initializeAudio();
    void startAudio();
    void stopAudio();

private:
    QAudioInput *m_audioInput;
    QAudioOutput *m_audioOutput;
};

#endif // AUDIOMANAGER_H 