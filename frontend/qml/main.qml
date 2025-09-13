import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Controls.Material 2.15
import QtQuick.Layouts 1.15
import QtQuick.Window 2.15
import VideoConference 1.0

ApplicationWindow {
    id: mainWindow
    
    width: 1280
    height: 720
    minimumWidth: 800
    minimumHeight: 600
    
    visible: true
    title: qsTr("Video Conference Client")
    
    // Material Design主题配置
    Material.theme: Material.Dark
    Material.primary: Material.Blue
    Material.accent: Material.LightBlue
    
    property bool isLoggedIn: authService ? authService.isLoggedIn : false
    property bool isInMeeting: meetingService ? meetingService.isInMeeting : false
    
    // 主要内容区域
    StackView {
        id: stackView
        anchors.fill: parent
        
        initialItem: isLoggedIn ? mainView : loginView
        
        // 页面切换动画
        pushEnter: Transition {
            PropertyAnimation {
                property: "opacity"
                from: 0
                to: 1
                duration: 300
            }
        }
        
        pushExit: Transition {
            PropertyAnimation {
                property: "opacity"
                from: 1
                to: 0
                duration: 300
            }
        }
    }
    
    // 登录视图组件
    Component {
        id: loginView
        
        LoginView {
            onLoginSuccessful: {
                stackView.replace(mainView)
            }
        }
    }
    
    // 主视图组件
    Component {
        id: mainView
        
        Item {
            RowLayout {
                anchors.fill: parent
                spacing: 0
                
                // 侧边栏
                Rectangle {
                    Layout.preferredWidth: 300
                    Layout.fillHeight: true
                    color: Material.color(Material.Grey, Material.Shade900)
                    
                    ColumnLayout {
                        anchors.fill: parent
                        anchors.margins: 16
                        spacing: 16
                        
                        // 用户信息
                        Rectangle {
                            Layout.fillWidth: true
                            Layout.preferredHeight: 80
                            color: Material.color(Material.Grey, Material.Shade800)
                            radius: 8
                            
                            RowLayout {
                                anchors.fill: parent
                                anchors.margins: 12
                                
                                // 头像
                                Rectangle {
                                    Layout.preferredWidth: 56
                                    Layout.preferredHeight: 56
                                    radius: 28
                                    color: Material.primary
                                    
                                    Text {
                                        anchors.centerIn: parent
                                        text: authService && authService.currentUser ? 
                                              authService.currentUser.fullName.charAt(0).toUpperCase() : "U"
                                        color: "white"
                                        font.pixelSize: 24
                                        font.bold: true
                                    }
                                }
                                
                                // 用户信息
                                ColumnLayout {
                                    Layout.fillWidth: true
                                    spacing: 4
                                    
                                    Text {
                                        text: authService && authService.currentUser ? 
                                              authService.currentUser.fullName : "Unknown User"
                                        color: "white"
                                        font.pixelSize: 16
                                        font.bold: true
                                    }
                                    
                                    Text {
                                        text: authService && authService.currentUser ? 
                                              authService.currentUser.email : ""
                                        color: Material.color(Material.Grey, Material.Shade400)
                                        font.pixelSize: 12
                                    }
                                }
                                
                                // 设置按钮
                                Button {
                                    Layout.preferredWidth: 32
                                    Layout.preferredHeight: 32
                                    flat: true
                                    
                                    text: "⚙"
                                    font.pixelSize: 16
                                    
                                    onClicked: settingsDialog.open()
                                }
                            }
                        }
                        
                        // 导航菜单
                        ColumnLayout {
                            Layout.fillWidth: true
                            spacing: 8
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("会议")
                                flat: true
                                highlighted: stackView.currentItem && stackView.currentItem.objectName === "meetingView"
                                
                                onClicked: {
                                    if (isInMeeting) {
                                        stackView.replace(meetingView)
                                    } else {
                                        stackView.replace(meetingListView)
                                    }
                                }
                            }
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("检测记录")
                                flat: true
                                highlighted: stackView.currentItem && stackView.currentItem.objectName === "detectionView"
                                
                                onClicked: stackView.replace(detectionView)
                            }
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("历史记录")
                                flat: true
                                
                                onClicked: stackView.replace(historyView)
                            }
                        }
                        
                        Item {
                            Layout.fillHeight: true
                        }
                        
                        // 底部按钮
                        ColumnLayout {
                            Layout.fillWidth: true
                            spacing: 8
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("创建会议")
                                Material.background: Material.primary
                                
                                onClicked: createMeetingDialog.open()
                            }
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("加入会议")
                                
                                onClicked: joinMeetingDialog.open()
                            }
                            
                            Button {
                                Layout.fillWidth: true
                                text: qsTr("退出登录")
                                flat: true
                                
                                onClicked: {
                                    authService.logout()
                                    stackView.replace(loginView)
                                }
                            }
                        }
                    }
                }
                
                // 主内容区域
                Rectangle {
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    color: Material.backgroundColor
                    
                    StackView {
                        id: contentStack
                        anchors.fill: parent
                        
                        initialItem: isInMeeting ? meetingView : meetingListView
                    }
                }
            }
        }
    }
    
    // 会议列表视图
    Component {
        id: meetingListView
        
        Item {
            objectName: "meetingListView"
            
            ScrollView {
                anchors.fill: parent
                anchors.margins: 16
                
                ColumnLayout {
                    width: parent.width
                    spacing: 16
                    
                    Text {
                        text: qsTr("我的会议")
                        font.pixelSize: 24
                        font.bold: true
                        color: Material.foreground
                    }
                    
                    // 会议列表
                    Repeater {
                        model: meetingService ? meetingService.meetings : null
                        
                        delegate: Rectangle {
                            Layout.fillWidth: true
                            Layout.preferredHeight: 80
                            color: Material.color(Material.Grey, Material.Shade800)
                            radius: 8
                            
                            RowLayout {
                                anchors.fill: parent
                                anchors.margins: 16
                                
                                ColumnLayout {
                                    Layout.fillWidth: true
                                    
                                    Text {
                                        text: model.title || ""
                                        font.pixelSize: 16
                                        font.bold: true
                                        color: Material.foreground
                                    }
                                    
                                    Text {
                                        text: model.startTime ? 
                                              Qt.formatDateTime(model.startTime, "yyyy-MM-dd hh:mm") : ""
                                        font.pixelSize: 12
                                        color: Material.color(Material.Grey, Material.Shade400)
                                    }
                                }
                                
                                Button {
                                    text: qsTr("加入")
                                    Material.background: Material.primary
                                    
                                    onClicked: {
                                        meetingService.joinMeeting(model.id)
                                        contentStack.replace(meetingView)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    // 会议视图
    Component {
        id: meetingView
        
        MeetingView {
            objectName: "meetingView"
            
            onLeaveMeeting: {
                contentStack.replace(meetingListView)
            }
        }
    }
    
    // 检测视图
    Component {
        id: detectionView
        
        DetectionView {
            objectName: "detectionView"
        }
    }
    
    // 历史记录视图
    Component {
        id: historyView
        
        Item {
            Text {
                anchors.centerIn: parent
                text: qsTr("历史记录")
                font.pixelSize: 24
                color: Material.foreground
            }
        }
    }
    
    // 对话框
    Dialog {
        id: createMeetingDialog
        title: qsTr("创建会议")
        modal: true
        anchors.centerIn: parent
        
        ColumnLayout {
            TextField {
                id: meetingTitleField
                placeholderText: qsTr("会议标题")
                Layout.fillWidth: true
            }
            
            TextField {
                id: meetingDescField
                placeholderText: qsTr("会议描述")
                Layout.fillWidth: true
            }
        }
        
        standardButtons: Dialog.Ok | Dialog.Cancel
        
        onAccepted: {
            if (meetingTitleField.text.trim() !== "") {
                meetingService.createMeeting(meetingTitleField.text, meetingDescField.text)
                meetingTitleField.clear()
                meetingDescField.clear()
            }
        }
    }
    
    Dialog {
        id: joinMeetingDialog
        title: qsTr("加入会议")
        modal: true
        anchors.centerIn: parent
        
        TextField {
            id: joinCodeField
            placeholderText: qsTr("会议代码")
        }
        
        standardButtons: Dialog.Ok | Dialog.Cancel
        
        onAccepted: {
            if (joinCodeField.text.trim() !== "") {
                meetingService.joinMeetingByCode(joinCodeField.text)
                joinCodeField.clear()
                contentStack.replace(meetingView)
            }
        }
    }
    
    Dialog {
        id: settingsDialog
        title: qsTr("设置")
        modal: true
        anchors.centerIn: parent
        
        Text {
            text: qsTr("设置功能开发中...")
            color: Material.foreground
        }
        
        standardButtons: Dialog.Close
    }
    
    // 连接信号
    Connections {
        target: authService
        
        function onLoginStateChanged() {
            if (authService.isLoggedIn) {
                stackView.replace(mainView)
            } else {
                stackView.replace(loginView)
            }
        }
    }
    
    Connections {
        target: meetingService
        
        function onMeetingJoined() {
            contentStack.replace(meetingView)
        }
        
        function onMeetingLeft() {
            contentStack.replace(meetingListView)
        }
    }
}
