import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Window 2.15
import MeetingSystem 1.0

ApplicationWindow {
    id: root
    visible: true
    width: 1280
    height: 800
    minimumWidth: 1024
    minimumHeight: 600
    title: qsTr("智能会议系统")

    // Color scheme (dark theme like Tencent Meeting)
    readonly property color backgroundColor: "#1F1F1F"
    readonly property color surfaceColor: "#2C2C2C"
    readonly property color primaryColor: "#1890FF"
    readonly property color textColor: "#FFFFFF"
    readonly property color secondaryTextColor: "#B0B0B0"

    color: backgroundColor

    // Stack view for navigation
    StackView {
        id: stackView
        anchors.fill: parent
        initialItem: authService.isAuthenticated ? mainWindowComponent : loginComponent

        // Smooth transitions
        pushEnter: Transition {
            PropertyAnimation {
                property: "opacity"
                from: 0
                to: 1
                duration: 200
            }
        }
        pushExit: Transition {
            PropertyAnimation {
                property: "opacity"
                from: 1
                to: 0
                duration: 200
            }
        }
    }

    // Login page component
    Component {
        id: loginComponent
        LoginPage {
            onLoginSuccess: {
                stackView.replace(mainWindowComponent)
            }
        }
    }

    // Main window component
    Component {
        id: mainWindowComponent
        MainWindow {
            onJoinMeeting: {
                stackView.push(meetingRoomComponent)
            }
            onLogout: {
                stackView.replace(loginComponent)
            }
        }
    }

    // Meeting room component
    Component {
        id: meetingRoomComponent
        MeetingRoom {
            onLeaveMeeting: {
                stackView.pop()
            }
        }
    }

    // Monitor authentication state
    Connections {
        target: authService
        function onAuthenticationChanged() {
            if (authService.isAuthenticated) {
                if (stackView.currentItem !== mainWindowComponent) {
                    stackView.replace(mainWindowComponent)
                }
            } else {
                if (stackView.currentItem !== loginComponent) {
                    stackView.replace(loginComponent)
                }
            }
        }
    }

    // Global error dialog
    Dialog {
        id: errorDialog
        title: qsTr("错误")
        modal: true
        anchors.centerIn: parent
        standardButtons: Dialog.Ok

        property string errorMessage: ""

        Label {
            text: errorDialog.errorMessage
            color: root.textColor
        }

        background: Rectangle {
            color: root.surfaceColor
            radius: 8
        }
    }

    // Show error function
    function showError(message) {
        errorDialog.errorMessage = message
        errorDialog.open()
    }

    // Connect to global error signals
    Connections {
        target: authService
        function onLoginFailed(error) {
            showError(qsTr("登录失败: ") + error)
        }
    }

    Connections {
        target: meetingService
        function onMeetingError(error) {
            showError(qsTr("会议错误: ") + error)
        }
    }
}

