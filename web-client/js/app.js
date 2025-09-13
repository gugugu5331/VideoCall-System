/**
 * 主应用逻辑
 */

// 全局变量
let currentTab = 'participants';

/**
 * 页面加载完成后初始化
 */
document.addEventListener('DOMContentLoaded', function() {
    console.log('视频会议系统初始化...');
    
    // 检查浏览器支持
    if (!checkBrowserSupport()) {
        alert('您的浏览器不支持WebRTC，请使用Chrome、Firefox或Safari的最新版本');
        return;
    }
    
    // 初始化检测系统
    window.detectionManager.initialize();
    
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
    
    // 文件拖拽支持
    setupFileDragDrop();
    
    // 键盘快捷键
    document.addEventListener('keydown', handleKeyboardShortcuts);
}

/**
 * 加入会议
 */
async function joinMeeting() {
    const username = document.getElementById('username').value.trim();
    const meetingId = document.getElementById('meetingId').value.trim();
    
    if (!username) {
        alert('请输入您的姓名');
        return;
    }
    
    if (!meetingId) {
        alert('请输入会议ID');
        return;
    }
    
    try {
        // 显示连接状态
        window.webrtcManager.updateStatus('正在连接...', 'connecting');
        
        // 初始化WebRTC
        await window.webrtcManager.initialize();
        
        // 加入会议
        await window.webrtcManager.joinMeeting(username, meetingId);
        
        // 隐藏登录模态框
        document.getElementById('loginModal').classList.add('hidden');
        
        // 更新状态
        window.webrtcManager.updateStatus('已加入会议', 'success');
        
        console.log(`用户 ${username} 已加入会议 ${meetingId}`);
        
    } catch (error) {
        console.error('加入会议失败:', error);
        window.webrtcManager.updateStatus('连接失败: ' + error.message, 'error');
    }
}

/**
 * 离开会议
 */
function leaveMeeting() {
    if (confirm('确定要离开会议吗？')) {
        window.webrtcManager.leaveMeeting();
        
        // 清空视频区域
        const videoArea = document.getElementById('videoArea');
        const remoteVideos = videoArea.querySelectorAll('.video-container:not(.local)');
        remoteVideos.forEach(video => video.remove());
        
        // 重置UI状态
        resetUIState();
    }
}

/**
 * 重置UI状态
 */
function resetUIState() {
    // 重置控制按钮
    document.getElementById('cameraBtn').classList.remove('off');
    document.getElementById('micBtn').classList.remove('off');
    
    // 清空聊天消息
    document.getElementById('chatMessages').innerHTML = '';
    
    // 清空参与者列表
    document.getElementById('participantList').innerHTML = '';
    document.getElementById('participantCount').textContent = '0';
    
    // 切换到参与者标签
    switchTab('participants');
}

/**
 * 切换摄像头
 */
async function toggleCamera() {
    await window.webrtcManager.toggleCamera();
}

/**
 * 切换麦克风
 */
async function toggleMicrophone() {
    await window.webrtcManager.toggleMicrophone();
}

/**
 * 切换屏幕共享
 */
async function toggleScreenShare() {
    await window.webrtcManager.toggleScreenShare();
}

/**
 * 切换本地视频
 */
function toggleLocalVideo() {
    toggleCamera();
}

/**
 * 切换本地音频
 */
function toggleLocalAudio() {
    toggleMicrophone();
}

/**
 * 静音远程用户
 */
function muteRemoteUser(userId) {
    const videoContainer = document.getElementById(`video-${userId}`);
    if (videoContainer) {
        const video = videoContainer.querySelector('video');
        if (video) {
            video.muted = !video.muted;
            const button = videoContainer.querySelector('.control-btn');
            if (button) {
                button.textContent = video.muted ? '🔊' : '🔇';
                button.title = video.muted ? '取消静音' : '静音';
            }
        }
    }
}

/**
 * 选择主视频
 */
function selectMainVideo(userId) {
    if (window.webrtcManager) {
        window.webrtcManager.selectMainVideo(userId);
    }
}

/**
 * 切换主视频静音
 */
function toggleMainVideoMute() {
    if (window.webrtcManager) {
        window.webrtcManager.toggleMainVideoMute();
    }
}

/**
 * 主视频全屏
 */
function toggleMainVideoFullscreen() {
    if (window.webrtcManager) {
        window.webrtcManager.toggleMainVideoFullscreen();
    }
}

/**
 * 刷新参与者列表
 */
function refreshParticipants() {
    if (window.webrtcManager) {
        const participants = window.webrtcManager.getAllParticipants();
        console.log('当前参与者:', participants);

        // 更新参与者计数
        updateParticipantCount(participants.length);

        // 可以在这里添加更多的刷新逻辑
        addInfoMessage('参与者列表已刷新');
    }
}

/**
 * 更新参与者计数
 */
function updateParticipantCount(count) {
    const participantCountElement = document.getElementById('participantCount');
    if (participantCountElement) {
        participantCountElement.textContent = count || 0;
    }
}

/**
 * 添加信息消息（用于调试）
 */
function addInfoMessage(message, type = 'info') {
    console.log(`[${type.toUpperCase()}] ${message}`);
}

/**
 * 切换侧边栏标签
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
 * 发送聊天消息
 */
function sendChatMessage() {
    const chatInput = document.getElementById('chatInput');
    const message = chatInput.value.trim();
    
    if (message && window.webrtcManager.isConnected) {
        window.webrtcManager.sendChatMessage(message);
        chatInput.value = '';
    }
}

/**
 * 处理聊天输入框回车键
 */
function handleChatKeyPress(event) {
    if (event.key === 'Enter') {
        sendChatMessage();
    }
}

/**
 * 切换检测状态
 */
function toggleDetection() {
    const checkbox = document.getElementById('enableDetection');
    window.detectionManager.toggleDetection(checkbox.checked);
}

/**
 * 切换设置面板
 */
function toggleSettings() {
    // 这里可以实现设置面板的显示/隐藏
    console.log('切换设置面板');
}

/**
 * 设置文件拖拽支持
 */
function setupFileDragDrop() {
    const detectionTab = document.getElementById('detectionTab');
    
    // 防止默认拖拽行为
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        detectionTab.addEventListener(eventName, preventDefaults, false);
    });
    
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    // 拖拽进入和离开的视觉反馈
    ['dragenter', 'dragover'].forEach(eventName => {
        detectionTab.addEventListener(eventName, highlight, false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        detectionTab.addEventListener(eventName, unhighlight, false);
    });
    
    function highlight(e) {
        detectionTab.style.background = '#4facfe20';
    }
    
    function unhighlight(e) {
        detectionTab.style.background = '';
    }
    
    // 处理文件放置
    detectionTab.addEventListener('drop', handleDrop, false);
    
    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        
        if (files.length > 0) {
            handleFileUpload(files[0]);
        }
    }
}

