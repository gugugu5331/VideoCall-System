// 标签页间通信管理类
class TabCommunication {
    constructor() {
        this.channel = null;
        this.tabId = this.generateTabId();
        this.isInitialized = false;
        this.init();
    }
    
    // 初始化
    init() {
        try {
            if ('BroadcastChannel' in window) {
                this.channel = new BroadcastChannel('videocall_tabs');
                this.setupListeners();
                this.isInitialized = true;
                console.log('标签页通信初始化成功，标签页ID:', this.tabId);
            } else {
                console.warn('浏览器不支持BroadcastChannel API');
            }
        } catch (error) {
            console.error('标签页通信初始化失败:', error);
        }
    }
    
    // 生成标签页唯一ID
    generateTabId() {
        return Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // 设置消息监听器
    setupListeners() {
        if (!this.channel) return;
        
        this.channel.onmessage = (event) => {
            const { type, data, sourceTabId } = event.data;
            
            // 忽略自己发送的消息
            if (sourceTabId === this.tabId) {
                return;
            }
            
            console.log('收到标签页消息:', type, data, '来自标签页:', sourceTabId);
            
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
                case 'tab_closing':
                    this.handleTabClosing(data);
                    break;
                case 'heartbeat':
                    this.handleHeartbeat(data);
                    break;
                default:
                    console.log('未知消息类型:', type);
            }
        };
    }
    
    // 通知设备被使用
    notifyDeviceInUse(deviceType) {
        this.sendMessage('device_in_use', {
            deviceType,
            timestamp: Date.now()
        });
    }
    
    // 通知通话开始
    notifyCallStarted(callData) {
        this.sendMessage('call_started', {
            callData,
            timestamp: Date.now()
        });
    }
    
    // 通知通话结束
    notifyCallEnded(callData) {
        this.sendMessage('call_ended', {
            callData,
            timestamp: Date.now()
        });
    }
    
    // 通知标签页关闭
    notifyTabClosing() {
        this.sendMessage('tab_closing', {
            timestamp: Date.now()
        });
    }
    
    // 发送心跳
    sendHeartbeat() {
        this.sendMessage('heartbeat', {
            timestamp: Date.now()
        });
    }
    
    // 发送消息
    sendMessage(type, data) {
        if (!this.channel || !this.isInitialized) {
            return;
        }
        
        const message = {
            type,
            data,
            sourceTabId: this.tabId,
            timestamp: Date.now()
        };
        
        try {
            this.channel.postMessage(message);
            console.log('发送标签页消息:', type, data);
        } catch (error) {
            console.error('发送标签页消息失败:', error);
        }
    }
    
    // 处理设备被使用消息
    handleDeviceInUse(data) {
        console.log('其他标签页正在使用设备:', data.deviceType);
        UI.showNotification('摄像头或麦克风被其他标签页占用', 'warning');
    }
    
    // 处理通话开始消息
    handleCallStarted(data) {
        console.log('其他标签页开始通话:', data.callData);
        UI.showNotification('其他标签页正在通话中', 'info');
    }
    
    // 处理通话结束消息
    handleCallEnded(data) {
        console.log('其他标签页结束通话:', data.callData);
    }
    
    // 处理标签页关闭消息
    handleTabClosing(data) {
        console.log('其他标签页关闭:', data);
    }
    
    // 处理心跳消息
    handleHeartbeat(data) {
        // 可以用于检测其他标签页的状态
        console.log('收到心跳:', data);
    }
    
    // 获取标签页ID
    getTabId() {
        return this.tabId;
    }
    
    // 检查是否支持
    isSupported() {
        return this.isInitialized && this.channel !== null;
    }
    
    // 关闭连接
    close() {
        if (this.channel) {
            this.notifyTabClosing();
            this.channel.close();
            this.channel = null;
            this.isInitialized = false;
            console.log('标签页通信已关闭');
        }
    }
}

// 页面卸载时通知其他标签页
window.addEventListener('beforeunload', () => {
    if (window.tabCommunication) {
        window.tabCommunication.notifyTabClosing();
    }
});

// 定期发送心跳
setInterval(() => {
    if (window.tabCommunication && window.tabCommunication.isSupported()) {
        window.tabCommunication.sendHeartbeat();
    }
}, 30000); // 每30秒发送一次心跳

// 创建全局实例
window.tabCommunication = new TabCommunication(); 