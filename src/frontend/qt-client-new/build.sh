#!/bin/bash

# VideoCall System Qt Client Build Script
# 构建功能完整的Qt客户端应用程序

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 显示横幅
show_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                VideoCall System Qt Client                   ║"
    echo "║                     Build Script v1.0                       ║"
    echo "║                                                              ║"
    echo "║  Features: WebRTC, AI Detection, Video Processing, OpenGL   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 检查依赖
check_dependencies() {
    print_step "检查构建依赖..."
    
    local missing_deps=()
    
    # 检查CMake
    if ! command -v cmake &> /dev/null; then
        missing_deps+=("cmake")
    else
        local cmake_version=$(cmake --version | head -n1 | cut -d' ' -f3)
        print_info "CMake version: $cmake_version"
    fi
    
    # 检查编译器
    if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
        missing_deps+=("g++ or clang++")
    else
        if command -v g++ &> /dev/null; then
            local gcc_version=$(g++ --version | head -n1)
            print_info "GCC: $gcc_version"
        fi
        if command -v clang++ &> /dev/null; then
            local clang_version=$(clang++ --version | head -n1)
            print_info "Clang: $clang_version"
        fi
    fi
    
    # 检查pkg-config
    if ! command -v pkg-config &> /dev/null; then
        print_warning "pkg-config not found - may affect dependency detection"
    fi
    
    # 检查Qt6
    if ! command -v qmake6 &> /dev/null && ! command -v qmake &> /dev/null; then
        print_warning "Qt6 qmake not found in PATH"
    else
        if command -v qmake6 &> /dev/null; then
            local qt_version=$(qmake6 -query QT_VERSION)
            print_info "Qt6 version: $qt_version"
        elif command -v qmake &> /dev/null; then
            local qt_version=$(qmake -query QT_VERSION)
            print_info "Qt version: $qt_version"
        fi
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        return 1
    fi
    
    print_success "All required dependencies found"
    return 0
}

# 安装依赖包
install_dependencies() {
    print_step "安装依赖包..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Ubuntu/Debian
        if command -v apt-get &> /dev/null; then
            print_info "Installing dependencies on Ubuntu/Debian..."
            sudo apt-get update
            sudo apt-get install -y \
                build-essential \
                cmake \
                pkg-config \
                qt6-base-dev \
                qt6-multimedia-dev \
                qt6-webengine-dev \
                qt6-websockets-dev \
                qt6-charts-dev \
                libqt6opengl6-dev \
                libqt6sql6-dev \
                libopencv-dev \
                libgl1-mesa-dev \
                libglu1-mesa-dev \
                libzmq3-dev \
                libprotobuf-dev \
                protobuf-compiler \
                libavcodec-dev \
                libavformat-dev \
                libavutil-dev \
                libswscale-dev \
                libswresample-dev \
                libasound2-dev \
                libpulse-dev
                
        # CentOS/RHEL/Fedora
        elif command -v dnf &> /dev/null; then
            print_info "Installing dependencies on Fedora..."
            sudo dnf groupinstall -y "Development Tools"
            sudo dnf install -y \
                cmake \
                qt6-qtbase-devel \
                qt6-qtmultimedia-devel \
                qt6-qtwebengine-devel \
                qt6-qtwebsockets-devel \
                qt6-qtcharts-devel \
                opencv-devel \
                mesa-libGL-devel \
                mesa-libGLU-devel \
                zeromq-devel \
                protobuf-devel \
                ffmpeg-devel \
                alsa-lib-devel \
                pulseaudio-libs-devel
                
        elif command -v yum &> /dev/null; then
            print_info "Installing dependencies on CentOS/RHEL..."
            sudo yum groupinstall -y "Development Tools"
            sudo yum install -y \
                cmake \
                opencv-devel \
                mesa-libGL-devel \
                mesa-libGLU-devel \
                zeromq-devel \
                protobuf-devel \
                ffmpeg-devel \
                alsa-lib-devel \
                pulseaudio-libs-devel
        fi
        
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            print_info "Installing dependencies on macOS..."
            brew install \
                cmake \
                qt6 \
                opencv \
                zeromq \
                protobuf \
                ffmpeg
        else
            print_error "Homebrew not found. Please install Homebrew first: https://brew.sh/"
            return 1
        fi
        
    else
        print_warning "Unknown operating system. Please install dependencies manually."
    fi
    
    print_success "Dependencies installation completed"
}

