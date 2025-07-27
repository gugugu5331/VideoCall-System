// 全局变量
let currentUser = null;
let currentCall = null;
let callDuration = 0;
let callTimer = null;
let localStream = null;
let remoteStream = null;
let peerConnection = null;

// API配置
const API_BASE_URL = 'http://localhost:8000/api/v1';
const AI_SERVICE_URL = 'http://localhost:5001';

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

// 初始化应用
function initializeApp() {
    // 检查登录状态
    checkLoginStatus();
    
    // 绑定事件监听器
    bindEventListeners();
    
    // 检查系统状态
    checkSystemStatus();
    
    // 定期检查连接状态
    setInterval(checkConnectionStatus, 30000);
}

// 绑定事件监听器
function bindEventListeners() {
    // 登录表单
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    
    // 注册表单
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }
    
    // 导航菜单
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const tab = this.getAttribute('data-tab');
            switchTab(tab);
        });
    });
}

// 检查登录状态
function checkLoginStatus() {
    const token = localStorage.getItem('authToken');
    if (token) {
        // 验证token有效性
        validateToken(token);
    } else {
        showPage('loginPage');
    }
}

// 验证token
async function validateToken(token) {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/validate`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            currentUser = data.user;
            showPage('mainPage');
            updateUserInfo();
        } else {
            localStorage.removeItem('authToken');
            showPage('loginPage');
        }
    } catch (error) {
        console.error('Token验证失败:', error);
        localStorage.removeItem('authToken');
        showPage('loginPage');
    }
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const statusElement = document.getElementById('loginStatus');
    
    if (!username || !password) {
        showStatus(statusElement, '请输入用户名和密码', 'error');
        return;
    }
    
    try {
        showStatus(statusElement, '正在登录...', '');
        
        const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            localStorage.setItem('authToken', data.token);
            currentUser = data.user;
            showStatus(statusElement, '登录成功！', 'success');
            console.log('登录成功，用户信息:', currentUser);
            setTimeout(() => {
                console.log('切换到主页面');
                showPage('mainPage');
                updateUserInfo();
            }, 1000);
        } else {
            showStatus(statusElement, data.message || '登录失败', 'error');
        }
    } catch (error) {
        console.error('登录错误:', error);
        showStatus(statusElement, '网络错误，请检查连接', 'error');
    }
}

// 处理注册
async function handleRegister(e) {
    e.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    const statusElement = document.getElementById('registerStatus');
    
    if (!username || !email || !password || !confirmPassword) {
        showStatus(statusElement, '请填写所有字段', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showStatus(statusElement, '两次输入的密码不一致', 'error');
        return;
    }
    
    try {
        showStatus(statusElement, '正在注册...', '');
        
        const response = await fetch(`${API_BASE_URL}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, email, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showStatus(statusElement, '注册成功！请登录', 'success');
            setTimeout(() => {
                showLogin();
            }, 1500);
        } else {
            showStatus(statusElement, data.message || '注册失败', 'error');
        }
    } catch (error) {
        console.error('注册错误:', error);
        showStatus(statusElement, '网络错误，请检查连接', 'error');
    }
}

// 显示状态消息
function showStatus(element, message, type) {
    if (element) {
        element.textContent = message;
        element.className = `status-message ${type}`;
    }
}

// 页面切换
function showPage(pageId) {
    console.log('显示页面:', pageId);
    // 隐藏所有页面
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    
    // 显示指定页面
    const targetPage = document.getElementById(pageId);
    if (targetPage) {
        targetPage.classList.add('active');
        console.log('页面已显示:', pageId);
    } else {
        console.error('页面不存在:', pageId);
    }
}

// 显示登录页面
function showLogin() {
    showPage('loginPage');
}

// 显示注册页面
function showRegister() {
    showPage('registerPage');
}

// 切换标签页
function switchTab(tabName) {
    // 更新导航状态
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    
    // 更新页面标题
    const pageTitle = document.getElementById('pageTitle');
    const titles = {
        'dashboard': '仪表板',
        'calls': '通话管理',
        'contacts': '联系人',
        'history': '通话记录',
        'security': '安全检测',
        'settings': '系统设置'
    };
    if (pageTitle && titles[tabName]) {
        pageTitle.textContent = titles[tabName];
    }
    
    // 切换内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`${tabName}Tab`).classList.add('active');
    
    // 加载对应数据
    loadTabData(tabName);
}

