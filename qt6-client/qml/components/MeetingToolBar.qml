import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

/**
 * MeetingToolBar - 会议工具栏组件
 * 
 * 提供会议控制功能：静音、视频、屏幕共享、聊天、参与者列表、离开会议等
 */
Rectangle {
    id: root

    // 公共属性
    property bool audioEnabled: true
    property bool videoEnabled: true
    property bool isScreenSharing: false
    property bool showChat: false
    property bool showParticipants: false
    property int participantCount: 0
    property int unreadMessageCount: 0

    // 样式属性 - 商务风格
    property color backgroundColor: "#0f172a"
    property int toolBarHeight: 90

    // 信号
    signal toggleAudio()
    signal toggleVideo()
    signal toggleScreenShare()
    signal toggleChat()
    signal toggleParticipants()
    signal showSettings()
    signal leaveMeeting()

    height: toolBarHeight
    color: backgroundColor
    border.color: "#1e293b"
    border.width: 1
    
    RowLayout {
        anchors.centerIn: parent
        spacing: 15
        
        // 静音/取消静音按钮 - 商务风格
        ToolButton {
            id: audioButton
            Layout.preferredWidth: 80
            Layout.preferredHeight: 70

            background: Rectangle {
                radius: 10

                gradient: Gradient {
                    GradientStop {
                        position: 0.0
                        color: {
                            if (!audioEnabled) return "#dc2626"
                            if (parent.pressed) return "#1e40af"
                            if (parent.hovered) return "#1e293b"
                            return "#0f172a"
                        }
                    }
                    GradientStop {
                        position: 1.0
                        color: {
                            if (!audioEnabled) return "#991b1b"
                            if (parent.pressed) return "#1e3a8a"
                            if (parent.hovered) return "#0f172a"
                            return "#020617"
                        }
                    }
                }

                border.color: {
                    if (!audioEnabled) return "#ef4444"
                    if (parent.hovered) return "#3b82f6"
                    return "#334155"
                }
                border.width: 2
            }

            contentItem: ColumnLayout {
                spacing: 6

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: audioEnabled ? "🎤" : "🔇"
                    font.pixelSize: 28
                }

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: audioEnabled ? "静音" : "取消静音"
                    font.pixelSize: 12
                    font.family: "Microsoft YaHei"
                    color: "#f1f5f9"
                }
            }

            onClicked: root.toggleAudio()

            ToolTip.visible: hovered
            ToolTip.text: audioEnabled ? "关闭麦克风" : "开启麦克风"
            ToolTip.delay: 500
        }
        
        // 视频开关按钮 - 商务风格
        ToolButton {
            id: videoButton
            Layout.preferredWidth: 80
            Layout.preferredHeight: 70

            background: Rectangle {
                radius: 10

                gradient: Gradient {
                    GradientStop {
                        position: 0.0
                        color: {
                            if (!videoEnabled) return "#dc2626"
                            if (parent.pressed) return "#1e40af"
                            if (parent.hovered) return "#1e293b"
                            return "#0f172a"
                        }
                    }
                    GradientStop {
                        position: 1.0
                        color: {
                            if (!videoEnabled) return "#991b1b"
                            if (parent.pressed) return "#1e3a8a"
                            if (parent.hovered) return "#0f172a"
                            return "#020617"
                        }
                    }
                }

                border.color: {
                    if (!videoEnabled) return "#ef4444"
                    if (parent.hovered) return "#3b82f6"
                    return "#334155"
                }
                border.width: 2
            }

            contentItem: ColumnLayout {
                spacing: 6

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: videoEnabled ? "📹" : "📷"
                    font.pixelSize: 28
                }

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: videoEnabled ? "停止视频" : "开启视频"
                    font.pixelSize: 12
                    font.family: "Microsoft YaHei"
                    color: "#f1f5f9"
                }
            }

            onClicked: root.toggleVideo()

            ToolTip.visible: hovered
            ToolTip.text: videoEnabled ? "关闭摄像头" : "开启摄像头"
            ToolTip.delay: 500
        }

        // 屏幕共享按钮 - 商务风格
        ToolButton {
            id: screenShareButton
            Layout.preferredWidth: 80
            Layout.preferredHeight: 70

            background: Rectangle {
                radius: 10

                gradient: Gradient {
                    GradientStop {
                        position: 0.0
                        color: {
                            if (isScreenSharing) return "#10b981"
                            if (parent.pressed) return "#1e40af"
                            if (parent.hovered) return "#1e293b"
                            return "#0f172a"
                        }
                    }
                    GradientStop {
                        position: 1.0
                        color: {
                            if (isScreenSharing) return "#059669"
                            if (parent.pressed) return "#1e3a8a"
                            if (parent.hovered) return "#0f172a"
                            return "#020617"
                        }
                    }
                }

                border.color: {
                    if (isScreenSharing) return "#34d399"
                    if (parent.hovered) return "#3b82f6"
                    return "#334155"
                }
                border.width: 2
            }

            contentItem: ColumnLayout {
                spacing: 6

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "🖥️"
                    font.pixelSize: 28
                }

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: isScreenSharing ? "停止共享" : "共享屏幕"
                    font.pixelSize: 12
                    font.family: "Microsoft YaHei"
                    color: "#f1f5f9"
                }
            }
            
            onClicked: root.toggleScreenShare()
            
            ToolTip.visible: hovered
            ToolTip.text: isScreenSharing ? "停止屏幕共享" : "开始屏幕共享"
            ToolTip.delay: 500
        }
        
        // 分隔线
        Rectangle {
            Layout.preferredWidth: 1
            Layout.preferredHeight: 40
            color: "#404040"
        }
        
        // 参与者列表按钮
        ToolButton {
            id: participantsButton
            Layout.preferredWidth: 70
            Layout.preferredHeight: 60
            
            background: Rectangle {
                color: {
                    if (showParticipants) return "#1890FF"
                    if (parent.pressed) return "#1890FF"
                    if (parent.hovered) return "#3A3A3A"
                    return "#2C2C2C"
                }
                radius: 8
            }
            
            contentItem: ColumnLayout {
                spacing: 4
                
                Item {
                    Layout.alignment: Qt.AlignHCenter
                    Layout.preferredWidth: 24
                    Layout.preferredHeight: 24
                    
                    Text {
                        anchors.centerIn: parent
                        text: "👥"
                        font.pixelSize: 24
                    }
                    
                    // 参与者数量徽章
                    Rectangle {
                        anchors.top: parent.top
                        anchors.right: parent.right
                        anchors.topMargin: -4
                        anchors.rightMargin: -8
                        width: 20
                        height: 20
                        radius: 10
                        color: "#F44336"
                        visible: participantCount > 0
                        
                        Text {
                            anchors.centerIn: parent
                            text: participantCount > 99 ? "99+" : participantCount.toString()
                            font.pixelSize: 10
                            font.bold: true
                            color: "#FFFFFF"
                        }
                    }
                }
                
                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "参与者"
                    font.pixelSize: 11
                    color: "#FFFFFF"
                }
            }
            
            onClicked: root.toggleParticipants()
            
            ToolTip.visible: hovered
            ToolTip.text: "查看参与者列表"
            ToolTip.delay: 500
        }
        
        // 聊天按钮
        ToolButton {
            id: chatButton
            Layout.preferredWidth: 70
            Layout.preferredHeight: 60
            
            background: Rectangle {
                color: {
                    if (showChat) return "#1890FF"
                    if (parent.pressed) return "#1890FF"
                    if (parent.hovered) return "#3A3A3A"
                    return "#2C2C2C"
                }
                radius: 8
            }
            
            contentItem: ColumnLayout {
                spacing: 4
                
                Item {
                    Layout.alignment: Qt.AlignHCenter
                    Layout.preferredWidth: 24
                    Layout.preferredHeight: 24
                    
                    Text {
                        anchors.centerIn: parent
                        text: "💬"
                        font.pixelSize: 24
                    }
                    
                    // 未读消息徽章
                    Rectangle {
                        anchors.top: parent.top
                        anchors.right: parent.right
                        anchors.topMargin: -4
                        anchors.rightMargin: -8
                        width: 20
                        height: 20
                        radius: 10
                        color: "#F44336"
                        visible: unreadMessageCount > 0 && !showChat
                        
                        Text {
                            anchors.centerIn: parent
                            text: unreadMessageCount > 99 ? "99+" : unreadMessageCount.toString()
                            font.pixelSize: 10
                            font.bold: true
                            color: "#FFFFFF"
                        }
                    }
                }
                
                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "聊天"
                    font.pixelSize: 11
                    color: "#FFFFFF"
                }
            }
            
            onClicked: root.toggleChat()
            
            ToolTip.visible: hovered
            ToolTip.text: "打开聊天面板"
            ToolTip.delay: 500
        }
        
        // 设置按钮
        ToolButton {
            id: settingsButton
            Layout.preferredWidth: 70
            Layout.preferredHeight: 60
            
            background: Rectangle {
                color: {
                    if (parent.pressed) return "#1890FF"
                    if (parent.hovered) return "#3A3A3A"
                    return "#2C2C2C"
                }
                radius: 8
            }
            
            contentItem: ColumnLayout {
                spacing: 4
                
                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "⚙️"
                    font.pixelSize: 24
                }
                
                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "设置"
                    font.pixelSize: 11
                    color: "#FFFFFF"
                }
            }
            
            onClicked: root.showSettings()
            
            ToolTip.visible: hovered
            ToolTip.text: "会议设置"
            ToolTip.delay: 500
        }
        
        // 分隔线
        Rectangle {
            Layout.preferredWidth: 1
            Layout.preferredHeight: 40
            color: "#404040"
        }
        
        // 离开会议按钮
        Button {
            id: leaveButton
            Layout.preferredWidth: 100
            Layout.preferredHeight: 50
            text: "离开会议"
            
            background: Rectangle {
                color: {
                    if (parent.pressed) return "#D32F2F"
                    if (parent.hovered) return "#E53935"
                    return "#F44336"
                }
                radius: 8
            }
            
            contentItem: Text {
                text: parent.text
                color: "#FFFFFF"
                horizontalAlignment: Text.AlignHCenter
                verticalAlignment: Text.AlignVCenter
                font.pixelSize: 14
                font.bold: true
            }
            
            onClicked: root.leaveMeeting()
            
            ToolTip.visible: hovered
            ToolTip.text: "离开当前会议"
            ToolTip.delay: 500
        }
    }
}

