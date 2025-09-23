/**
 * 增强版视频会议应用逻辑
 */

// 全局变量
let currentTab = 'participants';
let isInMeeting = false;

/**
 * 页面加载完成后初始化
 */
document.addEventListener('DOMContentLoaded', function() {
    console.log('增强版视频会议系统初始化...');

    // 检查浏览器支持
    if (!checkBrowserSupport()) {
        showNotification('您的浏览器不支持WebRTC，请使用Chrome、Firefox或Safari的最新版本', 'error');
        return;
    }

    // 检查必要的全局对象
    console.log('检查全局对象...');
    console.log('window.webrtcManager:', window.webrtcManager);
    console.log('window.detectionManager:', window.detectionManager);

    // 初始化检测系统
    if (window.detectionManager) {
        window.detectionManager.initialize();
    } else {
        console.warn('detectionManager未找到，跳过检测系统初始化');
    }

    // 绑定事件监听器
    bindEventListeners();

    console.log('系统初始化完成');
});

/**
 * 检查浏览器支持
 */
function checkBrowserSupport() {
    return !!(navigator.mediaDevices && 
              navigator.mediaDevices.getUserMedia && 
              window.RTCPeerConnection &&
              window.WebSocket);
}

/**
 * 绑定事件监听器
 */
function bindEventListeners() {
    // 登录表单回车键支持
    document.getElementById('username').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            document.getElementById('meetingId').focus();
        }
    });
    
    document.getElementById('meetingId').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            joinMeeting();
        }
    });
    
    // 键盘快捷键
    document.addEventListener('keydown', handleKeyboardShortcuts);
}

/**
 * 测试函数
 */
function testFunction() {
    console.log('测试函数被调用');
    alert('测试函数工作正常！');
    showNotification('测试通知功能', 'success');
}

/**
 * 测试加入会议函数
 */
function testJoinMeeting() {
    console.log('testJoinMeeting被调用');
    joinMeeting();
}

/**
 * 加入会议
 */
async function joinMeeting() {
    console.log('joinMeeting函数被调用');

    const username = document.getElementById('username').value.trim();
    const meetingId = document.getElementById('meetingId').value.trim();

    console.log('用户名:', username, '会议ID:', meetingId);

    if (!username) {
        showNotification('请输入您的姓名', 'error');
        return;
    }

    if (!meetingId) {
        showNotification('请输入会议ID', 'error');
        return;
    }

    try {
        showNotification('正在连接...', 'info');

        // 检查webrtcManager是否存在
        if (!window.webrtcManager) {
            console.error('webrtcManager未初始化');
            showNotification('系统初始化失败，请刷新页面重试', 'error');
            return;
        }

        console.log('开始初始化WebRTC...');

        // 初始化WebRTC
        await window.webrtcManager.initialize();

        console.log('WebRTC初始化完成，开始加入会议...');

        // 加入会议
        await window.webrtcManager.joinMeeting(username, meetingId);

        console.log('会议加入成功，更新UI...');

        // 等待一小段时间确保WebRTC初始化完成
        await new Promise(resolve => setTimeout(resolve, 500));

        // 隐藏登录模态框
        document.getElementById('loginModal').classList.add('hidden');

        // 更新UI状态
        isInMeeting = true;
        document.getElementById('currentMeetingId').textContent = meetingId;

        // 设置本地视频为默认主视频
        setTimeout(() => {
            if (window.webrtcManager && window.webrtcManager.selectMainVideo) {
                window.webrtcManager.selectMainVideo('local');
            }
        }, 1000);

        showNotification('已成功加入会议', 'success');

        console.log(`用户 ${username} 已加入会议 ${meetingId}`);

    } catch (error) {
        console.error('加入会议失败:', error);
        showNotification('连接失败: ' + error.message, 'error');
    }
}

/**
 * 离开会议
 */
