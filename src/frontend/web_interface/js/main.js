// 主应用类
class App {
    constructor() {
        this.isInitialized = false;
        this.init();
    }

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
            
            // 检查后端服务状态（不阻塞初始化）
            this.checkBackendStatus().then(() => {
                console.log('后端服务检查完成');
            }).catch((error) => {
                console.warn('后端服务检查失败:', error);
                UI.showNotification('后端服务连接失败，部分功能可能受限', 'warning');
            });
            
            // 初始化完成
            this.isInitialized = true;
            
            // 隐藏加载动画
            setTimeout(() => {
                UI.hideLoading();
                UI.showNotification('应用加载完成', 'success');
            }, 1500);
            
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

    // 检查浏览器兼容性
    checkBrowserCompatibility() {
        const issues = [];

        // 检查WebRTC支持
        if (!UI.isWebRTCSupported()) {
            issues.push('浏览器不支持WebRTC，视频通话功能可能无法正常工作');
        }

        // 检查WebSocket支持
        if (!UI.isWebSocketSupported()) {
            issues.push('浏览器不支持WebSocket，实时通信功能可能无法正常工作');
        }

        // 检查HTTPS（生产环境）
        if (location.protocol !== 'https:' && location.hostname !== 'localhost') {
            issues.push('建议使用HTTPS协议以确保安全');
        }

        // 显示兼容性问题
        if (issues.length > 0) {
            console.warn('浏览器兼容性问题:', issues);
            UI.showNotification('检测到兼容性问题，某些功能可能受限', 'warning');
        }
    }

    // 检查后端服务状态
    async checkBackendStatus() {
        try {
            // 检查主后端服务（添加超时）
            const backendPromise = api.healthCheck();
            const timeoutPromise = new Promise((_, reject) => 
                setTimeout(() => reject(new Error('请求超时')), 5000)
            );
            
            const backendHealth = await Promise.race([backendPromise, timeoutPromise]);
            console.log('后端服务状态:', backendHealth);

            // 检查AI服务（添加超时）
            try {
                const aiPromise = fetch(`${CONFIG.AI_SERVICE_URL}/health`);
                const aiTimeoutPromise = new Promise((_, reject) => 
                    setTimeout(() => reject(new Error('请求超时')), 3000)
                );
                
                const aiHealth = await Promise.race([aiPromise, aiTimeoutPromise]);
                if (aiHealth.ok) {
                    console.log('AI服务状态: 正常');
                } else {
                    console.warn('AI服务状态: 异常');
                }
            } catch (error) {
                console.warn('AI服务连接失败:', error);
                // 不抛出错误，AI服务是可选的
            }

        } catch (error) {
            console.error('后端服务检查失败:', error);
            // 不抛出错误，允许应用继续运行
            console.warn('后端服务连接失败，但应用将继续运行');
        }
    }

    // 设置事件监听器
    setupEventListeners() {
        // 窗口大小变化
        window.addEventListener('resize', this.handleResize.bind(this));

        // 页面可见性变化
        document.addEventListener('visibilitychange', this.handleVisibilityChange.bind(this));

        // 键盘快捷键
        document.addEventListener('keydown', this.handleKeyboardShortcuts.bind(this));

        // 在线/离线状态
        window.addEventListener('online', this.handleOnline.bind(this));
        window.addEventListener('offline', this.handleOffline.bind(this));

        // 错误处理
        window.addEventListener('error', this.handleGlobalError.bind(this));
        window.addEventListener('unhandledrejection', this.handleUnhandledRejection.bind(this));

        // 页面卸载
        window.addEventListener('beforeunload', this.handleBeforeUnload.bind(this));
    }

    // 处理窗口大小变化
    handleResize() {
        // 更新移动端状态
        if (UI.isMobile()) {
            document.body.classList.add('mobile');
        } else {
            document.body.classList.remove('mobile');
        }
    }

    // 处理页面可见性变化
    handleVisibilityChange() {
        if (document.hidden) {
            // 页面隐藏时的处理
            console.log('页面已隐藏');
        } else {
            // 页面显示时的处理
            console.log('页面已显示');
            
            // 刷新数据
            if (auth.isAuthenticated) {
                this.refreshData();
            }
        }
    }

