/**
 * 多用户AI检测管理器
 */

class DetectionManager {
    constructor(userIndex) {
        this.userIndex = userIndex;
        this.isEnabled = true;
        this.detectionInterval = null;
        this.detectionHistory = [];
        
        this.config = {
            videoDetectionInterval: 5000,
            audioDetectionInterval: 3000,
            maxHistorySize: 50
        };
    }
    
    /**
     * 初始化检测系统
     */
    initialize() {
        this.startRealTimeDetection();
        this.addInfoMessage('AI检测系统已启动');
    }
    
    /**
     * 启动实时检测
     */
    startRealTimeDetection() {
        if (!this.isEnabled) return;
        
        // 视频检测
        this.detectionInterval = setInterval(() => {
            this.detectVideoContent();
        }, this.config.videoDetectionInterval);
        
        // 音频检测
        setTimeout(() => {
            setInterval(() => {
                this.detectAudioContent();
            }, this.config.audioDetectionInterval);
        }, 1000);
    }
    
    /**
     * 停止实时检测
     */
    stopRealTimeDetection() {
        if (this.detectionInterval) {
            clearInterval(this.detectionInterval);
            this.detectionInterval = null;
        }
    }
    
    /**
     * 检测视频内容
     */
    async detectVideoContent() {
        try {
            const videoFrame = await this.captureVideoFrame();
            if (!videoFrame) return;
            
            const result = await this.simulateVideoDetection(videoFrame);
            this.handleDetectionResult('video', result);
            
        } catch (error) {
            console.error(`用户${this.userIndex}视频检测失败:`, error);
        }
    }
    
    /**
     * 检测音频内容
     */
    async detectAudioContent() {
        try {
            const audioData = await this.captureAudioData();
            if (!audioData) return;
            
            const result = await this.simulateAudioDetection(audioData);
            this.handleDetectionResult('audio', result);
            
        } catch (error) {
            console.error(`用户${this.userIndex}音频检测失败:`, error);
        }
    }
    
    /**
     * 捕获视频帧
     */
    async captureVideoFrame() {
        const localVideo = document.getElementById(`localVideo-${this.userIndex}`);
        if (!localVideo || !localVideo.srcObject) return null;
        
        try {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            
            canvas.width = localVideo.videoWidth || 640;
            canvas.height = localVideo.videoHeight || 480;
            
            ctx.drawImage(localVideo, 0, 0, canvas.width, canvas.height);
            const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
            
            return {
                width: canvas.width,
                height: canvas.height,
                data: imageData.data,
                timestamp: Date.now()
            };
            
        } catch (error) {
            return null;
        }
    }
    
    /**
     * 捕获音频数据
     */
    async captureAudioData() {
        return {
            sampleRate: 48000,
            channels: 2,
            duration: 1.0,
            timestamp: Date.now(),
            data: new Float32Array(48000)
        };
    }
    
    /**
     * 模拟视频检测
     */
    async simulateVideoDetection(videoFrame) {
        await new Promise(resolve => setTimeout(resolve, 50 + Math.random() * 150));
        
        const isFake = Math.random() < 0.08; // 8%概率检测到伪造
        const confidence = isFake ? 
            0.7 + Math.random() * 0.3 : 
            Math.random() * 0.6;
        
        return {
            type: 'face_swap',
            isFake: isFake,
            confidence: confidence,
            details: {
                faces_detected: Math.floor(Math.random() * 3) + 1,
                processing_time: Math.floor(Math.random() * 200) + 50,
                model_version: 'v2.1.0'
            },
            timestamp: Date.now()
        };
    }
    
    /**
     * 模拟音频检测
     */
    async simulateAudioDetection(audioData) {
        await new Promise(resolve => setTimeout(resolve, 30 + Math.random() * 120));
        
        const isFake = Math.random() < 0.05; // 5%概率检测到伪造
        const confidence = isFake ? 
            0.75 + Math.random() * 0.25 : 
            Math.random() * 0.5;
        
        return {
            type: 'voice_synthesis',
            isFake: isFake,
            confidence: confidence,
            details: {
                sample_rate: audioData.sampleRate,
                duration: audioData.duration,
                processing_time: Math.floor(Math.random() * 150) + 30,
                model_version: 'v1.8.2'
            },
            timestamp: Date.now()
        };
    }
    
    /**
     * 处理检测结果
     */
    handleDetectionResult(mediaType, result) {
        this.addToHistory(mediaType, result);
        
        if (result.isFake && result.confidence > 0.7) {
            this.addInfoMessage(`⚠️ 检测到可疑${mediaType === 'video' ? '视频' : '音频'}内容 (${(result.confidence * 100).toFixed(1)}%)`, 'error');
        }
        
        this.reportDetectionResult(mediaType, result);
    }
    
    /**
     * 添加到历史记录
     */
    addToHistory(mediaType, result) {
        const historyItem = {
            mediaType: mediaType,
            ...result
        };
        
        this.detectionHistory.unshift(historyItem);
        
        if (this.detectionHistory.length > this.config.maxHistorySize) {
            this.detectionHistory = this.detectionHistory.slice(0, this.config.maxHistorySize);
        }
    }
    
    /**
     * 上报检测结果
     */
    async reportDetectionResult(mediaType, result) {
        try {
            const user = users[this.userIndex];
            const reportData = {
                media_type: mediaType,
                detection_type: result.type,
                is_fake: result.isFake,
                confidence: result.confidence,
                timestamp: result.timestamp,
                meeting_id: user.webrtcManager?.meetingId,
                user_id: user.webrtcManager?.currentUser?.id,
                details: result.details
            };
            
            const response = await fetch('http://localhost:8080/api/v1/detection/report', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(reportData)
            });
            
            if (!response.ok) {
                console.warn(`用户${this.userIndex}上报检测结果失败:`, response.status);
            }
            
        } catch (error) {
            console.error(`用户${this.userIndex}上报检测结果错误:`, error);
        }
    }
    
    /**
     * 切换检测状态
     */
    toggleDetection(enabled) {
        this.isEnabled = enabled;
        
        if (enabled) {
            this.startRealTimeDetection();
            this.addInfoMessage('AI检测已启用');
        } else {
            this.stopRealTimeDetection();
            this.addInfoMessage('AI检测已禁用');
        }
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