function leaveMeeting() {
    if (!isInMeeting) return;
    
    if (confirm('确定要离开会议吗？')) {
        window.webrtcManager.leaveMeeting();
        
        // 重置UI状态
        isInMeeting = false;
        document.getElementById('currentMeetingId').textContent = '-';
        updateParticipantCount(0);
        
        // 清空参与者列表和缩略图
        document.getElementById('participantList').innerHTML = '';
        document.getElementById('thumbnailsGrid').innerHTML = '';
        document.getElementById('chatMessages').innerHTML = '';
        
        // 显示登录模态框
        document.getElementById('loginModal').classList.remove('hidden');
        
        showNotification('已离开会议', 'info');
    }
}

/**
 * 添加本地参与者
 */
function addLocalParticipant(username) {
    // 添加到参与者列表
    const participantList = document.getElementById('participantList');
    const participantElement = document.createElement('div');
    participantElement.className = 'participant-item main-speaker';
    participantElement.id = 'participant-local';
    participantElement.onclick = () => selectMainVideo('local');
    
    participantElement.innerHTML = `
        <div class="participant-avatar">${username.charAt(0).toUpperCase()}</div>
        <div class="participant-info">
            <div class="participant-name">${username} (您)</div>
            <div class="participant-status">
                <div class="status-indicator"></div>
                <span>在线 • 主讲人</span>
            </div>
        </div>
    `;
    
    participantList.appendChild(participantElement);
    
    // 添加到缩略图
    if (window.webrtcManager && window.webrtcManager.addThumbnailVideo) {
        window.webrtcManager.addThumbnailVideo('local', username, true);
    }

    // 更新参与者计数
    if (window.webrtcManager && window.webrtcManager.updateParticipantCount) {
        window.webrtcManager.updateParticipantCount();
    }
}



/**
 * 选择主视频
 */
function selectMainVideo(userId) {
    if (window.webrtcManager) {
        window.webrtcManager.selectMainVideo(userId);
        
        // 更新缩略图选中状态
        document.querySelectorAll('.thumbnail-video').forEach(thumb => {
            thumb.classList.remove('selected');
        });
        document.getElementById(`thumbnail-${userId}`).classList.add('selected');
        
        // 更新参与者列表主讲人状态
        document.querySelectorAll('.participant-item').forEach(item => {
            item.classList.remove('main-speaker');
            const statusSpan = item.querySelector('.participant-status span');
            if (statusSpan) {
                statusSpan.textContent = statusSpan.textContent.replace(' • 主讲人', '');
            }
        });
        
        const mainParticipant = document.getElementById(`participant-${userId}`);
        if (mainParticipant) {
            mainParticipant.classList.add('main-speaker');
            const statusSpan = mainParticipant.querySelector('.participant-status span');
            if (statusSpan) {
                statusSpan.textContent += ' • 主讲人';
            }
        }
    }
}

/**
 * 切换标签页
 */
function switchTab(tabName) {
    // 更新标签状态
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelector(`.tab:nth-child(${getTabIndex(tabName)})`).classList.add('active');
    
    // 显示对应内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.add('hidden');
    });
    document.getElementById(`${tabName}Tab`).classList.remove('hidden');
    
    currentTab = tabName;
}

/**
 * 获取标签索引
 */
function getTabIndex(tabName) {
    const tabMap = {
        'participants': 1,
        'chat': 2,
        'detection': 3
    };
    return tabMap[tabName] || 1;
}

/**
 * 更新参与者计数
 */
function updateParticipantCount(count) {
    document.getElementById('participantCount').textContent = count;
}

/**
 * 显示通知
 */
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = 'notification';
    
    const icons = {
        'info': 'ℹ️',
        'success': '✅',
        'error': '❌',
        'warning': '⚠️'
    };
    
    notification.innerHTML = `${icons[type] || 'ℹ️'} ${message}`;
    document.body.appendChild(notification);
    
    // 自动移除通知
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 3000);
}

/**
 * 媒体控制函数
 */
