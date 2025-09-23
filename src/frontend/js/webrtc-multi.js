/**
 * 多用户WebRTC管理器
 * 支持在同一页面中创建多个WebRTC实例
 */

class WebRTCManager {
    constructor(userIndex) {
        this.userIndex = userIndex;
        this.localStream = null;
        this.peerConnections = new Map();
        this.signalingSocket = null;
        this.isConnected = false;
        this.currentUser = null;
        this.meetingId = null;
        this.pendingIceCandidates = new Map();
        
        // WebRTC配置
        this.rtcConfig = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' }
            ]
        };
        
        // 媒体约束
        this.mediaConstraints = {
            video: {
                width: { ideal: 640 },
                height: { ideal: 480 },
                frameRate: { ideal: 15 }
            },
            audio: {
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true
            }
        };
        
        this.isVideoEnabled = true;
        this.isAudioEnabled = true;

        // 主视频管理
        this.currentMainVideo = 'local';
        this.participants = new Map();
    }
    
    /**
     * 初始化WebRTC
     */
    async initialize() {
        try {
            await this.getUserMedia();
            this.connectSignaling();
            this.addInfoMessage('WebRTC初始化成功');
            return true;
        } catch (error) {
            this.addInfoMessage(`WebRTC初始化失败: ${error.message}`, 'error');
            throw error;
        }
    }
    
    /**
     * 获取用户媒体
     */
    async getUserMedia() {
        try {
            this.localStream = await navigator.mediaDevices.getUserMedia(this.mediaConstraints);
            
            const localVideo = document.getElementById(`localVideo-${this.userIndex}`);
            if (localVideo) {
                localVideo.srcObject = this.localStream;
            }
            
            this.addInfoMessage('本地媒体流获取成功');
            return this.localStream;
        } catch (error) {
            this.addInfoMessage(`媒体流获取失败: ${error.message}`, 'error');
            throw new Error('无法访问摄像头或麦克风');
        }
    }
    
    /**
     * 连接信令服务器
     */
    connectSignaling() {
        const wsUrl = `ws://localhost:8080/signaling`;
        this.signalingSocket = new WebSocket(wsUrl);
        
        this.signalingSocket.onopen = () => {
            this.isConnected = true;
            this.addInfoMessage('信令服务器连接成功');
        };
        
        this.signalingSocket.onmessage = (event) => {
            this.handleSignalingMessage(JSON.parse(event.data));
        };
        
        this.signalingSocket.onclose = () => {
            this.isConnected = false;
            this.addInfoMessage('信令服务器连接断开', 'error');
        };
        
        this.signalingSocket.onerror = (error) => {
            this.addInfoMessage('信令服务器错误', 'error');
        };
    }
    
    /**
     * 处理信令消息
     */
    async handleSignalingMessage(message) {
        console.log(`用户${this.userIndex}收到信令消息:`, message);
        
        switch (message.type) {
            case 'welcome':
                this.addInfoMessage('收到服务器欢迎消息');
                break;
                
            case 'user-joined':
                await this.handleUserJoined(message.data);
                break;
                
            case 'user-left':
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
                
            default:
                console.log(`用户${this.userIndex}未知信令消息:`, message.type);
        }
    }
    
    /**
     * 加入会议
     */
    async joinMeeting(username, meetingId) {
        this.currentUser = { 
            id: `user_${this.userIndex}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`, 
            name: username 
        };
        this.meetingId = meetingId;
        
        this.sendSignalingMessage({
            type: 'join-meeting',
            data: {
                meetingId: meetingId,
                user: this.currentUser
            }
        });
        
        this.addInfoMessage(`加入会议: ${meetingId}`);
    }
    
    /**
     * 处理用户加入
     */
    async handleUserJoined(userData) {
        if (userData.id === this.currentUser.id) {
            return; // 忽略自己
        }

        this.addInfoMessage(`用户加入: ${userData.name}`);

        // 添加到参与者列表
        this.participants.set(userData.id, {
            id: userData.id,
            name: userData.name,
            stream: null
        });

        // 添加到主视频选择列表
        if (typeof addUserToMainVideoSelect === 'function') {
            addUserToMainVideoSelect(this.userIndex, userData.id, userData.name);
        }

        if (this.peerConnections.has(userData.id)) {
            return; // 连接已存在
        }

        try {
            const peerConnection = await this.createPeerConnection(userData.id);

            if (this.localStream) {
                this.localStream.getTracks().forEach(track => {
                    peerConnection.addTrack(track, this.localStream);
                });
            }

            const offer = await peerConnection.createOffer({
                offerToReceiveAudio: true,
                offerToReceiveVideo: true
            });
            await peerConnection.setLocalDescription(offer);

            this.sendSignalingMessage({
                type: 'offer',
                to: userData.id,
                data: offer
            });

            this.addInfoMessage(`发送offer给: ${userData.name}`);

        } catch (error) {
            this.addInfoMessage(`创建连接失败: ${error.message}`, 'error');
        }
    }
    
    /**
     * 处理用户离开
     */
    handleUserLeft(userData) {
        this.addInfoMessage(`用户离开: ${userData.name}`);

        // 从参与者列表移除
        this.participants.delete(userData.id);

        // 从主视频选择列表移除
        if (typeof removeUserFromMainVideoSelect === 'function') {
            removeUserFromMainVideoSelect(this.userIndex, userData.id);
        }

        if (this.peerConnections.has(userData.id)) {
            this.peerConnections.get(userData.id).close();
            this.peerConnections.delete(userData.id);
        }

        // 移除远程视频
        const remoteVideo = document.getElementById(`remote-${this.userIndex}-${userData.id}`);
        if (remoteVideo) {
            remoteVideo.remove();
        }
    }
    
    /**
     * 处理offer
     */
    async handleOffer(message) {
        this.addInfoMessage(`收到offer from: ${message.from}`);
        
        try {
            let peerConnection = this.peerConnections.get(message.from);
            if (!peerConnection) {
                peerConnection = await this.createPeerConnection(message.from);
            }
            
            if (this.localStream) {
                this.localStream.getTracks().forEach(track => {
                    peerConnection.addTrack(track, this.localStream);
                });
            }
            
            await peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
            
            const answer = await peerConnection.createAnswer();
            await peerConnection.setLocalDescription(answer);
            
            this.sendSignalingMessage({
                type: 'answer',
                to: message.from,
                data: answer
            });
            
            this.addInfoMessage(`发送answer给: ${message.from}`);
            
        } catch (error) {
            this.addInfoMessage(`处理offer失败: ${error.message}`, 'error');
        }
    }
    
    /**
     * 处理answer
     */
    async handleAnswer(message) {
        this.addInfoMessage(`收到answer from: ${message.from}`);
        
        try {
            const peerConnection = this.peerConnections.get(message.from);
            if (peerConnection) {
                await peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
                this.addInfoMessage(`设置远程描述成功: ${message.from}`);
            }
        } catch (error) {
            this.addInfoMessage(`处理answer失败: ${error.message}`, 'error');
        }
    }
    
    /**
     * 处理ICE候选
     */
    async handleIceCandidate(message) {
        try {
            const peerConnection = this.peerConnections.get(message.from);
            if (peerConnection && peerConnection.remoteDescription) {
                await peerConnection.addIceCandidate(new RTCIceCandidate(message.data));
            } else {
                // 暂存ICE候选
                if (!this.pendingIceCandidates.has(message.from)) {
                    this.pendingIceCandidates.set(message.from, []);
                }
                this.pendingIceCandidates.get(message.from).push(message.data);
            }
        } catch (error) {
            console.error(`用户${this.userIndex}处理ICE候选失败:`, error);
        }
    }
    
    /**
     * 创建对等连接
     */
    async createPeerConnection(userId) {
        const peerConnection = new RTCPeerConnection(this.rtcConfig);
        
        peerConnection.onicecandidate = (event) => {
            if (event.candidate) {
                this.sendSignalingMessage({
                    type: 'ice-candidate',
                    to: userId,
                    data: event.candidate
                });
            }
        };
        
        peerConnection.ontrack = (event) => {
            this.handleRemoteStream(userId, event.streams[0]);
        };
        
        peerConnection.onconnectionstatechange = () => {
            const state = peerConnection.connectionState;
            this.addInfoMessage(`连接状态 ${userId}: ${state}`);
            
            if (state === 'connected') {
                this.processPendingIceCandidates(userId);
            }
        };
        
        this.peerConnections.set(userId, peerConnection);
        return peerConnection;
    }
    
    /**
     * 处理远程流
     */
    handleRemoteStream(userId, stream) {
        this.addInfoMessage(`收到远程流: ${userId}`);

        // 更新参与者流信息
        if (this.participants.has(userId)) {
            this.participants.get(userId).stream = stream;
        }

        const remoteVideosContainer = document.getElementById(`remoteVideos-${this.userIndex}`);

        let remoteVideo = document.getElementById(`remote-${this.userIndex}-${userId}`);
        if (!remoteVideo) {
            remoteVideo = document.createElement('video');
            remoteVideo.id = `remote-${this.userIndex}-${userId}`;
            remoteVideo.className = 'remote-video';
            remoteVideo.autoplay = true;
            remoteVideo.playsinline = true;
            remoteVideo.title = `远程用户 ${userId}`;
            remoteVideosContainer.appendChild(remoteVideo);
        }

        remoteVideo.srcObject = stream;
        remoteVideo.play().catch(e => console.log('远程视频播放失败:', e));

        // 如果当前主视频是这个用户，更新主视频显示
        if (this.currentMainVideo === userId) {
            this.updateMainVideoDisplay();
        }
    }

    /**
     * 选择主视频
     */
    selectMainVideo(userId) {
        this.currentMainVideo = userId;
        this.updateMainVideoDisplay();
        this.addInfoMessage(`切换主视频到: ${userId}`);
    }

    /**
     * 更新主视频显示
     */
    updateMainVideoDisplay() {
        const localVideo = document.getElementById(`localVideo-${this.userIndex}`);

        if (this.currentMainVideo === 'local') {
            // 显示本地视频作为主视频
            if (localVideo && this.localStream) {
                localVideo.style.border = '3px solid #ffc107';
                localVideo.style.transform = 'scale(1.05)';
            }
            // 重置其他视频样式
            this.resetRemoteVideoStyles();
        } else {
            // 显示远程视频作为主视频
            const remoteVideo = document.getElementById(`remote-${this.userIndex}-${this.currentMainVideo}`);
            if (remoteVideo) {
                remoteVideo.style.border = '3px solid #ffc107';
                remoteVideo.style.transform = 'scale(1.05)';
            }
            // 重置本地视频样式
            if (localVideo) {
                localVideo.style.border = '';
                localVideo.style.transform = '';
            }
            // 重置其他远程视频样式
            this.resetRemoteVideoStyles(this.currentMainVideo);
        }
    }

    /**
     * 重置远程视频样式
     */
    resetRemoteVideoStyles(excludeUserId = null) {
        for (const userId of this.participants.keys()) {
            if (userId !== excludeUserId) {
                const remoteVideo = document.getElementById(`remote-${this.userIndex}-${userId}`);
                if (remoteVideo) {
                    remoteVideo.style.border = '';
                    remoteVideo.style.transform = '';
                }
            }
        }
    }
    
    /**
     * 处理暂存的ICE候选
     */
    async processPendingIceCandidates(userId) {
        if (this.pendingIceCandidates.has(userId)) {
            const candidates = this.pendingIceCandidates.get(userId);
            const peerConnection = this.peerConnections.get(userId);
            
            if (peerConnection && peerConnection.remoteDescription) {
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
     * 发送信令消息
     */
    sendSignalingMessage(message) {
        if (this.signalingSocket && this.signalingSocket.readyState === WebSocket.OPEN) {
            this.signalingSocket.send(JSON.stringify(message));
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
        }
    }
    
    /**
     * 离开会议
     */
    leaveMeeting() {
        // 关闭所有连接
        for (const [userId, peerConnection] of this.peerConnections) {
            peerConnection.close();
        }
        this.peerConnections.clear();
        
        // 停止本地流
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
        }
        
        // 关闭信令连接
        if (this.signalingSocket) {
            this.sendSignalingMessage({
                type: 'leave-meeting',
                data: {
                    meetingId: this.meetingId,
                    user: this.currentUser
                }
            });
            this.signalingSocket.close();
        }
        
        this.isConnected = false;
        this.addInfoMessage('已离开会议');
    }
    
    /**
     * 添加信息消息
     */
    addInfoMessage(message, type = 'info') {
        if (typeof addInfoMessage === 'function') {
            addInfoMessage(this.userIndex, message, type);
        }
    }
}
