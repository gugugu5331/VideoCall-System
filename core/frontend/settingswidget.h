#ifndef SETTINGSWIDGET_H
#define SETTINGSWIDGET_H

#include <QWidget>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QTabWidget>
#include <QPushButton>
#include <QGroupBox>

class SettingsWidget : public QWidget
{
    Q_OBJECT

public:
    explicit SettingsWidget(QWidget *parent = nullptr);
    ~SettingsWidget();

private:
    void setupUI();

private:
    QVBoxLayout *m_mainLayout;
    QTabWidget *m_tabWidget;
    QPushButton *m_saveButton;
    QPushButton *m_cancelButton;
};

#endif // SETTINGSWIDGET_H 