async function toggleCamera() {
    await window.webrtcManager.toggleCamera();
    const btn = document.getElementById('cameraBtn');
    const isEnabled = window.webrtcManager.isVideoEnabled;
    btn.classList.toggle('off', !isEnabled);
    showNotification(`摄像头已${isEnabled ? '开启' : '关闭'}`, 'info');
}

async function toggleMicrophone() {
    await window.webrtcManager.toggleMicrophone();
    const btn = document.getElementById('micBtn');
    const isEnabled = window.webrtcManager.isAudioEnabled;
    btn.classList.toggle('off', !isEnabled);
    showNotification(`麦克风已${isEnabled ? '开启' : '关闭'}`, 'info');
}

async function toggleScreenShare() {
    await window.webrtcManager.toggleScreenShare();
    const isSharing = window.webrtcManager.isScreenSharing;
    showNotification(`屏幕共享已${isSharing ? '开启' : '关闭'}`, 'info');
}

function toggleMainVideoMute() {
    if (window.webrtcManager) {
        window.webrtcManager.toggleMainVideoMute();
    }
}

function toggleMainVideoFullscreen() {
    if (window.webrtcManager) {
        window.webrtcManager.toggleMainVideoFullscreen();
    }
}

/**
 * 聊天功能
 */
function sendChatMessage() {
    const chatInput = document.getElementById('chatInput');
    const message = chatInput.value.trim();

    console.log('尝试发送聊天消息:', {
        message: message,
        hasWebrtcManager: !!window.webrtcManager,
        isConnected: window.webrtcManager?.isConnected,
        currentUser: window.webrtcManager?.currentUser
    });

    if (message && window.webrtcManager && window.webrtcManager.isConnected) {
        window.webrtcManager.sendChatMessage(message);
        chatInput.value = '';
        console.log('聊天消息已发送:', message);
    } else {
        console.log('无法发送聊天消息 - 原因:', {
            hasMessage: !!message,
            hasWebrtcManager: !!window.webrtcManager,
            isConnected: window.webrtcManager?.isConnected
        });
    }
}

function handleChatKeyPress(event) {
    if (event.key === 'Enter') {
        sendChatMessage();
    }
}