// 加载标签页数据
function loadTabData(tabName) {
    switch (tabName) {
        case 'dashboard':
            loadDashboardData();
            break;
        case 'calls':
            loadCallsData();
            break;
        case 'contacts':
            loadContactsData();
            break;
        case 'history':
            loadHistoryData();
            break;
        case 'security':
            loadSecurityData();
            break;
        case 'settings':
            loadSettingsData();
            break;
    }
}

// 加载仪表板数据
async function loadDashboardData() {
    await checkSystemStatus();
}

// 检查系统状态
async function checkSystemStatus() {
    try {
        // 检查后端服务
        const backendResponse = await fetch(`${API_BASE_URL}/health`);
        const backendStatus = backendResponse.ok ? '正常' : '异常';
        document.getElementById('backendStatus').textContent = backendStatus;
        document.getElementById('backendStatus').style.color = backendResponse.ok ? '#28a745' : '#dc3545';
        
        // 检查AI服务
        const aiResponse = await fetch(`${AI_SERVICE_URL}/health`);
        const aiStatus = aiResponse.ok ? '正常' : '异常';
        document.getElementById('aiStatus').textContent = aiStatus;
        document.getElementById('aiStatus').style.color = aiResponse.ok ? '#28a745' : '#dc3545';
        
        // 检查数据库（通过后端API）
        const dbResponse = await fetch(`${API_BASE_URL}/api/health/database`);
        const dbStatus = dbResponse.ok ? '正常' : '异常';
        document.getElementById('dbStatus').textContent = dbStatus;
        document.getElementById('dbStatus').style.color = dbResponse.ok ? '#28a745' : '#dc3545';
        
    } catch (error) {
        console.error('系统状态检查失败:', error);
        document.getElementById('backendStatus').textContent = '异常';
        document.getElementById('backendStatus').style.color = '#dc3545';
        document.getElementById('aiStatus').textContent = '异常';
        document.getElementById('aiStatus').style.color = '#dc3545';
        document.getElementById('dbStatus').textContent = '异常';
        document.getElementById('dbStatus').style.color = '#dc3545';
    }
}

// 检查连接状态
async function checkConnectionStatus() {
    try {
        const response = await fetch(`http://localhost:8000/health`);
        const statusIndicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.connection-status span:last-child');
        
        if (response.ok) {
            statusIndicator.className = 'status-indicator online';
            statusText.textContent = '已连接';
        } else {
            statusIndicator.className = 'status-indicator offline';
            statusText.textContent = '连接断开';
        }
    } catch (error) {
        const statusIndicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.connection-status span:last-child');
        statusIndicator.className = 'status-indicator offline';
        statusText.textContent = '连接断开';
    }
}

// 更新用户信息
function updateUserInfo() {
    if (currentUser) {
        document.getElementById('currentUser').textContent = currentUser.username;
    }
}

// 快速通话
async function quickCall() {
    const target = document.getElementById('callTarget').value.trim();
    if (!target) {
        alert('请输入通话目标');
        return;
    }
    
    await startCall(target);
}

// 发起通话
async function startCall(target) {
    try {
        const token = localStorage.getItem('authToken');
        if (!token) {
            alert('请先登录');
            return;
        }
        
        const response = await fetch(`${API_BASE_URL}/calls/start`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ target })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            currentCall = data.call;
            showPage('callPage');
            initializeCall();
        } else {
            alert(data.message || '发起通话失败');
        }
    } catch (error) {
        console.error('发起通话错误:', error);
        alert('网络错误，请检查连接');
    }
}

// 初始化通话
async function initializeCall() {
    try {
        // 获取媒体流
        localStream = await navigator.mediaDevices.getUserMedia({
            video: true,
            audio: true
        });
        
        // 显示本地视频
        const localVideo = document.getElementById('localVideo');
        if (localVideo) {
            localVideo.srcObject = localStream;
        }
        
        // 开始计时
        startCallTimer();
        
        // 更新通话信息
        document.getElementById('callTitle').textContent = `正在通话 - ${currentCall.target}`;
        document.getElementById('remoteUser').textContent = currentCall.target;
        
    } catch (error) {
        console.error('初始化通话失败:', error);
        alert('无法访问摄像头或麦克风');
        endCall();
    }
}