/**
 * 处理文件上传
 */
async function handleFileUpload(file) {
    try {
        // 检查文件类型
        const allowedTypes = ['image/', 'video/', 'audio/'];
        const isAllowed = allowedTypes.some(type => file.type.startsWith(type));
        
        if (!isAllowed) {
            alert('只支持图片、视频和音频文件');
            return;
        }
        
        // 检查文件大小 (50MB限制)
        if (file.size > 50 * 1024 * 1024) {
            alert('文件大小不能超过50MB');
            return;
        }
        
        // 显示上传状态
        window.webrtcManager.updateStatus('正在检测文件...', 'connecting');
        
        // 执行检测
        const result = await window.detectionManager.detectFile(file);
        
        // 更新状态
        window.webrtcManager.updateStatus('文件检测完成', 'success');
        
        console.log('文件检测结果:', result);
        
    } catch (error) {
        console.error('文件检测失败:', error);
        window.webrtcManager.updateStatus('文件检测失败: ' + error.message, 'error');
    }
}

/**
 * 处理键盘快捷键
 */
function handleKeyboardShortcuts(event) {
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
    
    // Ctrl/Cmd + Enter: 发送聊天消息
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
        if (currentTab === 'chat') {
            event.preventDefault();
            sendChatMessage();
        }
    }
    
    // Esc: 离开会议
    if (event.key === 'Escape') {
        if (!document.getElementById('loginModal').classList.contains('hidden')) {
            return; // 如果登录模态框显示，不处理ESC
        }
        event.preventDefault();
        leaveMeeting();
    }
}

/**
 * 获取媒体设备列表
 */
async function getMediaDevices() {
    try {
        const devices = await navigator.mediaDevices.enumerateDevices();
        
        const cameras = devices.filter(device => device.kind === 'videoinput');
        const microphones = devices.filter(device => device.kind === 'audioinput');
        const speakers = devices.filter(device => device.kind === 'audiooutput');
        
        return { cameras, microphones, speakers };
        
    } catch (error) {
        console.error('获取媒体设备失败:', error);
        return { cameras: [], microphones: [], speakers: [] };
    }
}

/**
 * 切换摄像头设备
 */
async function switchCamera(deviceId) {
    try {
        const constraints = {
            video: { deviceId: { exact: deviceId } },
            audio: false
        };
        
        const newStream = await navigator.mediaDevices.getUserMedia(constraints);
        const videoTrack = newStream.getVideoTracks()[0];
        
        await window.webrtcManager.replaceVideoTrack(videoTrack);
        
        console.log('摄像头切换成功');
        
    } catch (error) {
        console.error('切换摄像头失败:', error);
    }
}

/**
 * 显示网络质量信息
 */
function showNetworkQuality() {
    // 这里可以实现网络质量检测和显示
    console.log('显示网络质量信息');
}

/**
 * 显示系统信息
 */
function showSystemInfo() {
    const info = {
        browser: navigator.userAgent,
        webrtc: !!window.RTCPeerConnection,
        websocket: !!window.WebSocket,
        mediaDevices: !!navigator.mediaDevices,
        screen: screen.width + 'x' + screen.height
    };
    
    console.log('系统信息:', info);
    return info;
}

/**
 * 导出会议记录
 */
function exportMeetingRecord() {
    const stats = window.detectionManager.getDetectionStats();
    const record = {
        meeting_id: window.webrtcManager.meetingId,
        user: window.webrtcManager.currentUser,
        start_time: new Date().toISOString(),
        detection_stats: stats,
        participants: document.querySelectorAll('.participant-item').length
    };
    
    const blob = new Blob([JSON.stringify(record, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `meeting-record-${record.meeting_id}-${Date.now()}.json`;
    a.click();
    
    URL.revokeObjectURL(url);
}

// 导出全局函数供HTML调用
window.joinMeeting = joinMeeting;
window.leaveMeeting = leaveMeeting;
window.toggleCamera = toggleCamera;
window.toggleMicrophone = toggleMicrophone;
window.toggleScreenShare = toggleScreenShare;
window.toggleLocalVideo = toggleLocalVideo;
window.toggleLocalAudio = toggleLocalAudio;
window.muteRemoteUser = muteRemoteUser;
window.switchTab = switchTab;
window.sendChatMessage = sendChatMessage;
window.handleChatKeyPress = handleChatKeyPress;
window.toggleDetection = toggleDetection;
window.toggleSettings = toggleSettings;
