/**
 * WebRTC音视频通信核心模块
 */

class WebRTCManager {
    constructor() {
        this.localStream = null;
        this.peerConnections = new Map();
        this.signalingSocket = null;
        this.isConnected = false;
        this.isLeavingMeeting = false;
        this.currentUser = null;
        this.meetingId = null;
        
        // WebRTC配置
        this.rtcConfig = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun2.l.google.com:19302' }
            ]
        };
        
        // 媒体约束
        this.mediaConstraints = {
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
        
        this.isVideoEnabled = true;
        this.isAudioEnabled = true;
        this.isScreenSharing = false;

        // 主视频管理
        this.currentMainVideo = 'local'; // 当前主视频的用户ID
        this.mainVideoMuted = false;
        this.participants = new Map(); // 参与者信息
        this.participantInfo = new Map(); // 参与者详细信息
        this.isConnected = false; // 信令服务器连接状态
        this.pendingIceCandidates = new Map(); // 暂存ICE候选
    }
    
    /**
     * 初始化WebRTC
     */
    async initialize() {
        try {
            // 获取本地媒体流
            await this.getUserMedia();

            // 连接信令服务器
            await this.connectSignaling();

            console.log('WebRTC初始化成功');
            return true;
        } catch (error) {
            console.error('WebRTC初始化失败:', error);
            throw error;
        }
    }
    
    /**
     * 获取用户媒体
     */
    async getUserMedia() {
        try {
            this.localStream = await navigator.mediaDevices.getUserMedia(this.mediaConstraints);
            
            // 显示本地视频
            const localVideo = document.getElementById('localVideo');
            if (localVideo) {
                localVideo.srcObject = this.localStream;
            }
            
            console.log('获取本地媒体流成功');
            return this.localStream;
        } catch (error) {
            console.error('获取媒体流失败:', error);
            throw new Error('无法访问摄像头或麦克风，请检查权限设置');
        }
    }
    
    /**
     * 连接信令服务器
     */
    connectSignaling() {
        if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
            console.log('信令服务器已连接，跳过重复连接');
            return Promise.resolve();
        }

        if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.CONNECTING) {
            console.log('信令服务器正在连接中，跳过重复连接');
            return Promise.resolve();
        }

        console.log('连接信令服务器...');
        this.updateStatus('连接中...', 'info');

        const wsUrl = `ws://localhost:8081/signaling`;
        this.signalingSocket = new WebSocket(wsUrl);

        return new Promise((resolve, reject) => {
            this.signalingSocket.onopen = () => {
                console.log('信令服务器连接成功');
                this.isConnected = true;
                this.updateStatus('已连接', 'success');

                // 发送暂存的ICE候选
                this.sendPendingIceCandidates();
                resolve();
            };
        
            this.signalingSocket.onmessage = (event) => {
                this.handleSignalingMessage(JSON.parse(event.data));
            };

            this.signalingSocket.onclose = () => {
                console.log('信令服务器连接断开');
                this.isConnected = false;
                this.updateStatus('连接断开', 'error');

                // 清理连接状态
                this.signalingSocket = null;

                // 只在意外断开时才重连（不是主动离开）
                if (!document.hidden && !this.isLeavingMeeting) {
                    setTimeout(() => {
                        if (!this.isConnected && !this.signalingSocket && !this.isLeavingMeeting) {
                            console.log('检测到意外断开，尝试重新连接信令服务器...');
                            this.connectSignaling();
                        }
                    }, 3000);
                }
            };

            this.signalingSocket.onerror = (error) => {
                console.error('信令服务器错误:', error);
                this.updateStatus('连接错误', 'error');
                reject(error);
            };
        });
    }
    
    /**
     * 处理信令消息
     */
    async handleSignalingMessage(message) {
        console.log('收到信令消息:', message);

        switch (message.type) {
            case 'welcome':
                console.log('服务器欢迎消息:', message.message);
                break;

            case 'user-joined':
                console.log('处理用户加入消息:', message.data);
                await this.handleUserJoined(message.data);
                break;

            case 'user-left':
                console.log('处理用户离开消息:', message.data);
                this.handleUserLeft(message.data);
                break;
                
            case 'offer':
                await this.handleOffer(message);
                break;
                
            case 'answer':
                await this.handleAnswer(message);
                break;
                
            case 'ice-candidate':
                await this.handleIceCandidate(message);
                break;
                
            case 'chat-message':
                this.handleChatMessage(message.data);
                break;

            case 'presenter-set':
            case 'presenter-removed':
            case 'presenter-changed':
                this.handlePresenterMessage(message.type, message.data);
                break;

            default:
                console.log('未知信令消息类型:', message.type);
        }
    }
    
    /**
     * 加入会议
     */
    async joinMeeting(username, meetingId) {
        this.currentUser = {
            id: 'user_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9),
            name: username
        };
        this.meetingId = meetingId;
        this.pendingIceCandidates = new Map();

        console.log('加入会议:', this.currentUser, meetingId);

        // 确保信令服务器连接
        if (!this.signalingSocket || this.signalingSocket.readyState !== WebSocket.OPEN) {
            console.log('建立信令服务器连接...');
            await this.connectSignaling();
        }

        // 发送加入会议消息
        console.log('发送加入会议消息...', {
            type: 'join-meeting',
            data: {
                meetingId: meetingId,
                user: this.currentUser
            }
        });
        this.sendSignalingMessage({
            type: 'join-meeting',
            data: {
                meetingId: meetingId,
                user: this.currentUser
            }
        });

        // 更新UI - 安全地更新DOM元素
        const localUserNameElement = document.getElementById('localUserName');
        if (localUserNameElement) {
            localUserNameElement.textContent = username;
        }

        this.updateParticipantsList();
        this.isConnected = true;

        console.log('会议加入请求已发送');
    }

    /**
     * 等待信令服务器连接
     */
    async waitForSignalingConnection(timeout = 5000) {
        return new Promise((resolve, reject) => {
            if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
                resolve();
                return;
            }

            const startTime = Date.now();
            const checkConnection = () => {
                if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
                    resolve();
                } else if (Date.now() - startTime > timeout) {
                    reject(new Error('信令服务器连接超时'));
                } else {
                    setTimeout(checkConnection, 100);
                }
            };

            checkConnection();
        });
    }

    /**
     * 发送暂存的ICE候选
     */
    sendPendingIceCandidates() {
        console.log('发送暂存的ICE候选，暂存数量:', this.pendingIceCandidates.size);

        for (const [userId, candidates] of this.pendingIceCandidates) {
            console.log(`发送用户 ${userId} 的 ${candidates.length} 个暂存ICE候选`);

            candidates.forEach(candidate => {
                this.sendSignalingMessage(candidate);
            });
        }

        // 清空暂存的ICE候选
        this.pendingIceCandidates.clear();
    }
    
    /**
     * 处理用户加入
     */
    async handleUserJoined(userData) {
        console.log('用户加入:', userData);

        // 检查是否是自己
        if (userData.id === this.currentUser.id) {
            console.log('忽略自己的加入消息');
            return;
        }

        // 检查是否已经存在连接
        if (this.peerConnections.has(userData.id)) {
            console.log('用户连接已存在:', userData.id);
            return;
        }

        // 添加参与者信息
        this.addParticipantInfo(userData.id, userData.name);

        // 添加参与者到UI
        this.addParticipant(userData);

        // 创建对等连接
        const peerConnection = await this.createPeerConnection(userData.id);

        // 添加本地流
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                console.log('添加本地轨道:', track.kind);
                peerConnection.addTrack(track, this.localStream);
            });
        }

        // 创建并发送offer
        try {
            const offer = await peerConnection.createOffer({
                offerToReceiveAudio: true,
                offerToReceiveVideo: true
            });
            await peerConnection.setLocalDescription(offer);

            console.log('发送offer给:', userData.id);
            this.sendSignalingMessage({
                type: 'offer',
                to: userData.id,
                data: offer
            });
        } catch (error) {
            console.error('创建offer失败:', error);
        }
    }
    
    /**
     * 处理用户离开
     */
    handleUserLeft(userData) {
        console.log('用户离开:', userData);

        // 移除参与者信息
        this.removeParticipantInfo(userData.id);

        // 关闭对等连接
        if (this.peerConnections.has(userData.id)) {
            this.peerConnections.get(userData.id).close();
            this.peerConnections.delete(userData.id);
        }

        // 移除视频元素
        const videoContainer = document.getElementById(`video-${userData.id}`);
        if (videoContainer) {
            videoContainer.remove();
        }

        // 更新参与者列表
        this.removeParticipant(userData.id);
    }
    
    /**
     * 处理offer
     */
    async handleOffer(message) {
        console.log('收到offer from:', message.from);

        try {
            // 检查是否已经存在连接
            let peerConnection = this.peerConnections.get(message.from);
            if (!peerConnection) {
                peerConnection = await this.createPeerConnection(message.from);

                // 添加本地流到新创建的连接
                if (this.localStream) {
                    this.localStream.getTracks().forEach(track => {
                        console.log('添加本地轨道到新连接:', track.kind);
                        peerConnection.addTrack(track, this.localStream);
                    });
                }
            }

            // 设置远程描述
            await peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
            console.log('设置远程描述成功');

            // 创建并发送answer
            const answer = await peerConnection.createAnswer();
            await peerConnection.setLocalDescription(answer);

            console.log('发送answer给:', message.from);
            this.sendSignalingMessage({
                type: 'answer',
                to: message.from,
                data: answer
            });

            // 确保参与者信息存在
            if (!this.participants.has(message.from)) {
                // 从消息中获取用户信息，或使用默认值
                const userName = message.userName || `用户 ${message.from}`;
                console.log('从offer消息添加参与者:', message.from, userName);
                this.addParticipantInfo(message.from, userName);
                this.addParticipant({ id: message.from, name: userName });
            }

        } catch (error) {
            console.error('处理offer失败:', error);
        }
    }
    
    /**
     * 处理answer
     */
    async handleAnswer(message) {
        console.log('收到answer from:', message.from);

        try {
            const peerConnection = this.peerConnections.get(message.from);
            if (peerConnection) {
                await peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
                console.log('设置远程answer描述成功');
            } else {
                console.error('找不到对应的peer connection:', message.from);
            }
        } catch (error) {
            console.error('处理answer失败:', error);
        }
    }
    
    /**
     * 处理ICE候选
     */
    async handleIceCandidate(message) {
        console.log('收到ICE候选 from:', message.from);

        try {
            const peerConnection = this.peerConnections.get(message.from);
            if (peerConnection && peerConnection.remoteDescription) {
                await peerConnection.addIceCandidate(new RTCIceCandidate(message.data));
                console.log('添加ICE候选成功');
            } else {
                console.log('等待远程描述设置完成，暂存ICE候选');
                // 暂存ICE候选，等待远程描述设置完成
                if (!this.pendingIceCandidates) {
                    this.pendingIceCandidates = new Map();
                }
                if (!this.pendingIceCandidates.has(message.from)) {
                    this.pendingIceCandidates.set(message.from, []);
                }
                this.pendingIceCandidates.get(message.from).push(message.data);
            }
        } catch (error) {
            console.error('处理ICE候选失败:', error);
        }
    }
    
    /**
     * 创建对等连接
     */
    async createPeerConnection(userId) {
        const peerConnection = new RTCPeerConnection(this.rtcConfig);
        
        // ICE候选事件
        peerConnection.onicecandidate = (event) => {
            if (event.candidate) {
                // 检查信令服务器连接状态
                if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
                    this.sendSignalingMessage({
                        type: 'ice-candidate',
                        to: userId,
                        data: event.candidate
                    });
                } else {
                    // 如果信令服务器未连接，暂存ICE候选
                    console.log('信令服务器未连接，暂存ICE候选');
                    if (!this.pendingIceCandidates.has(userId)) {
                        this.pendingIceCandidates.set(userId, []);
                    }
                    this.pendingIceCandidates.get(userId).push({
                        type: 'ice-candidate',
                        to: userId,
                        data: event.candidate
                    });
                }
            }
        };
        
        // 远程流事件
        peerConnection.ontrack = (event) => {
            console.log('收到远程流:', event);
            this.handleRemoteStream(userId, event.streams[0]);
        };

        // 连接状态变化时处理暂存的ICE候选
        peerConnection.addEventListener('signalingstatechange', () => {
            if (peerConnection.signalingState === 'stable') {
                this.processPendingIceCandidates(userId);
            }
        });
        
        // 连接状态变化
        peerConnection.onconnectionstatechange = () => {
            console.log(`连接状态变化 ${userId}:`, peerConnection.connectionState);
        };
        
        this.peerConnections.set(userId, peerConnection);
        return peerConnection;
    }

    /**
     * 处理暂存的ICE候选
     */
    async processPendingIceCandidates(userId) {
        if (this.pendingIceCandidates && this.pendingIceCandidates.has(userId)) {
            const candidates = this.pendingIceCandidates.get(userId);
            const peerConnection = this.peerConnections.get(userId);

            if (peerConnection && peerConnection.remoteDescription) {
                console.log(`处理${candidates.length}个暂存的ICE候选:`, userId);
                for (const candidate of candidates) {
                    try {
                        await peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
                    } catch (error) {
                        console.error('添加暂存ICE候选失败:', error);
                    }
                }
                this.pendingIceCandidates.delete(userId);
            }
        }
    }
    
    /**
     * 处理远程流
     */
    handleRemoteStream(userId, stream) {
        console.log('处理远程流:', userId, stream);

        // 存储参与者流信息
        if (this.participants.has(userId)) {
            this.participants.get(userId).stream = stream;
        }

        // 更新缩略图视频
        const thumbnailVideo = document.getElementById(`thumbnail-${userId}`);
        if (thumbnailVideo) {
            const video = thumbnailVideo.querySelector('video');
            if (video) {
                video.srcObject = stream;
                video.play().catch(e => console.log('缩略图视频播放失败:', e));
                console.log('缩略图视频流设置完成:', userId);
            }
        } else {
            console.log('缩略图视频元素不存在:', userId);
        }

        // 如果当前主视频是这个用户，更新主视频
        if (this.currentMainVideo === userId) {
            const mainVideo = document.getElementById('mainVideo');
            if (mainVideo) {
                mainVideo.srcObject = stream;
                mainVideo.play().catch(e => console.log('主视频播放失败:', e));
                console.log('主视频流设置完成:', userId);
            }
        }

        // 如果当前没有主视频或主视频是本地，自动设置第一个远程用户为主视频
        if (this.currentMainVideo === 'local' && this.participants.size === 1) {
            this.selectMainVideo(userId);
        }

        // 检测音频活动
        this.detectAudioActivity(userId, stream);

        console.log('远程视频流设置完成:', userId);
    }
    
    /**
     * 创建远程视频元素
     */
    createRemoteVideoElement(userId) {
        const thumbnailArea = document.getElementById('thumbnailsGrid');
        const videoContainer = document.createElement('div');
        videoContainer.className = 'video-container thumbnail';
        videoContainer.id = `video-${userId}`;
        videoContainer.onclick = () => this.selectMainVideo(userId);

        const participantInfo = this.participants.get(userId);
        const userName = participantInfo ? participantInfo.name : `用户 ${userId}`;

        videoContainer.innerHTML = `
            <video autoplay playsinline></video>
            <div class="video-overlay">
                <span>${userName}</span>
                <span class="status"></span>
            </div>
            <div class="video-controls">
                <button class="control-btn" onclick="muteRemoteUser('${userId}')" title="静音">🔇</button>
                <button class="control-btn" onclick="selectMainVideo('${userId}')" title="设为主视频">📺</button>
            </div>
        `;

        thumbnailArea.appendChild(videoContainer);
        return videoContainer;
    }
    
    /**
     * 检测音频活动
     */
    detectAudioActivity(userId, stream) {
        const audioContext = new AudioContext();
        const analyser = audioContext.createAnalyser();
        const source = audioContext.createMediaStreamSource(stream);
        
        source.connect(analyser);
        analyser.fftSize = 256;
        
        const dataArray = new Uint8Array(analyser.frequencyBinCount);
        
        const checkAudioLevel = () => {
            analyser.getByteFrequencyData(dataArray);
            const average = dataArray.reduce((a, b) => a + b) / dataArray.length;
            
            const videoContainer = document.getElementById(`video-${userId}`);
            if (videoContainer) {
                if (average > 20) {
                    videoContainer.classList.add('speaking');
                } else {
                    videoContainer.classList.remove('speaking');
                }
            }
            
            requestAnimationFrame(checkAudioLevel);
        };
        
        checkAudioLevel();
    }
    
    /**
     * 发送信令消息
     */
    sendSignalingMessage(message) {
        if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
            console.log('发送信令消息:', message.type, message);
            this.signalingSocket.send(JSON.stringify(message));
        } else {
            console.error('信令服务器未连接，无法发送消息:', message.type);
            console.error('WebSocket状态:', this.signalingSocket ? this.signalingSocket.readyState : 'null');
        }
    }
    
    /**
     * 切换摄像头
     */
    async toggleCamera() {
        if (!this.localStream) return;
        
        const videoTrack = this.localStream.getVideoTracks()[0];
        if (videoTrack) {
            this.isVideoEnabled = !this.isVideoEnabled;
            videoTrack.enabled = this.isVideoEnabled;
            
            // 更新UI
            const cameraBtn = document.getElementById('cameraBtn');
            if (this.isVideoEnabled) {
                cameraBtn.classList.remove('off');
            } else {
                cameraBtn.classList.add('off');
            }
            
            // 通知其他用户
            this.broadcastMediaState();
        }
    }
    
    /**
     * 切换麦克风
     */
    async toggleMicrophone() {
        if (!this.localStream) return;
        
        const audioTrack = this.localStream.getAudioTracks()[0];
        if (audioTrack) {
            this.isAudioEnabled = !this.isAudioEnabled;
            audioTrack.enabled = this.isAudioEnabled;
            
            // 更新UI
            const micBtn = document.getElementById('micBtn');
            if (this.isAudioEnabled) {
                micBtn.classList.remove('off');
            } else {
                micBtn.classList.add('off');
            }
            
            // 通知其他用户
            this.broadcastMediaState();
        }
    }
    
    /**
     * 屏幕共享
     */
    async toggleScreenShare() {
        try {
            if (!this.isScreenSharing) {
                // 开始屏幕共享
                const screenStream = await navigator.mediaDevices.getDisplayMedia({
                    video: true,
                    audio: true
                });
                
                // 替换视频轨道
                const videoTrack = screenStream.getVideoTracks()[0];
                this.replaceVideoTrack(videoTrack);
                
                this.isScreenSharing = true;
                
                // 监听屏幕共享结束
                videoTrack.onended = () => {
                    this.stopScreenShare();
                };
                
            } else {
                // 停止屏幕共享
                this.stopScreenShare();
            }
        } catch (error) {
            console.error('屏幕共享失败:', error);
        }
    }
    
    /**
     * 停止屏幕共享
     */
    async stopScreenShare() {
        try {
            // 重新获取摄像头
            const cameraStream = await navigator.mediaDevices.getUserMedia({
                video: this.mediaConstraints.video,
                audio: false
            });
            
            const videoTrack = cameraStream.getVideoTracks()[0];
            this.replaceVideoTrack(videoTrack);
            
            this.isScreenSharing = false;
        } catch (error) {
            console.error('停止屏幕共享失败:', error);
        }
    }
    
    /**
     * 替换视频轨道
     */
    async replaceVideoTrack(newTrack) {
        const oldTrack = this.localStream.getVideoTracks()[0];
        if (oldTrack) {
            this.localStream.removeTrack(oldTrack);
            oldTrack.stop();
        }
        
        this.localStream.addTrack(newTrack);
        
        // 更新本地视频
        const localVideo = document.getElementById('localVideo');
        if (localVideo) {
            localVideo.srcObject = this.localStream;
        }
        
        // 更新所有对等连接
        for (const [userId, peerConnection] of this.peerConnections) {
            const sender = peerConnection.getSenders().find(s => 
                s.track && s.track.kind === 'video'
            );
            if (sender) {
                await sender.replaceTrack(newTrack);
            }
        }
    }
    
    /**
     * 广播媒体状态
     */
    broadcastMediaState() {
        this.sendSignalingMessage({
            type: 'media-state',
            data: {
                video: this.isVideoEnabled,
                audio: this.isAudioEnabled,
                screen: this.isScreenSharing
            }
        });
    }
    
    /**
     * 处理聊天消息
     */
    handleChatMessage(data) {
        console.log('收到聊天消息:', data);

        const chatMessages = document.getElementById('chatMessages');
        if (!chatMessages) {
            console.log('chatMessages元素不存在，跳过聊天消息显示');
            return;
        }

        const messageElement = document.createElement('div');
        messageElement.className = 'chat-message';
        messageElement.style.cssText = `
            margin-bottom: 10px;
            padding: 8px 12px;
            background: ${data.sender === this.currentUser?.name ? 'rgba(79, 172, 254, 0.3)' : 'rgba(255, 255, 255, 0.1)'};
            border-radius: 8px;
            ${data.sender === this.currentUser?.name ? 'margin-left: 20px;' : ''}
        `;

        messageElement.innerHTML = `
            <div style="font-weight: bold; font-size: 12px; margin-bottom: 2px;">${data.sender}</div>
            <div>${data.message}</div>
        `;

        chatMessages.appendChild(messageElement);
        chatMessages.scrollTop = chatMessages.scrollHeight;

        console.log('聊天消息显示完成');
    }

    /**
     * 发送聊天消息
     */
    sendChatMessage(message) {
        if (!this.isConnected || !this.currentUser) {
            console.log('未连接或用户信息不存在，无法发送聊天消息');
            return;
        }

        const chatData = {
            sender: this.currentUser.name,
            message: message,
            timestamp: Date.now()
        };

        console.log('发送聊天消息:', chatData);

        this.sendSignalingMessage({
            type: 'chat-message',
            data: chatData
        });
    }
    
    /**
     * 更新状态指示器
     */
    updateStatus(message, type = 'info') {
        console.log(`状态更新: ${message} (${type})`);

        const indicator = document.getElementById('statusIndicator');
        if (indicator) {
            indicator.textContent = message;
            indicator.className = `status-indicator ${type}`;
            indicator.classList.remove('hidden');

            if (type === 'success') {
                setTimeout(() => {
                    indicator.classList.add('hidden');
                }, 3000);
            }
        } else {
            // 如果状态指示器不存在，只在控制台输出
            console.log(`状态指示器不存在，消息: ${message}`);
        }
    }
    
    /**
     * 添加参与者
     */
    addParticipant(user) {
        console.log('添加参与者到UI:', user);

        const participantList = document.getElementById('participantList');
        if (!participantList) {
            console.log('participantList元素不存在，跳过添加参与者');
            return;
        }

        // 检查是否已经存在
        const existingParticipant = document.getElementById(`participant-${user.id}`);
        if (existingParticipant) {
            console.log('参与者已存在:', user.id);
            return;
        }

        const participantElement = document.createElement('div');
        participantElement.className = 'participant-item';
        participantElement.id = `participant-${user.id}`;
        participantElement.onclick = () => this.selectMainVideo(user.id);

        participantElement.innerHTML = `
            <div class="participant-avatar">${user.name.charAt(0).toUpperCase()}</div>
            <div class="participant-info">
                <div class="participant-name">${user.name}</div>
                <div class="participant-status">
                    <div class="status-indicator"></div>
                    <span>在线</span>
                </div>
            </div>
        `;

        participantList.appendChild(participantElement);

        // 添加缩略图视频
        this.addThumbnailVideo(user.id, user.name, false);

        this.updateParticipantCount();

        console.log('参与者添加完成:', user.id);
    }

    /**
     * 添加缩略图视频
     */
    addThumbnailVideo(userId, userName, isLocal = false) {
        console.log('添加缩略图视频:', userId, userName, isLocal);

        const thumbnailsGrid = document.getElementById('thumbnailsGrid');
        if (!thumbnailsGrid) {
            console.log('thumbnailsGrid元素不存在，跳过添加缩略图');
            return;
        }

        // 检查是否已经存在
        const existingThumbnail = document.getElementById(`thumbnail-${userId}`);
        if (existingThumbnail) {
            console.log('缩略图已存在:', userId);
            return;
        }

        const thumbnailElement = document.createElement('div');
        thumbnailElement.className = `thumbnail-video ${isLocal ? 'local selected' : ''}`;
        thumbnailElement.id = `thumbnail-${userId}`;
        thumbnailElement.onclick = () => this.selectMainVideo(userId);

        thumbnailElement.innerHTML = `
            <video autoplay ${isLocal ? 'muted' : ''} playsinline></video>
            <div class="thumbnail-overlay">${userName}${isLocal ? ' (您)' : ''}</div>
        `;

        thumbnailsGrid.appendChild(thumbnailElement);

        // 如果是本地视频，设置视频源
        if (isLocal && this.localStream) {
            const video = thumbnailElement.querySelector('video');
            if (video) {
                video.srcObject = this.localStream;
            }
        }

        console.log('缩略图添加完成:', userId);
    }
    
    /**
     * 移除参与者
     */
    removeParticipant(userId) {
        const participantElement = document.getElementById(`participant-${userId}`);
        if (participantElement) {
            participantElement.remove();
        }
        this.updateParticipantCount();
    }
    
    /**
     * 更新参与者列表
     */
    updateParticipantsList() {
        const participantList = document.getElementById('participantList');
        if (!participantList) {
            console.log('participantList元素不存在，跳过更新');
            return;
        }

        if (!this.currentUser) {
            console.log('currentUser不存在，跳过更新');
            return;
        }

        // 检查本地用户是否已存在
        const existingLocalParticipant = document.getElementById('participant-local');
        if (existingLocalParticipant) {
            console.log('本地用户已存在于参与者列表中，跳过添加');
            return;
        }

        // 只添加本地用户，不清空整个列表
        const localParticipantElement = document.createElement('div');
        localParticipantElement.className = 'participant-item main-speaker';
        localParticipantElement.id = 'participant-local';
        localParticipantElement.onclick = () => this.selectMainVideo('local');

        localParticipantElement.innerHTML = `
            <div class="participant-avatar">${this.currentUser.name.charAt(0).toUpperCase()}</div>
            <div class="participant-info">
                <div class="participant-name">${this.currentUser.name} (您)</div>
                <div class="participant-status">
                    <div class="status-indicator"></div>
                    <span>在线 • 主讲人</span>
                </div>
            </div>
        `;

        participantList.appendChild(localParticipantElement);

        // 添加本地用户的缩略图
        this.addThumbnailVideo('local', this.currentUser.name, true);

        // 设置本地用户为默认主视频
        if (this.selectMainVideo) {
            this.selectMainVideo('local');
        }
        this.updateParticipantCount();
    }
    
    /**
     * 更新参与者计数
     */
    updateParticipantCount() {
        const count = document.querySelectorAll('.participant-item').length;
        const participantCountElement = document.getElementById('participantCount');
        if (participantCountElement) {
            participantCountElement.textContent = count;
        } else {
            console.log('participantCount元素不存在');
        }
    }
    
    /**
     * 选择主视频
     */
    selectMainVideo(userId) {
        console.log('选择主视频:', userId);

        // 更新当前主视频
        this.currentMainVideo = userId;

        // 获取主视频元素
        const mainVideo = document.getElementById('mainVideo');
        const mainVideoUserName = document.getElementById('mainVideoUserName');

        if (!mainVideo || !mainVideoUserName) {
            console.log('主视频元素不存在，跳过主视频设置');
            return;
        }

        // 移除所有缩略图的选中状态
        document.querySelectorAll('.video-container').forEach(container => {
            container.classList.remove('main-selected');
        });

        if (userId === 'local') {
            // 显示本地视频
            if (this.localStream && mainVideo) {
                mainVideo.srcObject = this.localStream;
                mainVideo.muted = true; // 本地视频静音
            }
            if (mainVideoUserName && this.currentUser) {
                mainVideoUserName.textContent = `${this.currentUser.name} (您)`;
            }
            const localContainer = document.getElementById('localVideoContainer');
            if (localContainer) {
                localContainer.classList.add('main-selected');
            }
        } else {
            // 显示远程视频
            const participantInfo = this.participants.get(userId);
            if (participantInfo && participantInfo.stream && mainVideo) {
                mainVideo.srcObject = participantInfo.stream;
                mainVideo.muted = this.mainVideoMuted;
                mainVideoUserName.textContent = participantInfo.name;
                const remoteContainer = document.getElementById(`video-${userId}`);
                if (remoteContainer) {
                    remoteContainer.classList.add('main-selected');
                }
            }
        }

        // 更新参与者列表中的主讲人标识
        this.updateParticipantMainSpeaker(userId);

        console.log('主视频已切换到:', userId);
    }

    /**
     * 切换主视频静音状态
     */
    toggleMainVideoMute() {
        this.mainVideoMuted = !this.mainVideoMuted;
        const mainVideo = document.getElementById('mainVideo');
        const muteBtn = document.getElementById('mainVideoMuteBtn');

        if (this.currentMainVideo !== 'local') {
            mainVideo.muted = this.mainVideoMuted;
        }

        muteBtn.textContent = this.mainVideoMuted ? '🔇' : '🔊';
        muteBtn.title = this.mainVideoMuted ? '取消静音' : '静音';
    }

    /**
     * 主视频全屏
     */
    toggleMainVideoFullscreen() {
        const mainVideoContainer = document.getElementById('mainVideoContainer');

        if (!document.fullscreenElement) {
            mainVideoContainer.requestFullscreen().catch(err => {
                console.error('无法进入全屏模式:', err);
            });
        } else {
            document.exitFullscreen();
        }
    }

    /**
     * 更新参与者主讲人标识
     */
    updateParticipantMainSpeaker(mainUserId) {
        // 移除所有主讲人标识
        document.querySelectorAll('.participant-item').forEach(item => {
            item.classList.remove('main-speaker');
        });

        // 添加新的主讲人标识
        const mainParticipant = document.getElementById(`participant-${mainUserId}`);
        if (mainParticipant) {
            mainParticipant.classList.add('main-speaker');
        }
    }

    /**
     * 添加参与者信息
     */
    addParticipantInfo(userId, userName) {
        this.participants.set(userId, {
            id: userId,
            name: userName,
            stream: null,
            isAudioEnabled: true,
            isVideoEnabled: true,
            isSpeaking: false
        });

        console.log('添加参与者信息:', userId, userName);
    }

    /**
     * 移除参与者信息
     */
    removeParticipantInfo(userId) {
        this.participants.delete(userId);

        // 如果移除的是当前主视频，切换到本地视频
        if (this.currentMainVideo === userId) {
            this.selectMainVideo('local');
        }

        console.log('移除参与者信息:', userId);
    }

    /**
     * 获取所有参与者信息
     */
    getAllParticipants() {
        const participants = [];

        // 添加本地用户
        if (this.currentUser) {
            participants.push({
                id: 'local',
                name: this.currentUser.name,
                isLocal: true,
                isMainSpeaker: this.currentMainVideo === 'local',
                isAudioEnabled: this.isAudioEnabled,
                isVideoEnabled: this.isVideoEnabled,
                isSpeaking: false
            });
        }

        // 添加远程用户
        for (const [userId, info] of this.participants) {
            participants.push({
                id: userId,
                name: info.name,
                isLocal: false,
                isMainSpeaker: this.currentMainVideo === userId,
                isAudioEnabled: info.isAudioEnabled,
                isVideoEnabled: info.isVideoEnabled,
                isSpeaking: info.isSpeaking
            });
        }

        return participants;
    }

    /**
     * 离开会议
     */
    leaveMeeting() {
        // 设置离开标志，防止自动重连
        this.isLeavingMeeting = true;

        // 关闭所有对等连接
        for (const [userId, peerConnection] of this.peerConnections) {
            peerConnection.close();
        }
        this.peerConnections.clear();

        // 清理参与者信息
        this.participants.clear();

        // 停止本地流
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
        }

        // 发送离开消息
        this.sendSignalingMessage({
            type: 'leave-meeting',
            data: {
                meetingId: this.meetingId,
                user: this.currentUser
            }
        });

        // 等待消息发送后关闭信令连接
        setTimeout(() => {
            if (this.signalingSocket) {
                this.signalingSocket.close();
            }
        }, 100);

        // 重置状态
        this.isConnected = false;
        this.currentUser = null;
        this.meetingId = null;
        this.currentMainVideo = 'local';
        this.mainVideoMuted = false;

        // 清空主视频
        const mainVideo = document.getElementById('mainVideo');
        const mainVideoUserName = document.getElementById('mainVideoUserName');
        mainVideo.srcObject = null;
        mainVideoUserName.textContent = '选择一个参与者作为主视频';

        // 显示登录模态框
        document.getElementById('loginModal').classList.remove('hidden');
    }

    /**
     * 申请主讲人权限
     */
    requestPresenter() {
        console.log('申请主讲人权限');
        this.sendSignalingMessage({
            type: 'request-presenter',
            data: {
                meetingId: this.meetingId,
                userId: this.currentUser.id
            }
        });
    }

    /**
     * 释放主讲人权限
     */
    releasePresenter() {
        console.log('释放主讲人权限');
        this.sendSignalingMessage({
            type: 'release-presenter',
            data: {
                meetingId: this.meetingId,
                userId: this.currentUser.id
            }
        });
    }

    /**
     * 处理主讲人消息
     */
    handlePresenterMessage(type, data) {
        console.log('处理主讲人消息:', type, data);

        // 调用全局处理函数
        if (window.handlePresenterStatusChange) {
            window.handlePresenterStatusChange(type, data);
        }
    }
}

// 全局WebRTC管理器实例
window.webrtcManager = new WebRTCManager();
