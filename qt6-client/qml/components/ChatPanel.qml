import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

/**
 * ChatPanel - 聊天面板组件
 * 
 * 显示聊天消息历史，支持发送文本消息
 */
Rectangle {
    id: root

    // 公共属性
    property var messagesModel: null  // ListModel
    property int currentUserId: 0
    property string currentUsername: "我"

    // 样式属性 - 商务风格
    property color backgroundColor: "#0f172a"
    property color headerColor: "#1e293b"
    property color inputBackgroundColor: "#1e293b"

    // 信号
    signal sendMessage(string content)
    signal loadMoreMessages()

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
                    text: "💬 聊天"
                    font.pixelSize: 20
                    font.bold: true
                    font.family: "Microsoft YaHei"
                    color: "#f1f5f9"
                }

                Item { Layout.fillWidth: true }

                // 清空聊天按钮 - 商务风格
                Button {
                    Layout.preferredWidth: 40
                    Layout.preferredHeight: 40
                    text: "🗑️"

                    background: Rectangle {
                        color: parent.hovered ? "#334155" : "transparent"
                        radius: 8
                        border.color: parent.hovered ? "#3b82f6" : "#475569"
                        border.width: 1
                    }
                    
                    contentItem: Text {
                        text: parent.text
                        color: "#FFFFFF"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 16
                    }
                    
                    ToolTip.visible: hovered
                    ToolTip.text: "清空聊天记录"
                }
            }
        }
        
        // 消息列表
        ListView {
            id: messageListView
            Layout.fillWidth: true
            Layout.fillHeight: true
            clip: true
            spacing: 10
            
            model: messagesModel
            
            ScrollBar.vertical: ScrollBar {
                policy: ScrollBar.AsNeeded
            }
            
            // 自动滚动到底部
            onCountChanged: {
                if (count > 0) {
                    positionViewAtEnd()
                }
            }
            
            delegate: Item {
                width: messageListView.width
                height: messageContainer.height + 10
                
                ColumnLayout {
                    id: messageContainer
                    anchors.left: parent.left
                    anchors.right: parent.right
                    anchors.margins: 10
                    spacing: 5
                    
                    // 消息头部（发送者和时间）
                    RowLayout {
                        Layout.fillWidth: true
                        spacing: 8
                        
                        Text {
                            text: model.fromUsername || "未知用户"
                            font.pixelSize: 12
                            font.bold: true
                            color: model.fromUserId === currentUserId ? "#1890FF" : "#52C41A"
                        }
                        
                        Text {
                            text: Qt.formatDateTime(model.timestamp, "hh:mm")
                            font.pixelSize: 11
                            color: "#808080"
                        }
                        
                        Item { Layout.fillWidth: true }
                    }
                    
                    // 消息内容
                    Rectangle {
                        Layout.fillWidth: true
                        Layout.preferredHeight: messageText.height + 20
                        color: model.fromUserId === currentUserId ? "#1890FF" : "#3A3A3A"
                        radius: 8
                        
                        Text {
                            id: messageText
                            anchors.fill: parent
                            anchors.margins: 10
                            text: model.content || ""
                            font.pixelSize: 14
                            color: "#FFFFFF"
                            wrapMode: Text.Wrap
                        }
                    }
                }
            }
            
            // 空状态
            Label {
                anchors.centerIn: parent
                text: "暂无消息\n开始聊天吧！"
                font.pixelSize: 14
                color: "#808080"
                horizontalAlignment: Text.AlignHCenter
                visible: messageListView.count === 0
            }
            
            // 加载更多指示器
            header: Item {
                width: messageListView.width
                height: 40
                visible: messageListView.count > 20
                
                Button {
                    anchors.centerIn: parent
                    text: "加载更多消息"
                    
                    background: Rectangle {
                        color: parent.hovered ? "#3A3A3A" : "transparent"
                        radius: 4
                    }
                    
                    contentItem: Text {
                        text: parent.text
                        color: "#1890FF"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 12
                    }
                    
                    onClicked: root.loadMoreMessages()
                }
            }
        }
        
        // 输入区域
        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 100
            color: inputBackgroundColor
            
            ColumnLayout {
                anchors.fill: parent
                anchors.margins: 10
                spacing: 8
                
                // 输入框
                ScrollView {
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    
                    TextArea {
                        id: messageInput
                        placeholderText: "输入消息..."
                        wrapMode: TextArea.Wrap
                        
                        background: Rectangle {
                            color: "#2C2C2C"
                            radius: 4
                            border.color: messageInput.activeFocus ? "#1890FF" : "#404040"
                            border.width: 1
                        }
                        
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        
                        // 支持Ctrl+Enter发送
                        Keys.onPressed: {
                            if ((event.key === Qt.Key_Return || event.key === Qt.Key_Enter) && 
                                (event.modifiers & Qt.ControlModifier)) {
                                sendButton.clicked()
                                event.accepted = true
                            }
                        }
                    }
                }
                
                // 底部工具栏
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 8
                    
                    // 表情按钮
                    Button {
                        Layout.preferredWidth: 36
                        Layout.preferredHeight: 36
                        text: "😊"
                        
                        background: Rectangle {
                            color: parent.hovered ? "#3A3A3A" : "transparent"
                            radius: 4
                        }
                        
                        contentItem: Text {
                            text: parent.text
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                            font.pixelSize: 18
                        }
                        
                        ToolTip.visible: hovered
                        ToolTip.text: "插入表情"
                    }
                    
                    // 文件按钮
                    Button {
                        Layout.preferredWidth: 36
                        Layout.preferredHeight: 36
                        text: "📎"
                        
                        background: Rectangle {
                            color: parent.hovered ? "#3A3A3A" : "transparent"
                            radius: 4
                        }
                        
                        contentItem: Text {
                            text: parent.text
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                            font.pixelSize: 18
                        }
                        
                        ToolTip.visible: hovered
                        ToolTip.text: "发送文件"
                    }
                    
                    Item { Layout.fillWidth: true }
                    
                    // 字数统计
                    Text {
                        text: messageInput.length + " / 500"
                        font.pixelSize: 11
                        color: messageInput.length > 500 ? "#F44336" : "#808080"
                    }
                    
                    // 发送按钮
                    Button {
                        id: sendButton
                        Layout.preferredWidth: 80
                        Layout.preferredHeight: 36
                        text: "发送"
                        enabled: messageInput.length > 0 && messageInput.length <= 500
                        
                        background: Rectangle {
                            color: {
                                if (!parent.enabled) return "#404040"
                                if (parent.pressed) return "#096DD9"
                                if (parent.hovered) return "#40A9FF"
                                return "#1890FF"
                            }
                            radius: 4
                        }
                        
                        contentItem: Text {
                            text: parent.text
                            color: parent.enabled ? "#FFFFFF" : "#808080"
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                            font.pixelSize: 14
                            font.bold: true
                        }
                        
                        onClicked: {
                            if (messageInput.length > 0) {
                                root.sendMessage(messageInput.text)
                                messageInput.clear()
                            }
                        }
                        
                        ToolTip.visible: hovered
                        ToolTip.text: "发送消息 (Ctrl+Enter)"
                    }
                }
            }
        }
    }
}

