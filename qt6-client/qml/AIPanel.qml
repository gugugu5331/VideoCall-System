import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

Rectangle {
    id: root
    color: "#0f172a"
    border.color: "#1e293b"
    border.width: 1

    property var controller

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        // 标签栏 - 商务风格
        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 60
            color: "#1e293b"
            border.color: "#334155"
            border.width: 1

            RowLayout {
                anchors.fill: parent
                spacing: 0

                TabButton {
                    id: detectionTab
                    text: "🤖 合成检测"
                    checked: true
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    background: Rectangle {
                        color: parent.checked ? "#0f172a" : "#1e293b"
                        border.color: parent.checked ? "#3b82f6" : "transparent"
                        border.width: parent.checked ? 2 : 0
                    }
                    contentItem: Text {
                        text: parent.text
                        color: parent.checked ? "#60a5fa" : "#94a3b8"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 14
                        font.family: "Microsoft YaHei"
                        font.bold: parent.checked
                    }
                }

                TabButton {
                    id: asrTab
                    text: "🎤 语音识别"
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    background: Rectangle {
                        color: parent.checked ? "#0f172a" : "#1e293b"
                        border.color: parent.checked ? "#3b82f6" : "transparent"
                        border.width: parent.checked ? 2 : 0
                    }
                    contentItem: Text {
                        text: parent.text
                        color: parent.checked ? "#60a5fa" : "#94a3b8"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 14
                        font.family: "Microsoft YaHei"
                        font.bold: parent.checked
                    }
                }
                
                TabButton {
                    id: emotionTab
                    text: "😊 情绪识别"
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    background: Rectangle {
                        color: parent.checked ? "#2C2C2C" : "#3A3A3A"
                    }
                    contentItem: Text {
                        text: parent.text
                        color: "#FFFFFF"
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                        font.pixelSize: 13
                    }
                }
            }
        }
        
        // 内容区
        StackLayout {
            Layout.fillWidth: true
            Layout.fillHeight: true
            currentIndex: detectionTab.checked ? 0 : (asrTab.checked ? 1 : 2)
            
            // 合成检测面板
            Rectangle {
                color: "#2C2C2C"
                
                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 15
                    spacing: 10
                    
                    RowLayout {
                        Layout.fillWidth: true
                        
                        Text {
                            text: "实时合成检测"
                            font.pixelSize: 16
                            font.bold: true
                            color: "#FFFFFF"
                        }
                        
                        Item { Layout.fillWidth: true }
                        
                        Switch {
                            checked: controller ? controller.detectionEnabled : false
                            onToggled: {
                                if (controller) {
                                    controller.enableDetection(checked)
                                }
                            }
                        }
                    }
                    
                    ScrollView {
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                        
                        ListView {
                            model: controller ? controller.detectionResults : []
                            spacing: 10
                            
                            delegate: Rectangle {
                                width: ListView.view.width
                                height: 80
                                color: "#3A3A3A"
                                radius: 6
                                
                                ColumnLayout {
                                    anchors.fill: parent
                                    anchors.margins: 10
                                    spacing: 5
                                    
                                    RowLayout {
                                        Layout.fillWidth: true
                                        
                                        Text {
                                            text: "用户 " + modelData.userId
                                            font.pixelSize: 14
                                            font.bold: true
                                            color: "#FFFFFF"
                                        }
                                        
                                        Item { Layout.fillWidth: true }
                                        
                                        Rectangle {
                                            width: 60
                                            height: 24
                                            radius: 12
                                            color: modelData.isReal ? "#4CAF50" : "#F44336"
                                            
                                            Text {
                                                anchors.centerIn: parent
                                                text: modelData.isReal ? "真实" : "合成"
                                                font.pixelSize: 12
                                                color: "#FFFFFF"
                                            }
                                        }
                                    }
                                    
                                    Text {
                                        text: "置信度: " + (modelData.confidence * 100).toFixed(1) + "%"
                                        font.pixelSize: 12
                                        color: "#B0B0B0"
                                    }
                                    
                                    Text {
                                        text: Qt.formatDateTime(modelData.timestamp, "hh:mm:ss")
                                        font.pixelSize: 11
                                        color: "#808080"
                                    }
                                }
                            }
                        }
                    }
                }
            }
            
            // ASR面板
            Rectangle {
                color: "#2C2C2C"
                
                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 15
                    spacing: 10
                    
                    RowLayout {
                        Layout.fillWidth: true
                        
                        Text {
                            text: "实时字幕"
                            font.pixelSize: 16
                            font.bold: true
                            color: "#FFFFFF"
                        }
                        
                        Item { Layout.fillWidth: true }
                        
                        Switch {
                            checked: controller ? controller.asrEnabled : false
                            onToggled: {
                                if (controller) {
                                    controller.enableASR(checked)
                                }
                            }
                        }
                    }
                    
                    ScrollView {
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                        
                        ListView {
                            model: controller ? controller.asrResults : []
                            spacing: 8
                            
                            delegate: Rectangle {
                                width: ListView.view.width
                                height: textContent.height + 30
                                color: "#3A3A3A"
                                radius: 6
                                
                                ColumnLayout {
                                    anchors.fill: parent
                                    anchors.margins: 10
                                    spacing: 5
                                    
                                    RowLayout {
                                        Layout.fillWidth: true
                                        
                                        Text {
                                            text: "用户 " + modelData.userId
                                            font.pixelSize: 12
                                            font.bold: true
                                            color: "#1890FF"
                                        }
                                        
                                        Item { Layout.fillWidth: true }
                                        
                                        Text {
                                            text: Qt.formatDateTime(modelData.timestamp, "hh:mm:ss")
                                            font.pixelSize: 11
                                            color: "#808080"
                                        }
                                    }
                                    
                                    Text {
                                        id: textContent
                                        Layout.fillWidth: true
                                        text: modelData.text
                                        font.pixelSize: 13
                                        color: "#FFFFFF"
                                        wrapMode: Text.WordWrap
                                    }
                                }
                            }
                        }
                    }
                }
            }
            
            // 情绪识别面板
            Rectangle {
                color: "#2C2C2C"
                
                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 15
                    spacing: 10
                    
                    RowLayout {
                        Layout.fillWidth: true
                        
                        Text {
                            text: "情绪分析"
                            font.pixelSize: 16
                            font.bold: true
                            color: "#FFFFFF"
                        }
                        
                        Item { Layout.fillWidth: true }
                        
                        Switch {
                            checked: controller ? controller.emotionEnabled : false
                            onToggled: {
                                if (controller) {
                                    controller.enableEmotion(checked)
                                }
                            }
                        }
                    }
                    
                    ScrollView {
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                        
                        ListView {
                            model: controller ? controller.emotionResults : []
                            spacing: 10
                            
                            delegate: Rectangle {
                                width: ListView.view.width
                                height: 100
                                color: "#3A3A3A"
                                radius: 6
                                
                                ColumnLayout {
                                    anchors.fill: parent
                                    anchors.margins: 10
                                    spacing: 5
                                    
                                    RowLayout {
                                        Layout.fillWidth: true
                                        
                                        Text {
                                            text: "用户 " + modelData.userId
                                            font.pixelSize: 14
                                            font.bold: true
                                            color: "#FFFFFF"
                                        }
                                        
                                        Item { Layout.fillWidth: true }
                                        
                                        Text {
                                            text: getEmotionEmoji(modelData.emotion)
                                            font.pixelSize: 24
                                        }
                                    }
                                    
                                    Text {
                                        text: "情绪: " + getEmotionText(modelData.emotion)
                                        font.pixelSize: 13
                                        color: "#FFFFFF"
                                    }
                                    
                                    Text {
                                        text: "置信度: " + (modelData.confidence * 100).toFixed(1) + "%"
                                        font.pixelSize: 12
                                        color: "#B0B0B0"
                                    }
                                    
                                    Text {
                                        text: Qt.formatDateTime(modelData.timestamp, "hh:mm:ss")
                                        font.pixelSize: 11
                                        color: "#808080"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    // 底部清空按钮
    Rectangle {
        anchors.bottom: parent.bottom
        anchors.left: parent.left
        anchors.right: parent.right
        height: 50
        color: "#3A3A3A"

        Button {
            anchors.centerIn: parent
            text: "清空所有结果"

            background: Rectangle {
                color: parent.pressed ? "#D32F2F" : (parent.hovered ? "#E53935" : "#F44336")
                radius: 4
            }

            contentItem: Text {
                text: parent.text
                color: "#FFFFFF"
                horizontalAlignment: Text.AlignHCenter
                verticalAlignment: Text.AlignVCenter
                font.pixelSize: 13
            }

            onClicked: {
                if (controller) {
                    controller.clearResults()
                }
            }
        }
    }

    // 连接控制器信号
    Connections {
        target: controller

        function onDetectionResultsChanged() {
            console.log("Detection results updated:", controller.detectionResults.length)
        }

        function onAsrResultsChanged() {
            console.log("ASR results updated:", controller.asrResults.length)
        }

        function onEmotionResultsChanged() {
            console.log("Emotion results updated:", controller.emotionResults.length)
        }
    }

    function getEmotionEmoji(emotion) {
        const emojiMap = {
            "happy": "😊",
            "sad": "😢",
            "angry": "😠",
            "surprised": "😲",
            "neutral": "😐",
            "fear": "😨",
            "disgust": "🤢"
        }
        return emojiMap[emotion] || "😐"
    }

    function getEmotionText(emotion) {
        const textMap = {
            "happy": "开心",
            "sad": "悲伤",
            "angry": "生气",
            "surprised": "惊讶",
            "neutral": "平静",
            "fear": "恐惧",
            "disgust": "厌恶"
        }
        return textMap[emotion] || "未知"
    }
}

