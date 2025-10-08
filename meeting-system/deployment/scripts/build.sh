#!/bin/bash

# 构建脚本 - 构建整个会议系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 配置变量
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
BUILD_TYPE=${BUILD_TYPE:-Release}
PARALLEL_JOBS=${PARALLEL_JOBS:-$(nproc)}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-""}
IMAGE_TAG=${IMAGE_TAG:-latest}

# 显示构建信息
show_build_info() {
    log_info "=== Meeting System Build Script ==="
    log_info "Project Root: $PROJECT_ROOT"
    log_info "Build Type: $BUILD_TYPE"
    log_info "Parallel Jobs: $PARALLEL_JOBS"
    log_info "Docker Registry: ${DOCKER_REGISTRY:-"(local)"}"
    log_info "Image Tag: $IMAGE_TAG"
    echo
}

# 检查依赖
check_dependencies() {
    log_step "Checking dependencies..."
    
    local missing_deps=()
    
    # 检查Go
    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi
    
    # 检查Docker
    if ! command -v docker >/dev/null 2>&1; then
        missing_deps+=("docker")
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose >/dev/null 2>&1; then
        missing_deps+=("docker-compose")
    fi
    
    # 检查CMake (用于AI节点)
    if ! command -v cmake >/dev/null 2>&1; then
        missing_deps+=("cmake")
    fi
    
    # 检查Make
    if ! command -v make >/dev/null 2>&1; then
        missing_deps+=("make")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again"
        exit 1
    fi
    
    log_info "All dependencies are available"
}

# 构建Go微服务
build_go_services() {
    log_step "Building Go microservices..."
    
    cd "$PROJECT_ROOT/backend"
    
    # 服务列表
    local services=(
        "user-service"
        "meeting-service"
        "signaling-service"
        "media-service"
        "ai-service"
        "notification-service"
    )
    
    for service in "${services[@]}"; do
        if [ -d "$service" ]; then
            log_info "Building $service..."
            cd "$service"
            
            # 下载依赖
            go mod download
            
            # 构建
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
            
            log_info "$service built successfully"
            cd ..
        else
            log_warn "$service directory not found, skipping..."
        fi
    done
    
    log_info "All Go services built successfully"
}

# 构建AI节点 (C++)
build_ai_node() {
    log_step "Building AI Node (C++)..."
    
    cd "$PROJECT_ROOT/ai-node"
    
    # 创建构建目录
    mkdir -p build
    cd build
    
    # CMake配置
    cmake -DCMAKE_BUILD_TYPE=$BUILD_TYPE ..
    
    # 构建
    make -j$PARALLEL_JOBS
    
    log_info "AI Node built successfully"
}

# 构建Docker镜像
build_docker_images() {
    log_step "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # 构建Go服务镜像
    local services=(
        "user-service"
        "meeting-service"
        "signaling-service"
        "media-service"
        "ai-service"
        "notification-service"
    )
    
    for service in "${services[@]}"; do
        if [ -d "backend/$service" ]; then
            log_info "Building Docker image for $service..."
            
            local image_name="meeting-system/$service"
            if [ -n "$DOCKER_REGISTRY" ]; then
                image_name="$DOCKER_REGISTRY/$image_name"
            fi
            
            docker build -t "$image_name:$IMAGE_TAG" -f "backend/$service/Dockerfile" backend/
            
            log_info "Docker image for $service built successfully"
        fi
    done
    
    # 构建AI节点镜像
    if [ -d "ai-node" ]; then
        log_info "Building Docker image for AI Node..."
        
        local image_name="meeting-system/ai-node"
        if [ -n "$DOCKER_REGISTRY" ]; then
            image_name="$DOCKER_REGISTRY/$image_name"
        fi
        
        docker build -t "$image_name:$IMAGE_TAG" ai-node/
        
        log_info "Docker image for AI Node built successfully"
    fi
    
    log_info "All Docker images built successfully"
}