// 开始通话计时
function startCallTimer() {
    callDuration = 0;
    callTimer = setInterval(() => {
        callDuration++;
        const minutes = Math.floor(callDuration / 60);
        const seconds = callDuration % 60;
        document.getElementById('callDuration').textContent = 
            `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }, 1000);
}

// 结束通话
async function endCall() {
    if (callTimer) {
        clearInterval(callTimer);
        callTimer = null;
    }
    
    if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
        localStream = null;
    }
    
    if (currentCall) {
        try {
            const token = localStorage.getItem('authToken');
            await fetch(`${API_BASE_URL}/calls/end`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ callId: currentCall.id })
            });
        } catch (error) {
            console.error('结束通话错误:', error);
        }
        
        currentCall = null;
    }
    
    showPage('mainPage');
}

// 静音切换
function toggleMute() {
    if (localStream) {
        const audioTrack = localStream.getAudioTracks()[0];
        if (audioTrack) {
            audioTrack.enabled = !audioTrack.enabled;
            const muteBtn = document.getElementById('muteBtn');
            const icon = muteBtn.querySelector('i');
            if (audioTrack.enabled) {
                icon.className = 'fas fa-microphone';
            } else {
                icon.className = 'fas fa-microphone-slash';
            }
        }
    }
}

// 视频切换
function toggleVideo() {
    if (localStream) {
        const videoTrack = localStream.getVideoTracks()[0];
        if (videoTrack) {
            videoTrack.enabled = !videoTrack.enabled;
            const videoBtn = document.getElementById('videoBtn');
            const icon = videoBtn.querySelector('i');
            if (videoTrack.enabled) {
                icon.className = 'fas fa-video';
            } else {
                icon.className = 'fas fa-video-slash';
            }
        }
    }
}

// 全屏切换
function toggleFullscreen() {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
}

// 截图
function takeScreenshot() {
    const canvas = document.createElement('canvas');
    const video = document.getElementById('localVideo');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    
    const ctx = canvas.getContext('2d');
    ctx.drawImage(video, 0, 0);
    
    const link = document.createElement('a');
    link.download = `screenshot_${Date.now()}.png`;
    link.href = canvas.toDataURL();
    link.click();
}

// 运行安全检测
async function runSecurityCheck() {
    const resultsContainer = document.getElementById('securityResults');
    resultsContainer.innerHTML = '<div class="loading">正在运行安全检测...</div>';
    
    try {
        const response = await fetch(`${AI_SERVICE_URL}/api/v1/detect`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                audio_data: "test_audio",
                video_data: "test_video"
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            const riskLevel = data.risk_score > 0.7 ? 'danger' : data.risk_score > 0.3 ? 'warning' : 'safe';
            resultsContainer.innerHTML = `
                <div class="security-result ${riskLevel}">
                    <h4>检测结果</h4>
                    <p>风险评分: ${(data.risk_score * 100).toFixed(1)}%</p>
                    <p>置信度: ${(data.confidence * 100).toFixed(1)}%</p>
                    <p>状态: ${data.is_spoofed ? '检测到伪造' : '正常'}</p>
                </div>
            `;
        } else {
            resultsContainer.innerHTML = '<div class="error">安全检测失败</div>';
        }
    } catch (error) {
        console.error('安全检测错误:', error);
        resultsContainer.innerHTML = '<div class="error">网络错误，请检查连接</div>';
    }
}

// 加载通话数据
async function loadCallsData() {
    const callsList = document.getElementById('callsList');
    callsList.innerHTML = '<div class="loading">加载中...</div>';
    
    try {
        const token = localStorage.getItem('authToken');
        const response = await fetch(`${API_BASE_URL}/calls/history`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            displayCallsList(data.calls || []);
        } else {
            callsList.innerHTML = '<div class="error">加载失败</div>';
        }
    } catch (error) {
        console.error('加载通话数据错误:', error);
        callsList.innerHTML = '<div class="error">网络错误</div>';
    }
}

// 显示通话列表
function displayCallsList(calls) {
    const callsList = document.getElementById('callsList');
    
    if (calls.length === 0) {
        callsList.innerHTML = '<div class="empty">暂无通话记录</div>';
        return;
    }
    
    const html = calls.map(call => `
        <div class="call-item">
            <div class="call-info">
                <h4>${call.target}</h4>
                <p>${new Date(call.start_time).toLocaleString()}</p>
            </div>
            <div class="call-actions">
                <button class="btn btn-primary" onclick="startCall('${call.target}')">
                    <i class="fas fa-phone"></i>
                    重新通话
                </button>
            </div>
        </div>
    `).join('');
    
    callsList.innerHTML = html;
}

// 加载联系人数据
async function loadContactsData() {
    const contactsList = document.getElementById('contactsList');
    contactsList.innerHTML = '<div class="loading">加载中...</div>';
    
    try {
        const token = localStorage.getItem('authToken');
        const response = await fetch(`${API_BASE_URL}/user/contacts`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            displayContactsList(data.contacts || []);
        } else {
            contactsList.innerHTML = '<div class="error">加载失败</div>';
        }
    } catch (error) {
        console.error('加载联系人数据错误:', error);
        contactsList.innerHTML = '<div class="error">网络错误</div>';
    }
}

// 显示联系人列表
function displayContactsList(contacts) {
    const contactsList = document.getElementById('contactsList');
    
    if (contacts.length === 0) {
        contactsList.innerHTML = '<div class="empty">暂无联系人</div>';
        return;
    }
    
    const html = contacts.map(contact => `
        <div class="contact-item">
            <div class="contact-info">
                <h4>${contact.username}</h4>
                <p>${contact.status || '离线'}</p>
            </div>
            <div class="contact-actions">
                <button class="btn btn-primary" onclick="startCall('${contact.username}')">
                    <i class="fas fa-phone"></i>
                    通话
                </button>
            </div>
        </div>
    `).join('');
    
    contactsList.innerHTML = html;
}

// 加载通话记录
async function loadHistoryData() {
    const historyList = document.getElementById('historyList');
    historyList.innerHTML = '<div class="loading">加载中...</div>';
    
    try {
        const token = localStorage.getItem('authToken');
        const response = await fetch(`${API_BASE_URL}/calls/history`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            displayHistoryList(data.history || []);
        } else {
            historyList.innerHTML = '<div class="error">加载失败</div>';
        }
    } catch (error) {
        console.error('加载通话记录错误:', error);
        historyList.innerHTML = '<div class="error">网络错误</div>';
    }
}

// 显示通话记录
function displayHistoryList(history) {
    const historyList = document.getElementById('historyList');
    
    if (history.length === 0) {
        historyList.innerHTML = '<div class="empty">暂无通话记录</div>';
        return;
    }
    
    const html = history.map(record => `
        <div class="history-item">
            <div class="history-info">
                <h4>${record.target}</h4>
                <p>${new Date(record.start_time).toLocaleString()}</p>
                <p>时长: ${record.duration || 0}秒</p>
            </div>
            <div class="history-status ${record.status}">
                ${record.status === 'completed' ? '已完成' : 
                  record.status === 'missed' ? '未接' : '已拒绝'}
            </div>
        </div>
    `).join('');
    
    historyList.innerHTML = html;
}

// 加载设置数据
function loadSettingsData() {
    // 加载设备列表
    loadDeviceOptions();
}

// 加载设备选项
async function loadDeviceOptions() {
    try {
        const devices = await navigator.mediaDevices.enumerateDevices();
        
        const audioInput = document.getElementById('audioInput');
        const audioOutput = document.getElementById('audioOutput');
        const videoInput = document.getElementById('videoInput');
        
        // 音频输入设备
        const audioInputs = devices.filter(device => device.kind === 'audioinput');
        audioInput.innerHTML = '<option value="">选择麦克风</option>' +
            audioInputs.map(device => `<option value="${device.deviceId}">${device.label}</option>`).join('');
        
        // 音频输出设备
        const audioOutputs = devices.filter(device => device.kind === 'audiooutput');
        audioOutput.innerHTML = '<option value="">选择扬声器</option>' +
            audioOutputs.map(device => `<option value="${device.deviceId}">${device.label}</option>`).join('');
        
        // 视频输入设备
        const videoInputs = devices.filter(device => device.kind === 'videoinput');
        videoInput.innerHTML = '<option value="">选择摄像头</option>' +
            videoInputs.map(device => `<option value="${device.deviceId}">${device.label}</option>`).join('');
        
    } catch (error) {
        console.error('加载设备列表失败:', error);
    }
}

// 添加联系人
function addContact() {
    const username = prompt('请输入联系人用户名:');
    if (username) {
        // 这里可以添加联系人的逻辑
        alert('联系人功能开发中...');
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('authToken');
    currentUser = null;
    showPage('loginPage');
}

// 新建通话
function startNewCall() {
    const target = prompt('请输入通话目标用户名:');
    if (target) {
        startCall(target);
    }
}

// 加载安全检测数据
function loadSecurityData() {
    // 安全检测页面初始化
    console.log('安全检测页面已加载');
} 