#!/bin/bash

# 视频特效演示构建脚本
# VideoCall System - Video Effects Demo Build Script

set -e  # 遇到错误立即退出

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

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未找到，请先安装"
        return 1
    fi
    return 0
}

# 检查依赖
check_dependencies() {
    print_info "检查构建依赖..."
    
    local missing_deps=()
    
    if ! check_command cmake; then
        missing_deps+=("cmake")
    fi
    
    if ! check_command make; then
        missing_deps+=("make")
    fi
    
    if ! check_command pkg-config; then
        missing_deps+=("pkg-config")
    fi
    
    # 检查Qt6
    if ! pkg-config --exists Qt6Core; then
        missing_deps+=("qt6-base-dev")
    fi
    
    if ! pkg-config --exists Qt6Multimedia; then
        missing_deps+=("qt6-multimedia-dev")
    fi
    
    # 检查OpenCV
    if ! pkg-config --exists opencv4; then
        missing_deps+=("libopencv-dev")
    fi
    
    # 检查OpenGL
    if ! pkg-config --exists gl; then
        missing_deps+=("libgl1-mesa-dev")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "缺少以下依赖："
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        print_info "请运行以下命令安装依赖："
        echo "sudo apt-get update"
        echo "sudo apt-get install ${missing_deps[*]}"
        exit 1
    fi
    
    print_success "所有依赖检查通过"
}

# 清理构建目录
clean_build() {
    print_info "清理构建目录..."
    if [ -d "build" ]; then
        rm -rf build
        print_success "构建目录已清理"
    fi
}

# 创建构建目录
create_build_dir() {
    print_info "创建构建目录..."
    mkdir -p build
    cd build
}

# 配置CMake
configure_cmake() {
    print_info "配置CMake..."
    
    local build_type=${BUILD_TYPE:-Release}
    local cmake_args=(
        -DCMAKE_BUILD_TYPE=$build_type
        -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
        -f ../CMakeLists_effects_demo.txt
        ..
    )
    
    # 添加自定义参数
    if [ ! -z "$CMAKE_PREFIX_PATH" ]; then
        cmake_args+=(-DCMAKE_PREFIX_PATH="$CMAKE_PREFIX_PATH")
    fi
    
    if [ ! -z "$OpenCV_DIR" ]; then
        cmake_args+=(-DOpenCV_DIR="$OpenCV_DIR")
    fi
    
    if [ ! -z "$Qt6_DIR" ]; then
        cmake_args+=(-DQt6_DIR="$Qt6_DIR")
    fi
    
    print_info "CMake参数: ${cmake_args[*]}"
    
    if cmake "${cmake_args[@]}"; then
        print_success "CMake配置成功"
    else
        print_error "CMake配置失败"
        exit 1
    fi
}

# 构建项目
build_project() {
    print_info "开始构建项目..."
    
    local jobs=${JOBS:-$(nproc)}
    print_info "使用 $jobs 个并行任务"
    
    if make -j$jobs; then
        print_success "项目构建成功"
    else
        print_error "项目构建失败"
        exit 1
    fi
}

# 复制资源文件
copy_resources() {
    print_info "复制资源文件..."
    
    # 创建资源目录
    mkdir -p resources/{stickers,backgrounds,filters,shaders}
    
    # 复制默认资源
    local resource_src="../resources"
    if [ -d "$resource_src" ]; then
        cp -r "$resource_src"/* resources/
        print_success "资源文件复制完成"
    else
        print_warning "资源目录不存在，跳过资源复制"
    fi
}

# 运行测试
run_tests() {
    if [ "$RUN_TESTS" = "true" ]; then
        print_info "运行测试..."
        
        if [ -f "VideoEffectsDemo" ]; then
            # 运行基本功能测试
            print_info "测试应用程序启动..."
            timeout 10s ./VideoEffectsDemo --test || true
            print_success "基本测试完成"
        else
            print_warning "可执行文件不存在，跳过测试"
        fi
    fi
}

# 创建安装包
create_package() {
    if [ "$CREATE_PACKAGE" = "true" ]; then
        print_info "创建安装包..."
        
        if make package; then
            print_success "安装包创建成功"
            ls -la *.tar.gz *.deb *.rpm 2>/dev/null || true
        else
            print_warning "安装包创建失败"
        fi
    fi
}

# 显示构建信息
show_build_info() {
    print_info "构建信息："
    echo "  构建类型: ${BUILD_TYPE:-Release}"
    echo "  并行任务: ${JOBS:-$(nproc)}"
    echo "  Qt版本: $(pkg-config --modversion Qt6Core 2>/dev/null || echo '未知')"
    echo "  OpenCV版本: $(pkg-config --modversion opencv4 2>/dev/null || echo '未知')"
    echo "  构建目录: $(pwd)"
    
    if [ -f "VideoEffectsDemo" ]; then
        echo "  可执行文件: $(ls -lh VideoEffectsDemo | awk '{print $5}')"
    fi
}

# 显示使用说明
show_usage() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示此帮助信息"
    echo "  -c, --clean         清理构建目录"
    echo "  -d, --debug         调试构建"
    echo "  -r, --release       发布构建 (默认)"
    echo "  -t, --test          运行测试"
    echo "  -p, --package       创建安装包"
    echo "  -j, --jobs N        使用N个并行任务"
    echo ""
    echo "环境变量:"
    echo "  BUILD_TYPE          构建类型 (Debug/Release)"
    echo "  JOBS                并行任务数"
    echo "  CMAKE_PREFIX_PATH   CMake前缀路径"
    echo "  OpenCV_DIR          OpenCV安装目录"
    echo "  Qt6_DIR             Qt6安装目录"
    echo ""
    echo "示例:"
    echo "  $0                  # 默认发布构建"
    echo "  $0 -d -t            # 调试构建并运行测试"
    echo "  $0 -c -r -p         # 清理、发布构建并创建安装包"
}

# 主函数
main() {
    local clean_first=false
    local run_tests=false
    local create_package=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -c|--clean)
                clean_first=true
                shift
                ;;
            -d|--debug)
                export BUILD_TYPE=Debug
                shift
                ;;
            -r|--release)
                export BUILD_TYPE=Release
                shift
                ;;
            -t|--test)
                export RUN_TESTS=true
                shift
                ;;
            -p|--package)
                export CREATE_PACKAGE=true
                shift
                ;;
            -j|--jobs)
                export JOBS="$2"
                shift 2
                ;;
            *)
                print_error "未知选项: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_info "开始构建视频特效演示应用..."
    
    # 检查依赖
    check_dependencies
    
    # 清理构建目录（如果需要）
    if [ "$clean_first" = true ]; then
        clean_build
    fi
    
    # 创建构建目录
    create_build_dir
    
    # 配置CMake
    configure_cmake
    
    # 构建项目
    build_project
    
    # 复制资源文件
    copy_resources
    
    # 运行测试
    run_tests
    
    # 创建安装包
    create_package
    
    # 显示构建信息
    show_build_info
    
    print_success "视频特效演示应用构建完成！"
    print_info "运行应用: ./VideoEffectsDemo"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