# 构建前端应用
build_frontend() {
    log_step "Building frontend applications..."
    
    # Web前端
    if [ -d "$PROJECT_ROOT/frontend/web" ]; then
        log_info "Building Web frontend..."
        cd "$PROJECT_ROOT/frontend/web"
        
        if [ -f "package.json" ]; then
            npm install
            npm run build
            log_info "Web frontend built successfully"
        else
            log_warn "Web frontend package.json not found, skipping..."
        fi
    fi
    
    # Qt前端
    if [ -d "$PROJECT_ROOT/frontend/qt" ]; then
        log_info "Building Qt frontend..."
        cd "$PROJECT_ROOT/frontend/qt"
        
        if [ -f "CMakeLists.txt" ]; then
            mkdir -p build
            cd build
            cmake -DCMAKE_BUILD_TYPE=$BUILD_TYPE ..
            make -j$PARALLEL_JOBS
            log_info "Qt frontend built successfully"
        else
            log_warn "Qt frontend CMakeLists.txt not found, skipping..."
        fi
    fi
    
    # 管理界面
    if [ -d "$PROJECT_ROOT/admin-web" ]; then
        log_info "Building admin web interface..."
        cd "$PROJECT_ROOT/admin-web"
        
        if [ -f "package.json" ]; then
            npm install
            npm run build
            log_info "Admin web interface built successfully"
        else
            log_warn "Admin web interface package.json not found, skipping..."
        fi
    fi
}

# 运行测试
run_tests() {
    log_step "Running tests..."
    
    # Go服务测试
    cd "$PROJECT_ROOT/backend"
    
    local services=(
        "user-service"
        "meeting-service"
        "signaling-service"
        "media-service"
        "ai-service"
        "notification-service"
    )
    
    for service in "${services[@]}"; do
        if [ -d "$service" ]; then
            log_info "Testing $service..."
            cd "$service"
            
            if ls *_test.go >/dev/null 2>&1; then
                go test -v ./...
                log_info "$service tests passed"
            else
                log_warn "No tests found for $service"
            fi
            
            cd ..
        fi
    done
    
    # AI节点测试
    if [ -d "$PROJECT_ROOT/ai-node/tests" ]; then
        log_info "Testing AI Node..."
        cd "$PROJECT_ROOT/ai-node/build"
        
        if [ -f "test_runner" ]; then
            ./test_runner
            log_info "AI Node tests passed"
        else
            log_warn "AI Node test runner not found"
        fi
    fi
    
    log_info "All tests completed"
}