# 设置构建环境
setup_build_environment() {
    print_step "设置构建环境..."
    
    # 设置构建类型
    BUILD_TYPE=${BUILD_TYPE:-Release}
    BUILD_DIR="build-${BUILD_TYPE,,}"
    
    # 清理旧的构建目录
    if [ -d "$BUILD_DIR" ]; then
        print_info "Cleaning existing build directory..."
        rm -rf "$BUILD_DIR"
    fi
    
    # 创建构建目录
    mkdir -p "$BUILD_DIR"
    
    # 设置环境变量
    export CMAKE_BUILD_TYPE="$BUILD_TYPE"
    export CMAKE_EXPORT_COMPILE_COMMANDS=ON
    
    print_info "Build type: $BUILD_TYPE"
    print_info "Build directory: $BUILD_DIR"
    
    print_success "Build environment setup completed"
}

# 配置CMake
configure_cmake() {
    print_step "配置CMake..."
    
    cd "$BUILD_DIR"
    
    # CMake配置参数
    CMAKE_ARGS=(
        -DCMAKE_BUILD_TYPE="$BUILD_TYPE"
        -DCMAKE_CXX_STANDARD=17
        -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
        -DCMAKE_INSTALL_PREFIX="../install"
    )
    
    # 添加Qt6路径（如果需要）
    if [ ! -z "$QT6_DIR" ]; then
        CMAKE_ARGS+=(-DCMAKE_PREFIX_PATH="$QT6_DIR")
    fi
    
    # 添加OpenCV路径（如果需要）
    if [ ! -z "$OpenCV_DIR" ]; then
        CMAKE_ARGS+=(-DOpenCV_DIR="$OpenCV_DIR")
    fi
    
    # 添加其他自定义路径
    if [ ! -z "$CMAKE_PREFIX_PATH" ]; then
        CMAKE_ARGS+=(-DCMAKE_PREFIX_PATH="$CMAKE_PREFIX_PATH")
    fi
    
    print_info "CMake arguments: ${CMAKE_ARGS[*]}"
    
    # 运行CMake配置
    cmake "${CMAKE_ARGS[@]}" ..
    
    cd ..
    print_success "CMake configuration completed"
}

# 编译项目
build_project() {
    print_step "编译项目..."
    
    cd "$BUILD_DIR"
    
    # 获取CPU核心数
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        CORES=$(nproc)
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        CORES=$(sysctl -n hw.ncpu)
    else
        CORES=4
    fi
    
    print_info "Using $CORES cores for compilation"
    
    # 编译
    cmake --build . --config "$BUILD_TYPE" --parallel "$CORES"
    
    cd ..
    print_success "Project compilation completed"
}

# 运行测试
run_tests() {
    print_step "运行测试..."
    
    cd "$BUILD_DIR"
    
    if [ -f "VideoCallSystemClient" ] || [ -f "VideoCallSystemClient.exe" ]; then
        print_info "Executable found, running basic tests..."
        
        # 运行帮助命令测试
        if ./VideoCallSystemClient --help > /dev/null 2>&1; then
            print_success "Help command test passed"
        else
            print_warning "Help command test failed"
        fi
        
        # 运行版本命令测试
        if ./VideoCallSystemClient --version > /dev/null 2>&1; then
            print_success "Version command test passed"
        else
            print_warning "Version command test failed"
        fi
        
    else
        print_error "Executable not found"
        return 1
    fi
    
    cd ..
    print_success "Tests completed"
}

