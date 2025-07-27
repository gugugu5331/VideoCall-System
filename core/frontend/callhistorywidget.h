#ifndef CALLHISTORYWIDGET_H
#define CALLHISTORYWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QListWidget>
#include <QPushButton>
#include <QGroupBox>

class CallHistoryWidget : public QWidget
{
    Q_OBJECT

public:
    explicit CallHistoryWidget(QWidget *parent = nullptr);
    ~CallHistoryWidget();

private:
    void setupUI();

private:
    QVBoxLayout *m_mainLayout;
    QListWidget *m_historyList;
    QPushButton *m_clearButton;
    QPushButton *m_exportButton;
};

#endif // CALLHISTORYWIDGET_H 