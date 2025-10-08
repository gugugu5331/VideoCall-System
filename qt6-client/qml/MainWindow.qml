import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

Item {
    id: root

    signal joinMeeting()
    signal logout()

    Rectangle {
        anchors.fill: parent

        // 商务风格渐变背景
        gradient: Gradient {
            GradientStop { position: 0.0; color: "#1a1f2e" }
            GradientStop { position: 1.0; color: "#2d3748" }
        }

        Row {
            anchors.fill: parent

            // Sidebar - 商务风格
            Rectangle {
                width: 80
                height: parent.height
                color: "#0f172a"
                border.color: "#1e293b"
                border.width: 1

                Column {
                    anchors.fill: parent
                    anchors.topMargin: 20
                    spacing: 0

                    // Avatar - 商务风格
                    Rectangle {
                        width: 52
                        height: 52
                        radius: 26
                        anchors.horizontalCenter: parent.horizontalCenter

                        gradient: Gradient {
                            GradientStop { position: 0.0; color: "#2563eb" }
                            GradientStop { position: 1.0; color: "#1e40af" }
                        }

                        border.color: "#3b82f6"
                        border.width: 2

                        Text {
                            anchors.centerIn: parent
                            text: authService.currentUser ? authService.currentUser.username.charAt(0).toUpperCase() : "U"
                            font.pixelSize: 22
                            font.bold: true
                            font.family: "Microsoft YaHei"
                            color: "#ffffff"
                        }
                    }

                    // Navigation buttons
                    Repeater {
                        model: [
                            { icon: "📹", text: qsTr("会议"), page: "meeting" },
                            { icon: "👥", text: qsTr("通讯录"), page: "contacts" },
                            { icon: "📼", text: qsTr("录制"), page: "recordings" },
                            { icon: "💬", text: qsTr("聊天"), page: "chat" },
                            { icon: "⚙️", text: qsTr("设置"), page: "settings" }
                        ]

                        delegate: Rectangle {
                            width: parent.width
                            height: 80
                            color: currentPage === modelData.page ? "#1e293b" : "transparent"

                            property string currentPage: "meeting"

                            // 左侧高亮条
                            Rectangle {
                                width: 3
                                height: parent.height
                                color: "#3b82f6"
                                visible: currentPage === modelData.page
                            }

                            Column {
                                anchors.centerIn: parent
                                spacing: 6

                                Text {
                                    text: modelData.icon
                                    font.pixelSize: 26
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }

                                Text {
                                    text: modelData.text
                                    font.pixelSize: 11
                                    font.family: "Microsoft YaHei"
                                    color: currentPage === modelData.page ? "#60a5fa" : "#94a3b8"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    parent.currentPage = modelData.page
                                }
                            }
                        }
                    }

                    Item { Layout.fillHeight: true }
                }
            }

            // Main content - 商务风格
            Rectangle {
                width: parent.width - 80
                height: parent.height
                color: "transparent"

                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 40
                    spacing: 30

                    // Header - 商务风格
                    RowLayout {
                        Layout.fillWidth: true

                        Column {
                            spacing: 4

                            Text {
                                text: qsTr("10月3日 周五")
                                font.pixelSize: 24
                                font.bold: true
                                font.family: "Microsoft YaHei"
                                color: "#f1f5f9"
                            }

                            Text {
                                text: qsTr("农历八月十二")
                                font.pixelSize: 13
                                font.family: "Microsoft YaHei"
                                color: "#94a3b8"
                            }
                        }

                        Item { Layout.fillWidth: true }

                        Text {
                            text: qsTr("全部会议 >")
                            font.pixelSize: 14
                            font.family: "Microsoft YaHei"
                            color: "#60a5fa"
                            font.underline: allMeetingsMouseArea.containsMouse

                            MouseArea {
                                id: allMeetingsMouseArea
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    // TODO: Show all meetings
                                }
                            }
                        }
                    }

                    // Quick actions
                    GridLayout {
                        Layout.fillWidth: true
                        columns: 4
                        rowSpacing: 20
                        columnSpacing: 20

                        // Join meeting - 商务风格
                        Rectangle {
                            Layout.preferredWidth: 160
                            Layout.preferredHeight: 160
                            radius: 12
                            color: "#1e293b"
                            border.color: joinMouseArea.containsMouse ? "#3b82f6" : "#334155"
                            border.width: 2

                            gradient: Gradient {
                                GradientStop { position: 0.0; color: "#1e293b" }
                                GradientStop { position: 1.0; color: "#0f172a" }
                            }

                            Column {
                                anchors.centerIn: parent
                                spacing: 15

                                Text {
                                    text: "➕"
                                    font.pixelSize: 52
                                    color: "#60a5fa"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }

                                Text {
                                    text: qsTr("加入会议")
                                    font.pixelSize: 16
                                    font.bold: true
                                    font.family: "Microsoft YaHei"
                                    color: "#f1f5f9"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }
                            }

                            MouseArea {
                                id: joinMouseArea
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    joinMeetingDialog.open()
                                }
                            }
                        }

                        // Quick meeting - 商务风格
                        Rectangle {
                            Layout.preferredWidth: 160
                            Layout.preferredHeight: 160
                            radius: 12
                            color: "#1e293b"
                            border.color: quickMouseArea.containsMouse ? "#3b82f6" : "#334155"
                            border.width: 2

                            gradient: Gradient {
                                GradientStop { position: 0.0; color: "#1e293b" }
                                GradientStop { position: 1.0; color: "#0f172a" }
                            }

                            Column {
                                anchors.centerIn: parent
                                spacing: 15

                                Text {
                                    text: "⚡"
                                    font.pixelSize: 52
                                    color: "#60a5fa"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }

                                Text {
                                    text: qsTr("快速会议")
                                    font.pixelSize: 16
                                    font.bold: true
                                    font.family: "Microsoft YaHei"
                                    color: "#f1f5f9"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }
                            }

                            MouseArea {
                                id: quickMouseArea
                                anchors.fill: parent
                                hoverEnabled: true
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    // Create and join meeting immediately
                                    meetingService.createMeeting(
                                        qsTr("快速会议"),
                                        "",
                                        new Date(),
                                        60
                                    )
                                }
                            }
                        }

                        // Schedule meeting
                        Rectangle {
                            Layout.preferredWidth: 150
                            Layout.preferredHeight: 150
                            color: "#1890FF"
                            radius: 12

                            Column {
                                anchors.centerIn: parent
                                spacing: 12

                                Text {
                                    text: "✓"
                                    font.pixelSize: 48
                                    color: "#FFFFFF"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }

                                Text {
                                    text: qsTr("预定会议")
                                    font.pixelSize: 16
                                    font.bold: true
                                    color: "#FFFFFF"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    scheduleMeetingDialog.open()
                                }
                            }
                        }

                        // Screen share
                        Rectangle {
                            Layout.preferredWidth: 150
                            Layout.preferredHeight: 150
                            color: "#1890FF"
                            radius: 12

                            Column {
                                anchors.centerIn: parent
                                spacing: 12

                                Text {
                                    text: "🖥️"
                                    font.pixelSize: 48
                                    color: "#FFFFFF"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }

                                Text {
                                    text: qsTr("共享屏幕")
                                    font.pixelSize: 16
                                    font.bold: true
                                    color: "#FFFFFF"
                                    anchors.horizontalCenter: parent.horizontalCenter
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    // TODO: Start screen share
                                }
                            }
                        }
                    }

                    // Meeting list
                    Rectangle {
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                        color: "#F8F9FA"
                        radius: 12

                        Column {
                            anchors.centerIn: parent
                            spacing: 20

                            Text {
                                text: "💼"
                                font.pixelSize: 64
                                anchors.horizontalCenter: parent.horizontalCenter
                            }

                            Text {
                                text: qsTr("暂无会议")
                                font.pixelSize: 16
                                color: "#B0B0B0"
                                anchors.horizontalCenter: parent.horizontalCenter
                            }
                        }
                    }
                }
            }
        }
    }

    // Join meeting dialog
    Dialog {
        id: joinMeetingDialog
        title: qsTr("加入会议")
        modal: true
        anchors.centerIn: parent
        width: 400

        ColumnLayout {
            width: parent.width
            spacing: 20

            TextField {
                id: meetingIdField
                Layout.fillWidth: true
                placeholderText: qsTr("请输入会议ID")
            }

            TextField {
                id: meetingPasswordField
                Layout.fillWidth: true
                placeholderText: qsTr("会议密码(可选)")
                echoMode: TextInput.Password
            }
        }

        standardButtons: Dialog.Ok | Dialog.Cancel

        onAccepted: {
            if (meetingIdField.text.length > 0) {
                meetingService.joinMeeting(
                    parseInt(meetingIdField.text),
                    meetingPasswordField.text
                )
                root.joinMeeting()
            }
        }
    }

    // Schedule meeting dialog
    Dialog {
        id: scheduleMeetingDialog
        title: qsTr("预定会议")
        modal: true
        anchors.centerIn: parent
        width: 500

        // TODO: Add meeting scheduling form

        standardButtons: Dialog.Ok | Dialog.Cancel
    }
}

