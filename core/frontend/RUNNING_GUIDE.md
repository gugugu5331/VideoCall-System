# 音视频通话系统 - 运行指南

## 🎯 项目状态

由于当前环境的Qt库配置问题，程序无法直接编译运行，但项目结构和代码已经完整创建。本文档提供了多种运行方案来解决这个问题。

## 🚀 快速运行方案

### 方案一：一键启动（推荐）

运行智能启动脚本，它会自动检测环境并选择合适的启动方式：

```bash
.\start_app.bat
```

这个脚本会：
1. 检查是否有已构建的可执行文件
2. 检测Qt和OpenCV环境
3. 根据环境自动选择构建方式
4. 启动程序

### 方案二：简化版本（新手友好）

如果您只是想快速体验界面，不需要完整功能：

```bash
# 构建简化版本（只需要Qt6）
.\build_simple.bat

# 运行简化版本
.\bin\VideoCallApp_simple.exe
```

**简化版本特点：**
- ✅ 只需要Qt6环境
- ✅ 包含基本界面框架
- ✅ 登录和主窗口功能
- ❌ 不包含视频通话功能
- ❌ 不包含安全检测功能

### 方案三：完整版本（需要完整环境）

如果您需要完整功能，需要先安装所有依赖：

```bash
# 1. 安装Qt6和OpenCV（参考INSTALLATION_GUIDE.md）
# 2. 运行完整构建
.\build_qt6.bat

# 3. 运行完整版本
.\bin\VideoCallApp.exe
```

## 📋 环境要求

### 最低要求（简化版本）
- Windows 10/11
- Qt6.5+ 基础组件
- 4GB RAM
- 2GB 可用磁盘空间

### 完整要求（完整版本）
- Windows 10/11
- Qt6.5+ 完整组件
- OpenCV 4.x
- MinGW 11.2.0 或 MSVC 2019+
- 8GB RAM
- 10GB 可用磁盘空间
- 摄像头和麦克风

## 🛠️ 安装步骤

### 1. 安装Qt6

