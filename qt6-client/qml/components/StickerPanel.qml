import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15
import QtQuick.Dialogs

// 贴图面板 - 商务风格
Rectangle {
    id: root
    color: "#0f172a"
    border.color: "#1e293b"
    border.width: 1

    property var controller: null  // VideoEffectsController

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 25
        spacing: 25

        // ========================================
        // 标题
        // ========================================
        Text {
            text: "🎭 贴图效果"
            font.pixelSize: 22
            font.bold: true
            font.family: "Microsoft YaHei"
            color: "#f1f5f9"
            Layout.fillWidth: true
        }

        // ========================================
        // 贴图开关
        // ========================================
        GroupBox {
            title: "⚙️ 设置"
            Layout.fillWidth: true

            background: Rectangle {
                color: "#1e293b"
                radius: 12
                border.color: "#334155"
                border.width: 2
            }

            label: Text {
                text: parent.title
                color: "#f1f5f9"
                font.pixelSize: 16
                font.bold: true
                font.family: "Microsoft YaHei"
                padding: 12
            }
            
            ColumnLayout {
                anchors.fill: parent
                spacing: 15
                
                // 启用贴图
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    
                    Text {
                        text: "启用贴图"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        Layout.fillWidth: true
                    }
                    
                    Switch {
                        id: stickerSwitch
                        checked: controller ? controller.stickerEnabled : false
                        onToggled: {
                            if (controller) {
                                controller.stickerEnabled = checked
                            }
                        }
                    }
                }
                
                // 当前贴图数量
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    visible: stickerSwitch.checked
                    
                    Text {
                        text: "当前贴图"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        Layout.fillWidth: true
                    }
                    
                    Text {
                        text: controller ? controller.stickerCount : 0
                        color: "#1890FF"
                        font.pixelSize: 16
                        font.bold: true
                    }
                }
            }
        }

        // ========================================
        // 预设贴图
        // ========================================
        GroupBox {
            title: "📦 预设贴图"
            Layout.fillWidth: true
            visible: stickerSwitch.checked

            background: Rectangle {
                color: "#1e293b"
                radius: 12
                border.color: "#334155"
                border.width: 2
            }

            label: Text {
                text: parent.title
                color: "#f1f5f9"
                font.pixelSize: 16
                font.bold: true
                font.family: "Microsoft YaHei"
                padding: 12
            }
            
            ColumnLayout {
                anchors.fill: parent
                spacing: 15
                
                // 锚点类型选择
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    
                    Text {
                        text: "位置"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                    }
                    
                    ComboBox {
                        id: anchorTypeCombo
                        Layout.fillWidth: true
                        model: ["固定位置", "人脸中心", "左眼", "右眼", "鼻子", "嘴巴"]
                        currentIndex: 1  // 默认人脸中心
                        
                        background: Rectangle {
                            color: "#334155"
                            radius: 6
                            border.color: anchorTypeCombo.pressed ? "#1890FF" : "#475569"
                            border.width: 1
                        }
                        
                        contentItem: Text {
                            text: anchorTypeCombo.displayText
                            color: "#FFFFFF"
                            verticalAlignment: Text.AlignVCenter
                            leftPadding: 10
                            font.pixelSize: 14
                        }
                    }
                }
                
                // 预设贴图网格
                GridView {
                    id: presetGrid
                    Layout.fillWidth: true
                    Layout.preferredHeight: 200
                    cellWidth: 80
                    cellHeight: 100
                    clip: true
                    
                    model: controller ? controller.getPresetStickers() : []
                    
                    delegate: Item {
                        width: presetGrid.cellWidth
                        height: presetGrid.cellHeight
                        
                        Column {
                            anchors.centerIn: parent
                            spacing: 5
                            
                            Rectangle {
                                width: 60
                                height: 60
                                color: "#334155"
                                radius: 8
                                border.color: stickerMouseArea.containsMouse ? "#1890FF" : "#475569"
                                border.width: 2
                                
                                Text {
                                    anchors.centerIn: parent
                                    text: modelData.split(" ")[0]  // 只显示emoji
                                    font.pixelSize: 32
                                }
                                
                                MouseArea {
                                    id: stickerMouseArea
                                    anchors.fill: parent
                                    hoverEnabled: true
                                    cursorShape: Qt.PointingHandCursor
                                    
                                    onClicked: {
                                        if (controller) {
                                            var stickerId = controller.addPresetSticker(
                                                modelData,
                                                anchorTypeCombo.currentIndex
                                            )
                                            if (stickerId) {
                                                console.log("Added sticker:", stickerId)
                                            }
                                        }
                                    }
                                }
                            }
                            
                            Text {
                                text: modelData.split(" ")[1] || modelData  // 显示名称
                                color: "#94a3b8"
                                font.pixelSize: 11
                                anchors.horizontalCenter: parent.horizontalCenter
                            }
                        }
                    }
                    
                    ScrollBar.vertical: ScrollBar {
                        policy: ScrollBar.AsNeeded
                    }
                }
            }
        }

        // ========================================
        // 自定义贴图
        // ========================================
        GroupBox {
            title: "🖼️ 自定义贴图"
            Layout.fillWidth: true
            visible: stickerSwitch.checked

            background: Rectangle {
                color: "#1e293b"
                radius: 12
                border.color: "#334155"
                border.width: 2
            }

            label: Text {
                text: parent.title
                color: "#f1f5f9"
                font.pixelSize: 16
                font.bold: true
                font.family: "Microsoft YaHei"
                padding: 12
            }
            
            ColumnLayout {
                anchors.fill: parent
                spacing: 15
                
                Button {
                    text: "📁 选择图片"
                    Layout.fillWidth: true
                    
                    background: Rectangle {
                        color: parent.pressed ? "#0c4a6e" : (parent.hovered ? "#0369a1" : "#0284c7")
                        radius: 8
                    }
                    
                    contentItem: Text {
                        text: parent.text
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        font.bold: true
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                    }
                    
                    onClicked: {
                        fileDialog.open()
                    }
                }
                
                Button {
                    text: "🗑️ 清除所有贴图"
                    Layout.fillWidth: true
                    
                    background: Rectangle {
                        color: parent.pressed ? "#7f1d1d" : (parent.hovered ? "#991b1b" : "#dc2626")
                        radius: 8
                    }
                    
                    contentItem: Text {
                        text: parent.text
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        font.bold: true
                        horizontalAlignment: Text.AlignHCenter
                        verticalAlignment: Text.AlignVCenter
                    }
                    
                    onClicked: {
                        if (controller) {
                            controller.clearStickers()
                        }
                    }
                }
            }
        }

        // 填充空白
        Item {
            Layout.fillHeight: true
        }
    }

    // 文件选择对话框
    FileDialog {
        id: fileDialog
        title: "选择贴图图片"
        nameFilters: ["图片文件 (*.png *.jpg *.jpeg *.bmp)"]
        
        onAccepted: {
            if (controller) {
                var stickerId = controller.addSticker(
                    fileDialog.selectedFile,
                    anchorTypeCombo.currentIndex
                )
                if (stickerId) {
                    console.log("Added custom sticker:", stickerId)
                }
            }
        }
    }
}

