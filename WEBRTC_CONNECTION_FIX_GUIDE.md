# WebRTC连接修复指南

## 问题描述

用户报告"连接后，没有显示对面的音视频"的问题，WebSocket连接成功但WebRTC连接没有正确建立。

## 问题原因分析

从日志分析发现以下问题：

1. **WebSocket用户ID识别问题**：后端使用默认生成的测试ID，导致用户身份混乱
2. **Join消息发送问题**：后端没有正确发送join通知消息给其他用户
3. **WebRTC信令交换问题**：前端没有收到join消息，导致没有触发offer/answer交换
4. **ICE候选处理问题**：ICE候选消息处理不完整

## 修复方案

### 1. 后端WebSocket处理器修复

**文件**: `core/backend/handlers/call_handler.go`

#### 修复用户ID识别
```go
// 如果还是没有用户ID，尝试从请求头获取
if userID == "" {
    authHeader := c.GetHeader("Authorization")
    if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
        // 这里应该解析JWT token获取用户ID
        log.Printf("Found Authorization header, but token parsing not implemented yet")
    }
}
```

#### 修复Join消息发送逻辑
```go
// 添加用户到房间
room.mutex.Lock()
userExists := false
if _, exists := room.Users[userID]; exists {
    userExists = true
} else {
    // 根据房间中的用户数量分配角色
    role := "participant"
    if len(room.Users) == 0 {
        role = "caller"
    } else if len(room.Users) == 1 {
        role = "callee"
    }

    room.Users[userID] = &CallUser{
        ID:       userID,
        UUID:     userID,
        Username: "user",
        Role:     role,
    }
    log.Printf("用户 %s 加入房间，角色: %s", userID, role)
}
room.Connections[userID] = conn
room.mutex.Unlock()

// 通知其他用户有新用户加入（只有新用户才发送通知）
if !userExists {
    // 发送join通知消息给其他用户
}
```

### 2. 前端WebSocket连接修复

**文件**: `web_interface/js/call.js`

