#!/bin/bash

# 视频会议系统 Kubernetes 部署脚本
# 使用方法: ./deploy.sh [command]
# 命令: deploy, undeploy, status, logs, restart

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
NAMESPACE="video-conference"
K8S_DIR="$(dirname "$0")/.."

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查kubectl是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在PATH中"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    log_success "Kubernetes集群连接正常"
}

# 检查必要的镜像
check_images() {
    log_info "检查Docker镜像..."
    
    local images=(
        "video-conference/gateway-service:latest"
        "video-conference/user-service:latest"
        "video-conference/meeting-service:latest"
        "video-conference/signaling-service:latest"
        "video-conference/ai-detection-service:latest"
    )
    
    for image in "${images[@]}"; do
        if ! docker image inspect "$image" &> /dev/null; then
            log_warning "镜像 $image 不存在，需要先构建"
        fi
    done
}

# 部署函数
deploy() {
    log_info "开始部署视频会议系统到Kubernetes..."
    
    check_kubectl
    check_images
    
    # 1. 创建命名空间
    log_info "创建命名空间..."
    kubectl apply -f "$K8S_DIR/base/namespace.yaml"
    
    # 2. 创建存储
    log_info "创建存储卷..."
    kubectl apply -f "$K8S_DIR/storage/persistent-volumes.yaml"
    kubectl apply -f "$K8S_DIR/storage/persistent-volume-claims.yaml"
    
    # 3. 创建配置
    log_info "创建配置和密钥..."
    kubectl apply -f "$K8S_DIR/base/configmap.yaml"
    kubectl apply -f "$K8S_DIR/base/secrets.yaml"
    
    # 4. 部署数据库服务
    log_info "部署数据库服务..."
    kubectl apply -f "$K8S_DIR/databases/"
    
    # 等待数据库服务就绪
    log_info "等待数据库服务启动..."
    kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=mongodb -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=rabbitmq -n $NAMESPACE --timeout=300s
    
    # 5. 部署微服务
    log_info "部署微服务..."
    kubectl apply -f "$K8S_DIR/services/"
    
    # 6. 创建Ingress
    log_info "创建Ingress..."
    kubectl apply -f "$K8S_DIR/base/ingress.yaml"
    
    # 等待服务就绪
    log_info "等待服务启动..."
    sleep 30
    
    log_success "部署完成！"
    show_status
}

# 卸载函数
undeploy() {
    log_info "开始卸载视频会议系统..."
    
    check_kubectl
    
    # 删除Ingress
    kubectl delete -f "$K8S_DIR/base/ingress.yaml" --ignore-not-found=true
    
    # 删除微服务
    kubectl delete -f "$K8S_DIR/services/" --ignore-not-found=true
    
    # 删除数据库服务
    kubectl delete -f "$K8S_DIR/databases/" --ignore-not-found=true
    
    # 删除配置
    kubectl delete -f "$K8S_DIR/base/configmap.yaml" --ignore-not-found=true
    kubectl delete -f "$K8S_DIR/base/secrets.yaml" --ignore-not-found=true
    
    # 删除存储（可选，保留数据）
    read -p "是否删除存储卷？这将删除所有数据 (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete -f "$K8S_DIR/storage/persistent-volume-claims.yaml" --ignore-not-found=true
        kubectl delete -f "$K8S_DIR/storage/persistent-volumes.yaml" --ignore-not-found=true
    fi
    
    # 删除命名空间
    kubectl delete -f "$K8S_DIR/base/namespace.yaml" --ignore-not-found=true
    
    log_success "卸载完成！"
}

# 显示状态
show_status() {
    log_info "系统状态："
    
    echo
    echo "=== 命名空间 ==="
    kubectl get namespaces | grep video-conference || echo "命名空间不存在"
    
    echo
    echo "=== Pod状态 ==="
    kubectl get pods -n $NAMESPACE -o wide
    
    echo
    echo "=== 服务状态 ==="
    kubectl get services -n $NAMESPACE
    
    echo
    echo "=== Ingress状态 ==="
    kubectl get ingress -n $NAMESPACE
    
    echo
    echo "=== 存储状态 ==="
    kubectl get pv,pvc -n $NAMESPACE
}

# 查看日志
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        log_error "请指定服务名称"
        echo "可用服务: gateway-service, user-service, meeting-service, signaling-service, ai-detection-service"
        return 1
    fi
    
    log_info "查看 $service 日志..."
    kubectl logs -f deployment/$service -n $NAMESPACE
}

# 重启服务
restart_service() {
    local service=$1
    if [ -z "$service" ]; then
        log_error "请指定服务名称"
        return 1
    fi
    
    log_info "重启 $service..."
    kubectl rollout restart deployment/$service -n $NAMESPACE
    kubectl rollout status deployment/$service -n $NAMESPACE
    log_success "$service 重启完成"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "deploy")
            deploy
            ;;
        "undeploy")
            undeploy
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs "$2"
            ;;
        "restart")
            restart_service "$2"
            ;;
        "help"|"-h"|"--help")
            echo "用法: $0 [command] [options]"
            echo "命令:"
            echo "  deploy    - 部署系统 (默认)"
            echo "  undeploy  - 卸载系统"
            echo "  status    - 显示系统状态"
            echo "  logs      - 查看服务日志 (需要指定服务名)"
            echo "  restart   - 重启服务 (需要指定服务名)"
            echo "  help      - 显示帮助信息"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看帮助信息"
            exit 1
            ;;
    esac
}

main "$@"
