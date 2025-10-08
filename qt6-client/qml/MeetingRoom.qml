import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15
import MeetingSystem 1.0
import "components"

Rectangle {
    id: root

    // 商务风格渐变背景
    gradient: Gradient {
        GradientStop { position: 0.0; color: "#0f172a" }
        GradientStop { position: 1.0; color: "#1e293b" }
    }

    // Signals
    signal leaveMeeting()

    property var meetingRoomController: MeetingRoomController {}
    property var aiPanelController: AIPanelController {}
    property var videoEffectsController: VideoEffectsController {}
    property bool showAIPanel: true
    property bool showChatPanel: false
    property bool showParticipantList: false
    property bool showVideoEffects: false

    // 顶部信息栏 - 商务风格
    Rectangle {
        id: topBar
        anchors.top: parent.top
        anchors.left: parent.left
        anchors.right: parent.right
        height: 70
        color: "#0f172a"
        border.color: "#1e293b"
        border.width: 1

        RowLayout {
            anchors.fill: parent
            anchors.margins: 20
            spacing: 25

            Text {
                text: "智能会议系统"
                font.pixelSize: 20
                font.bold: true
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
            }

            Rectangle {
                width: 2
                height: 35
                color: "#334155"
            }

            Text {
                text: "会议ID: " + (meetingRoomController.meetingId || "未加入")
                font.pixelSize: 14
                font.family: "Microsoft YaHei"
                color: "#94a3b8"
            }

            Text {
                text: "参会人数: " + meetingRoomController.participantCount
                font.pixelSize: 14
                font.family: "Microsoft YaHei"
                color: "#94a3b8"
            }

            Text {
                text: "时长: " + meetingRoomController.meetingDuration
                font.pixelSize: 14
                color: "#B0B0B0"
            }

            Item { Layout.fillWidth: true }

            Button {
                text: "🎨 效果"
                onClicked: showVideoEffects = !showVideoEffects
                background: Rectangle {
                    color: parent.hovered ? "#3A3A3A" : "#2C2C2C"
                    radius: 4
                    border.color: showVideoEffects ? "#1890FF" : "transparent"
                    border.width: 2
                }
                contentItem: Text {
                    text: parent.text
                    color: showVideoEffects ? "#1890FF" : "#FFFFFF"
                    horizontalAlignment: Text.AlignHCenter
                    verticalAlignment: Text.AlignVCenter
                }
            }

            Button {
                text: showAIPanel ? "隐藏AI面板" : "显示AI面板"
                onClicked: showAIPanel = !showAIPanel
                background: Rectangle {
                    color: parent.hovered ? "#3A3A3A" : "#2C2C2C"
                    radius: 4
                }
                contentItem: Text {
                    text: parent.text
                    color: "#FFFFFF"
                    horizontalAlignment: Text.AlignHCenter
                    verticalAlignment: Text.AlignVCenter
                }
            }
        }
    }
    
    // 主内容区
    Row {
        anchors.top: topBar.bottom
        anchors.bottom: bottomBar.top
        anchors.left: parent.left
        anchors.right: parent.right

        // 左侧：视频画面网格
        Rectangle {
            width: {
                var baseWidth = parent.width
                var rightPanelCount = 0

                if (showAIPanel) rightPanelCount++
                if (showVideoEffects) rightPanelCount++
                if (showChatPanel || showParticipantList) rightPanelCount++

                if (rightPanelCount === 0) return baseWidth
                if (rightPanelCount === 1) return baseWidth * 0.7
                if (rightPanelCount === 2) return baseWidth * 0.5
                return baseWidth * 0.4
            }
            height: parent.height
            color: "#1F1F1F"

            GridView {
                id: videoGrid
                anchors.fill: parent
                anchors.margins: 10
                cellWidth: width / 2
                cellHeight: height / 2

                model: meetingRoomController.participants

                delegate: VideoTile {
                    width: videoGrid.cellWidth - 10
                    height: videoGrid.cellHeight - 10

                    userId: model.userId
                    username: model.username
                    videoEnabled: model.videoEnabled
                    audioEnabled: model.audioEnabled
                    isScreenSharing: model.isScreenSharing
                    isLocalUser: false  // 需要判断是否是本地用户
                    aiPanelController: root.aiPanelController  // 传递AI面板控制器

                    onClicked: {
                        console.log("Video tile clicked:", username)
                    }

                    onDoubleClicked: {
                        console.log("Video tile double clicked:", username)
                        // 可以实现全屏功能
                    }
                }

                // 空状态
                Label {
                    anchors.centerIn: parent
                    text: "等待参与者加入..."
                    font.pixelSize: 16
                    color: "#808080"
                    visible: videoGrid.count === 0
                }
            }
        }

        // 中间：聊天面板或参与者列表
        Loader {
            id: sidePanelLoader
            width: parent.width * 0.2
            height: parent.height
            visible: showChatPanel || showParticipantList

            sourceComponent: {
                if (showChatPanel) {
                    return chatPanelComponent
                } else if (showParticipantList) {
                    return participantListComponent
                }
                return null
            }
        }
        
        // 右侧：AI面板
        AIPanel {
            id: aiPanel
            width: parent.width * 0.3
            height: parent.height
            visible: showAIPanel
            controller: aiPanelController
        }
    }
    
    // 底部工具栏
    MeetingToolBar {
        id: bottomBar
        anchors.bottom: parent.bottom
        anchors.left: parent.left
        anchors.right: parent.right

        audioEnabled: meetingRoomController.audioEnabled
        videoEnabled: meetingRoomController.videoEnabled
        isScreenSharing: meetingRoomController.isScreenSharing
        showChat: showChatPanel
        showParticipants: showParticipantList
        participantCount: meetingRoomController.participantCount
        unreadMessageCount: meetingRoomController.unreadMessageCount

        onToggleAudio: {
            meetingRoomController.toggleAudio()
        }

        onToggleVideo: {
            meetingRoomController.toggleVideo()
        }

        onToggleScreenShare: {
            meetingRoomController.toggleScreenShare()
        }

        onToggleChat: {
            showChatPanel = !showChatPanel
            if (showChatPanel) {
                showParticipantList = false
                meetingRoomController.clearUnreadMessages()
            }
        }

        onToggleParticipants: {
            showParticipantList = !showParticipantList
            if (showParticipantList) {
                showChatPanel = false
            }
        }

        onShowSettings: {
            console.log("Show settings")
            // 打开设置对话框
        }

        onLeaveMeeting: {
            meetingRoomController.leaveMeeting()
            // 触发信号返回主界面
            root.leaveMeeting()
        }
    }

    // 聊天面板组件
    Component {
        id: chatPanelComponent

        ChatPanel {
            messagesModel: meetingRoomController.chatMessages
            currentUserId: 0  // 需要从某处获取当前用户ID
            currentUsername: "我"

            onSendMessage: function(content) {
                meetingRoomController.sendChatMessage(content)
            }

            onLoadMoreMessages: {
                console.log("Load more messages")
                // 加载更多历史消息
            }
        }
    }

    // 参与者列表组件
    Component {
        id: participantListComponent

        ParticipantList {
            participantsModel: meetingRoomController.participants
            isHost: meetingRoomController.isHost
            currentUserId: 0  // 需要从某处获取当前用户ID

            onMuteParticipant: function(userId) {
                meetingRoomController.muteParticipant(userId)
            }

            onKickParticipant: function(userId) {
                meetingRoomController.kickParticipant(userId)
            }

            onMakeHost: function(userId) {
                meetingRoomController.makeHost(userId)
            }

            onPinParticipant: function(userId) {
                console.log("Pin participant:", userId)
                // 固定参与者视频
            }
        }
    }

    // 连接控制器信号
    Connections {
        target: meetingRoomController

        function onMeetingJoined() {
            console.log("Meeting joined")
        }

        function onMeetingLeft() {
            console.log("Meeting left")
        }

        function onMeetingError(error) {
            console.error("Meeting error:", error)
            // 显示错误对话框
        }

        function onParticipantJoined(userId, username) {
            console.log("Participant joined:", username)
        }

        function onParticipantLeft(userId) {
            console.log("Participant left:", userId)
        }

        function onChatMessageReceived(fromUserId, username, message) {
            console.log("Chat message from", username, ":", message)
            // 如果聊天面板未打开，显示通知
        }
    }

    // 视频效果面板
    Loader {
        id: videoEffectsPanelLoader
        anchors.right: parent.right
        anchors.top: topBar.bottom
        anchors.bottom: bottomBar.top
        width: showVideoEffects ? 400 : 0
        visible: showVideoEffects

        sourceComponent: showVideoEffects ? videoEffectsPanelComponent : null

        Behavior on width {
            NumberAnimation { duration: 300; easing.type: Easing.OutCubic }
        }
    }

    // 视频效果面板组件
    Component {
        id: videoEffectsPanelComponent

        VideoEffectsPanel {
            controller: videoEffectsController
        }
    }
}

