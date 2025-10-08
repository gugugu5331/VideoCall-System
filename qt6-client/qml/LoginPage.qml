import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

Item {
    id: root

    signal loginSuccess()

    // Dialogs
    RegisterDialog {
        id: registerDialog
    }

    ForgotPasswordDialog {
        id: forgotPasswordDialog
    }

    Rectangle {
        anchors.fill: parent

        // 商务风格渐变背景
        gradient: Gradient {
            GradientStop { position: 0.0; color: "#1a1f2e" }
            GradientStop { position: 1.0; color: "#2d3748" }
        }

        // Logo and title
        Column {
            anchors.centerIn: parent
            spacing: 50
            width: 450

            // Logo - 商务风格
            Rectangle {
                width: 100
                height: 100
                radius: 8
                anchors.horizontalCenter: parent.horizontalCenter

                gradient: Gradient {
                    GradientStop { position: 0.0; color: "#2563eb" }
                    GradientStop { position: 1.0; color: "#1e40af" }
                }

                border.color: "#3b82f6"
                border.width: 2

                Text {
                    anchors.centerIn: parent
                    text: "🎥"
                    font.pixelSize: 48
                }
            }

            // Title - 商务风格
            Column {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: 8

                Text {
                    text: qsTr("智能会议系统")
                    font.pixelSize: 32
                    font.bold: true
                    font.family: "Microsoft YaHei"
                    color: "#f8fafc"
                    anchors.horizontalCenter: parent.horizontalCenter
                }

                Text {
                    text: qsTr("Enterprise Video Conference Platform")
                    font.pixelSize: 13
                    font.family: "Arial"
                    color: "#94a3b8"
                    anchors.horizontalCenter: parent.horizontalCenter
                }
            }

            // Login form - 商务风格
            Rectangle {
                width: parent.width
                height: 380
                color: "#1e293b"
                radius: 12
                border.color: "#3b82f6"
                border.width: 2

                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 40
                    spacing: 20

                    // Username field
                    Column {
                        Layout.fillWidth: true
                        spacing: 8

                        Text {
                            text: qsTr("用户名/邮箱")
                            font.pixelSize: 13
                            font.family: "Microsoft YaHei"
                            color: "#cbd5e1"
                            font.weight: Font.Medium
                        }

                        TextField {
                            id: usernameField
                            width: parent.width
                            placeholderText: qsTr("请输入用户名或邮箱")
                            font.pixelSize: 14
                            font.family: "Microsoft YaHei"
                            color: "#f1f5f9"
                            placeholderTextColor: "#64748b"

                            background: Rectangle {
                                color: "#0f172a"
                                radius: 8
                                border.color: usernameField.activeFocus ? "#3b82f6" : "#475569"
                                border.width: 2
                            }
                        }
                    }

                    // Password field
                    Column {
                        Layout.fillWidth: true
                        spacing: 8

                        Text {
                            text: qsTr("密码")
                            font.pixelSize: 13
                            font.family: "Microsoft YaHei"
                            color: "#cbd5e1"
                            font.weight: Font.Medium
                        }

                        TextField {
                            id: passwordField
                            width: parent.width
                            placeholderText: qsTr("请输入密码")
                            echoMode: TextInput.Password
                            font.pixelSize: 14
                            font.family: "Microsoft YaHei"
                            color: "#f1f5f9"
                            placeholderTextColor: "#64748b"

                            background: Rectangle {
                                color: "#0f172a"
                                radius: 8
                                border.color: passwordField.activeFocus ? "#3b82f6" : "#475569"
                                border.width: 2
                            }

                            Keys.onReturnPressed: loginButton.clicked()
                        }
                    }

                    // Error message - 商务风格
                    Rectangle {
                        id: errorMessage
                        Layout.fillWidth: true
                        height: 45
                        color: "#7f1d1d"
                        border.color: "#dc2626"
                        border.width: 2
                        radius: 8
                        visible: false

                        Row {
                            anchors.fill: parent
                            anchors.margins: 12
                            spacing: 10

                            Text {
                                text: "⚠"
                                font.pixelSize: 16
                                color: "#fca5a5"
                                anchors.verticalCenter: parent.verticalCenter
                            }

                            Text {
                                id: errorText
                                width: parent.width - 30
                                text: ""
                                font.pixelSize: 12
                                font.family: "Microsoft YaHei"
                                color: "#fecaca"
                                wrapMode: Text.WordWrap
                                anchors.verticalCenter: parent.verticalCenter
                            }
                        }
                    }

                    // Remember password - 商务风格
                    CheckBox {
                        id: rememberCheckbox
                        text: qsTr("记住密码")
                        font.pixelSize: 13
                        font.family: "Microsoft YaHei"

                        contentItem: Text {
                            text: rememberCheckbox.text
                            font: rememberCheckbox.font
                            color: "#cbd5e1"
                            leftPadding: rememberCheckbox.indicator.width + 8
                            verticalAlignment: Text.AlignVCenter
                        }
                    }

                    // Login button - 商务风格
                    Button {
                        id: loginButton
                        Layout.fillWidth: true
                        Layout.preferredHeight: 50
                        text: qsTr("登录")
                        font.pixelSize: 16
                        font.bold: true
                        font.family: "Microsoft YaHei"
                        enabled: usernameField.text.length > 0 && passwordField.text.length > 0

                        background: Rectangle {
                            radius: 8

                            gradient: Gradient {
                                GradientStop {
                                    position: 0.0
                                    color: loginButton.enabled ? (loginButton.pressed ? "#1e40af" : "#2563eb") : "#475569"
                                }
                                GradientStop {
                                    position: 1.0
                                    color: loginButton.enabled ? (loginButton.pressed ? "#1e3a8a" : "#1e40af") : "#334155"
                                }
                            }

                            border.color: loginButton.enabled ? "#3b82f6" : "#64748b"
                            border.width: 2
                        }

                        contentItem: Text {
                            text: loginButton.text
                            font: loginButton.font
                            color: loginButton.enabled ? "#ffffff" : "#94a3b8"
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                        }

                        onClicked: {
                            console.log("Login button clicked")
                            console.log("Username:", usernameField.text)
                            console.log("Calling authService.login()")
                            authService.login(usernameField.text, passwordField.text)
                        }
                    }

                    // Links - 商务风格
                    Row {
                        Layout.fillWidth: true
                        Layout.topMargin: 10
                        spacing: 30

                        Text {
                            text: qsTr("注册新账号")
                            font.pixelSize: 13
                            font.family: "Microsoft YaHei"
                            color: "#60a5fa"
                            font.underline: linkMouseArea1.containsMouse

                            MouseArea {
                                id: linkMouseArea1
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    console.log("Register link clicked")
                                    registerDialog.open()
                                }
                            }
                        }

                        Text {
                            text: qsTr("忘记密码?")
                            font.pixelSize: 13
                            font.family: "Microsoft YaHei"
                            color: "#60a5fa"
                            font.underline: linkMouseArea2.containsMouse

                            MouseArea {
                                id: linkMouseArea2
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    console.log("Forgot password link clicked")
                                    forgotPasswordDialog.open()
                                }
                            }
                        }
                    }
                }
            }

            // Footer - 商务风格
            Row {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: 15

                Text {
                    text: qsTr("版本 v1.0.2")
                    font.pixelSize: 11
                    font.family: "Arial"
                    color: "#64748b"
                }

                Text {
                    text: "|"
                    font.pixelSize: 11
                    color: "#475569"
                }

                Text {
                    text: qsTr("隐私政策")
                    font.pixelSize: 11
                    font.family: "Microsoft YaHei"
                    color: "#64748b"
                }

                Text {
                    text: "|"
                    font.pixelSize: 11
                    color: "#475569"
                }

                Text {
                    text: qsTr("使用条款")
                    font.pixelSize: 11
                    font.family: "Microsoft YaHei"
                    color: "#64748b"
                }
            }
        }
    }

    // Monitor login success and failure
    Connections {
        target: authService
        function onLoginSuccess() {
            console.log("Login success signal received")
            errorMessage.visible = false
            root.loginSuccess()
        }

        function onLoginFailed(message) {
            console.log("Login failed signal received:", message)
            errorText.text = message || qsTr("登录失败，请检查用户名和密码")
            errorMessage.visible = true
        }
    }

    // Loading indicator
    BusyIndicator {
        anchors.centerIn: parent
        running: false // TODO: Connect to loading state
        visible: running
    }
}

