# 音视频通话系统 - Qt前端

这是一个基于Qt6开发的音视频通话系统前端界面，提供用户友好的图形界面来管理视频通话、安全检测和用户管理功能。

## 🚀 快速开始

### 方法一：简化版本（推荐新手）

如果您只是想快速体验界面，可以运行简化版本：

```bash
# 运行简化构建脚本
.\build_simple.bat

# 运行程序
.\bin\VideoCallApp_simple.exe
```

**简化版本特点：**
- ✅ 只需要Qt6环境
- ✅ 包含基本界面框架
- ✅ 登录和主窗口功能
- ❌ 不包含视频通话功能
- ❌ 不包含安全检测功能

### 方法二：完整版本

如果您需要完整功能，请按以下步骤操作：

1. **安装Qt6**
   ```bash
   # 访问 https://www.qt.io/download
   # 下载并安装Qt6.5+ 和 MinGW编译器
   ```

2. **安装OpenCV**
   ```bash
   # 方法1: 使用vcpkg（推荐）
   .\vcpkg install opencv4[contrib]
   
   # 方法2: 下载预编译包
   # 访问 https://opencv.org/releases/
   ```

3. **运行完整构建**
   ```bash
   .\build_qt6.bat
   ```

4. **运行程序**
   ```bash
   .\bin\VideoCallApp.exe
   ```

## 📋 系统要求

### 最低要求
- Windows 10/11
- 4GB RAM
- 2GB 可用磁盘空间
- Qt6.5+ 基础组件

### 推荐配置
- Windows 10/11
- 8GB RAM
- 10GB 可用磁盘空间
- Qt6.5+ 完整组件
- OpenCV 4.x
- 摄像头和麦克风

## 🛠️ 安装指南

### 详细安装步骤

1. **安装Qt6**
   - 访问 [Qt官网](https://www.qt.io/download)
   - 下载Qt在线安装器
   - 选择以下组件：
     - Qt 6.5.x (最新稳定版)
     - MinGW 11.2.0 64-bit
     - Qt Creator
     - Qt Multimedia
     - Qt WebEngine

2. **安装OpenCV**
   - 访问 [OpenCV官网](https://opencv.org/releases/)
   - 下载Windows版本
   - 解压到 `C:\opencv`
   - 将 `C:\opencv\build\x64\vc15\bin` 添加到系统PATH

3. **环境配置**
   - 将Qt的bin目录添加到系统PATH
   - 重启命令提示符
   - 运行环境检查：`.\check_environment.bat`

## 🔧 构建脚本说明

### 主要脚本

| 脚本 | 功能 | 依赖 |
|------|------|------|
| `build_simple.bat` | 构建简化版本 | 仅Qt6 |
| `build_qt6.bat` | 构建完整版本 | Qt6 + OpenCV |
| `run_qt_frontend.bat` | 运行程序 | 已构建的可执行文件 |
| `check_environment.bat` | 检查环境 | 无 |
| `fix_opencv_paths.bat` | 修复OpenCV路径 | OpenCV安装 |
| `quick_setup.bat` | 快速设置向导 | 无 |

### 使用建议

1. **首次使用**：运行 `.\quick_setup.bat`
2. **环境问题**：运行 `.\check_environment.bat`
3. **OpenCV问题**：运行 `.\fix_opencv_paths.bat`
4. **快速体验**：运行 `.\build_simple.bat`

## 📁 项目结构

```
core/frontend/
├── main.cpp                 # 程序入口
├── mainwindow.cpp/h        # 主窗口
├── loginwidget.cpp/h       # 登录界面
├── videocallwidget.cpp/h   # 视频通话界面
├── securitydetectionwidget.cpp/h  # 安全检测界面
├── userprofilewidget.cpp/h # 用户资料界面
├── callhistorywidget.cpp/h # 通话历史界面
├── settingswidget.cpp/h    # 设置界面
├── networkmanager.cpp/h    # 网络管理
├── audiomanager.cpp/h      # 音频管理
├── videomanager.cpp/h      # 视频管理
├── securitymanager.cpp/h   # 安全管理
├── VideoCallApp.pro        # 完整项目文件
├── VideoCallApp_simple.pro # 简化项目文件
├── resources.qrc           # 资源文件
├── build_qt6.bat          # 完整构建脚本
├── build_simple.bat       # 简化构建脚本
├── run_qt_frontend.bat    # 运行脚本
├── check_environment.bat  # 环境检查脚本
├── fix_opencv_paths.bat   # OpenCV路径修复脚本
├── quick_setup.bat        # 快速设置脚本
└── INSTALLATION_GUIDE.md  # 详细安装指南
```

## 🎯 功能特性

### 完整版本功能
- ✅ 用户登录和注册
- ✅ 视频通话界面
- ✅ 音频管理
- ✅ 安全检测（人脸识别）
- ✅ 通话历史记录
- ✅ 用户资料管理
- ✅ 系统设置
- ✅ 网络连接管理

### 简化版本功能
- ✅ 用户登录界面
- ✅ 主窗口框架
- ✅ 基本UI组件
- ❌ 视频通话功能
- ❌ 安全检测功能
- ❌ 网络通信功能

## 🔍 故障排除

### 常见问题

1. **找不到qmake**
   ```
   解决方案：确保Qt6已安装并添加到PATH
   ```

2. **OpenCV链接错误**
   ```
   解决方案：运行 .\fix_opencv_paths.bat
   ```

3. **编译器错误**
   ```
   解决方案：安装MinGW或Visual Studio
   ```

4. **缺少DLL文件**
   ```
   解决方案：将Qt和OpenCV的bin目录添加到PATH
   ```

### 调试步骤

1. 运行环境检查：`.\check_environment.bat`
2. 查看错误信息
3. 参考安装指南：`INSTALLATION_GUIDE.md`
4. 尝试简化版本：`.\build_simple.bat`

## 📚 开发指南

### 编译选项

```bash
# 调试版本
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=debug

# 发布版本
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=release

# 简化版本
qmake VideoCallApp_simple.pro -spec win32-g++ CONFIG+=debug
```

### 代码结构

- **UI层**：基于Qt Widgets的界面组件
- **业务层**：各种管理器类处理具体功能
- **数据层**：与后端API交互的网络管理

### 扩展开发

1. 添加新的界面组件
2. 实现新的管理器类
3. 更新项目文件配置
4. 重新构建项目

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如果您遇到问题或有建议，请：

1. 查看 [故障排除](#故障排除) 部分
2. 运行环境检查脚本
3. 参考详细安装指南
4. 提交 Issue 或联系开发团队

---

**注意：** 此项目需要配合后端服务使用才能实现完整功能。请确保后端服务正在运行。 