    // 处理键盘快捷键
    handleKeyboardShortcuts(event) {
        // Ctrl/Cmd + K: 快速开始通话
        if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
            event.preventDefault();
            if (auth.isAuthenticated) {
                navigateToPage('call');
            }
        }

        // Ctrl/Cmd + H: 查看历史记录
        if ((event.ctrlKey || event.metaKey) && event.key === 'h') {
            event.preventDefault();
            if (auth.isAuthenticated) {
                navigateToPage('history');
            }
        }

        // Ctrl/Cmd + P: 个人资料
        if ((event.ctrlKey || event.metaKey) && event.key === 'p') {
            event.preventDefault();
            if (auth.isAuthenticated) {
                navigateToPage('profile');
            }
        }

        // Ctrl/Cmd + L: 登出
        if ((event.ctrlKey || event.metaKey) && event.key === 'l') {
            event.preventDefault();
            if (auth.isAuthenticated) {
                logout();
            }
        }
    }

    // 处理在线状态
    handleOnline() {
        UI.showNotification('网络连接已恢复', 'success');
        this.refreshData();
    }

    // 处理离线状态
    handleOffline() {
        UI.showNotification('网络连接已断开', 'warning');
    }

    // 处理全局错误
    handleGlobalError(event) {
        console.error('全局错误:', event.error);
        UI.showNotification('发生未知错误', 'error');
    }

    // 处理未处理的Promise拒绝
    handleUnhandledRejection(event) {
        console.error('未处理的Promise拒绝:', event.reason);
        UI.showNotification('操作失败，请重试', 'error');
    }

    // 处理页面卸载
    handleBeforeUnload(event) {
        // 如果正在通话中，提示用户
        if (callManager.isInCall) {
            event.preventDefault();
            event.returnValue = '您正在通话中，确定要离开吗？';
            return event.returnValue;
        }
    }

    // 刷新数据
    async refreshData() {
        try {
            // 刷新用户信息
            if (auth.isAuthenticated) {
                const userProfile = await api.getUserProfile();
                auth.updateUserInfo(userProfile);
            }

            // 刷新当前页面数据
            switch (UI.currentPage) {
                case 'history':
                    UI.loadCallHistory();
                    break;
                case 'profile':
                    UI.loadUserProfile();
                    break;
            }
        } catch (error) {
            console.error('数据刷新失败:', error);
        }
    }

    // 获取应用状态
    getAppStatus() {
        return {
            isInitialized: this.isInitialized,
            isAuthenticated: auth.isAuthenticated,
            currentPage: UI.currentPage,
            callStatus: callManager.getCallStatus(),
            userAgent: navigator.userAgent,
            online: navigator.onLine
        };
    }
}

// 创建全局应用实例
const app = new App();

// 全局工具函数
function formatTime(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    } else {
        return `${minutes}:${secs.toString().padStart(2, '0')}`;
    }
}

function formatDate(date) {
    return new Date(date).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// 开发模式下的调试工具
if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
    window.debugApp = {
        getStatus: () => app.getAppStatus(),
        getAuth: () => auth,
        getCallManager: () => callManager,
        getUI: () => UI,
        getAPI: () => api,
        refreshData: () => app.refreshData(),
        testNotification: (message, type) => UI.showNotification(message, type),
        clearStorage: () => {
            localStorage.clear();
            location.reload();
        }
    };
    
    console.log('开发模式已启用，使用 window.debugApp 访问调试工具');
}

// 全局变量
let searchTimeout = null;
let selectedUser = null;

// 用户搜索功能
async function searchUsers(event = null) {
    const searchInput = document.getElementById('user-search-input');
    const searchResults = document.getElementById('search-results');
    
    // 如果是键盘事件且不是回车键，则延迟搜索
    if (event && event.type === 'keyup' && event.key !== 'Enter') {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => searchUsers(), 500);
        return;
    }
    
    const query = searchInput.value.trim();
    
    if (query.length < 2) {
        searchResults.innerHTML = '<div class="no-results">请输入至少2个字符进行搜索</div>';
        return;
    }
    
    try {
        UI.showLoading();
        const response = await api.searchUsers(query, 10);
        
        if (response.users && response.users.length > 0) {
            displaySearchResults(response.users);
        } else {
            searchResults.innerHTML = '<div class="no-results">未找到匹配的用户</div>';
        }
    } catch (error) {
        console.error('搜索用户失败:', error);
        searchResults.innerHTML = '<div class="no-results">搜索失败，请重试</div>';
        UI.showNotification('搜索用户失败: ' + error.message, 'error');
    } finally {
        UI.hideLoading();
    }
}

