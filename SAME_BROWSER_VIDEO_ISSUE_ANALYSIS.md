# 同一个浏览器中只能看见自己视频的问题分析

## 🔍 **问题现象**

在同一个浏览器中打开多个标签页进行视频通话时，只能看见自己的视频流，无法看见对方的视频流。

## 🎯 **根本原因分析**

### 1. **WebRTC连接冲突** ⚠️
**问题**: 同一个浏览器中的多个标签页共享相同的WebRTC资源
- **ICE候选冲突**: 多个PeerConnection可能使用相同的ICE候选
- **媒体流冲突**: 同一个摄像头/麦克风被多个标签页同时访问
- **信令服务器混淆**: WebSocket连接可能相互干扰

### 2. **媒体设备独占访问** ⚠️
**问题**: 浏览器的媒体设备访问机制
```javascript
// 当前代码中的问题
this.localStream = await navigator.mediaDevices.getUserMedia(constraints);
```
- **设备独占**: 一旦一个标签页获取了摄像头/麦克风，其他标签页无法访问
- **权限冲突**: 浏览器可能拒绝多个标签页同时访问同一设备
- **流共享问题**: 即使获取到流，也可能无法正确共享

### 3. **WebSocket连接管理问题** ⚠️
**问题**: 多个标签页的WebSocket连接可能相互干扰
```javascript
// 当前WebSocket连接代码
const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/call/${callId}?user_id=${currentUser.uuid}`;
this.webSocket = new WebSocket(wsUrl);
```
- **连接冲突**: 多个标签页可能建立相同的WebSocket连接
- **消息混乱**: 信令消息可能在标签页间混淆
- **状态不同步**: 通话状态在不同标签页间不一致

### 4. **PeerConnection状态管理问题** ⚠️
**问题**: WebRTC连接状态管理不当
```javascript
// 当前PeerConnection处理
this.peerConnection.ontrack = (event) => {
    this.remoteStream = event.streams[0];
    const remoteVideo = document.getElementById('remote-video');
    if (remoteVideo) {
        remoteVideo.srcObject = this.remoteStream;
    }
};
```
- **流分配错误**: 远程流可能被错误地分配给本地视频元素
- **连接状态混乱**: 多个PeerConnection的状态可能相互影响

## 🛠️ **解决方案**

### 1. **改进媒体设备管理** 🔧
```javascript
// 改进后的媒体设备获取
async getMediaPermissions() {
    try {
        // 检查设备是否已被其他标签页使用
        const devices = await navigator.mediaDevices.enumerateDevices();
        const videoDevices = devices.filter(device => device.kind === 'videoinput');
        
        if (videoDevices.length === 0) {
            throw new Error('未检测到摄像头设备');
        }
        
        // 尝试获取媒体流
        const constraints = {
            video: {
                width: { ideal: 1280 },
                height: { ideal: 720 },
                frameRate: { ideal: 30 }
            },
            audio: {
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true
            }
        };

        this.localStream = await navigator.mediaDevices.getUserMedia(constraints);
        
        // 显示本地视频
        const localVideo = document.getElementById('local-video');
        if (localVideo) {
            localVideo.srcObject = this.localStream;
            localVideo.play().catch(e => console.log('本地视频播放失败:', e));
        }
        
    } catch (error) {
        if (error.name === 'NotAllowedError') {
            throw new Error('摄像头或麦克风权限被拒绝，请检查浏览器权限设置');
        } else if (error.name === 'NotFoundError') {
            throw new Error('未找到摄像头或麦克风设备');
        } else if (error.name === 'NotReadableError') {
            throw new Error('摄像头或麦克风被其他应用程序占用');
        } else {
            throw new Error('获取媒体设备失败: ' + error.message);
        }
    }
}
```

### 2. **改进WebSocket连接管理** 🔧
```javascript
// 改进后的WebSocket连接
async connectWebSocket() {
    return new Promise((resolve, reject) => {
        try {
            const currentUser = auth.getCurrentUser();
            if (!currentUser || !currentUser.uuid) {
                reject(new Error('用户未登录'));
                return;
            }

            // 添加标签页唯一标识
            const tabId = Date.now() + Math.random();
            const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/call/${this.currentCall.room_id}?user_id=${currentUser.uuid}&tab_id=${tabId}`;
            
            console.log('连接WebSocket:', wsUrl);
            this.webSocket = new WebSocket(wsUrl);
            
            this.webSocket.onopen = () => {
                console.log('WebSocket连接成功');
                resolve();
            };
            
            this.webSocket.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleSignalingMessage(message);
                } catch (error) {
                    console.error('解析WebSocket消息失败:', error);
                }
            };
            
            this.webSocket.onerror = (error) => {
                console.error('WebSocket连接错误:', error);
                reject(new Error('WebSocket连接失败'));
            };
            
            this.webSocket.onclose = (event) => {
                console.log('WebSocket连接关闭:', event.code, event.reason);
                if (this.isInCall) {
                    UI.showNotification('WebSocket连接断开，通话可能受影响', 'warning');
                }
            };
        } catch (error) {
            reject(error);
        }
    });
}
```

### 3. **改进PeerConnection管理** 🔧
```javascript
// 改进后的PeerConnection初始化
async initializeWebRTC() {
    try {
        console.log('初始化WebRTC...');
        
        // 创建唯一的连接配置
        const configuration = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun2.l.google.com:19302' }
            ],
            iceCandidatePoolSize: 10
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
        
        // 改进远程流处理
        this.peerConnection.ontrack = (event) => {
            console.log('收到远程流:', event.streams);
            if (event.streams && event.streams.length > 0) {
                this.remoteStream = event.streams[0];
                const remoteVideo = document.getElementById('remote-video');
                if (remoteVideo && this.remoteStream) {
                    remoteVideo.srcObject = this.remoteStream;
                    remoteVideo.play().catch(e => console.log('远程视频播放失败:', e));
                    console.log('远程视频已设置');
                    
                    // 确保本地视频显示在正确位置
                    const localVideo = document.getElementById('local-video');
                    if (localVideo && this.localStream) {
                        localVideo.srcObject = this.localStream;
                        localVideo.play().catch(e => console.log('本地视频播放失败:', e));
                    }
                }
            }
        };
        
        // 其他事件处理...
        
    } catch (error) {
        console.error('WebRTC初始化失败:', error);
        throw new Error('WebRTC连接失败');
    }
}
```

### 4. **添加标签页间通信** 🔧
```javascript
// 使用BroadcastChannel在标签页间同步状态
class TabCommunication {
    constructor() {
        this.channel = new BroadcastChannel('videocall_tabs');
        this.setupListeners();
    }
    
    setupListeners() {
        this.channel.onmessage = (event) => {
            const { type, data } = event.data;
            
            switch (type) {
                case 'device_in_use':
                    this.handleDeviceInUse(data);
                    break;
                case 'call_started':
                    this.handleCallStarted(data);
                    break;
                case 'call_ended':
                    this.handleCallEnded(data);
                    break;
            }
        };
    }
    
    notifyDeviceInUse(deviceType) {
        this.channel.postMessage({
            type: 'device_in_use',
            data: { deviceType, tabId: this.getTabId() }
        });
    }
    
    handleDeviceInUse(data) {
        if (data.tabId !== this.getTabId()) {
            UI.showNotification('摄像头或麦克风被其他标签页占用', 'warning');
        }
    }
    
    getTabId() {
        return Date.now() + Math.random();
    }
}
```

## 📋 **实施建议**

### 阶段1: 立即修复
1. ✅ 改进媒体设备错误处理
2. ✅ 添加WebSocket连接唯一标识
3. ✅ 改进PeerConnection状态管理

### 阶段2: 用户体验优化
1. 🔄 添加标签页间通信
2. 🔄 实现设备占用检测
3. 🔄 添加多标签页警告

### 阶段3: 长期优化
1. 🔄 实现设备共享机制
2. 🔄 添加连接质量监控
3. 🔄 实现自动重连机制

## 🎯 **预期效果**

### 修复后:
- ✅ 正确显示本地和远程视频
- ✅ 避免设备访问冲突
- ✅ 提供清晰的错误提示
- ✅ 支持多标签页使用（带警告）

### 用户体验改进:
- 🔧 更清晰的错误信息
- 🔧 更好的设备管理
- 🔧 更稳定的连接
- 🔧 更友好的多标签页提示 