# Qt6 和 OpenCV 安装指南

## 系统要求
- Windows 10/11
- 至少 8GB RAM
- 至少 10GB 可用磁盘空间

## 1. 安装 Qt6

### 方法一：使用 Qt 在线安装器（推荐）

1. 访问 [Qt 官网下载页面](https://www.qt.io/download)
2. 下载 "Qt Online Installer"
3. 运行安装器，选择以下组件：
   - Qt 6.5.x (最新稳定版)
   - MinGW 11.2.0 64-bit
   - Qt Creator
   - Qt Debug Information Files
   - Qt WebEngine
   - Qt Multimedia

### 方法二：使用包管理器

#### 使用 vcpkg（推荐）
```bash
# 安装 vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat

# 安装 Qt6
.\vcpkg install qt6-base qt6-multimedia qt6-webengine

# 安装 OpenCV
.\vcpkg install opencv4
```

#### 使用 Chocolatey
```bash
# 安装 Chocolatey（如果未安装）
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 安装 Qt6
choco install qt6

# 安装 OpenCV
choco install opencv
```

## 2. 安装 OpenCV

### 方法一：使用预编译包
1. 访问 [OpenCV 官网](https://opencv.org/releases/)
2. 下载 Windows 版本
3. 解压到 `C:\opencv`
4. 将 `C:\opencv\build\x64\vc15\bin` 添加到系统 PATH

### 方法二：使用 vcpkg（推荐）
```bash
.\vcpkg install opencv4[contrib]
```

### 方法三：从源码编译
```bash
# 克隆 OpenCV
git clone https://github.com/opencv/opencv.git
git clone https://github.com/opencv/opencv_contrib.git

# 创建构建目录
mkdir opencv_build
cd opencv_build

# 配置 CMake
cmake -DOPENCV_EXTRA_MODULES_PATH=../opencv_contrib/modules -DCMAKE_BUILD_TYPE=Release ../opencv

# 编译
cmake --build . --config Release
```

## 3. 环境配置

### 设置 PATH 环境变量
将以下路径添加到系统 PATH：
```
C:\Qt\6.5.x\mingw_64\bin
C:\Qt\Tools\mingw1120_64\bin
C:\opencv\build\x64\vc15\bin
```

### 验证安装
```bash
# 检查 Qt
qmake -v

# 检查编译器
g++ --version

# 检查 OpenCV
pkg-config --modversion opencv4
```

## 4. 项目构建

### 自动构建
```bash
# 运行构建脚本
.\build_qt6.bat

# 或运行完整构建脚本
.\run_qt_frontend.bat
```

### 手动构建
```bash
# 生成 Makefile
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=debug

# 编译
make -j4

# 运行
.\debug\VideoCallApp.exe
```

## 5. 常见问题解决

### 问题1：找不到 qmake
**解决方案：**
- 确保 Qt 已正确安装
- 检查 PATH 环境变量
- 重启命令提示符

### 问题2：编译器错误
**解决方案：**
- 安装 MinGW 或 Visual Studio
- 确保编译器在 PATH 中
- 检查 Qt 和编译器版本兼容性

### 问题3：OpenCV 链接错误
**解决方案：**
- 确保 OpenCV 库文件路径正确
- 检查库文件版本匹配
- 更新项目文件中的库路径

### 问题4：缺少 DLL 文件
**解决方案：**
- 将 Qt 和 OpenCV 的 bin 目录添加到 PATH
- 复制必要的 DLL 文件到可执行文件目录
- 使用 windeployqt 工具部署 Qt 依赖

## 6. 开发工具推荐

### IDE
- **Qt Creator**（推荐）：Qt 官方 IDE
- **Visual Studio**：强大的 C++ IDE
- **CLion**：JetBrains 的 C++ IDE

### 调试工具
- **Qt Creator 调试器**
- **Visual Studio 调试器**
- **GDB**（MinGW）

## 7. 性能优化

### 编译优化
```bash
# 发布版本构建
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=release
make -j8
```

### 运行时优化
- 启用硬件加速
- 优化视频编解码器
- 使用适当的缓冲区大小

## 8. 部署

### 使用 windeployqt
```bash
# 部署 Qt 依赖
windeployqt bin\VideoCallApp.exe

# 部署 OpenCV 依赖
copy C:\opencv\build\x64\vc15\bin\*.dll bin\
```

### 创建安装包
- 使用 NSIS 创建安装程序
- 使用 Inno Setup 打包
- 使用 Qt Installer Framework

## 9. 测试

### 单元测试
```bash
# 运行测试
.\test_qt_frontend.bat
```

### 集成测试
```bash
# 运行完整系统测试
.\test_system.bat
```

## 10. 维护

### 更新依赖
- 定期更新 Qt 版本
- 更新 OpenCV 版本
- 检查安全补丁

### 备份
- 备份项目源码
- 备份构建配置
- 备份依赖库

---

**注意：** 如果遇到任何问题，请查看项目根目录的 `TROUBLESHOOTING.md` 文件或提交 Issue。 