# 安装项目
install_project() {
    print_step "安装项目..."
    
    cd "$BUILD_DIR"
    cmake --install . --config "$BUILD_TYPE"
    cd ..
    
    print_success "Project installation completed"
}

# 打包项目
package_project() {
    print_step "打包项目..."
    
    cd "$BUILD_DIR"
    
    # 创建打包目录
    PACKAGE_DIR="VideoCallSystemClient-$(date +%Y%m%d)"
    mkdir -p "$PACKAGE_DIR"
    
    # 复制可执行文件
    if [ -f "VideoCallSystemClient" ]; then
        cp VideoCallSystemClient "$PACKAGE_DIR/"
    elif [ -f "VideoCallSystemClient.exe" ]; then
        cp VideoCallSystemClient.exe "$PACKAGE_DIR/"
    fi
    
    # 复制资源文件
    if [ -d "../shaders" ]; then
        cp -r ../shaders "$PACKAGE_DIR/"
    fi
    
    if [ -d "../assets" ]; then
        cp -r ../assets "$PACKAGE_DIR/"
    fi
    
    if [ -d "../config" ]; then
        cp -r ../config "$PACKAGE_DIR/"
    fi
    
    # 创建压缩包
    tar -czf "${PACKAGE_DIR}.tar.gz" "$PACKAGE_DIR"
    
    print_info "Package created: ${PACKAGE_DIR}.tar.gz"
    
    cd ..
    print_success "Project packaging completed"
}

# 清理构建文件
clean_build() {
    print_step "清理构建文件..."
    
    if [ -d "build-debug" ]; then
        rm -rf build-debug
        print_info "Removed build-debug directory"
    fi
    
    if [ -d "build-release" ]; then
        rm -rf build-release
        print_info "Removed build-release directory"
    fi
    
    if [ -d "install" ]; then
        rm -rf install
        print_info "Removed install directory"
    fi
    
    print_success "Build cleanup completed"
}

# 显示帮助信息
show_help() {
    echo "VideoCall System Qt Client Build Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -d, --deps          Install dependencies"
    echo "  -c, --clean         Clean build files"
    echo "  -b, --build         Build project"
    echo "  -t, --test          Run tests"
    echo "  -i, --install       Install project"
    echo "  -p, --package       Package project"
    echo "  -a, --all           Full build process (deps + build + test)"
    echo "  --debug             Build in debug mode"
    echo "  --release           Build in release mode (default)"
    echo ""
    echo "Environment Variables:"
    echo "  QT6_DIR             Qt6 installation directory"
    echo "  OpenCV_DIR          OpenCV installation directory"
    echo "  CMAKE_PREFIX_PATH   Additional CMake prefix paths"
    echo "  BUILD_TYPE          Build type (Debug/Release)"
    echo ""
    echo "Examples:"
    echo "  $0 --deps           # Install dependencies"
    echo "  $0 --build          # Build project"
    echo "  $0 --all            # Full build process"
    echo "  BUILD_TYPE=Debug $0 --build  # Debug build"
}

# 主函数
main() {
    show_banner
    
    case "$1" in
        -h|--help)
            show_help
            ;;
        -d|--deps)
            check_dependencies || install_dependencies
            ;;
        -c|--clean)
            clean_build
            ;;
        -b|--build)
            check_dependencies
            setup_build_environment
            configure_cmake
            build_project
            ;;
        -t|--test)
            run_tests
            ;;
        -i|--install)
            install_project
            ;;
        -p|--package)
            package_project
            ;;
        -a|--all)
            check_dependencies || install_dependencies
            setup_build_environment
            configure_cmake
            build_project
            run_tests
            ;;
        --debug)
            export BUILD_TYPE=Debug
            check_dependencies
            setup_build_environment
            configure_cmake
            build_project
            ;;
        --release)
            export BUILD_TYPE=Release
            check_dependencies
            setup_build_environment
            configure_cmake
            build_project
            ;;
        "")
            # 默认构建
            check_dependencies
            setup_build_environment
            configure_cmake
            build_project
            run_tests
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
