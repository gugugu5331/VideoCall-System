#!/bin/bash

# Video Processing Build Script
# 使用OpenCV和OpenGL构建视频处理应用

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    # 检查CMake
    if ! command -v cmake &> /dev/null; then
        print_error "CMake 未安装，请先安装 CMake"
        exit 1
    fi
    
    # 检查编译器
    if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
        print_error "未找到 C++ 编译器，请安装 g++ 或 clang++"
        exit 1
    fi
    
    # 检查pkg-config
    if ! command -v pkg-config &> /dev/null; then
        print_warning "pkg-config 未安装，可能会影响依赖检测"
    fi
    
    print_success "依赖检查完成"
}

# 安装依赖包
install_dependencies() {
    print_info "安装依赖包..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Ubuntu/Debian
        if command -v apt-get &> /dev/null; then
            sudo apt-get update
            sudo apt-get install -y \
                build-essential \
                cmake \
                pkg-config \
                libopencv-dev \
                libgl1-mesa-dev \
                libglu1-mesa-dev \
                libglfw3-dev \
                libglew-dev \
                libglm-dev \
                libasound2-dev \
                libpulse-dev
        # CentOS/RHEL/Fedora
        elif command -v yum &> /dev/null; then
            sudo yum groupinstall -y "Development Tools"
            sudo yum install -y \
                cmake \
                opencv-devel \
                mesa-libGL-devel \
                mesa-libGLU-devel \
                glfw-devel \
                glew-devel \
                glm-devel \
                alsa-lib-devel \
                pulseaudio-libs-devel
        elif command -v dnf &> /dev/null; then
            sudo dnf groupinstall -y "Development Tools"
            sudo dnf install -y \
                cmake \
                opencv-devel \
                mesa-libGL-devel \
                mesa-libGLU-devel \
                glfw-devel \
                glew-devel \
                glm-devel \
                alsa-lib-devel \
                pulseaudio-libs-devel
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install \
                cmake \
                opencv \
                glfw \
                glew \
                glm
        else
            print_error "请先安装 Homebrew: https://brew.sh/"
            exit 1
        fi
    else
        print_warning "未知操作系统，请手动安装依赖"
    fi
    
    print_success "依赖安装完成"
}

# 创建构建目录
setup_build_dir() {
    print_info "设置构建目录..."
    
    BUILD_DIR="build"
    
    if [ -d "$BUILD_DIR" ]; then
        print_warning "构建目录已存在，清理中..."
        rm -rf "$BUILD_DIR"
    fi
    
    mkdir -p "$BUILD_DIR"
    print_success "构建目录创建完成"
}

# 配置CMake
configure_cmake() {
    print_info "配置 CMake..."
    
    cd "$BUILD_DIR"
    
    CMAKE_ARGS=(
        -DCMAKE_BUILD_TYPE=Release
        -DCMAKE_CXX_STANDARD=17
        -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
    )
    
    # 添加额外的CMake参数
    if [ ! -z "$CMAKE_PREFIX_PATH" ]; then
        CMAKE_ARGS+=(-DCMAKE_PREFIX_PATH="$CMAKE_PREFIX_PATH")
    fi
    
    cmake "${CMAKE_ARGS[@]}" ..
    
    cd ..
    print_success "CMake 配置完成"
}

# 编译项目
build_project() {
    print_info "编译项目..."
    
    cd "$BUILD_DIR"
    
    # 获取CPU核心数
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        CORES=$(nproc)
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        CORES=$(sysctl -n hw.ncpu)
    else
        CORES=4
    fi
    
    print_info "使用 $CORES 个核心进行编译"
    
    make -j"$CORES"
    
    cd ..
    print_success "编译完成"
}

# 运行测试
run_tests() {
    print_info "运行测试..."
    
    cd "$BUILD_DIR"
    
    if [ -f "VideoProcessing" ]; then
        print_info "可执行文件已生成: VideoProcessing"
        
        # 检查是否有摄像头
        if ls /dev/video* 1> /dev/null 2>&1; then
            print_info "检测到摄像头设备"
        else
            print_warning "未检测到摄像头设备，程序可能无法正常运行"
        fi
        
        print_info "运行程序: ./VideoProcessing --help"
        ./VideoProcessing --help
    else
        print_error "可执行文件未生成"
        exit 1
    fi
    
    cd ..
    print_success "测试完成"
}

# 安装程序
install_program() {
    print_info "安装程序..."
    
    cd "$BUILD_DIR"
    sudo make install
    cd ..
    
    print_success "安装完成"
}

# 清理构建文件
clean_build() {
    print_info "清理构建文件..."
    
    if [ -d "build" ]; then
        rm -rf build
        print_success "构建文件清理完成"
    else
        print_info "没有构建文件需要清理"
    fi
}

# 显示帮助信息
show_help() {
    echo "Video Processing Build Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示此帮助信息"
    echo "  -d, --deps          安装依赖包"
    echo "  -c, --clean         清理构建文件"
    echo "  -b, --build         构建项目"
    echo "  -t, --test          运行测试"
    echo "  -i, --install       安装程序"
    echo "  -a, --all           执行完整构建流程"
    echo ""
    echo "示例:"
    echo "  $0 --deps           # 安装依赖"
    echo "  $0 --build          # 构建项目"
    echo "  $0 --all            # 完整流程"
}

# 主函数
main() {
    case "$1" in
        -h|--help)
            show_help
            ;;
        -d|--deps)
            check_dependencies
            install_dependencies
            ;;
        -c|--clean)
            clean_build
            ;;
        -b|--build)
            check_dependencies
            setup_build_dir
            configure_cmake
            build_project
            ;;
        -t|--test)
            run_tests
            ;;
        -i|--install)
            install_program
            ;;
        -a|--all)
            check_dependencies
            install_dependencies
            setup_build_dir
            configure_cmake
            build_project
            run_tests
            ;;
        "")
            # 默认构建
            check_dependencies
            setup_build_dir
            configure_cmake
            build_project
            run_tests
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