// 显示搜索结果
function displaySearchResults(users) {
    const searchResults = document.getElementById('search-results');
    
    const resultsHTML = users.map(user => `
        <div class="search-result-item" onclick="selectUser('${user.uuid}', '${user.username}', '${user.full_name || ''}')">
            <div class="user-info-search">
                <div class="user-avatar-search">
                    ${user.avatar_url ? `<img src="${user.avatar_url}" alt="${user.username}">` : user.username.charAt(0).toUpperCase()}
                </div>
                <div class="user-details-search">
                    <div class="user-name-search">${user.username}</div>
                    <div class="user-fullname-search">${user.full_name || '未设置姓名'}</div>
                </div>
            </div>
            <button class="call-user-btn" onclick="callUser('${user.uuid}', '${user.username}', event)">
                <i class="fas fa-phone"></i> 通话
            </button>
        </div>
    `).join('');
    
    searchResults.innerHTML = resultsHTML;
}

// 选择用户
function selectUser(uuid, username, fullName) {
    selectedUser = { uuid, username, fullName };
    
    // 更新UI显示选中的用户
    const searchResults = document.getElementById('search-results');
    const items = searchResults.querySelectorAll('.search-result-item');
    
    items.forEach(item => {
        item.classList.remove('selected');
        if (item.querySelector(`[onclick*="${uuid}"]`)) {
            item.classList.add('selected');
        }
    });
    
    // 更新搜索输入框
    const searchInput = document.getElementById('user-search-input');
    searchInput.value = username;
    
    UI.showNotification(`已选择用户: ${username}`, 'success');
}

// 呼叫用户
async function callUser(uuid, username, event = null) {
    if (event) {
        event.stopPropagation();
    }
    
    if (!auth.isAuthenticated) {
        UI.showNotification('请先登录', 'error');
        return;
    }
    
    try {
        UI.showLoading();
        
        // 选择用户
        selectUser(uuid, username, '');
        
        // 使用通话管理器发起通话
        if (window.callManager) {
            await window.callManager.startCall({ uuid, username });
        } else {
            UI.showNotification('通话管理器未初始化', 'error');
        }
        
    } catch (error) {
        console.error('呼叫用户失败:', error);
        UI.showNotification('呼叫失败: ' + error.message, 'error');
    } finally {
        UI.hideLoading();
    }
}

// 全局开始通话函数（用于按钮点击）
async function startCall() {
    if (!auth.isAuthenticated) {
        UI.showNotification('请先登录', 'error');
        return;
    }
    
    // 检查是否有选中的用户
    if (!selectedUser) {
        UI.showNotification('请先搜索并选择要通话的用户', 'warning');
        return;
    }
    
    try {
        if (window.callManager) {
            await window.callManager.startCall(selectedUser);
        } else {
            UI.showNotification('通话管理器未初始化', 'error');
        }
    } catch (error) {
        console.error('开始通话失败:', error);
        UI.showNotification('开始通话失败: ' + error.message, 'error');
    }
}

// 全局结束通话函数
async function endCall() {
    if (window.callManager) {
        await window.callManager.endCall();
    }
}

// 全局静音切换函数
function toggleMute() {
    if (window.callManager) {
        window.callManager.toggleMute();
    }
}

// 全局视频切换函数
function toggleVideo() {
    if (window.callManager) {
        window.callManager.toggleVideo();
    }
}

// 更新通话状态
function updateCallStatus(status, username = '') {
    const statusText = document.getElementById('call-status')?.querySelector('.status-text');
    const callerName = document.getElementById('caller-name');
    
    if (statusText) {
        switch (status) {
            case 'calling':
                statusText.textContent = `正在呼叫 ${username}...`;
                break;
            case 'connected':
                statusText.textContent = '通话中';
                break;
            case 'ended':
                statusText.textContent = '通话结束';
                break;
            default:
                statusText.textContent = '准备就绪';
        }
    }
    
    if (callerName) {
        callerName.textContent = username || '对方用户';
    }
} 