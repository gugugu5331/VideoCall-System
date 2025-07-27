#ifndef SECURITYDETECTIONWIDGET_H
#define SECURITYDETECTIONWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QProgressBar>
#include <QTextEdit>
#include <QPushButton>
#include <QGroupBox>

class SecurityDetectionWidget : public QWidget
{
    Q_OBJECT

public:
    explicit SecurityDetectionWidget(QWidget *parent = nullptr);
    ~SecurityDetectionWidget();

private:
    void setupUI();

private:
    QVBoxLayout *m_mainLayout;
    QProgressBar *m_faceRiskBar;
    QProgressBar *m_voiceRiskBar;
    QProgressBar *m_videoRiskBar;
    QLabel *m_securityScoreLabel;
    QTextEdit *m_securityDetailsText;
    QPushButton *m_startDetectionButton;
    QPushButton *m_stopDetectionButton;
};

#endif // SECURITYDETECTIONWIDGET_H 