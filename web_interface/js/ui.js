// UI管理类
class UIManager {
    constructor() {
        this.currentPage = 'home';
        this.init();
    }

    // 初始化
    init() {
        this.setupNavigation();
        this.setupEventListeners();
    }

    // 设置导航
    setupNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const page = link.getAttribute('data-page');
                this.navigateToPage(page);
            });
        });
    }

    // 设置事件监听器
    setupEventListeners() {
        // 个人资料表单提交
        const profileForm = document.getElementById('profile-form');
        if (profileForm) {
            profileForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                await this.updateProfile();
            });
        }

        // 历史记录筛选
        const filterType = document.getElementById('filter-type');
        const filterDate = document.getElementById('filter-date');
        
        if (filterType) {
            filterType.addEventListener('change', () => this.loadCallHistory());
        }
        
        if (filterDate) {
            filterDate.addEventListener('change', () => this.loadCallHistory());
        }
    }

    // 页面导航
    navigateToPage(page) {
        // 检查认证状态
        if (page !== 'home' && !auth.isAuthenticated) {
            this.showNotification('请先登录', 'warning');
            showLoginModal();
            return;
        }

        // 隐藏所有页面
        const pages = document.querySelectorAll('.page');
        pages.forEach(p => p.classList.remove('active'));

        // 显示目标页面
        const targetPage = document.getElementById(`${page}-page`);
        if (targetPage) {
            targetPage.classList.add('active');
        }

        // 更新导航状态
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('data-page') === page) {
                link.classList.add('active');
            }
        });

        this.currentPage = page;

        // 页面特定初始化
        switch (page) {
            case 'history':
                this.loadCallHistory();
                break;
            case 'profile':
                this.loadUserProfile();
                break;
        }
    }

    // 加载通话历史
    async loadCallHistory() {
        if (!auth.isAuthenticated) return;

        try {
            const filters = {};
            const filterType = document.getElementById('filter-type');
            const filterDate = document.getElementById('filter-date');

            if (filterType && filterType.value !== 'all') {
                filters.type = filterType.value;
            }

            if (filterDate && filterDate.value) {
                filters.date = filterDate.value;
            }

            const history = await api.getCallHistory(filters);
            this.renderCallHistory(history);
        } catch (error) {
            console.error('加载通话历史失败:', error);
            this.showNotification('加载通话历史失败', 'error');
        }
    }

    // 渲染通话历史
    renderCallHistory(history) {
        const historyContainer = document.getElementById('call-history');
        if (!historyContainer) return;

        if (!history || history.length === 0) {
            historyContainer.innerHTML = '<p class="no-data">暂无通话记录</p>';
            return;
        }

        const historyHTML = history.map(call => `
            <div class="history-item">
                <div class="history-header">
                    <span class="call-type">${this.getCallStatusIcon(call.status)} ${this.getCallStatusText(call.status)}</span>
                    <span class="call-date">${formatDate(call.created_at)}</span>
                </div>
                <div class="history-details">
                    <span class="call-duration">${this.formatDuration(call.duration || 0)}</span>
                    <span class="call-participants">${call.participants || '未知'}</span>
                </div>
                ${call.security_alert ? `<div class="security-alert">⚠️ 安全警告: ${call.security_alert}</div>` : ''}
            </div>
        `).join('');

        historyContainer.innerHTML = historyHTML;
    }

    // 格式化通话时长
    formatDuration(seconds) {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        
        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        } else {
            return `${minutes}:${secs.toString().padStart(2, '0')}`;
        }
    }

    // 获取通话状态图标
    getCallStatusIcon(status) {
        const icons = {
            'completed': '✅',
            'missed': '❌',
            'ongoing': '📞',
            'failed': '⚠️'
        };
        return icons[status] || '❓';
    }

    // 获取通话状态文本
    getCallStatusText(status) {
        const texts = {
            'completed': '已完成',
            'missed': '未接',
            'ongoing': '进行中',
            'failed': '失败'
        };
        return texts[status] || '未知';
    }

    // 加载用户资料
    async loadUserProfile() {
        if (!auth.isAuthenticated) return;

        try {
            const profile = await api.getUserProfile();
            this.fillProfileForm(profile);
        } catch (error) {
            console.error('加载用户资料失败:', error);
            this.showNotification('加载用户资料失败', 'error');
        }
    }

    // 填充资料表单
    fillProfileForm(profile) {
        const username = document.getElementById('username');
        const email = document.getElementById('email');
        const phone = document.getElementById('phone');

        if (username) username.value = profile.username || '';
        if (email) email.value = profile.email || '';
        if (phone) phone.value = profile.phone || '';
    }

    // 更新用户资料
    async updateProfile() {
        if (!auth.isAuthenticated) return;

        try {
            const formData = new FormData(document.getElementById('profile-form'));
            const profileData = {
                username: formData.get('username'),
                email: formData.get('email'),
                phone: formData.get('phone')
            };

            await api.updateUserProfile(profileData);
            this.showNotification('资料更新成功', 'success');
        } catch (error) {
            console.error('更新资料失败:', error);
            this.showNotification('更新资料失败', 'error');
        }
    }

    // 更新通话状态
    updateCallStatus(status) {
        const statusElement = document.getElementById('call-status');
        if (!statusElement) return;

        const statusTexts = {
            'connecting': '正在连接...',
            'connected': '已连接',
            'disconnected': '已断开',
            'failed': '连接失败'
        };

        statusElement.textContent = statusTexts[status] || status;
        statusElement.className = `call-status ${status}`;
    }

    // 显示安全警告
    showSecurityAlert(data) {
        const alertContainer = document.getElementById('security-alert');
        if (!alertContainer) return;

        alertContainer.innerHTML = `
            <div class="alert alert-warning">
                <i class="fas fa-exclamation-triangle"></i>
                <div class="alert-content">
                    <h4>安全警告</h4>
                    <p>检测到可疑活动: ${data.message}</p>
                    <p>风险等级: ${data.risk_level}</p>
                    <button onclick="this.parentElement.parentElement.remove()" class="alert-close">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
        `;
        alertContainer.style.display = 'block';
    }

    // 显示加载动画
    showLoading() {
        const loading = document.getElementById('loading');
        if (loading) {
            loading.style.display = 'flex';
        }
    }

    // 隐藏加载动画
    hideLoading() {
        const loading = document.getElementById('loading');
        if (loading) {
            loading.style.display = 'none';
        }
    }

    // 显示通知
    showNotification(message, type = 'info') {
        const notification = document.getElementById('notification');
        if (!notification) return;

        // 设置通知内容
        notification.textContent = message;
        notification.className = `notification ${type}`;

        // 显示通知
        setTimeout(() => {
            notification.classList.add('show');
        }, 100);

        // 自动隐藏
        setTimeout(() => {
            notification.classList.remove('show');
        }, CONFIG.UI_CONFIG.notificationDuration);
    }

    // 显示确认对话框
    showConfirmDialog(message, onConfirm, onCancel) {
        const dialog = document.createElement('div');
        dialog.className = 'modal active';
        dialog.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>确认</h3>
                </div>
                <div style="padding: 24px;">
                    <p>${message}</p>
                    <div style="display: flex; gap: 12px; margin-top: 24px; justify-content: flex-end;">
                        <button class="auth-btn login-btn" onclick="this.closest('.modal').remove(); ${onCancel ? onCancel() : ''}">取消</button>
                        <button class="auth-btn register-btn" onclick="this.closest('.modal').remove(); ${onConfirm ? onConfirm() : ''}">确认</button>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(dialog);
    }

    // 显示输入对话框
    showInputDialog(title, placeholder, onConfirm, onCancel) {
        const dialog = document.createElement('div');
        dialog.className = 'modal active';
        dialog.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>${title}</h3>
                </div>
                <div style="padding: 24px;">
                    <input type="text" placeholder="${placeholder}" id="input-dialog-value" style="width: 100%; padding: 12px; border: 1px solid var(--border-color); border-radius: var(--border-radius);">
                    <div style="display: flex; gap: 12px; margin-top: 24px; justify-content: flex-end;">
                        <button class="auth-btn login-btn" onclick="this.closest('.modal').remove(); ${onCancel ? onCancel() : ''}">取消</button>
                        <button class="auth-btn register-btn" onclick="this.closest('.modal').remove(); ${onConfirm ? onConfirm() : ''}">确认</button>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(dialog);
        
        // 聚焦输入框
        setTimeout(() => {
            const input = dialog.querySelector('#input-dialog-value');
            if (input) input.focus();
        }, 100);
    }

    // 更新页面标题
    updatePageTitle(title) {
        document.title = `${title} - 智能视频通话系统`;
    }

    // 滚动到顶部
    scrollToTop() {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    }

    // 检查设备类型
    isMobile() {
        return window.innerWidth <= 768;
    }

    // 检查是否支持WebRTC
    isWebRTCSupported() {
        return !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia);
    }

    // 检查是否支持WebSocket
    isWebSocketSupported() {
        return 'WebSocket' in window;
    }
}

// 创建全局UI实例
const UI = new UIManager();

// 全局函数
function navigateToPage(page) {
    if (typeof UI !== 'undefined') {
        UI.navigateToPage(page);
    }
}

function filterHistory() {
    if (typeof UI !== 'undefined') {
        UI.loadCallHistory();
    }
} 