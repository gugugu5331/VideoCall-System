/**
 * AI检测模块
 * 负责实时检测音视频中的伪造内容
 */

class DetectionManager {
    constructor() {
        this.isEnabled = true;
        this.detectionInterval = null;
        this.lastDetectionTime = 0;
        this.detectionThreshold = 0.7; // 检测阈值
        this.detectionHistory = [];
        
        // 检测配置
        this.config = {
            videoDetectionInterval: 5000, // 5秒检测一次视频
            audioDetectionInterval: 3000, // 3秒检测一次音频
            maxHistorySize: 100
        };
        
        // 模拟AI模型状态
        this.models = {
            faceSwap: { loaded: true, accuracy: 0.95 },
            voiceSynthesis: { loaded: true, accuracy: 0.92 },
            deepfake: { loaded: true, accuracy: 0.88 }
        };
    }
    
    /**
     * 初始化检测系统
     */
    initialize() {
        console.log('初始化AI检测系统...');
        
        // 启动实时检测
        this.startRealTimeDetection();
        
        // 更新UI状态
        this.updateDetectionUI();
        
        console.log('AI检测系统初始化完成');
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
        
        console.log('实时检测已启动');
    }
    
    /**
     * 停止实时检测
     */
    stopRealTimeDetection() {
        if (this.detectionInterval) {
            clearInterval(this.detectionInterval);
            this.detectionInterval = null;
        }
        
        console.log('实时检测已停止');
    }
    
    /**
     * 检测视频内容
     */
    async detectVideoContent() {
        try {
            // 获取本地视频帧
            const videoFrame = await this.captureVideoFrame();
            if (!videoFrame) return;
            
            // 模拟AI检测过程
            const result = await this.simulateVideoDetection(videoFrame);
            
            // 处理检测结果
            this.handleDetectionResult('video', result);
            
        } catch (error) {
            console.error('视频检测失败:', error);
        }
    }
    
    /**
     * 检测音频内容
     */
    async detectAudioContent() {
        try {
            // 获取音频数据
            const audioData = await this.captureAudioData();
            if (!audioData) return;
            
            // 模拟AI检测过程
            const result = await this.simulateAudioDetection(audioData);
            
            // 处理检测结果
            this.handleDetectionResult('audio', result);
            
        } catch (error) {
            console.error('音频检测失败:', error);
        }
    }
    
    /**
     * 捕获视频帧
     */
    async captureVideoFrame() {
        const localVideo = document.getElementById('localVideo');
        if (!localVideo || !localVideo.srcObject) return null;
        
        try {
            // 创建canvas来捕获视频帧
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            
            canvas.width = localVideo.videoWidth || 640;
            canvas.height = localVideo.videoHeight || 480;
            
            ctx.drawImage(localVideo, 0, 0, canvas.width, canvas.height);
            
            // 转换为ImageData
            const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
            
            return {
                width: canvas.width,
                height: canvas.height,
                data: imageData.data,
                timestamp: Date.now()
            };
            
        } catch (error) {
            console.error('捕获视频帧失败:', error);
            return null;
        }
    }
    
    /**
     * 捕获音频数据
     */
    async captureAudioData() {
        // 这里应该从WebRTC音频流中获取音频数据
        // 为了演示，我们返回模拟数据
        return {
            sampleRate: 48000,
            channels: 2,
            duration: 1.0,
            timestamp: Date.now(),
            data: new Float32Array(48000) // 模拟1秒的音频数据
        };
    }
    
