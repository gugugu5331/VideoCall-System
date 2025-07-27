#ifndef USERPROFILEWIDGET_H
#define USERPROFILEWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QTextEdit>
#include <QGroupBox>

class UserProfileWidget : public QWidget
{
    Q_OBJECT

public:
    explicit UserProfileWidget(QWidget *parent = nullptr);
    ~UserProfileWidget();

private:
    void setupUI();

private:
    QVBoxLayout *m_mainLayout;
    QLabel *m_avatarLabel;
    QLineEdit *m_usernameEdit;
    QLineEdit *m_emailEdit;
    QTextEdit *m_bioEdit;
    QPushButton *m_saveButton;
    QPushButton *m_cancelButton;
};

#endif // USERPROFILEWIDGET_H 