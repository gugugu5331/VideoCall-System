import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

// 视频效果面板 - 商务风格
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
        // 标题 - 商务风格
        // ========================================
        Text {
            text: "🎨 视频效果"
            font.pixelSize: 22
            font.bold: true
            font.family: "Microsoft YaHei"
            color: "#f1f5f9"
            Layout.fillWidth: true
        }

        // ========================================
        // 美颜设置 - 商务风格
        // ========================================
        GroupBox {
            title: "✨ 美颜"
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
                
                // 美颜开关
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    
                    Text {
                        text: "启用美颜"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        Layout.fillWidth: true
                    }
                    
                    Switch {
                        id: beautySwitch
                        checked: controller ? controller.beautyEnabled : false
                        onToggled: {
                            if (controller) {
                                controller.beautyEnabled = checked
                            }
                        }
                    }
                }
                
                // 美颜预设
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    visible: beautySwitch.checked
                    
                    Text {
                        text: "预设"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                    }
                    
                    ComboBox {
                        id: beautyPresetCombo
                        Layout.fillWidth: true
                        model: controller ? controller.getBeautyPresets() : []
                        
                        background: Rectangle {
                            color: "#4A4A4A"
                            radius: 4
                            border.color: beautyPresetCombo.pressed ? "#1890FF" : "#5A5A5A"
                            border.width: 1
                        }
                        
                        contentItem: Text {
                            text: beautyPresetCombo.displayText
                            color: "#FFFFFF"
                            verticalAlignment: Text.AlignVCenter
                            leftPadding: 10
                        }
                        
                        onActivated: {
                            if (controller) {
                                controller.applyBeautyPreset(currentText)
                            }
                        }
                    }
                }
                
                // 磨皮强度
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 5
                    visible: beautySwitch.checked
                    
                    RowLayout {
                        Layout.fillWidth: true
                        
                        Text {
                            text: "磨皮强度"
                            color: "#FFFFFF"
                            font.pixelSize: 14
                            Layout.fillWidth: true
                        }
                        
                        Text {
                            text: beautyLevelSlider.value
                            color: "#1890FF"
                            font.pixelSize: 14
                            font.bold: true
                        }
                    }
                    
                    Slider {
                        id: beautyLevelSlider
                        Layout.fillWidth: true
                        from: 0
                        to: 100
                        value: controller ? controller.beautyLevel : 50
                        stepSize: 1
                        
                        onValueChanged: {
                            if (controller && pressed) {
                                controller.beautyLevel = value
                            }
                        }
                        
                        background: Rectangle {
                            x: beautyLevelSlider.leftPadding
                            y: beautyLevelSlider.topPadding + beautyLevelSlider.availableHeight / 2 - height / 2
                            width: beautyLevelSlider.availableWidth
                            height: 4
                            radius: 2
                            color: "#4A4A4A"
                            
                            Rectangle {
                                width: beautyLevelSlider.visualPosition * parent.width
                                height: parent.height
                                color: "#1890FF"
                                radius: 2
                            }
                        }
                        
                        handle: Rectangle {
                            x: beautyLevelSlider.leftPadding + beautyLevelSlider.visualPosition * (beautyLevelSlider.availableWidth - width)
                            y: beautyLevelSlider.topPadding + beautyLevelSlider.availableHeight / 2 - height / 2
                            width: 20
                            height: 20
                            radius: 10
                            color: beautyLevelSlider.pressed ? "#1890FF" : "#FFFFFF"
                            border.color: "#1890FF"
                            border.width: 2
                        }
                    }
                }
                
                // 美白强度
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 5
                    visible: beautySwitch.checked
                    
                    RowLayout {
                        Layout.fillWidth: true
                        
                        Text {
                            text: "美白强度"
                            color: "#FFFFFF"
                            font.pixelSize: 14
                            Layout.fillWidth: true
                        }
                        
                        Text {
                            text: whitenLevelSlider.value
                            color: "#1890FF"
                            font.pixelSize: 14
                            font.bold: true
                        }
                    }
                    
                    Slider {
                        id: whitenLevelSlider
                        Layout.fillWidth: true
                        from: 0
                        to: 100
                        value: controller ? controller.whitenLevel : 30
                        stepSize: 1
                        
                        onValueChanged: {
                            if (controller && pressed) {
                                controller.whitenLevel = value
                            }
                        }
                        
                        background: Rectangle {
                            x: whitenLevelSlider.leftPadding
                            y: whitenLevelSlider.topPadding + whitenLevelSlider.availableHeight / 2 - height / 2
                            width: whitenLevelSlider.availableWidth
                            height: 4
                            radius: 2
                            color: "#4A4A4A"
                            
                            Rectangle {
                                width: whitenLevelSlider.visualPosition * parent.width
                                height: parent.height
                                color: "#1890FF"
                                radius: 2
                            }
                        }
                        
                        handle: Rectangle {
                            x: whitenLevelSlider.leftPadding + whitenLevelSlider.visualPosition * (whitenLevelSlider.availableWidth - width)
                            y: whitenLevelSlider.topPadding + whitenLevelSlider.availableHeight / 2 - height / 2
                            width: 20
                            height: 20
                            radius: 10
                            color: whitenLevelSlider.pressed ? "#1890FF" : "#FFFFFF"
                            border.color: "#1890FF"
                            border.width: 2
                        }
                    }
                }
            }
        }
        
        // ========================================
        // 虚拟背景设置
        // ========================================
        GroupBox {
            title: "🖼️ 虚拟背景"
            Layout.fillWidth: true
            
            background: Rectangle {
                color: "#3A3A3A"
                radius: 8
                border.color: "#4A4A4A"
                border.width: 1
            }
            
            label: Text {
                text: parent.title
                color: "#FFFFFF"
                font.pixelSize: 16
                font.bold: true
                padding: 10
            }
            
            ColumnLayout {
                anchors.fill: parent
                spacing: 15
                
                // 虚拟背景开关
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    
                    Text {
                        text: "启用虚拟背景"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                        Layout.fillWidth: true
                    }
                    
                    Switch {
                        id: virtualBgSwitch
                        checked: controller ? controller.virtualBackgroundEnabled : false
                        onToggled: {
                            if (controller) {
                                controller.virtualBackgroundEnabled = checked
                            }
                        }
                    }
                }
                
                // 背景模式
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    visible: virtualBgSwitch.checked
                    
                    Text {
                        text: "背景模式"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                    }
                    
                    GridLayout {
                        Layout.fillWidth: true
                        columns: 2
                        rowSpacing: 10
                        columnSpacing: 10
                        
                        Button {
                            text: "🌫️ 模糊"
                            Layout.fillWidth: true
                            checkable: true
                            checked: controller ? controller.backgroundMode === 1 : false
                            onClicked: {
                                if (controller) {
                                    controller.backgroundMode = 1  // Blur
                                }
                            }
                            
                            background: Rectangle {
                                color: parent.checked ? "#1890FF" : "#4A4A4A"
                                radius: 4
                                border.color: parent.hovered ? "#1890FF" : "#5A5A5A"
                                border.width: 1
                            }
                            
                            contentItem: Text {
                                text: parent.text
                                color: "#FFFFFF"
                                horizontalAlignment: Text.AlignHCenter
                                verticalAlignment: Text.AlignVCenter
                            }
                        }
                        
                        Button {
                            text: "🖼️ 替换"
                            Layout.fillWidth: true
                            checkable: true
                            checked: controller ? controller.backgroundMode === 2 : false
                            onClicked: {
                                if (controller) {
                                    controller.backgroundMode = 2  // Replace
                                }
                            }
                            
                            background: Rectangle {
                                color: parent.checked ? "#1890FF" : "#4A4A4A"
                                radius: 4
                                border.color: parent.hovered ? "#1890FF" : "#5A5A5A"
                                border.width: 1
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
                
                // 背景图片选择
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 10
                    visible: virtualBgSwitch.checked && controller && controller.backgroundMode === 2
                    
                    Text {
                        text: "选择背景"
                        color: "#FFFFFF"
                        font.pixelSize: 14
                    }
                    
                    Button {
                        text: "📁 浏览图片..."
                        Layout.fillWidth: true
                        
                        background: Rectangle {
                            color: parent.pressed ? "#0D6EFD" : (parent.hovered ? "#1890FF" : "#4A4A4A")
                            radius: 4
                        }
                        
                        contentItem: Text {
                            text: parent.text
                            color: "#FFFFFF"
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                        }
                        
                        onClicked: {
                            // TODO: Open file dialog
                            console.log("Open file dialog for background image")
                        }
                    }
                    
                    Button {
                        text: "🗑️ 清除背景"
                        Layout.fillWidth: true
                        
                        background: Rectangle {
                            color: parent.pressed ? "#D32F2F" : (parent.hovered ? "#F44336" : "#4A4A4A")
                            radius: 4
                        }
                        
                        contentItem: Text {
                            text: parent.text
                            color: "#FFFFFF"
                            horizontalAlignment: Text.AlignHCenter
                            verticalAlignment: Text.AlignVCenter
                        }
                        
                        onClicked: {
                            if (controller) {
                                controller.clearBackgroundImage()
                            }
                        }
                    }
                }
            }
        }
        
        // 填充剩余空间
        Item {
            Layout.fillHeight: true
        }
        
        // ========================================
        // 状态信息
        // ========================================
        Rectangle {
            Layout.fillWidth: true
            height: 40
            color: "#3A3A3A"
            radius: 4
            visible: controller && controller.processing
            
            RowLayout {
                anchors.fill: parent
                anchors.margins: 10
                spacing: 10
                
                BusyIndicator {
                    running: true
                    Layout.preferredWidth: 20
                    Layout.preferredHeight: 20
                }
                
                Text {
                    text: "处理中..."
                    color: "#FFFFFF"
                    font.pixelSize: 12
                }
            }
        }
    }
}

