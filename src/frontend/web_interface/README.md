# 智能视频通话系统 - Web前端

## 概述

这是一个现代化的Web前端界面，为智能视频通话系统提供用户友好的交互体验。前端采用原生JavaScript开发，具备响应式设计，支持桌面和移动设备。

## 功能特性

### 🎯 核心功能
- **用户认证**: 注册、登录、登出功能
- **视频通话**: 基于WebRTC的高清视频通话
- **安全检测**: 实时AI伪造检测和风险提示
- **通话历史**: 完整的通话记录管理
- **个人资料**: 用户信息管理

### 🎨 界面特性
- **现代化设计**: 采用Material Design风格
- **响应式布局**: 完美适配桌面和移动设备
- **实时状态**: 通话状态、安全检测状态实时更新
- **通知系统**: 友好的用户通知和提示
- **键盘快捷键**: 提升操作效率

### 🔧 技术特性
- **WebRTC**: 点对点视频通话
- **WebSocket**: 实时通信
- **本地存储**: 用户数据持久化
- **错误处理**: 完善的错误处理机制
- **浏览器兼容性**: 支持现代浏览器

## 文件结构

```
web_interface/
├── index.html              # 主页面
├── styles/
│   └── main.css           # 主样式文件
├── js/
│   ├── config.js          # 配置文件
│   ├── api.js             # API接口
│   ├── auth.js            # 认证管理
│   ├── call.js            # 通话管理
│   ├── ui.js              # UI管理
│   └── main.js            # 主应用逻辑
├── assets/
│   └── default-avatar.png # 默认头像
└── README.md              # 说明文档
```

## 快速开始

### 1. 启动后端服务

确保后端服务正在运行：
- 主后端服务: `http://localhost:8080`
- AI服务: `http://localhost:8000`

### 2. 打开前端界面

直接在浏览器中打开 `index.html` 文件，或者使用本地服务器：

```bash
# 使用Python启动本地服务器
python -m http.server 8081

# 或使用Node.js
npx http-server -p 8081
```

然后访问 `http://localhost:8081`

### 3. 使用流程

1. **注册/登录**: 点击右上角的注册或登录按钮
2. **开始通话**: 登录后点击"通话"页面，然后点击"开始通话"
3. **安全检测**: 通话过程中会自动进行AI安全检测
4. **查看历史**: 在"历史"页面查看通话记录
5. **管理资料**: 在"个人"页面管理个人信息

## 配置说明

### 修改API端点

在 `js/config.js` 中修改配置：

```javascript
const CONFIG = {
    API_BASE_URL: 'http://localhost:8080',    // 主后端服务地址
    AI_SERVICE_URL: 'http://localhost:8000',  // AI服务地址
    // ... 其他配置
};
```

### 自定义样式

在 `styles/main.css` 中修改CSS变量：

```css
:root {
    --primary-color: #2563eb;    /* 主色调 */
    --success-color: #10b981;    /* 成功色 */
    --danger-color: #ef4444;     /* 危险色 */
    /* ... 其他颜色变量 */
}
```

## 浏览器兼容性

### 支持的浏览器
- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+

### 必需功能
- WebRTC支持
- WebSocket支持
- ES6+支持
- LocalStorage支持

## 开发指南

### 添加新功能

1. **创建新的JavaScript模块**:
   ```javascript
   // js/new-feature.js
   class NewFeature {
       constructor() {
           this.init();
       }
       
       init() {
           // 初始化逻辑
       }
   }
   ```

2. **在HTML中引入**:
   ```html
   <script src="js/new-feature.js"></script>
   ```

3. **在main.js中初始化**:
   ```javascript
   const newFeature = new NewFeature();
   ```

### 调试工具

在开发模式下，可以使用浏览器控制台访问调试工具：

```javascript
// 获取应用状态
window.debugApp.getStatus()

// 测试通知
window.debugApp.testNotification('测试消息', 'success')

// 清除本地存储
window.debugApp.clearStorage()
```

## 部署说明

### 生产环境部署

1. **配置HTTPS**: 确保使用HTTPS协议
2. **修改API地址**: 更新配置文件中的API端点
3. **优化资源**: 压缩CSS和JavaScript文件
4. **设置缓存**: 配置适当的缓存策略

### 静态文件服务器

推荐使用Nginx配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/web_interface;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

## 故障排除

### 常见问题

1. **无法获取摄像头权限**
   - 确保浏览器支持getUserMedia API
   - 检查浏览器权限设置
   - 确保使用HTTPS或localhost

2. **WebRTC连接失败**
   - 检查STUN服务器配置
   - 确保防火墙允许WebRTC流量
   - 检查网络连接

3. **API请求失败**
   - 检查后端服务是否运行
   - 验证API端点配置
   - 检查CORS设置

### 调试步骤

1. 打开浏览器开发者工具
2. 查看Console标签页的错误信息
3. 检查Network标签页的API请求
4. 使用调试工具检查应用状态

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础视频通话功能
- 用户认证系统
- 安全检测集成
- 响应式设计

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。 