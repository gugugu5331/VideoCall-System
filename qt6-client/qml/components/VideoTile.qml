import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15
import QtMultimedia

/**
 * VideoTile - 视频瓦片组件
 * 
 * 用于显示单个参与者的视频流、用户信息和状态
 */
Rectangle {
    id: root
    
    // 公共属性
    property int userId: 0
    property string username: "未知用户"
    property bool videoEnabled: true
    property bool audioEnabled: true
    property bool isScreenSharing: false
    property bool isLocalUser: false
    property var videoOutput: null  // VideoOutput对象
    property var aiPanelController: null  // AI面板控制器
    
    // 样式属性 - 商务风格
    property color backgroundColor: "#0f172a"
    property color borderColor: isLocalUser ? "#3b82f6" : "#334155"
    property int borderWidth: 2
    property int cornerRadius: 12
    
    // 信号
    signal clicked()
    signal doubleClicked()
    
    color: backgroundColor
    radius: cornerRadius
    border.color: borderColor
    border.width: borderWidth
    
    // 鼠标区域
    MouseArea {
        anchors.fill: parent
        onClicked: root.clicked()
        onDoubleClicked: root.doubleClicked()
        hoverEnabled: true
        
        onEntered: {
            root.borderColor = "#60a5fa"
        }

        onExited: {
            root.borderColor = isLocalUser ? "#3b82f6" : "#334155"
        }
    }
    
    // 视频显示区域 - 商务风格
    Rectangle {
        id: videoContainer
        anchors.fill: parent
        anchors.margins: 6
        color: "#1e293b"
        radius: cornerRadius - 2
        clip: true
        
        // VideoOutput - 用于显示视频流
        Loader {
            id: videoLoader
            anchors.fill: parent
            active: videoEnabled && videoOutput !== null
            
            sourceComponent: Component {
                VideoOutput {
                    id: videoOutputItem
                    anchors.fill: parent
                    fillMode: VideoOutput.PreserveAspectCrop
                    
                    // 绑定到传入的videoOutput
                    Component.onCompleted: {
                        if (root.videoOutput) {
                            // 这里需要根据实际的MediaStream实现来绑定
                            // videoOutputItem.source = root.videoOutput
                        }
                    }
                }
            }
        }
        
        // 视频关闭时的占位符 - 商务风格
        Rectangle {
            anchors.fill: parent
            color: "#1e293b"
            visible: !videoEnabled

            ColumnLayout {
                anchors.centerIn: parent
                spacing: 15

                // 用户头像占位符
                Rectangle {
                    Layout.alignment: Qt.AlignHCenter
                    width: 90
                    height: 90
                    radius: 45

                    gradient: Gradient {
                        GradientStop { position: 0.0; color: "#2563eb" }
                        GradientStop { position: 1.0; color: "#1e40af" }
                    }

                    border.color: "#3b82f6"
                    border.width: 2

                    Text {
                        anchors.centerIn: parent
                        text: username.length > 0 ? username.charAt(0).toUpperCase() : "?"
                        font.pixelSize: 40
                        font.bold: true
                        font.family: "Microsoft YaHei"
                        color: "#ffffff"
                    }
                }

                Text {
                    Layout.alignment: Qt.AlignHCenter
                    text: "📷 摄像头已关闭"
                    font.pixelSize: 14
                    font.family: "Microsoft YaHei"
                    color: "#94a3b8"
                }
            }
        }
        
        // 屏幕共享标识 - 商务风格
        Rectangle {
            anchors.top: parent.top
            anchors.right: parent.right
            anchors.margins: 12
            width: 110
            height: 32
            radius: 6
            visible: isScreenSharing

            gradient: Gradient {
                GradientStop { position: 0.0; color: "#2563eb" }
                GradientStop { position: 1.0; color: "#1e40af" }
            }

            border.color: "#3b82f6"
            border.width: 1

            Text {
                anchors.centerIn: parent
                text: "🖥️ 屏幕共享"
                font.pixelSize: 12
                font.family: "Microsoft YaHei"
                color: "#ffffff"
            }
        }
        
        // 网络质量指示器 - 商务风格
        Row {
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.margins: 12
            spacing: 3

            Repeater {
                model: 3
                Rectangle {
                    width: 5
                    height: 10 + index * 5
                    color: index < 2 ? "#10b981" : "#64748b"
                    radius: 2
                }
            }
        }
    }
    
    // 底部用户信息栏 - 商务风格
    Rectangle {
        id: userInfoBar
        anchors.bottom: parent.bottom
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.margins: 12
        height: 44
        color: "#0f172a"
        opacity: 0.95
        radius: 8
        border.color: "#334155"
        border.width: 1
        
        RowLayout {
            anchors.fill: parent
            anchors.margins: 8
            spacing: 8
            
            // 用户名
            Text {
                Layout.fillWidth: true
                text: username + (isLocalUser ? " (我)" : "")
                font.pixelSize: 14
                font.bold: isLocalUser
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                elide: Text.ElideRight
            }

            // 音频状态图标 - 商务风格
            Rectangle {
                width: 30
                height: 30
                radius: 15
                color: audioEnabled ? "transparent" : "#dc2626"
                border.color: audioEnabled ? "#3b82f6" : "transparent"
                border.width: 1

                Text {
                    anchors.centerIn: parent
                    text: audioEnabled ? "🎤" : "🔇"
                    font.pixelSize: 16
                }
            }

            // 视频状态图标（小）
            Text {
                text: videoEnabled ? "📹" : "📷"
                font.pixelSize: 14
                color: videoEnabled ? "#60a5fa" : "#64748b"
                visible: !isScreenSharing
            }
        }
    }
    
    // 加载动画 - 商务风格
    BusyIndicator {
        anchors.centerIn: parent
        running: videoEnabled && videoOutput === null
        visible: running

        contentItem: Item {
            implicitWidth: 56
            implicitHeight: 56

            Rectangle {
                width: parent.width
                height: parent.height
                radius: width / 2
                color: "transparent"
                border.color: "#3b82f6"
                border.width: 4

                RotationAnimation on rotation {
                    from: 0
                    to: 360
                    duration: 1000
                    loops: Animation.Infinite
                    running: parent.parent.running
                }
            }
        }
    }
    
    // AI分析结果叠加层
    Column {
        anchors.left: parent.left
        anchors.top: parent.top
        anchors.margins: 12
        spacing: 6
        visible: !isLocalUser && aiPanelController  // 只对远程用户显示

        // 深度伪造检测结果
        Rectangle {
            width: deepfakeText.width + 16
            height: 28
            radius: 14
            color: "#1e293b"
            opacity: 0.9
            visible: deepfakeText.text !== ""

            property var detectionResult: aiPanelController ? aiPanelController.getDetectionResultForUser(userId) : null

            Text {
                id: deepfakeText
                anchors.centerIn: parent
                text: {
                    if (!parent.detectionResult || Object.keys(parent.detectionResult).length === 0) return ""
                    var isReal = parent.detectionResult.isReal
                    var confidence = Math.round(parent.detectionResult.confidence * 100)
                    return (isReal ? "✅ 真实" : "⚠️ 合成") + " (" + confidence + "%)"
                }
                color: parent.detectionResult && parent.detectionResult.isReal ? "#10b981" : "#f59e0b"
                font.pixelSize: 12
                font.bold: true
            }
        }

        // 情绪识别结果
        Rectangle {
            width: emotionText.width + 16
            height: 28
            radius: 14
            color: "#1e293b"
            opacity: 0.9
            visible: emotionText.text !== ""

            property var emotionResult: aiPanelController ? aiPanelController.getEmotionResultForUser(userId) : null

            Text {
                id: emotionText
                anchors.centerIn: parent
                text: {
                    if (!parent.emotionResult || Object.keys(parent.emotionResult).length === 0) return ""
                    var emotion = parent.emotionResult.emotion
                    var confidence = Math.round(parent.emotionResult.confidence * 100)
                    var emoji = getEmotionEmoji(emotion)
                    return emoji + " " + emotion + " (" + confidence + "%)"
                }
                color: "#60a5fa"
                font.pixelSize: 12
                font.bold: true
            }
        }

        // 最新语音识别结果
        Rectangle {
            width: Math.min(asrText.implicitWidth + 16, root.width - 24)
            height: asrText.implicitHeight + 12
            radius: 14
            color: "#1e293b"
            opacity: 0.9
            visible: asrText.text !== ""

            property var asrResults: aiPanelController ? aiPanelController.getAsrResultsForUser(userId) : []

            Text {
                id: asrText
                anchors.centerIn: parent
                width: parent.width - 16
                text: {
                    if (!parent.asrResults || parent.asrResults.length === 0) return ""
                    var latestResult = parent.asrResults[parent.asrResults.length - 1]
                    return "💬 " + latestResult.text
                }
                color: "#e2e8f0"
                font.pixelSize: 12
                wrapMode: Text.WordWrap
                maximumLineCount: 2
                elide: Text.ElideRight
            }
        }
    }

    // 悬停时显示的操作按钮
    Row {
        anchors.top: parent.top
        anchors.right: parent.right
        anchors.margins: 10
        spacing: 5
        visible: false  // 暂时隐藏，后续可以添加更多操作

        // 固定视频按钮
        Button {
            width: 32
            height: 32
            text: "📌"

            background: Rectangle {
                color: parent.pressed ? "#1890FF" : (parent.hovered ? "#3A3A3A" : "#2C2C2C")
                radius: 4
                opacity: 0.9
            }

            contentItem: Text {
                text: parent.text
                color: "#FFFFFF"
                horizontalAlignment: Text.AlignHCenter
                verticalAlignment: Text.AlignVCenter
                font.pixelSize: 14
            }
        }
    }

    // 辅助函数：根据情绪返回对应的emoji
    function getEmotionEmoji(emotion) {
        var emojiMap = {
            "happy": "😊",
            "sad": "😢",
            "angry": "😠",
            "neutral": "😐",
            "surprised": "😲",
            "fear": "😨",
            "disgust": "🤢"
        }
        return emojiMap[emotion] || "😐"
    }
}

