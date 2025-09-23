// 通话管理类
class CallManager {
    constructor() {
        this.localStream = null;
        this.remoteStream = null;
        this.peerConnection = null;
        this.webSocket = null;
        this.currentCall = null;
        this.isInCall = false;
        this.isMuted = false;
        this.isVideoEnabled = true;
        this.callStartTime = null;
        this.callDurationInterval = null;
        this.detectionInterval = null;
        this.iceCandidates = [];
        this.isInitiator = false;
        this.remoteUser = null;
        this.init();
    }

    // 初始化
    init() {
        this.setupEventListeners();
        
        // 检查是否有未接来电
        if (auth.isAuthenticated) {
            setTimeout(() => {
                this.checkIncomingCalls();
            }, 2000); // 延迟2秒检查，确保用户信息已加载
            
            // 启动实时来电检测
            this.startRealTimeCallDetection();
        }
        
        // 添加调试工具（开发环境）
        if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
            this.addDebugButtons();
        }
    }

    // 启动实时来电检测
    startRealTimeCallDetection() {
        console.log('启动实时来电检测...');
        
        // 方法1: 定时检查通话历史
        this.callCheckInterval = setInterval(() => {
            if (auth.isAuthenticated && !this.isInCall) {
                this.checkIncomingCalls();
            }
        }, 5000); // 每5秒检查一次
        
        // 方法2: 建立专门的WebSocket连接监听来电
        this.connectCallNotificationWebSocket();
        
        console.log('实时来电检测已启动');
    }

    // 停止实时来电检测
    stopRealTimeCallDetection() {
        console.log('停止实时来电检测...');
        
        if (this.callCheckInterval) {
            clearInterval(this.callCheckInterval);
            this.callCheckInterval = null;
        }
        
        if (this.notificationWebSocket) {
            this.notificationWebSocket.close();
            this.notificationWebSocket = null;
        }
        
        console.log('实时来电检测已停止');
    }

    // 连接来电通知WebSocket
    connectCallNotificationWebSocket() {
        try {
            const currentUser = auth.getCurrentUser();
            if (!currentUser || !currentUser.uuid) {
                console.error('用户信息不完整，无法建立通知WebSocket');
                return;
            }

            // 建立专门的通知WebSocket连接
            const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/notifications?user_id=${currentUser.uuid}`;
            console.log('连接来电通知WebSocket:', wsUrl);
            
            this.notificationWebSocket = new WebSocket(wsUrl);
            
            this.notificationWebSocket.onopen = () => {
                console.log('来电通知WebSocket连接已建立');
                // 发送订阅消息
                this.notificationWebSocket.send(JSON.stringify({
                    type: 'subscribe',
                    user_id: currentUser.uuid,
                    event: 'incoming_call'
                }));
            };
            
            this.notificationWebSocket.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log('收到通知消息:', message);
                    
                    if (message.type === 'incoming_call') {
                        this.handleIncomingCallNotification(message.data);
                    }
                } catch (error) {
                    console.error('解析通知消息失败:', error);
                }
            };
            
            this.notificationWebSocket.onerror = (error) => {
                console.error('来电通知WebSocket连接错误:', error);
            };
            
            this.notificationWebSocket.onclose = (event) => {
                console.log('来电通知WebSocket连接关闭:', event.code, event.reason);
                // 尝试重新连接
                setTimeout(() => {
                    if (auth.isAuthenticated && !this.isInCall) {
                        this.connectCallNotificationWebSocket();
                    }
                }, 5000);
            };
            
        } catch (error) {
            console.error('建立来电通知WebSocket失败:', error);
        }
    }

    // 处理实时来电通知
    handleIncomingCallNotification(callData) {
        console.log('收到实时来电通知:', callData);
        
        // 检查是否已经在通话中
        if (this.isInCall) {
            console.log('当前正在通话中，忽略来电');
            return;
        }
        
        // 检查是否已经显示过这个来电通知
        if (this.currentIncomingCall && this.currentIncomingCall.uuid === callData.uuid) {
            console.log('来电通知已显示，忽略重复通知');
            return;
        }
        
        // 显示来电通知
        this.showIncomingCallNotification(callData);
        
        // 播放来电铃声
        this.playIncomingCallSound();
        
        // 更新当前来电信息
        this.currentIncomingCall = callData;
    }

    // 播放来电铃声
    playIncomingCallSound() {
        try {
            // 创建音频上下文
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            
            // 生成来电铃声（简单的蜂鸣声）
            const oscillator = audioContext.createOscillator();
            const gainNode = audioContext.createGain();
            
            oscillator.connect(gainNode);
            gainNode.connect(audioContext.destination);
            
            // 设置音频参数
            oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
            oscillator.frequency.setValueAtTime(600, audioContext.currentTime + 0.5);
            
            gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
            gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 1);
            
            // 播放铃声
            oscillator.start(audioContext.currentTime);
            oscillator.stop(audioContext.currentTime + 1);
            
            // 重复播放
            this.ringtoneInterval = setInterval(() => {
                if (this.currentIncomingCall && !this.isInCall) {
                    this.playIncomingCallSound();
                } else {
                    clearInterval(this.ringtoneInterval);
                }
            }, 2000);
            
        } catch (error) {
            console.error('播放来电铃声失败:', error);
        }
    }

    // 停止来电铃声
    stopIncomingCallSound() {
        if (this.ringtoneInterval) {
            clearInterval(this.ringtoneInterval);
            this.ringtoneInterval = null;
        }
    }

    // 设置事件监听器
    setupEventListeners() {
        // 页面可见性变化时处理媒体流
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.pauseMediaStreams();
            } else {
                this.resumeMediaStreams();
            }
        });
    }

    // 开始通话
    async startCall(selectedUser = null) {
        if (!auth.checkAuth()) {
            return;
        }

        try {
            UI.showLoading();
            
            // 获取媒体权限
            await this.getMediaPermissions();
            
            // 确定被叫用户
            let calleeUUID;
            let calleeUsername;
            
            if (selectedUser) {
                // 使用选中的用户
                calleeUUID = selectedUser.uuid;
                calleeUsername = selectedUser.username;
            } else {
                // 如果没有选中用户，提示用户先搜索并选择
                UI.showNotification('请先搜索并选择要通话的用户', 'warning');
                UI.hideLoading();
                return;
            }
            
            const callData = {
                callee_id: calleeUUID,
                callee_username: calleeUsername,
                call_type: 'video'
            };

            const response = await api.startCall(callData);
            this.currentCall = response.call;
            this.isInitiator = true;
            this.remoteUser = selectedUser;
            
            // 建立WebSocket连接
            await this.connectWebSocket();
            
            // 初始化WebRTC
            await this.initializeWebRTC();
            
            // 开始通话计时
            this.startCallTimer();
            
            // 开始安全检测
            this.startSecurityDetection();
            
            this.isInCall = true;
            UI.updateCallStatus('connected');
            UI.showNotification(`正在与 ${calleeUsername} 通话`, 'success');
            
            // 通知其他标签页通话开始
            if (window.tabCommunication) {
                window.tabCommunication.notifyCallStarted({ 
                    user: selectedUser, 
                    callId: this.currentCall.uuid 
                });
            }
            
            // 更新UI
            document.getElementById('start-call-btn').style.display = 'none';
            document.getElementById('end-call-btn').style.display = 'flex';
            
        } catch (error) {
            console.error('开始通话失败:', error);
            UI.showNotification('开始通话失败: ' + error.message, 'error');
        } finally {
            UI.hideLoading();
        }
    }

    // 结束通话
    async endCall() {
        if (!this.currentCall) {
            return;
        }

        try {
            // 停止通话计时
            this.stopCallTimer();
            
            // 停止安全检测
            this.stopSecurityDetection();
            
            // 关闭WebSocket连接
            if (this.webSocket) {
                this.webSocket.close();
                this.webSocket = null;
            }
            
            // 结束通话
            await api.endCall(this.currentCall.uuid);
            
            // 停止媒体流
            this.stopMediaStreams();
            
            // 重置状态
            this.isInCall = false;
            this.currentCall = null;
            this.isInitiator = false;
            this.remoteUser = null;
            this.iceCandidates = [];
            
            UI.updateCallStatus('disconnected');
            UI.showNotification('通话已结束', 'info');
            
            // 通知其他标签页通话结束
            if (window.tabCommunication) {
                window.tabCommunication.notifyCallEnded({ 
                    callId: this.currentCall.uuid 
                });
            }
            
            // 更新UI
            document.getElementById('start-call-btn').style.display = 'flex';
            document.getElementById('end-call-btn').style.display = 'none';
            
        } catch (error) {
            console.error('结束通话失败:', error);
            UI.showNotification('结束通话失败', 'error');
        }
    }

    // 连接WebSocket
    async connectWebSocket() {
        return new Promise((resolve, reject) => {
            if (!this.currentCall || !this.currentCall.uuid) {
                reject(new Error('通话信息不完整'));
                return;
            }

            const currentUser = auth.getCurrentUser();
            if (!currentUser || !currentUser.uuid) {
                reject(new Error('用户信息不完整'));
                return;
            }

            // 添加标签页唯一标识
            const tabId = Date.now() + Math.random();
            this.tabId = tabId; // 保存标签页ID
            
            // 构建WebSocket URL，包含用户认证信息和标签页ID
            const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/call/${this.currentCall.uuid}?user_id=${currentUser.uuid}&tab_id=${tabId}`;
            
            console.log('连接WebSocket:', wsUrl);
            console.log('当前用户:', currentUser);
            console.log('通话信息:', this.currentCall);
            console.log('标签页ID:', tabId);
            
            this.webSocket = new WebSocket(wsUrl);
            
            this.webSocket.onopen = () => {
                console.log('WebSocket连接已建立');
                resolve();
            };
            
            this.webSocket.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log('收到WebSocket消息:', message);
                    
                    // 检查消息是否属于当前标签页
                    if (message.tab_id && message.tab_id !== this.tabId) {
                        console.log('收到其他标签页的消息，忽略:', message.tab_id);
                        return;
                    }
                    
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
        });
    }

    // 处理信令消息
    handleSignalingMessage(message) {
        console.log('收到信令消息:', message);
        
        switch (message.type) {
            case 'connection':
                this.handleConnectionMessage(message);
                break;
            case 'join':
                this.handleJoinMessage(message);
                break;
            case 'offer':
                this.handleOfferMessage(message);
                break;
            case 'answer':
                this.handleAnswerMessage(message);
                break;
            case 'ice_candidate':
                this.handleICECandidateMessage(message);
                break;
            case 'leave':
                this.handleLeaveMessage(message);
                break;
            default:
                console.log('未知消息类型:', message.type);
        }
    }

    // 处理连接消息
    handleConnectionMessage(message) {
        console.log('WebSocket连接成功:', message.data);
        if (message.data && message.data.room) {
            this.updateRoomInfo(message.data.room);
        }
    }

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
                // 作为接收者，等待offer
            } else {
                console.log('WebRTC连接未初始化，等待初始化...');
            }
        }
    }

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

    // 处理Answer消息
    async handleAnswerMessage(message) {
        console.log('收到Answer消息:', message);
        
        try {
            if (!this.peerConnection) {
                console.error('PeerConnection未初始化');
                return;
            }
            
            console.log('设置远程描述...');
            await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
            console.log('远程描述设置成功');
            
        } catch (error) {
            console.error('处理Answer失败:', error);
            UI.showNotification('处理通话应答失败: ' + error.message, 'error');
        }
    }

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

    // 处理离开消息
    handleLeaveMessage(message) {
        console.log('用户离开:', message.user_id);
        if (this.remoteUser && this.remoteUser.uuid === message.user_id) {
            this.remoteUser = null;
            UI.showNotification('对方已离开通话', 'warning');
        }
    }

    // 发送信令消息
    sendSignalingMessage(type, data) {
        if (this.webSocket && this.webSocket.readyState === WebSocket.OPEN) {
            const message = {
                type: type,
                call_id: this.currentCall.room_id,
                user_id: auth.getCurrentUser().uuid,
                data: data,
                timestamp: Date.now()
            };
            this.webSocket.send(JSON.stringify(message));
        }
    }

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

    // 获取媒体权限
    async getMediaPermissions() {
        try {
            // 检查设备是否已被其他标签页使用
            const devices = await navigator.mediaDevices.enumerateDevices();
            const videoDevices = devices.filter(device => device.kind === 'videoinput');
            const audioDevices = devices.filter(device => device.kind === 'audioinput');
            
            console.log(`检测到 ${videoDevices.length} 个摄像头设备, ${audioDevices.length} 个麦克风设备`);
            
            if (videoDevices.length === 0) {
                throw new Error('未检测到摄像头设备');
            }
            
            if (audioDevices.length === 0) {
                throw new Error('未检测到麦克风设备');
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
            
            // 通知其他标签页设备已被使用
            if (window.tabCommunication) {
                window.tabCommunication.notifyDeviceInUse('camera_microphone');
            }
            
            console.log('媒体设备获取成功');
            
        } catch (error) {
            console.error('获取媒体权限失败:', error);
            
            if (error.name === 'NotAllowedError') {
                throw new Error('摄像头或麦克风权限被拒绝，请检查浏览器权限设置');
            } else if (error.name === 'NotFoundError') {
                throw new Error('未找到摄像头或麦克风设备');
            } else if (error.name === 'NotReadableError') {
                throw new Error('摄像头或麦克风被其他应用程序占用，请关闭其他使用摄像头的应用');
            } else if (error.name === 'OverconstrainedError') {
                throw new Error('摄像头或麦克风不满足要求，请检查设备设置');
            } else if (error.name === 'TypeError') {
                throw new Error('浏览器不支持媒体设备访问');
            } else {
                throw new Error('获取媒体设备失败: ' + error.message);
            }
        }
    }

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
                if (event.streams && event.streams.length > 0) {
                    this.remoteStream = event.streams[0];
                    const remoteVideo = document.getElementById('remote-video');
                    if (remoteVideo && this.remoteStream) {
                        // 确保远程视频不被静音
                        remoteVideo.muted = false;
                        remoteVideo.srcObject = this.remoteStream;
                        
                        // 添加错误处理
                        remoteVideo.onerror = (error) => {
                            console.error('远程视频播放错误:', error);
                            UI.showNotification('远程视频播放失败', 'error');
                        };
                        
                        // 添加加载处理
                        remoteVideo.onloadedmetadata = () => {
                            console.log('远程视频元数据加载完成');
                            remoteVideo.play().catch(e => {
                                console.log('远程视频自动播放失败，尝试用户交互后播放:', e);
                                // 显示提示让用户点击播放
                                UI.showNotification('请点击视频区域开始播放', 'info');
                            });
                        };
                        
                        console.log('远程视频已设置');
                        
                        // 确保本地视频显示在正确位置
                        const localVideo = document.getElementById('local-video');
                        if (localVideo && this.localStream) {
                            localVideo.srcObject = this.localStream;
                            localVideo.play().catch(e => console.log('本地视频播放失败:', e));
                            console.log('本地视频已设置');
                        }
                    }
                }
            };
            
            // 处理ICE候选
            this.peerConnection.onicecandidate = (event) => {
                if (event.candidate) {
                    console.log('发送ICE候选:', event.candidate);
                    this.sendSignalingMessage('ice_candidate', event.candidate);
                }
            };
            
            // 处理连接状态变化
            this.peerConnection.onconnectionstatechange = () => {
                console.log('连接状态:', this.peerConnection.connectionState);
                this.updateConnectionStatus(this.peerConnection.connectionState);
                
                // 根据连接状态显示不同提示
                switch (this.peerConnection.connectionState) {
                    case 'connected':
                        UI.showNotification('连接已建立，音视频流正常', 'success');
                        break;
                    case 'disconnected':
                        UI.showNotification('连接断开，正在尝试重连...', 'warning');
                        break;
                    case 'failed':
                        UI.showNotification('连接失败，请检查网络设置', 'error');
                        break;
                    case 'closed':
                        UI.showNotification('连接已关闭', 'info');
                        break;
                }
            };
            
            // 处理ICE连接状态变化
            this.peerConnection.oniceconnectionstatechange = () => {
                console.log('ICE连接状态:', this.peerConnection.iceConnectionState);
            };
            
            // 处理ICE收集状态变化
            this.peerConnection.onicegatheringstatechange = () => {
                console.log('ICE收集状态:', this.peerConnection.iceGatheringState);
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

    // 更新连接状态
    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connection-status');
        if (statusElement) {
            let statusText = '';
            let statusClass = '';
            
            switch (status) {
                case 'new':
                    statusText = '正在连接...';
                    statusClass = 'connecting';
                    break;
                case 'connecting':
                    statusText = '正在连接...';
                    statusClass = 'connecting';
                    break;
                case 'connected':
                    statusText = '已连接';
                    statusClass = 'connected';
                    break;
                case 'disconnected':
                    statusText = '连接断开';
                    statusClass = 'disconnected';
                    break;
                case 'failed':
                    statusText = '连接失败';
                    statusClass = 'failed';
                    break;
                case 'closed':
                    statusText = '连接已关闭';
                    statusClass = 'closed';
                    break;
                default:
                    statusText = status;
                    statusClass = 'unknown';
            }
            
            statusElement.textContent = statusText;
            statusElement.className = `connection-status ${statusClass}`;
        }
    }

    // 更新房间信息
    updateRoomInfo(room) {
        const roomInfoElement = document.getElementById('room-info');
        if (roomInfoElement) {
            roomInfoElement.innerHTML = `
                <div class="room-details">
                    <span class="room-id">房间: ${room.id}</span>
                    <span class="call-type">${room.call_type === 'video' ? '视频通话' : '音频通话'}</span>
                    <span class="participants">参与者: ${Object.keys(room.users).length}</span>
                </div>
            `;
        }
    }

    // 开始通话计时
    startCallTimer() {
        this.callStartTime = Date.now();
        this.callDurationInterval = setInterval(() => {
            const duration = Date.now() - this.callStartTime;
            const minutes = Math.floor(duration / 60000);
            const seconds = Math.floor((duration % 60000) / 1000);
            const durationText = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            
            const durationElement = document.getElementById('call-duration');
            if (durationElement) {
                durationElement.textContent = durationText;
            }
        }, 1000);
    }

    // 停止通话计时
    stopCallTimer() {
        if (this.callDurationInterval) {
            clearInterval(this.callDurationInterval);
            this.callDurationInterval = null;
        }
        
        const durationElement = document.getElementById('call-duration');
        if (durationElement) {
            durationElement.textContent = '00:00';
        }
    }

    // 开始安全检测
    startSecurityDetection() {
        this.detectionInterval = setInterval(async () => {
            if (this.isInCall && this.currentCall) {
                await this.performSecurityDetection();
            }
        }, 10000); // 每10秒检测一次
    }

    // 停止安全检测
    stopSecurityDetection() {
        if (this.detectionInterval) {
            clearInterval(this.detectionInterval);
            this.detectionInterval = null;
        }
    }

    // 执行安全检测
    async performSecurityDetection() {
        try {
            // 获取视频帧数据
            const videoFrame = await this.captureVideoFrame();
            
            // 获取音频数据
            const audioData = await this.captureAudioData();
            
            // 发送检测请求
            const detectionRequest = {
                detection_id: `${this.currentCall.id}_${Date.now()}`,
                detection_type: 'spoofing',
                video_data: videoFrame,
                audio_data: audioData,
                metadata: {
                    call_id: this.currentCall.id,
                    timestamp: new Date().toISOString()
                }
            };

            const response = await api.detectSpoofing(detectionRequest);
            
            // 更新安全检测UI
            this.updateSecurityUI(response);
            
        } catch (error) {
            console.error('安全检测失败:', error);
        }
    }

    // 捕获视频帧
    async captureVideoFrame() {
        try {
            const video = document.getElementById('remote-video');
            if (!video || !video.videoWidth) {
                return null;
            }

            const canvas = document.createElement('canvas');
            canvas.width = video.videoWidth;
            canvas.height = video.videoHeight;
            
            const ctx = canvas.getContext('2d');
            ctx.drawImage(video, 0, 0);
            
            return canvas.toDataURL('image/jpeg', 0.8);
        } catch (error) {
            console.error('视频帧捕获失败:', error);
            return null;
        }
    }

    // 捕获音频数据
    async captureAudioData() {
        try {
            // 这里应该实现音频数据捕获
            // 由于浏览器限制，这里返回模拟数据
            return {
                sample_rate: 44100,
                channels: 1,
                data: new Array(1024).fill(0) // 模拟音频数据
            };
        } catch (error) {
            console.error('音频数据捕获失败:', error);
            return null;
        }
    }

    // 更新安全检测UI
    updateSecurityUI(detectionResult) {
        const riskScoreElement = document.getElementById('risk-score');
        const confidenceElement = document.getElementById('confidence');
        const securityStatusElement = document.getElementById('security-status');
        
        if (riskScoreElement) {
            riskScoreElement.textContent = detectionResult.risk_score.toFixed(2);
        }
        
        if (confidenceElement) {
            confidenceElement.textContent = `${(detectionResult.confidence * 100).toFixed(1)}%`;
        }
        
        if (securityStatusElement) {
            const riskScore = detectionResult.risk_score;
            const confidence = detectionResult.confidence;
            
            if (riskScore > 0.7 && confidence > 0.8) {
                securityStatusElement.innerHTML = '<i class="fas fa-exclamation-triangle"></i><span>检测到风险</span>';
                securityStatusElement.style.color = '#ef4444';
                UI.showNotification('检测到潜在安全风险', 'warning');
            } else {
                securityStatusElement.innerHTML = '<i class="fas fa-shield-alt"></i><span>安全</span>';
                securityStatusElement.style.color = '#10b981';
            }
        }
    }

    // 切换静音状态
    toggleMute() {
        if (!this.localStream) return;
        
        const audioTrack = this.localStream.getAudioTracks()[0];
        if (audioTrack) {
            audioTrack.enabled = !audioTrack.enabled;
            this.isMuted = !audioTrack.enabled;
            
            const muteBtn = document.getElementById('mute-btn');
            if (muteBtn) {
                if (this.isMuted) {
                    muteBtn.classList.add('active');
                    muteBtn.innerHTML = '<i class="fas fa-microphone-slash"></i>';
                } else {
                    muteBtn.classList.remove('active');
                    muteBtn.innerHTML = '<i class="fas fa-microphone"></i>';
                }
            }
        }
    }

    // 切换视频状态
    toggleVideo() {
        if (!this.localStream) return;
        
        const videoTrack = this.localStream.getVideoTracks()[0];
        if (videoTrack) {
            videoTrack.enabled = !videoTrack.enabled;
            this.isVideoEnabled = videoTrack.enabled;
            
            const videoBtn = document.getElementById('video-btn');
            if (videoBtn) {
                if (!this.isVideoEnabled) {
                    videoBtn.classList.add('active');
                    videoBtn.innerHTML = '<i class="fas fa-video-slash"></i>';
                } else {
                    videoBtn.classList.remove('active');
                    videoBtn.innerHTML = '<i class="fas fa-video"></i>';
                }
            }
        }
    }

    // 停止媒体流
    stopMediaStreams() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
        }
        
        if (this.peerConnection) {
            this.peerConnection.close();
            this.peerConnection = null;
        }
        
        // 清除视频元素
        const localVideo = document.getElementById('local-video');
        const remoteVideo = document.getElementById('remote-video');
        
        if (localVideo) localVideo.srcObject = null;
        if (remoteVideo) remoteVideo.srcObject = null;
    }

    // 暂停媒体流
    pauseMediaStreams() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                track.enabled = false;
            });
        }
    }

    // 恢复媒体流
    resumeMediaStreams() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                track.enabled = true;
            });
        }
    }

    // 获取通话状态
    getCallStatus() {
        return {
            isInCall: this.isInCall,
            isMuted: this.isMuted,
            isVideoEnabled: this.isVideoEnabled,
            callId: this.currentCall?.id,
            connectionState: this.peerConnection?.connectionState
        };
    }

    // 检查流状态
    checkStreamStatus() {
        const remoteVideo = document.getElementById('remote-video');
        const localVideo = document.getElementById('local-video');
        
        console.log('=== 流状态检查 ===');
        console.log('本地流:', this.localStream);
        console.log('远程流:', this.remoteStream);
        console.log('本地视频元素:', localVideo);
        console.log('远程视频元素:', remoteVideo);
        
        if (this.localStream) {
            const localTracks = this.localStream.getTracks();
            console.log('本地轨道:', localTracks.map(track => ({
                kind: track.kind,
                enabled: track.enabled,
                readyState: track.readyState
            })));
        }
        
        if (this.remoteStream) {
            const remoteTracks = this.remoteStream.getTracks();
            console.log('远程轨道:', remoteTracks.map(track => ({
                kind: track.kind,
                enabled: track.enabled,
                readyState: track.readyState
            })));
        }
        
        if (remoteVideo) {
            console.log('远程视频状态:', {
                srcObject: remoteVideo.srcObject,
                muted: remoteVideo.muted,
                paused: remoteVideo.paused,
                readyState: remoteVideo.readyState,
                networkState: remoteVideo.networkState
            });
        }
    }

    // 强制播放远程视频
    forcePlayRemote() {
        const remoteVideo = document.getElementById('remote-video');
        if (remoteVideo && this.remoteStream) {
            remoteVideo.muted = false;
            remoteVideo.srcObject = this.remoteStream;
            remoteVideo.play().then(() => {
                console.log('强制播放成功');
                UI.showNotification('远程视频播放成功', 'success');
            }).catch(e => {
                console.error('强制播放失败:', e);
                UI.showNotification('播放失败: ' + e.message, 'error');
            });
        }
    }

    // 添加调试工具
    addDebugButtons() {
        const debugContainer = document.createElement('div');
        debugContainer.innerHTML = `
            <div style="position: fixed; top: 10px; right: 10px; z-index: 1000; background: rgba(0,0,0,0.8); color: white; padding: 10px; border-radius: 5px;">
                <h4>调试工具</h4>
                <button onclick="window.callManager.checkStreamStatus()">检查流状态</button>
                <button onclick="window.callManager.forcePlayRemote()">强制播放远程</button>
            </div>
        `;
        document.body.appendChild(debugContainer);
    }

    // 检查未接来电
    async checkIncomingCalls() {
        try {
            const response = await api.getCallHistory(1, 10);
            const calls = response.calls || [];
            
            // 查找状态为initiated且被叫方是当前用户的通话
            const incomingCalls = calls.filter(call => 
                call.status === 'initiated' && 
                call.callee_uuid === auth.getCurrentUser().uuid
            );
            
            if (incomingCalls.length > 0) {
                const latestCall = incomingCalls[0];
                console.log('发现未接来电:', latestCall);
                
                // 显示来电通知
                this.showIncomingCallNotification(latestCall);
            }
        } catch (error) {
            console.error('检查未接来电失败:', error);
        }
    }

    // 显示来电通知
    showIncomingCallNotification(call) {
        const notification = document.createElement('div');
        notification.className = 'incoming-call-notification';
        notification.innerHTML = `
            <div class="notification-content">
                <h3>来电</h3>
                <p>${call.caller_username || '未知用户'} 正在呼叫您</p>
                <div class="notification-buttons">
                    <button onclick="window.callManager.acceptCall('${call.uuid}')" class="accept-btn">
                        <i class="fas fa-phone"></i> 接听
                    </button>
                    <button onclick="window.callManager.rejectCall('${call.uuid}')" class="reject-btn">
                        <i class="fas fa-phone-slash"></i> 拒绝
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // 5秒后自动移除通知
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 5000);
    }

    // 接受通话
    async acceptCall(callUUID) {
        try {
            console.log('接受通话:', callUUID);
            
            // 停止来电铃声
            this.stopIncomingCallSound();
            
            // 获取通话详情
            const response = await api.getCallDetails(callUUID);
            const call = response.call;
            
            // 设置当前通话
            this.currentCall = call;
            this.isInitiator = false;
            this.remoteUser = {
                uuid: call.caller_uuid,
                username: call.caller_username
            };
            
            // 获取媒体权限
            await this.getMediaPermissions();
            
            // 建立WebSocket连接
            await this.connectWebSocket();
            
            // 初始化WebRTC
            await this.initializeWebRTC();
            
            // 开始通话计时
            this.startCallTimer();
            
            // 开始安全检测
            this.startSecurityDetection();
            
            this.isInCall = true;
            UI.updateCallStatus('connected');
            UI.showNotification(`正在与 ${call.caller_username} 通话`, 'success');
            
            // 更新UI
            document.getElementById('start-call-btn').style.display = 'none';
            document.getElementById('end-call-btn').style.display = 'flex';
            
            // 移除通知
            const notification = document.querySelector('.incoming-call-notification');
            if (notification) {
                notification.parentNode.removeChild(notification);
            }
            
            // 清除当前来电信息
            this.currentIncomingCall = null;
            
        } catch (error) {
            console.error('接受通话失败:', error);
            UI.showNotification('接受通话失败: ' + error.message, 'error');
        }
    }

    // 拒绝通话
    async rejectCall(callUUID) {
        try {
            console.log('拒绝通话:', callUUID);
            
            // 停止来电铃声
            this.stopIncomingCallSound();
            
            // 调用结束通话API
            await api.endCall(callUUID);
            
            // 移除通知
            const notification = document.querySelector('.incoming-call-notification');
            if (notification) {
                notification.parentNode.removeChild(notification);
            }
            
            // 清除当前来电信息
            this.currentIncomingCall = null;
            
            UI.showNotification('已拒绝通话', 'info');
            
        } catch (error) {
            console.error('拒绝通话失败:', error);
            UI.showNotification('拒绝通话失败: ' + error.message, 'error');
        }
    }
}

// 全局函数 - 这些函数将在main.js中通过window.callManager调用
function startCall() {
    if (window.callManager) {
        window.callManager.startCall();
    } else {
        console.error('通话管理器未初始化');
        UI.showNotification('通话管理器未初始化', 'error');
    }
}

function endCall() {
    if (window.callManager) {
        window.callManager.endCall();
    } else {
        console.error('通话管理器未初始化');
        UI.showNotification('通话管理器未初始化', 'error');
    }
}

function toggleMute() {
    if (window.callManager) {
        window.callManager.toggleMute();
    } else {
        console.error('通话管理器未初始化');
        UI.showNotification('通话管理器未初始化', 'error');
    }
}

function toggleVideo() {
    if (window.callManager) {
        window.callManager.toggleVideo();
    } else {
        console.error('通话管理器未初始化');
        UI.showNotification('通话管理器未初始化', 'error');
    }
} 