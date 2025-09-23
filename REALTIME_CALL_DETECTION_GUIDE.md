# 实时来电检测功能指南

## 功能概述

实时来电检测功能允许系统在用户发起通话时，立即通知被叫方，无需等待被叫方主动检查。这提供了类似手机来电的实时体验。

## 技术架构

### 1. 前端实时检测机制

#### 双重检测策略
- **定时检查**: 每5秒检查一次通话历史，查找未接来电
- **WebSocket监听**: 建立专门的通知WebSocket连接，接收实时通知

#### 核心功能
- `startRealTimeCallDetection()`: 启动实时检测
- `connectCallNotificationWebSocket()`: 建立通知WebSocket连接
- `handleIncomingCallNotification()`: 处理实时来电通知
- `playIncomingCallSound()`: 播放来电铃声

### 2. 后端通知系统

#### 通知WebSocket处理器
- `NotificationWebSocketHandler`: 处理通知WebSocket连接
- `addNotificationConnection()`: 管理用户通知连接
- `sendNotificationToUser()`: 发送通知给指定用户

#### 实时通知流程
1. 用户连接通知WebSocket
2. 系统记录用户连接
3. 当有来电时，立即发送通知
4. 前端接收通知并显示来电界面

## 功能特性

### ✅ 实时性
- **即时通知**: 通话发起后立即通知被叫方
- **低延迟**: WebSocket连接确保最小延迟
- **自动重连**: 连接断开时自动重新连接

### ✅ 用户体验
- **来电铃声**: 播放音频提示音
- **视觉通知**: 美观的来电通知界面
- **动画效果**: 滑入动画和脉冲效果
- **响应式设计**: 适配移动端和桌面端

### ✅ 可靠性
- **双重保障**: WebSocket + 定时检查
- **连接管理**: 自动管理WebSocket连接
- **错误处理**: 完善的错误处理和重试机制

## 使用方法

### 1. 自动启动
用户登录后，系统自动启动实时来电检测：
```javascript
// 在CallManager.init()中自动启动
if (auth.isAuthenticated) {
    this.startRealTimeCallDetection();
}
```

### 2. 手动控制
```javascript
// 启动实时检测
window.callManager.startRealTimeCallDetection();

// 停止实时检测
window.callManager.stopRealTimeCallDetection();
```

### 3. 浏览器测试
1. 打开两个浏览器窗口
2. 分别使用不同用户登录
3. 在主叫方发起通话
4. 被叫方立即收到来电通知

## 技术实现

### 前端实现

#### 通知WebSocket连接
```javascript
connectCallNotificationWebSocket() {
    const wsUrl = `${CONFIG.WS_URL.replace('http', 'ws')}/ws/notifications?user_id=${currentUser.uuid}`;
    this.notificationWebSocket = new WebSocket(wsUrl);
    
    this.notificationWebSocket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        if (message.type === 'incoming_call') {
            this.handleIncomingCallNotification(message.data);
        }
    };
}
```

#### 来电铃声播放
```javascript
playIncomingCallSound() {
    const audioContext = new (window.AudioContext || window.webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();
    
    oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
    oscillator.frequency.setValueAtTime(600, audioContext.currentTime + 0.5);
    
    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + 1);
}
```

### 后端实现

#### 通知连接管理
```go
var (
    notificationConnections = make(map[string]*websocket.Conn)
    notificationMutex       sync.RWMutex
)

func addNotificationConnection(userID string, conn *websocket.Conn) {
    notificationMutex.Lock()
    defer notificationMutex.Unlock()
    
    if oldConn, exists := notificationConnections[userID]; exists {
        oldConn.Close()
    }
    
    notificationConnections[userID] = conn
}
```

#### 发送实时通知
```go
func sendNotificationToUser(userID string, notificationType string, data interface{}) {
    notificationMutex.RLock()
    conn, exists := notificationConnections[userID]
    notificationMutex.RUnlock()

    if exists {
        message := gin.H{
            "type": notificationType,
            "data": data,
            "timestamp": time.Now().Unix(),
        }
        conn.WriteJSON(message)
    }
}
```

## 测试方法

### 1. 自动化测试
```bash
python test_realtime_call_detection.py
```

### 2. 手动测试
1. 启动后端服务
2. 打开两个浏览器窗口
3. 分别登录不同用户
4. 发起通话测试实时通知

### 3. 控制台检查
```javascript
// 检查通知WebSocket状态
console.log('通知WebSocket状态:', window.callManager?.notificationWebSocket?.readyState);

// 检查实时检测状态
console.log('实时检测间隔:', window.callManager?.callCheckInterval);

// 检查当前来电
console.log('当前来电:', window.callManager?.currentIncomingCall);
```

## 配置选项

### 检测间隔
```javascript
// 修改定时检查间隔（默认5秒）
this.callCheckInterval = setInterval(() => {
    this.checkIncomingCalls();
}, 5000); // 可调整此值
```

### 铃声设置
```javascript
// 修改铃声频率和持续时间
oscillator.frequency.setValueAtTime(800, audioContext.currentTime); // 频率
oscillator.stop(audioContext.currentTime + 1); // 持续时间
```

### 通知样式
```css
/* 修改通知样式 */
.incoming-call-notification {
    animation: slideInRight 0.5s ease-out, pulse 2s infinite;
    /* 可调整动画效果 */
}
```

## 故障排除

### 常见问题

#### 1. 没有收到实时通知
- 检查WebSocket连接状态
- 确认后端通知服务正常运行
- 查看浏览器控制台错误信息

#### 2. 通知延迟
- 检查网络连接
- 确认WebSocket连接稳定
- 查看后端日志

#### 3. 铃声不播放
- 检查浏览器音频权限
- 确认音频上下文支持
- 查看控制台错误信息

### 调试方法

#### 前端调试
```javascript
// 启用详细日志
console.log('实时检测状态:', {
    websocket: window.callManager?.notificationWebSocket?.readyState,
    interval: !!window.callManager?.callCheckInterval,
    currentCall: window.callManager?.currentIncomingCall
});
```

#### 后端调试
```go
// 查看通知连接状态
log.Printf("当前通知连接数: %d", len(notificationConnections))
log.Printf("用户 %s 的通知连接: %v", userID, notificationConnections[userID] != nil)
```

## 性能优化

### 1. 连接管理
- 自动清理断开的连接
- 限制每个用户的最大连接数
- 实现连接池管理

### 2. 消息优化
- 压缩通知消息
- 实现消息队列
- 添加消息优先级

### 3. 资源管理
- 及时清理定时器
- 优化音频资源使用
- 实现内存泄漏检测

## 扩展功能

### 1. 推送通知
- 集成浏览器推送API
- 支持离线通知
- 添加通知权限管理

### 2. 多设备同步
- 支持多设备同时在线
- 实现设备间通知同步
- 添加设备管理功能

### 3. 通知历史
- 保存通知历史记录
- 支持通知搜索和过滤
- 添加通知统计功能

## 状态

**开发状态**: ✅ 已完成
**测试状态**: 🔄 待测试
**部署状态**: 🔄 待部署
**文档状态**: ✅ 已完成 