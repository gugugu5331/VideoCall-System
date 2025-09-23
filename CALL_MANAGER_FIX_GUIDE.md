# 通话管理器修复指南

## 问题描述

用户报告"通话管理器未初始化"错误，导致无法进行多用户通话。

## 问题原因

1. **初始化冲突**：`call.js`文件中存在全局`callManager`实例，与`main.js`中的`window.callManager`冲突
2. **初始化时机**：`main.js`中没有在应用启动时初始化通话管理器
3. **脚本加载顺序**：可能存在脚本加载时序问题

## 修复方案

### 1. 修复main.js初始化

在`web_interface/js/main.js`的`App.init()`方法中添加通话管理器初始化：

```javascript
// 初始化应用
async init() {
    try {
        UI.showLoading();
        
        // 检查浏览器兼容性
        this.checkBrowserCompatibility();
        
        // 初始化认证状态
        auth.init();
        
        // 初始化通话管理器
        this.initCallManager();
        
        // 设置事件监听器
        this.setupEventListeners();
        
        // ... 其他初始化代码
    } catch (error) {
        console.error('应用初始化失败:', error);
        UI.hideLoading();
        UI.showNotification('应用初始化失败: ' + error.message, 'error');
    }
}

// 初始化通话管理器
initCallManager() {
    try {
        // 检查CallManager类是否存在
        if (typeof CallManager !== 'undefined') {
            // 创建全局通话管理器实例
            window.callManager = new CallManager();
            console.log('通话管理器初始化成功');
        } else {
            console.error('CallManager类未定义，请检查call.js文件是否正确加载');
            UI.showNotification('通话管理器初始化失败', 'error');
        }
    } catch (error) {
        console.error('通话管理器初始化失败:', error);
        UI.showNotification('通话管理器初始化失败: ' + error.message, 'error');
    }
}
```

### 2. 修复call.js全局函数

在`web_interface/js/call.js`中移除冲突的全局实例，更新全局函数：

```javascript
// 移除这行代码：
// const callManager = new CallManager();

// 更新全局函数：
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
```

## 测试方法

### 1. 使用测试页面

访问 `http://localhost:8080/test_call_manager.html` 来测试通话管理器初始化状态。

### 2. 浏览器控制台检查

打开浏览器开发者工具，在控制台中运行：

```javascript
// 检查CallManager类是否存在
console.log('CallManager类:', typeof CallManager);

// 检查全局通话管理器实例
console.log('window.callManager:', window.callManager);

// 检查通话管理器状态
if (window.callManager) {
    console.log('通话状态:', window.callManager.getCallStatus());
}
```

### 3. 功能测试

1. 打开两个浏览器窗口
2. 分别使用不同用户登录
3. 测试用户搜索功能
4. 测试通话发起功能
5. 验证WebRTC连接

## 预期结果

修复后应该看到：

1. ✅ 控制台显示"通话管理器初始化成功"
2. ✅ `window.callManager`对象存在
3. ✅ 可以正常发起多用户通话
4. ✅ 不再出现"通话管理器未初始化"错误

## 故障排除

如果问题仍然存在：

1. **检查脚本加载顺序**：确保`call.js`在`main.js`之前加载
2. **清除浏览器缓存**：强制刷新页面（Ctrl+F5）
3. **检查控制台错误**：查看是否有JavaScript错误
4. **验证后端服务**：确保后端服务正常运行

## 文件修改清单

- ✅ `web_interface/js/main.js` - 添加通话管理器初始化
- ✅ `web_interface/js/call.js` - 修复全局函数冲突
- ✅ `test_call_manager.html` - 创建测试页面

## 状态

**修复状态**: ✅ 已完成
**测试状态**: 🔄 待测试
**部署状态**: 🔄 待部署 