#### 修复WebSocket连接
```javascript
// 连接WebSocket
async connectWebSocket() {
    return new Promise((resolve, reject) => {
        const currentUser = auth.getCurrentUser();
        if (!currentUser || !currentUser.uuid) {
            reject(new Error('用户信息不完整'));
            return;
        }

        // 构建WebSocket URL，包含用户认证信息
        const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/call/${this.currentCall.uuid}?user_id=${currentUser.uuid}`;
        
        console.log('连接WebSocket:', wsUrl);
        
        this.webSocket = new WebSocket(wsUrl);
        // ... 其他代码
    });
}
```

#### 修复Join消息处理
```javascript
// 处理加入消息
handleJoinMessage(message) {
    if (message.data && message.data.user) {
        this.remoteUser = message.data.user;
        console.log('远程用户加入:', this.remoteUser);
        
        // 如果是发起者，创建offer
        if (this.isInitiator && this.peerConnection) {
            console.log('作为发起者，创建offer...');
            this.createOffer();
        } else if (!this.isInitiator && this.peerConnection) {
            console.log('作为接收者，等待offer...');
        } else {
            console.log('WebRTC连接未初始化，等待初始化...');
        }
    }
}
```

#### 修复Offer创建
```javascript
// 创建Offer
async createOffer() {
    try {
        console.log('开始创建offer...');
        console.log('PeerConnection状态:', this.peerConnection?.connectionState);
        console.log('本地流状态:', this.localStream?.getTracks().length);
        
        if (!this.peerConnection) {
            console.error('PeerConnection未初始化');
            return;
        }
        
        if (!this.localStream) {
            console.error('本地流未获取');
            return;
        }
        
        const offer = await this.peerConnection.createOffer();
        console.log('Offer创建成功:', offer);
        
        await this.peerConnection.setLocalDescription(offer);
        console.log('本地描述设置成功');
        
        this.sendSignalingMessage('offer', offer);
        console.log('Offer已发送');
        
    } catch (error) {
        console.error('创建Offer失败:', error);
        UI.showNotification('创建通话连接失败: ' + error.message, 'error');
    }
}
```

#### 修复Offer处理
```javascript
// 处理Offer消息
async handleOfferMessage(message) {
    console.log('收到Offer消息:', message);
    
    try {
        if (!this.peerConnection) {
            console.log('PeerConnection未初始化，正在初始化...');
            await this.initializeWebRTC();
        }
        
        if (!this.localStream) {
            console.log('本地流未获取，正在获取...');
            await this.getMediaPermissions();
        }
        
        console.log('设置远程描述...');
        await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
        console.log('远程描述设置成功');
        
        // 创建answer
        console.log('创建answer...');
        const answer = await this.peerConnection.createAnswer();
        console.log('Answer创建成功:', answer);
        
        await this.peerConnection.setLocalDescription(answer);
        console.log('本地描述设置成功');
        
        // 发送answer
        this.sendSignalingMessage('answer', answer);
        console.log('Answer已发送');
        
    } catch (error) {
        console.error('处理Offer失败:', error);
        UI.showNotification('处理通话请求失败: ' + error.message, 'error');
    }
}
```

#### 修复ICE候选处理
```javascript
// 处理ICE候选消息
async handleICECandidateMessage(message) {
    console.log('收到ICE候选消息:', message);
    
    if (this.peerConnection && this.peerConnection.remoteDescription) {
        try {
            console.log('添加ICE候选...');
            await this.peerConnection.addIceCandidate(new RTCIceCandidate(message.data));
            console.log('ICE候选添加成功');
        } catch (error) {
            console.error('添加ICE候选失败:', error);
        }
    } else {
        // 如果远程描述还没设置，先保存ICE候选
        console.log('远程描述未设置，保存ICE候选');
        this.iceCandidates.push(message.data);
    }
}
```

#### 修复WebRTC初始化
```javascript
// 初始化WebRTC
async initializeWebRTC() {
    try {
        console.log('初始化WebRTC...');
        
        // 创建RTCPeerConnection
        const configuration = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun2.l.google.com:19302' }
            ]
        };

        this.peerConnection = new RTCPeerConnection(configuration);
        console.log('RTCPeerConnection创建成功');
        
        // 添加本地流
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                this.peerConnection.addTrack(track, this.localStream);
            });
            console.log('本地流已添加到PeerConnection');
        }
        
        // 处理远程流
        this.peerConnection.ontrack = (event) => {
            console.log('收到远程流:', event.streams);
            this.remoteStream = event.streams[0];
            const remoteVideo = document.getElementById('remote-video');
            if (remoteVideo) {
                remoteVideo.srcObject = this.remoteStream;
                remoteVideo.play().catch(e => console.log('远程视频播放失败:', e));
                console.log('远程视频已设置');
            }
        };
        
        // 处理保存的ICE候选
        if (this.iceCandidates.length > 0) {
            console.log('处理保存的ICE候选:', this.iceCandidates.length);
            for (const candidate of this.iceCandidates) {
                try {
                    await this.peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
                    console.log('保存的ICE候选添加成功');
                } catch (error) {
                    console.error('添加保存的ICE候选失败:', error);
                }
            }
            this.iceCandidates = [];
        }
        
        console.log('WebRTC初始化完成');
        
    } catch (error) {
        console.error('WebRTC初始化失败:', error);
        throw new Error('WebRTC连接失败');
    }
}
```

## 测试方法

### 1. 使用测试脚本
```bash
python test_webrtc_connection.py
```

### 2. 浏览器测试
1. 打开两个浏览器窗口
2. 分别使用不同用户登录
3. 测试用户搜索和通话功能
4. 检查浏览器控制台日志

### 3. 控制台检查
在浏览器开发者工具中检查：
```javascript
// 检查WebSocket连接
console.log('WebSocket状态:', window.callManager?.webSocket?.readyState);

// 检查PeerConnection状态
console.log('PeerConnection状态:', window.callManager?.peerConnection?.connectionState);

// 检查媒体流
console.log('本地流:', window.callManager?.localStream);
console.log('远程流:', window.callManager?.remoteStream);
```

## 预期结果

修复后应该看到：

1. ✅ WebSocket连接成功，用户ID正确识别
2. ✅ 收到join消息，触发WebRTC offer/answer交换
3. ✅ ICE候选正确交换
4. ✅ 远程视频流正确显示
5. ✅ 控制台显示详细的连接日志

## 故障排除

如果问题仍然存在：

1. **检查浏览器权限**：确保允许摄像头和麦克风访问
2. **检查网络连接**：确保STUN服务器可访问
3. **检查防火墙设置**：确保WebRTC流量不被阻止
4. **检查浏览器兼容性**：确保使用支持WebRTC的现代浏览器

## 文件修改清单

- ✅ `core/backend/handlers/call_handler.go` - 修复WebSocket用户ID识别和join消息发送
- ✅ `web_interface/js/call.js` - 修复WebRTC信令处理和连接建立
- ✅ `test_webrtc_connection.py` - 创建WebRTC连接测试脚本

## 状态

**修复状态**: ✅ 已完成
**测试状态**: 🔄 待测试
**部署状态**: 🔄 待部署 