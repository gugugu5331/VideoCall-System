import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

Dialog {
    id: root
    title: qsTr("找回密码")
    modal: true
    standardButtons: Dialog.Ok | Dialog.Cancel

    width: 450
    height: 300

    x: (parent.width - width) / 2
    y: (parent.height - height) / 2

    // 商务风格背景
    background: Rectangle {
        color: "#1e293b"
        radius: 12
        border.color: "#334155"
        border.width: 2
    }

    // 标题样式
    header: Rectangle {
        height: 60
        color: "#0f172a"
        radius: 12

        Text {
            anchors.centerIn: parent
            text: root.title
            font.pixelSize: 18
            font.bold: true
            font.family: "Microsoft YaHei"
            color: "#f1f5f9"
        }
    }

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 20
        spacing: 20

        Text {
            Layout.fillWidth: true
            text: qsTr("请输入您的注册邮箱，我们将发送重置密码的链接到您的邮箱。")
            font.pixelSize: 13
            font.family: "Microsoft YaHei"
            color: "#94a3b8"
            wrapMode: Text.WordWrap
        }

        // Email
        Column {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: qsTr("邮箱地址")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: emailField
                width: parent.width
                placeholderText: qsTr("请输入注册邮箱")
                font.pixelSize: 14
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                placeholderTextColor: "#64748b"

                background: Rectangle {
                    color: "#0f172a"
                    radius: 8
                    border.color: emailField.activeFocus ? "#3b82f6" : "#475569"
                    border.width: 2
                }
            }
        }

        // Success/Error message
        Rectangle {
            id: messageBox
            Layout.fillWidth: true
            height: 40
            radius: 8
            visible: messageText.text.length > 0
            color: messageText.text.includes("成功") ? "#14532d" : "#7f1d1d"
            border.color: messageText.text.includes("成功") ? "#16a34a" : "#dc2626"
            border.width: 2

            Text {
                id: messageText
                anchors.fill: parent
                anchors.margins: 10
                text: ""
                font.pixelSize: 12
                font.family: "Microsoft YaHei"
                color: messageText.text.includes("成功") ? "#86efac" : "#fecaca"
                wrapMode: Text.WordWrap
                verticalAlignment: Text.AlignVCenter
            }
        }
        
        Item {
            Layout.fillHeight: true
        }
    }
    
    onAccepted: {
        // Validate email
        if (!emailField.text.includes("@")) {
            messageText.text = qsTr("请输入有效的邮箱地址")
            open()
            return
        }
        
        // Call forgot password service
        console.log("Requesting password reset for:", emailField.text)
        authService.requestPasswordReset(emailField.text)
        
        // Show success message (temporary)
        messageText.text = qsTr("密码重置邮件已发送，请查收邮箱。")
        
        // Close after 2 seconds
        Qt.callLater(function() {
            root.close()
        })
    }
    
    onRejected: {
        // Clear fields
        emailField.text = ""
        messageText.text = ""
    }
    
    // Monitor password reset result
    Connections {
        target: authService
        function onPasswordResetSuccess() {
            console.log("Password reset email sent")
            messageText.text = qsTr("密码重置邮件已发送成功！")
        }
        
        function onPasswordResetFailed(message) {
            console.log("Password reset failed:", message)
            messageText.text = message
            root.open()
        }
    }
}

