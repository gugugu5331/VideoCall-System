import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Layouts 1.15

Dialog {
    id: root
    title: qsTr("注册新账号")
    modal: true
    standardButtons: Dialog.Ok | Dialog.Cancel

    width: 480
    height: 550

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
        spacing: 15
        
        // Username
        Column {
            Layout.fillWidth: true
            spacing: 5
            
            Text {
                text: qsTr("用户名 *")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: usernameField
                width: parent.width
                placeholderText: qsTr("请输入用户名 (3-20个字符)")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                placeholderTextColor: "#64748b"

                background: Rectangle {
                    color: "#0f172a"
                    radius: 8
                    border.color: usernameField.activeFocus ? "#3b82f6" : "#475569"
                    border.width: 2
                }
            }
        }
        
        // Email
        Column {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: qsTr("邮箱 *")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: emailField
                width: parent.width
                placeholderText: qsTr("请输入邮箱地址")
                font.pixelSize: 13
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

        // Full Name
        Column {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: qsTr("姓名")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: fullNameField
                width: parent.width
                placeholderText: qsTr("请输入真实姓名")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                placeholderTextColor: "#64748b"

                background: Rectangle {
                    color: "#0f172a"
                    radius: 8
                    border.color: fullNameField.activeFocus ? "#3b82f6" : "#475569"
                    border.width: 2
                }
            }
        }

        // Password
        Column {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: qsTr("密码 *")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: passwordField
                width: parent.width
                placeholderText: qsTr("请输入密码 (至少6个字符)")
                echoMode: TextInput.Password
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                placeholderTextColor: "#64748b"

                background: Rectangle {
                    color: "#0f172a"
                    radius: 8
                    border.color: passwordField.activeFocus ? "#3b82f6" : "#475569"
                    border.width: 2
                }
            }
        }
        
        // Confirm Password
        Column {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: qsTr("确认密码 *")
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#cbd5e1"
                font.weight: Font.Medium
            }

            TextField {
                id: confirmPasswordField
                width: parent.width
                placeholderText: qsTr("请再次输入密码")
                echoMode: TextInput.Password
                font.pixelSize: 13
                font.family: "Microsoft YaHei"
                color: "#f1f5f9"
                placeholderTextColor: "#64748b"

                background: Rectangle {
                    color: "#0f172a"
                    radius: 8
                    border.color: confirmPasswordField.activeFocus ? "#3b82f6" : "#475569"
                    border.width: 2
                }
            }
        }

        // Error message
        Rectangle {
            id: errorBox
            Layout.fillWidth: true
            height: 40
            radius: 8
            visible: errorText.text.length > 0
            color: "#7f1d1d"
            border.color: "#dc2626"
            border.width: 2

            Text {
                id: errorText
                anchors.fill: parent
                anchors.margins: 10
                text: ""
                font.pixelSize: 12
                font.family: "Microsoft YaHei"
                color: "#fecaca"
                wrapMode: Text.WordWrap
                verticalAlignment: Text.AlignVCenter
            }
        }
        
        Item {
            Layout.fillHeight: true
        }
    }
    
    onAccepted: {
        // Validate inputs
        if (usernameField.text.length < 3) {
            errorText.text = qsTr("用户名至少需要3个字符")
            open()
            return
        }
        
        if (!emailField.text.includes("@")) {
            errorText.text = qsTr("请输入有效的邮箱地址")
            open()
            return
        }
        
        if (passwordField.text.length < 6) {
            errorText.text = qsTr("密码至少需要6个字符")
            open()
            return
        }
        
        if (passwordField.text !== confirmPasswordField.text) {
            errorText.text = qsTr("两次输入的密码不一致")
            open()
            return
        }
        
        // Call register service
        console.log("Registering user:", usernameField.text, emailField.text)
        authService.registerUser(
            usernameField.text,
            emailField.text,
            passwordField.text,
            fullNameField.text
        )
    }
    
    onRejected: {
        // Clear fields
        usernameField.text = ""
        emailField.text = ""
        fullNameField.text = ""
        passwordField.text = ""
        confirmPasswordField.text = ""
        errorText.text = ""
    }
    
    // Monitor registration result
    Connections {
        target: authService
        function onRegisterSuccess() {
            console.log("Registration successful")
            errorText.text = ""
            root.close()
        }
        
        function onRegisterFailed(message) {
            console.log("Registration failed:", message)
            errorText.text = message
            root.open()
        }
    }
}

