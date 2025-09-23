// 系统配置
const CONFIG = {
    // API配置
    API_BASE_URL: 'http://localhost:8000',
    AI_SERVICE_URL: 'http://localhost:5000',
    WS_URL: 'ws://localhost:8000',
    
    // 存储键名
    STORAGE_KEYS: {
        AUTH_TOKEN: 'auth_token',
        USER_INFO: 'user_info',
        CALL_SETTINGS: 'call_settings'
    },
    
    // 视频配置
    VIDEO_CONFIG: {
        width: { ideal: 1280 },
        height: { ideal: 720 },
        frameRate: { ideal: 30 }
    },
    
    // 音频配置
    AUDIO_CONFIG: {
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true
    },
    
    // 安全检测配置
    SECURITY_CONFIG: {
        detectionInterval: 10000, // 10秒
        riskThreshold: 0.7,
        confidenceThreshold: 0.8
    },
    
    // WebRTC配置
    WEBRTC_CONFIG: {
        iceServers: [
            { urls: 'stun:stun.l.google.com:19302' },
            { urls: 'stun:stun1.l.google.com:19302' },
            { urls: 'stun:stun2.l.google.com:19302' }
        ]
    },
    
    // UI配置
    UI_CONFIG: {
        notificationDuration: 5000,
        loadingTimeout: 30000,
        reconnectAttempts: 3
    }
};

// 导出配置
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CONFIG;
} 