function addChatMessage(sender, message, isOwn = false) {
    const chatMessages = document.getElementById('chatMessages');
    const messageElement = document.createElement('div');
    messageElement.style.cssText = `
        margin-bottom: 10px;
        padding: 8px 12px;
        background: ${isOwn ? 'rgba(79, 172, 254, 0.3)' : 'rgba(255, 255, 255, 0.1)'};
        border-radius: 8px;
        ${isOwn ? 'margin-left: 20px;' : ''}
    `;
    
    messageElement.innerHTML = `
        <div style="font-weight: bold; font-size: 12px; margin-bottom: 2px;">${sender}</div>
        <div>${message}</div>
    `;
    
    chatMessages.appendChild(messageElement);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

/**
 * 检测功能
 */
function toggleDetection() {
    const checkbox = document.getElementById('enableDetection');
    window.detectionManager.toggleDetection(checkbox.checked);
    showNotification(`AI检测已${checkbox.checked ? '启用' : '禁用'}`, 'info');
}

/**
 * 键盘快捷键处理
 */
function handleKeyboardShortcuts(event) {
    if (!isInMeeting) return;
    
    // Ctrl/Cmd + M: 切换麦克风
    if ((event.ctrlKey || event.metaKey) && event.key === 'm') {
        event.preventDefault();
        toggleMicrophone();
    }
    
    // Ctrl/Cmd + D: 切换摄像头
    if ((event.ctrlKey || event.metaKey) && event.key === 'd') {
        event.preventDefault();
        toggleCamera();
    }
    
    // Ctrl/Cmd + S: 切换屏幕共享
    if ((event.ctrlKey || event.metaKey) && event.key === 's') {
        event.preventDefault();
        toggleScreenShare();
    }
    
    // F: 全屏主视频
    if (event.key === 'f' || event.key === 'F') {
        event.preventDefault();
        toggleMainVideoFullscreen();
    }
    
    // 数字键1-9: 快速选择参与者
    if (event.key >= '1' && event.key <= '9') {
        const index = parseInt(event.key) - 1;
        const participants = window.webrtcManager ? window.webrtcManager.getAllParticipants() : [];
        if (participants[index]) {
            selectMainVideo(participants[index].id);
        }
    }
}

// 导出全局函数
window.testFunction = testFunction;
window.testJoinMeeting = testJoinMeeting;
window.joinMeeting = joinMeeting;
window.leaveMeeting = leaveMeeting;
window.selectMainVideo = selectMainVideo;
window.switchTab = switchTab;
window.toggleCamera = toggleCamera;
window.toggleMicrophone = toggleMicrophone;
window.toggleScreenShare = toggleScreenShare;
window.toggleMainVideoMute = toggleMainVideoMute;
window.toggleMainVideoFullscreen = toggleMainVideoFullscreen;
window.sendChatMessage = sendChatMessage;
window.handleChatKeyPress = handleChatKeyPress;
window.toggleDetection = toggleDetection;

// 主讲人状态
let isPresenter = false;

/**
 * 切换主讲人状态
 */
function togglePresenter() {
    if (!window.webrtcManager) {
        console.log('WebRTC管理器未初始化');
        return;
    }

    const presenterBtn = document.getElementById('presenterBtn');

    if (isPresenter) {
        // 释放主讲人权限
        window.webrtcManager.releasePresenter();
        isPresenter = false;
        presenterBtn.classList.remove('active');
        presenterBtn.title = '申请主讲';
        console.log('释放主讲人权限');
    } else {
        // 申请主讲人权限
        window.webrtcManager.requestPresenter();
        console.log('申请主讲人权限');
    }
}

/**
 * 处理主讲人状态变更
 */
function handlePresenterStatusChange(status, data) {
    const presenterBtn = document.getElementById('presenterBtn');

    switch (status) {
        case 'presenter-set':
            isPresenter = true;
            presenterBtn.classList.add('active');
            presenterBtn.title = '释放主讲';
            showNotification('您现在是主讲人', 'success');
            break;

        case 'presenter-removed':
            isPresenter = false;
            presenterBtn.classList.remove('active');
            presenterBtn.title = '申请主讲';
            if (data && data.message) {
                showNotification(data.message, 'info');
            }
            break;

        case 'presenter-changed':
            if (data && data.presenter_name) {
                showNotification(`${data.presenter_name} 现在是主讲人`, 'info');
            }
            break;
    }
}

/**
 * 显示通知
 */
function showNotification(message, type = 'info') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;

    // 添加样式
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${type === 'success' ? '#28a745' : type === 'error' ? '#dc3545' : '#17a2b8'};
        color: white;
        padding: 15px 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 10000;
        font-weight: 500;
        max-width: 300px;
        word-wrap: break-word;
        animation: slideIn 0.3s ease-out;
    `;

    // 添加动画样式
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        @keyframes slideOut {
            from { transform: translateX(0); opacity: 1; }
            to { transform: translateX(100%); opacity: 0; }
        }
    `;
    document.head.appendChild(style);

    document.body.appendChild(notification);

    // 3秒后自动移除
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease-in';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 3000);
}

// 导出主讲人功能
window.togglePresenter = togglePresenter;
window.handlePresenterStatusChange = handlePresenterStatusChange;

// 页面生命周期管理
window.addEventListener('beforeunload', function(event) {
    if (isInMeeting && window.webrtcManager) {
        // 发送离开消息
        window.webrtcManager.sendSignalingMessage({
            type: 'leave-meeting',
            data: {
                meetingId: window.webrtcManager.meetingId,
                user: window.webrtcManager.currentUser
            }
        });
    }
});

// 防止意外刷新
window.addEventListener('beforeunload', function(event) {
    if (isInMeeting) {
        event.preventDefault();
        event.returnValue = '您正在会议中，确定要离开吗？';
        return '您正在会议中，确定要离开吗？';
    }
});