**方法一：Qt在线安装器（推荐）**
1. 访问 [Qt官网](https://www.qt.io/download)
2. 下载Qt在线安装器
3. 选择以下组件：
   - Qt 6.5.x (最新稳定版)
   - MinGW 11.2.0 64-bit
   - Qt Creator
   - Qt Multimedia
   - Qt WebEngine

**方法二：使用包管理器**
```bash
# 使用vcpkg
.\vcpkg install qt6-base qt6-multimedia qt6-webengine

# 使用Chocolatey
choco install qt6
```

### 2. 安装OpenCV（仅完整版本需要）

**方法一：使用vcpkg（推荐）**
```bash
.\vcpkg install opencv4[contrib]
```

**方法二：下载预编译包**
1. 访问 [OpenCV官网](https://opencv.org/releases/)
2. 下载Windows版本
3. 解压到 `C:\opencv`
4. 将 `C:\opencv\build\x64\vc15\bin` 添加到系统PATH

### 3. 环境配置

将以下路径添加到系统PATH：
```
C:\Qt\6.5.x\mingw_64\bin
C:\Qt\Tools\mingw1120_64\bin
C:\opencv\build\x64\vc15\bin  # 仅完整版本需要
```

## 🔧 构建脚本说明

### 主要脚本

| 脚本 | 功能 | 依赖 | 适用场景 |
|------|------|------|----------|
| `start_app.bat` | 智能启动 | 自动检测 | 推荐使用 |
| `build_simple.bat` | 构建简化版本 | 仅Qt6 | 快速体验 |
| `build_qt6.bat` | 构建完整版本 | Qt6 + OpenCV | 完整功能 |
| `run_qt_frontend.bat` | 运行程序 | 已构建文件 | 直接运行 |
| `check_environment.bat` | 检查环境 | 无 | 诊断问题 |
| `fix_opencv_paths.bat` | 修复OpenCV路径 | OpenCV安装 | 解决路径问题 |
| `quick_setup.bat` | 快速设置向导 | 无 | 首次设置 |

### 使用建议

1. **首次使用**：运行 `.\start_app.bat`
2. **环境问题**：运行 `.\check_environment.bat`
3. **OpenCV问题**：运行 `.\fix_opencv_paths.bat`
4. **快速体验**：运行 `.\build_simple.bat`

## 🔍 故障排除

### 常见问题及解决方案

#### 1. 找不到qmake
```
错误：未找到 qmake
解决方案：
1. 确保Qt6已正确安装
2. 将Qt的bin目录添加到系统PATH
3. 重启命令提示符
```

#### 2. OpenCV链接错误
```
错误：无法链接OpenCV库
解决方案：
1. 运行 .\fix_opencv_paths.bat
2. 确保OpenCV版本与编译器兼容
3. 检查库文件路径是否正确
```

#### 3. 编译器错误
```
错误：找不到编译器
解决方案：
1. 安装MinGW或Visual Studio
2. 确保编译器在PATH中
3. 检查Qt和编译器版本兼容性
```

#### 4. 缺少DLL文件
```
错误：缺少Qt6Core.dll等文件
解决方案：
1. 将Qt和OpenCV的bin目录添加到PATH
2. 复制必要的DLL文件到可执行文件目录
3. 使用windeployqt工具部署Qt依赖
```

### 调试步骤

1. **运行环境检查**
   ```bash
   .\check_environment.bat
   ```

2. **查看详细错误信息**
   - 检查编译输出
   - 查看错误日志
   - 运行调试版本

3. **尝试简化版本**
   ```bash
   .\build_simple.bat
   ```

4. **参考安装指南**
   - 查看 `INSTALLATION_GUIDE.md`
   - 按照步骤重新安装依赖

## 📁 项目文件说明

### 项目文件

| 文件 | 用途 | 依赖 |
|------|------|------|
| `VideoCallApp.pro` | 完整项目文件 | Qt6 + OpenCV |
| `VideoCallApp_simple.pro` | 简化项目文件 | 仅Qt6 |
| `main.cpp` | 程序入口 | Qt6 |
| `mainwindow.cpp/h` | 主窗口 | Qt6 |
| `loginwidget.cpp/h` | 登录界面 | Qt6 |
| `videocallwidget.cpp/h` | 视频通话界面 | Qt6 + OpenCV |
| `securitydetectionwidget.cpp/h` | 安全检测界面 | Qt6 + OpenCV |

### 资源文件

| 文件 | 用途 |
|------|------|
| `resources.qrc` | Qt资源文件 |
| `*.ui` | Qt Designer界面文件 |
| `*.qss` | Qt样式表文件 |

## 🎯 功能对比

### 简化版本 vs 完整版本

| 功能 | 简化版本 | 完整版本 |
|------|----------|----------|
| 用户登录 | ✅ | ✅ |
| 主窗口界面 | ✅ | ✅ |
| 基本UI组件 | ✅ | ✅ |
| 视频通话 | ❌ | ✅ |
| 音频处理 | ❌ | ✅ |
| 安全检测 | ❌ | ✅ |
| 网络通信 | ❌ | ✅ |
| 通话历史 | ❌ | ✅ |
| 用户管理 | ❌ | ✅ |

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

## 🚀 部署指南

### 使用windeployqt部署

```bash
# 部署Qt依赖
windeployqt bin\VideoCallApp.exe

# 部署OpenCV依赖
copy C:\opencv\build\x64\vc15\bin\*.dll bin\
```

### 创建安装包

- 使用NSIS创建安装程序
- 使用Inno Setup打包
- 使用Qt Installer Framework

## 📞 支持

如果您遇到问题或有建议，请：

1. 查看本文档的故障排除部分
2. 运行环境检查脚本
3. 参考详细安装指南
4. 提交Issue或联系开发团队

---

**注意：** 此项目需要配合后端服务使用才能实现完整功能。请确保后端服务正在运行。 