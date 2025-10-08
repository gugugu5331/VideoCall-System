import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

/**
 * ParticipantList - 参与者列表组件
 * 
 * 显示所有会议参与者及其状态，支持参与者管理操作
 */
Rectangle {
    id: root

    // 公共属性
    property var participantsModel: null  // ListModel
    property bool isHost: false
    property int currentUserId: 0

    // 样式属性 - 商务风格
    property color backgroundColor: "#0f172a"
    property color headerColor: "#1e293b"

    // 信号
    signal muteParticipant(int userId)
    signal kickParticipant(int userId)
    signal makeHost(int userId)
    signal pinParticipant(int userId)

    color: backgroundColor
    border.color: "#334155"
    border.width: 1

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        // 头部 - 商务风格
        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 70
            color: headerColor
            border.color: "#334155"
            border.width: 1

            RowLayout {
                anchors.fill: parent
                anchors.margins: 20
                spacing: 15

                Text {
                    text: "👥 参与者"
                    font.pixelSize: 20
                    font.bold: true
                    font.family: "Microsoft YaHei"
                    color: "#f1f5f9"
                }

                Rectangle {
                    Layout.preferredWidth: 35
                    Layout.preferredHeight: 28
                    radius: 14

                    gradient: Gradient {
                        GradientStop { position: 0.0; color: "#2563eb" }
                        GradientStop { position: 1.0; color: "#1e40af" }
                    }

                    border.color: "#3b82f6"
                    border.width: 1

                    Text {
                        anchors.centerIn: parent
                        text: participantsModel ? participantsModel.count : 0
                        font.pixelSize: 13
                        font.bold: true
                        font.family: "Microsoft YaHei"
                        color: "#ffffff"
                    }
                }
                
                Item { Layout.fillWidth: true }
                
                // 搜索按钮
                Button {
                    Layout.preferredWidth: 36
                    Layout.preferredHeight: 36
                    text: "🔍"
                    
                    background: Rectangle {
                        color: parent.hovered ? "#3A3A3A" : "transparent"
                        radius: 4
                    }
                    
                    contentItem: Text {
                        text: parent.text
                        color: "#FFFFFF"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 16
                    }
                }
            }
        }
        
        // 搜索框
        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 50
            color: headerColor
            
            TextField {
                id: searchField
                anchors.fill: parent
                anchors.margins: 10
                placeholderText: "搜索参与者..."
                
                background: Rectangle {
                    color: "#1F1F1F"
                    radius: 4
                    border.color: searchField.activeFocus ? "#1890FF" : "#404040"
                    border.width: 1
                }
                
                color: "#FFFFFF"
                font.pixelSize: 14
                leftPadding: 35
                
                // 搜索图标
                Text {
                    anchors.left: parent.left
                    anchors.leftMargin: 10
                    anchors.verticalCenter: parent.verticalCenter
                    text: "🔍"
                    font.pixelSize: 16
                }
            }
        }
        
        // 参与者列表
        ListView {
            id: participantListView
            Layout.fillWidth: true
            Layout.fillHeight: true
            clip: true
            
            model: participantsModel
            
            ScrollBar.vertical: ScrollBar {
                policy: ScrollBar.AsNeeded
            }
            
            delegate: Rectangle {
                width: participantListView.width
                height: 70
                color: mouseArea.containsMouse ? "#3A3A3A" : "transparent"
                
                MouseArea {
                    id: mouseArea
                    anchors.fill: parent
                    hoverEnabled: true
                }
                
                RowLayout {
                    anchors.fill: parent
                    anchors.margins: 10
                    spacing: 12
                    
                    // 用户头像
                    Rectangle {
                        Layout.preferredWidth: 50
                        Layout.preferredHeight: 50
                        radius: 25
                        color: "#1890FF"
                        
                        Text {
                            anchors.centerIn: parent
                            text: model.username ? model.username.charAt(0).toUpperCase() : "?"
                            font.pixelSize: 20
                            font.bold: true
                            color: "#FFFFFF"
                        }
                        
                        // 在线状态指示器
                        Rectangle {
                            anchors.bottom: parent.bottom
                            anchors.right: parent.right
                            width: 14
                            height: 14
                            radius: 7
                            color: model.status === "online" ? "#52C41A" : "#808080"
                            border.color: root.backgroundColor
                            border.width: 2
                        }
                    }
                    
                    // 用户信息
                    ColumnLayout {
                        Layout.fillWidth: true
                        spacing: 4
                        
                        RowLayout {
                            spacing: 6
                            
                            Text {
                                text: model.username || "未知用户"
                                font.pixelSize: 14
                                font.bold: model.userId === currentUserId
                                color: "#FFFFFF"
                                elide: Text.ElideRight
                                Layout.maximumWidth: 150
                            }
                            
                            // 主持人标识
                            Rectangle {
                                Layout.preferredWidth: 50
                                Layout.preferredHeight: 20
                                radius: 3
                                color: "#FFA940"
                                visible: model.role === "host"
                                
                                Text {
                                    anchors.centerIn: parent
                                    text: "主持人"
                                    font.pixelSize: 10
                                    color: "#FFFFFF"
                                }
                            }
                            
                            // "我"标识
                            Rectangle {
                                Layout.preferredWidth: 30
                                Layout.preferredHeight: 20
                                radius: 3
                                color: "#1890FF"
                                visible: model.userId === currentUserId
                                
                                Text {
                                    anchors.centerIn: parent
                                    text: "我"
                                    font.pixelSize: 10
                                    color: "#FFFFFF"
                                }
                            }
                        }
                        
                        // 状态信息
                        RowLayout {
                            spacing: 8
                            
                            Text {
                                text: model.audioEnabled ? "🎤" : "🔇"
                                font.pixelSize: 14
                                color: model.audioEnabled ? "#52C41A" : "#F44336"
                            }
                            
                            Text {
                                text: model.videoEnabled ? "📹" : "📷"
                                font.pixelSize: 14
                                color: model.videoEnabled ? "#52C41A" : "#808080"
                            }
                            
                            Text {
                                text: model.isScreenSharing ? "🖥️" : ""
                                font.pixelSize: 14
                                visible: model.isScreenSharing
                            }
                            
                            // 网络质量
                            Row {
                                spacing: 2
                                
                                Repeater {
                                    model: 3
                                    Rectangle {
                                        width: 3
                                        height: 6 + index * 3
                                        color: {
                                            var quality = participantListView.model.get(index).networkQuality || 2
                                            return index < quality ? "#52C41A" : "#404040"
                                        }
                                        radius: 1
                                    }
                                }
                            }
                        }
                    }
                    
                    // 操作按钮（仅主持人可见）
                    Row {
                        spacing: 5
                        visible: isHost && model.userId !== currentUserId
                        
                        // 静音按钮
                        Button {
                            width: 32
                            height: 32
                            text: model.audioEnabled ? "🔇" : "🎤"
                            
                            background: Rectangle {
                                color: parent.hovered ? "#3A3A3A" : "transparent"
                                radius: 4
                            }
                            
                            contentItem: Text {
                                text: parent.text
                                color: "#FFFFFF"
                                horizontalAlignment: Text.AlignHCenter
                                verticalAlignment: Text.AlignVCenter
                                font.pixelSize: 14
                            }
                            
                            onClicked: root.muteParticipant(model.userId)
                            
                            ToolTip.visible: hovered
                            ToolTip.text: model.audioEnabled ? "静音该参与者" : "取消静音"
                        }
                        
                        // 更多操作按钮
                        Button {
                            width: 32
                            height: 32
                            text: "⋮"
                            
                            background: Rectangle {
                                color: parent.hovered ? "#3A3A3A" : "transparent"
                                radius: 4
                            }
                            
                            contentItem: Text {
                                text: parent.text
                                color: "#FFFFFF"
                                horizontalAlignment: Text.AlignHCenter
                                verticalAlignment: Text.AlignVCenter
                                font.pixelSize: 18
                            }
                            
                            onClicked: contextMenu.popup()
                            
                            Menu {
                                id: contextMenu
                                
                                MenuItem {
                                    text: "设为主持人"
                                    onTriggered: root.makeHost(model.userId)
                                }
                                
                                MenuItem {
                                    text: "固定视频"
                                    onTriggered: root.pinParticipant(model.userId)
                                }
                                
                                MenuSeparator {}
                                
                                MenuItem {
                                    text: "移出会议"
                                    onTriggered: root.kickParticipant(model.userId)
                                }
                            }
                        }
                    }
                }
                
                // 分隔线
                Rectangle {
                    anchors.bottom: parent.bottom
                    anchors.left: parent.left
                    anchors.right: parent.right
                    anchors.leftMargin: 10
                    anchors.rightMargin: 10
                    height: 1
                    color: "#404040"
                }
            }
            
            // 空状态
            Label {
                anchors.centerIn: parent
                text: "暂无参与者"
                font.pixelSize: 14
                color: "#808080"
                visible: participantListView.count === 0
            }
        }
    }
}