    /**
     * 模拟视频检测
     */
    async simulateVideoDetection(videoFrame) {
        // 模拟网络请求延迟
        await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 200));
        
        // 模拟检测结果
        const isFake = Math.random() < 0.1; // 10%概率检测到伪造
        const confidence = isFake ? 
            0.7 + Math.random() * 0.3 : // 伪造时置信度0.7-1.0
            Math.random() * 0.6; // 正常时置信度0.0-0.6
        
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
        // 模拟网络请求延迟
        await new Promise(resolve => setTimeout(resolve, 80 + Math.random() * 150));
        
        // 模拟检测结果
        const isFake = Math.random() < 0.05; // 5%概率检测到伪造
        const confidence = isFake ? 
            0.75 + Math.random() * 0.25 : // 伪造时置信度0.75-1.0
            Math.random() * 0.5; // 正常时置信度0.0-0.5
        
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
        // 添加到历史记录
        this.addToHistory(mediaType, result);
        
        // 更新UI
        this.updateDetectionUI();
        
        // 检查是否需要告警
        if (result.isFake && result.confidence > this.detectionThreshold) {
            this.triggerAlert(mediaType, result);
        }
        
        // 发送检测结果到服务器
        this.reportDetectionResult(mediaType, result);
        
        console.log(`${mediaType}检测结果:`, result);
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
        
        // 限制历史记录大小
        if (this.detectionHistory.length > this.config.maxHistorySize) {
            this.detectionHistory = this.detectionHistory.slice(0, this.config.maxHistorySize);
        }
    }
    
    /**
     * 触发告警
     */
    triggerAlert(mediaType, result) {
        console.log('触发告警:', mediaType, result);

        const alertElement = document.getElementById('detectionAlert');
        const mediaTypeText = mediaType === 'video' ? '视频' : '音频';
        const detectionTypeText = result.type === 'face_swap' ? '人脸交换' : '语音合成';

        if (alertElement) {
            alertElement.innerHTML = `
                <strong>⚠️ 检测到可疑内容</strong>
                <p>在${mediaTypeText}中检测到可能的${detectionTypeText}，置信度: ${(result.confidence * 100).toFixed(1)}%</p>
            `;

            alertElement.classList.remove('hidden');

            // 5秒后自动隐藏
            setTimeout(() => {
                alertElement.classList.add('hidden');
            }, 5000);
        } else {
            // 如果告警元素不存在，在控制台显示告警
            console.warn(`⚠️ 检测到可疑${mediaTypeText}内容: ${detectionTypeText}，置信度: ${(result.confidence * 100).toFixed(1)}%`);
        }
        
        // 播放告警声音（如果需要）
        this.playAlertSound();
        
        console.warn(`检测告警: ${mediaTypeText} ${detectionTypeText}, 置信度: ${result.confidence}`);
    }
    
    /**
     * 播放告警声音
     */
    playAlertSound() {
        try {
            // 创建音频上下文
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            
            // 创建振荡器
            const oscillator = audioContext.createOscillator();
            const gainNode = audioContext.createGain();
            
            oscillator.connect(gainNode);
            gainNode.connect(audioContext.destination);
            
            // 设置音频参数
            oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
            oscillator.frequency.setValueAtTime(600, audioContext.currentTime + 0.1);
            oscillator.frequency.setValueAtTime(800, audioContext.currentTime + 0.2);
            
            gainNode.gain.setValueAtTime(0.1, audioContext.currentTime);
            gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3);
            
            // 播放声音
            oscillator.start(audioContext.currentTime);
            oscillator.stop(audioContext.currentTime + 0.3);
            
        } catch (error) {
            console.error('播放告警声音失败:', error);
        }
    }
    
    /**
     * 上报检测结果
     */
    async reportDetectionResult(mediaType, result) {
        try {
            const reportData = {
                media_type: mediaType,
                detection_type: result.type,
                is_fake: result.isFake,
                confidence: result.confidence,
                timestamp: result.timestamp,
                meeting_id: window.webrtcManager?.meetingId,
                user_id: window.webrtcManager?.currentUser?.id,
                details: result.details
            };
            
            // 发送到后端API
            const response = await fetch('http://localhost:8080/api/v1/detection/report', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(reportData)
            });
            
            if (!response.ok) {
                console.warn('上报检测结果失败:', response.status);
            }
            
        } catch (error) {
            console.error('上报检测结果错误:', error);
        }
    }
    
    /**
     * 更新检测UI
     */
    updateDetectionUI() {
        const detectionResults = document.getElementById('detectionResults');
        if (!detectionResults) return;
        
        // 获取最近的检测结果
        const recentResults = this.detectionHistory.slice(0, 5);
        
        let html = '';
        
        if (recentResults.length === 0) {
            html = `
                <div style="background: #3d3d3d; padding: 10px; border-radius: 5px; margin-bottom: 10px;">
                    <div style="font-size: 12px; color: #888;">检测状态</div>
                    <div style="color: #28a745;">✅ 系统正常运行</div>
                    <div style="font-size: 12px; color: #888;">等待检测数据...</div>
                </div>
            `;
        } else {
            recentResults.forEach(result => {
                const timeStr = new Date(result.timestamp).toLocaleTimeString();
                const mediaTypeText = result.mediaType === 'video' ? '视频' : '音频';
                const statusIcon = result.isFake ? '⚠️' : '✅';
                const statusText = result.isFake ? '检测到异常' : '未发现异常';
                const statusColor = result.isFake ? '#dc3545' : '#28a745';
                
                html += `
                    <div style="background: #3d3d3d; padding: 10px; border-radius: 5px; margin-bottom: 10px;">
                        <div style="font-size: 12px; color: #888;">${timeStr} - ${mediaTypeText}</div>
                        <div style="color: ${statusColor};">${statusIcon} ${statusText}</div>
                        <div style="font-size: 12px; color: #888;">置信度: ${(result.confidence * 100).toFixed(1)}%</div>
                    </div>
                `;
            });
        }
        
        detectionResults.innerHTML = html;
    }
    
    /**
     * 切换检测状态
     */
    toggleDetection(enabled) {
        this.isEnabled = enabled;
        
        if (enabled) {
            this.startRealTimeDetection();
            console.log('AI检测已启用');
        } else {
            this.stopRealTimeDetection();
            console.log('AI检测已禁用');
        }
        
        this.updateDetectionUI();
    }
    
    /**
     * 获取检测统计
     */
    getDetectionStats() {
        const totalDetections = this.detectionHistory.length;
        const fakeDetections = this.detectionHistory.filter(r => r.isFake).length;
        const videoDetections = this.detectionHistory.filter(r => r.mediaType === 'video').length;
        const audioDetections = this.detectionHistory.filter(r => r.mediaType === 'audio').length;
        
        return {
            total: totalDetections,
            fake: fakeDetections,
            real: totalDetections - fakeDetections,
            video: videoDetections,
            audio: audioDetections,
            accuracy: totalDetections > 0 ? ((totalDetections - fakeDetections) / totalDetections * 100).toFixed(1) : 0
        };
    }
    
    /**
     * 手动检测文件
     */
    async detectFile(file) {
        try {
            const formData = new FormData();
            formData.append('file', file);
            
            const response = await fetch('http://localhost:8080/api/v1/detection/analyze', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                throw new Error(`检测请求失败: ${response.status}`);
            }
            
            const result = await response.json();
            
            // 显示检测结果
            this.showFileDetectionResult(result);
            
            return result;
            
        } catch (error) {
            console.error('文件检测失败:', error);
            throw error;
        }
    }
    
    /**
     * 显示文件检测结果
     */
    showFileDetectionResult(result) {
        const alertElement = document.getElementById('detectionAlert');
        
        if (result.is_fake) {
            alertElement.innerHTML = `
                <strong>⚠️ 文件检测结果</strong>
                <p>检测到可疑内容，置信度: ${(result.confidence * 100).toFixed(1)}%</p>
                <p>类型: ${result.detection_type}</p>
            `;
        } else {
            alertElement.innerHTML = `
                <strong>✅ 文件检测结果</strong>
                <p>未发现异常内容，置信度: ${(result.confidence * 100).toFixed(1)}%</p>
            `;
            alertElement.style.background = '#28a745';
        }
        
        alertElement.classList.remove('hidden');
        
        setTimeout(() => {
            alertElement.classList.add('hidden');
            alertElement.style.background = '#dc3545';
        }, 8000);
    }
}

// 全局检测管理器实例
window.detectionManager = new DetectionManager();