# 创建部署包
create_deployment_package() {
    log_step "Creating deployment package..."
    
    local package_dir="$PROJECT_ROOT/dist"
    local package_name="meeting-system-$IMAGE_TAG.tar.gz"
    
    # 清理并创建目录
    rm -rf "$package_dir"
    mkdir -p "$package_dir"
    
    # 复制部署文件
    cp -r "$PROJECT_ROOT/deployment" "$package_dir/"
    
    # 复制配置文件
    mkdir -p "$package_dir/config"
    cp -r "$PROJECT_ROOT/backend/config" "$package_dir/"
    
    # 复制前端构建产物
    if [ -d "$PROJECT_ROOT/frontend/web/dist" ]; then
        mkdir -p "$package_dir/web"
        cp -r "$PROJECT_ROOT/frontend/web/dist"/* "$package_dir/web/"
    fi
    
    if [ -d "$PROJECT_ROOT/admin-web/dist" ]; then
        mkdir -p "$package_dir/admin"
        cp -r "$PROJECT_ROOT/admin-web/dist"/* "$package_dir/admin/"
    fi
    
    # 创建版本信息
    cat > "$package_dir/VERSION" << EOF
Meeting System v$IMAGE_TAG
Build Date: $(date -Iseconds)
Build Type: $BUILD_TYPE
Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
EOF
    
    # 创建压缩包
    cd "$PROJECT_ROOT"
    tar -czf "$package_name" -C dist .
    
    log_info "Deployment package created: $package_name"
}

# 清理构建产物
clean() {
    log_step "Cleaning build artifacts..."
    
    # 清理Go构建产物
    cd "$PROJECT_ROOT/backend"
    find . -name "main" -type f -delete
    
    # 清理AI节点构建产物
    rm -rf "$PROJECT_ROOT/ai-node/build"
    
    # 清理前端构建产物
    rm -rf "$PROJECT_ROOT/frontend/web/dist"
    rm -rf "$PROJECT_ROOT/frontend/web/node_modules"
    rm -rf "$PROJECT_ROOT/admin-web/dist"
    rm -rf "$PROJECT_ROOT/admin-web/node_modules"
    rm -rf "$PROJECT_ROOT/frontend/qt/build"
    
    # 清理部署包
    rm -rf "$PROJECT_ROOT/dist"
    rm -f "$PROJECT_ROOT"/meeting-system-*.tar.gz
    
    log_info "Clean completed"
}

# 显示帮助信息
show_help() {
    echo "Usage: $0 [OPTIONS] [TARGETS]"
    echo
    echo "OPTIONS:"
    echo "  -h, --help              Show this help message"
    echo "  -c, --clean             Clean build artifacts"
    echo "  -t, --test              Run tests"
    echo "  --build-type TYPE       Set build type (Debug|Release) [default: Release]"
    echo "  --jobs N                Set parallel jobs [default: $(nproc)]"
    echo "  --registry REGISTRY     Set Docker registry"
    echo "  --tag TAG               Set image tag [default: latest]"
    echo
    echo "TARGETS:"
    echo "  all                     Build everything (default)"
    echo "  go                      Build Go microservices only"
    echo "  ai                      Build AI node only"
    echo "  docker                  Build Docker images only"
    echo "  frontend                Build frontend applications only"
    echo "  package                 Create deployment package"
    echo
    echo "Examples:"
    echo "  $0                      # Build everything"
    echo "  $0 --clean             # Clean build artifacts"
    echo "  $0 go docker           # Build Go services and Docker images"
    echo "  $0 --test              # Build and run tests"
    echo "  $0 --tag v1.0.0 package # Build and create deployment package"
}

# 主函数
main() {
    local targets=()
    local run_tests=false
    local clean_only=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--clean)
                clean_only=true
                shift
                ;;
            -t|--test)
                run_tests=true
                shift
                ;;
            --build-type)
                BUILD_TYPE="$2"
                shift 2
                ;;
            --jobs)
                PARALLEL_JOBS="$2"
                shift 2
                ;;
            --registry)
                DOCKER_REGISTRY="$2"
                shift 2
                ;;
            --tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            *)
                targets+=("$1")
                shift
                ;;
        esac
    done
    
    # 如果只是清理，执行清理并退出
    if [ "$clean_only" = true ]; then
        clean
        exit 0
    fi
    
    # 如果没有指定目标，默认构建所有
    if [ ${#targets[@]} -eq 0 ]; then
        targets=("all")
    fi
    
    show_build_info
    check_dependencies
    
    # 执行构建目标
    for target in "${targets[@]}"; do
        case $target in
            all)
                build_go_services
                build_ai_node
                build_docker_images
                build_frontend
                if [ "$run_tests" = true ]; then
                    run_tests
                fi
                create_deployment_package
                ;;
            go)
                build_go_services
                ;;
            ai)
                build_ai_node
                ;;
            docker)
                build_docker_images
                ;;
            frontend)
                build_frontend
                ;;
            package)
                create_deployment_package
                ;;
            *)
                log_error "Unknown target: $target"
                show_help
                exit 1
                ;;
        esac
    done
    
    log_info "Build completed successfully!"
}

# 执行主函数
main "$@"
