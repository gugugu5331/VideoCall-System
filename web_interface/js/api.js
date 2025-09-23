// API接口类
class API {
    constructor() {
        this.baseURL = CONFIG.API_BASE_URL;
        this.aiServiceURL = CONFIG.AI_SERVICE_URL;
        this.token = localStorage.getItem(CONFIG.STORAGE_KEYS.AUTH_TOKEN);
    }

    // 设置认证token
    setToken(token) {
        this.token = token;
        localStorage.setItem(CONFIG.STORAGE_KEYS.AUTH_TOKEN, token);
    }

    // 清除token
    clearToken() {
        this.token = null;
        localStorage.removeItem(CONFIG.STORAGE_KEYS.AUTH_TOKEN);
    }

    // 获取请求头
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    // 通用请求方法
    async request(url, options = {}) {
        try {
            const response = await fetch(url, {
                ...options,
                headers: this.getHeaders()
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || `HTTP ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API请求失败:', error);
            throw error;
        }
    }

    // 健康检查
    async healthCheck() {
        try {
            return await this.request(`${this.baseURL}/health`);
        } catch (error) {
            console.warn('健康检查失败:', error);
            return { status: 'error', message: error.message };
        }
    }

    // 用户认证相关API
    async register(userData) {
        return this.request(`${this.baseURL}/api/v1/auth/register`, {
            method: 'POST',
            body: JSON.stringify(userData)
        });
    }

    async login(credentials) {
        const response = await this.request(`${this.baseURL}/api/v1/auth/login`, {
            method: 'POST',
            body: JSON.stringify(credentials)
        });
        
        if (response.token) {
            this.setToken(response.token);
        }
        
        return response;
    }

    async logout() {
        this.clearToken();
        return { success: true };
    }

    // 用户资料相关API
    async getUserProfile() {
        return this.request(`${this.baseURL}/api/v1/user/profile`);
    }

    async updateUserProfile(profileData) {
        return this.request(`${this.baseURL}/api/v1/user/profile`, {
            method: 'PUT',
            body: JSON.stringify(profileData)
        });
    }

    // 用户搜索API
    async searchUsers(query, limit = 10) {
        const params = new URLSearchParams({
            query: query,
            limit: limit
        });
        return this.request(`${this.baseURL}/api/v1/users/search?${params}`);
    }

    // 通话相关API
    async startCall(callData) {
        // 支持通过用户名或UUID发起通话
        const requestData = {
            call_type: callData.call_type
        };
        
        if (callData.callee_username) {
            requestData.callee_username = callData.callee_username;
        } else if (callData.callee_id) {
            requestData.callee_id = callData.callee_id;
        }
        
        return this.request(`${this.baseURL}/api/v1/calls/start`, {
            method: 'POST',
            body: JSON.stringify(requestData)
        });
    }

    async endCall(callId) {
        return this.request(`${this.baseURL}/api/v1/calls/end`, {
            method: 'POST',
            body: JSON.stringify({ call_uuid: callId })
        });
    }

    async getCallHistory(filters = {}) {
        const params = new URLSearchParams(filters);
        return this.request(`${this.baseURL}/api/v1/calls/history?${params}`);
    }

    async getCallDetails(callId) {
        return this.request(`${this.baseURL}/api/v1/calls/${callId}`);
    }

    async getActiveCalls() {
        return this.request(`${this.baseURL}/api/v1/calls/active`);
    }

    // 安全检测相关API
    async triggerDetection(detectionData) {
        return this.request(`${this.baseURL}/api/v1/security/detect`, {
            method: 'POST',
            body: JSON.stringify(detectionData)
        });
    }

    async getDetectionStatus(callId) {
        return this.request(`${this.baseURL}/api/v1/security/status/${callId}`);
    }

    async getDetectionHistory(filters = {}) {
        const params = new URLSearchParams(filters);
        return this.request(`${this.baseURL}/api/v1/security/history?${params}`);
    }

    // AI服务相关API
    async detectSpoofing(detectionRequest) {
        return this.request(`${this.aiServiceURL}/detect`, {
            method: 'POST',
            body: JSON.stringify(detectionRequest)
        });
    }

    async getDetectionStatusAI(detectionId) {
        return this.request(`${this.aiServiceURL}/status/${detectionId}`);
    }

    async getAvailableModels() {
        return this.request(`${this.aiServiceURL}/models`);
    }

    // 文件上传
    async uploadFile(file, type = 'avatar') {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('type', type);

        try {
            const response = await fetch(`${this.baseURL}/api/v1/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                },
                body: formData
            });

            if (!response.ok) {
                throw new Error(`上传失败: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('文件上传失败:', error);
            throw error;
        }
    }

    // WebSocket连接
    createWebSocketConnection(callId) {
        const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/call/${callId}`;
        const ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('WebSocket连接已建立');
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleWebSocketMessage(data);
            } catch (error) {
                console.error('WebSocket消息解析失败:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket错误:', error);
        };

        ws.onclose = () => {
            console.log('WebSocket连接已关闭');
        };

        return ws;
    }

    // 处理WebSocket消息
    handleWebSocketMessage(data) {
        switch (data.type) {
            case 'call_connected':
                UI.updateCallStatus('connected');
                break;
            case 'call_disconnected':
                UI.updateCallStatus('disconnected');
                break;
            case 'security_alert':
                UI.showSecurityAlert(data.data);
                break;
            case 'user_joined':
                UI.showNotification(`${data.username} 加入了通话`, 'info');
                break;
            case 'user_left':
                UI.showNotification(`${data.username} 离开了通话`, 'info');
                break;
            default:
                console.log('未知的WebSocket消息类型:', data.type);
        }
    }
}

// 创建全局API实例
const api = new API(); 