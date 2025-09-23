// 认证管理类
class Auth {
    constructor() {
        this.isAuthenticated = false;
        this.currentUser = null;
        this.init();
    }

    // 初始化认证状态
    init() {
        const token = localStorage.getItem(CONFIG.STORAGE_KEYS.AUTH_TOKEN);
        const userInfo = localStorage.getItem(CONFIG.STORAGE_KEYS.USER_INFO);
        
        if (token && userInfo) {
            try {
                this.currentUser = JSON.parse(userInfo);
                this.isAuthenticated = true;
                this.updateUI();
            } catch (error) {
                console.error('用户信息解析失败:', error);
                this.logout();
            }
        }
    }

    // 用户注册
    async register(userData) {
        try {
            UI.showLoading();
            
            // 验证密码
            if (userData.password !== userData.confirm_password) {
                throw new Error('密码不匹配');
            }

            // 验证邮箱格式
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(userData.email)) {
                throw new Error('邮箱格式不正确');
            }

            const response = await api.register({
                username: userData.username,
                email: userData.email,
                password: userData.password
            });

            UI.hideLoading();
            UI.showNotification('注册成功！请登录', 'success');
            closeModal('register-modal');
            
            // 清空注册表单
            document.getElementById('register-form').reset();
            
            return response;
        } catch (error) {
            UI.hideLoading();
            UI.showNotification(error.message || '注册失败', 'error');
            throw error;
        }
    }

    // 用户登录
    async login(credentials) {
        try {
            UI.showLoading();
            
            const response = await api.login({
                username: credentials.username,
                password: credentials.password
            });

            // 保存用户信息
            this.currentUser = response.user;
            this.isAuthenticated = true;
            localStorage.setItem(CONFIG.STORAGE_KEYS.USER_INFO, JSON.stringify(response.user));

            UI.hideLoading();
            UI.showNotification('登录成功！', 'success');
            closeModal('login-modal');
            
            // 更新UI
            this.updateUI();
            
            // 清空登录表单
            document.getElementById('login-form').reset();
            
            return response;
        } catch (error) {
            UI.hideLoading();
            UI.showNotification(error.message || '登录失败', 'error');
            throw error;
        }
    }

    // 用户登出
    async logout() {
        try {
            await api.logout();
            
            this.isAuthenticated = false;
            this.currentUser = null;
            
            // 清除本地存储
            localStorage.removeItem(CONFIG.STORAGE_KEYS.AUTH_TOKEN);
            localStorage.removeItem(CONFIG.STORAGE_KEYS.USER_INFO);
            
            // 更新UI
            this.updateUI();
            
            // 跳转到首页
            navigateToPage('home');
            
            UI.showNotification('已退出登录', 'info');
        } catch (error) {
            console.error('登出失败:', error);
            UI.showNotification('登出失败', 'error');
        }
    }

    // 更新UI显示
    updateUI() {
        const navUser = document.getElementById('nav-user');
        const navAuth = document.getElementById('nav-auth');
        const navMenu = document.getElementById('nav-menu');
        const userName = document.getElementById('user-name');

        if (this.isAuthenticated && this.currentUser) {
            // 显示用户信息
            navUser.style.display = 'flex';
            navAuth.style.display = 'none';
            navMenu.style.display = 'flex';
            
            if (userName) {
                userName.textContent = this.currentUser.username || '用户';
            }
            
            // 更新用户头像
            const userAvatar = document.querySelector('.user-avatar');
            if (userAvatar && this.currentUser.avatar) {
                userAvatar.src = this.currentUser.avatar;
            }
        } else {
            // 显示登录/注册按钮
            navUser.style.display = 'none';
            navAuth.style.display = 'flex';
            navMenu.style.display = 'none';
        }
    }

    // 检查认证状态
    checkAuth() {
        if (!this.isAuthenticated) {
            UI.showNotification('请先登录', 'warning');
            showLoginModal();
            return false;
        }
        return true;
    }

    // 获取当前用户信息
    getCurrentUser() {
        return this.currentUser;
    }

    // 更新用户信息
    updateUserInfo(userInfo) {
        this.currentUser = { ...this.currentUser, ...userInfo };
        localStorage.setItem(CONFIG.STORAGE_KEYS.USER_INFO, JSON.stringify(this.currentUser));
        this.updateUI();
    }
}

// 创建全局认证实例
const auth = new Auth();

// 全局函数
function showLoginModal() {
    const modal = document.getElementById('login-modal');
    modal.classList.add('active');
}

function showRegisterModal() {
    const modal = document.getElementById('register-modal');
    modal.classList.add('active');
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    modal.classList.remove('active');
}

function logout() {
    auth.logout();
}

// 事件监听器
document.addEventListener('DOMContentLoaded', function() {
    // 登录表单提交
    document.getElementById('login-form').addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const formData = new FormData(this);
        const credentials = {
            username: formData.get('username'),
            password: formData.get('password')
        };

        try {
            await auth.login(credentials);
        } catch (error) {
            console.error('登录失败:', error);
        }
    });

    // 注册表单提交
    document.getElementById('register-form').addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const formData = new FormData(this);
        const userData = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password'),
            confirm_password: formData.get('confirm_password')
        };

        try {
            await auth.register(userData);
        } catch (error) {
            console.error('注册失败:', error);
        }
    });

    // 模态框背景点击关闭
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', function(e) {
            if (e.target === this) {
                this.classList.remove('active');
            }
        });
    });

    // ESC键关闭模态框
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            document.querySelectorAll('.modal.active').forEach(modal => {
                modal.classList.remove('active');
            });
        }
    });
}); 