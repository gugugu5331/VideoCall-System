#!/bin/bash

# 视频会议系统在线访问部署脚本
# 使用方法: ./deploy-online.sh [method] [domain]
# 方法: loadbalancer, nodeport, ingress, proxy

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
NAMESPACE="video-conference"
K8S_DIR="$(dirname "$0")/.."
METHOD="${1:-ingress}"
DOMAIN="${2:-your-domain.com}"

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

# 检查前置条件
check_prerequisites() {
    log_info "检查前置条件..."
    
    # 检查kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    # 检查命名空间
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        log_error "命名空间 $NAMESPACE 不存在，请先运行基础部署"
        exit 1
    fi
    
    log_success "前置条件检查通过"
}

# 更新域名配置
update_domain_config() {
    local domain=$1
    log_info "更新域名配置为: $domain"
    
    # 更新Ingress配置中的域名
    if [ -f "$K8S_DIR/base/ingress.yaml" ]; then
        sed -i.bak "s/your-domain\.com/$domain/g" "$K8S_DIR/base/ingress.yaml"
        log_success "Ingress域名配置已更新"
    fi
    
    # 更新代理配置中的域名
    if [ -f "$K8S_DIR/proxy/nginx-proxy.yaml" ]; then
        sed -i.bak "s/your-domain\.com/$domain/g" "$K8S_DIR/proxy/nginx-proxy.yaml"
        log_success "Nginx代理域名配置已更新"
    fi
    
    # 更新云服务商配置中的域名
    for cloud_config in "$K8S_DIR/cloud"/*.yaml; do
        if [ -f "$cloud_config" ]; then
            sed -i.bak "s/your-domain\.com/$domain/g" "$cloud_config"
        fi
    done
    
    log_success "所有域名配置已更新为: $domain"
}

# 部署LoadBalancer方式
deploy_loadbalancer() {
    log_info "部署LoadBalancer服务..."
    
    kubectl apply -f "$K8S_DIR/services/loadbalancer-services.yaml"
    
    log_info "等待LoadBalancer获取外部IP..."
    kubectl wait --for=condition=ready service/video-conference-lb -n $NAMESPACE --timeout=300s
    
    # 获取外部IP
    EXTERNAL_IP=$(kubectl get service video-conference-lb -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -z "$EXTERNAL_IP" ]; then
        EXTERNAL_IP=$(kubectl get service video-conference-lb -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    
    if [ -n "$EXTERNAL_IP" ]; then
        log_success "LoadBalancer部署完成！"
        echo "外部访问地址: http://$EXTERNAL_IP"
        echo "HTTPS访问地址: https://$EXTERNAL_IP"
        echo ""
        echo "请将以下DNS记录添加到您的域名提供商："
        echo "A记录: $DOMAIN -> $EXTERNAL_IP"
        echo "A记录: www.$DOMAIN -> $EXTERNAL_IP"
        echo "A记录: api.$DOMAIN -> $EXTERNAL_IP"
    else
        log_warning "LoadBalancer外部IP尚未分配，请稍后检查"
    fi
}

# 部署NodePort方式
deploy_nodeport() {
    log_info "部署NodePort服务..."
    
    kubectl apply -f "$K8S_DIR/services/nodeport-services.yaml"
    
    # 获取节点IP
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    if [ -z "$NODE_IP" ]; then
        NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    fi
    
    log_success "NodePort部署完成！"
    echo "访问地址："
    echo "Web界面: http://$NODE_IP:30081"
    echo "API接口: http://$NODE_IP:30800"
    echo "WebSocket: ws://$NODE_IP:30083"
    echo ""
    echo "注意：NodePort方式需要确保防火墙允许这些端口的访问"
}

# 部署Ingress方式
deploy_ingress() {
    log_info "部署Ingress方式..."
    
    # 检查Ingress Controller
    if ! kubectl get ingressclass nginx &> /dev/null; then
        log_warning "未检测到Nginx Ingress Controller，正在安装..."
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml
        
        log_info "等待Ingress Controller就绪..."
        kubectl wait --namespace ingress-nginx \
            --for=condition=ready pod \
            --selector=app.kubernetes.io/component=controller \
            --timeout=300s
    fi
    
    # 部署SSL证书
    if command -v cert-manager &> /dev/null; then
        log_info "部署Let's Encrypt证书..."
        kubectl apply -f "$K8S_DIR/base/ssl-certificates.yaml"
    else
        log_warning "cert-manager未安装，将使用自签名证书"
    fi
    
    # 部署Ingress
    kubectl apply -f "$K8S_DIR/base/ingress.yaml"
    
    # 获取Ingress IP
    log_info "等待Ingress获取外部IP..."
    sleep 30
    
    INGRESS_IP=$(kubectl get ingress video-conference-ingress -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -z "$INGRESS_IP" ]; then
        INGRESS_IP=$(kubectl get ingress video-conference-ingress -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    
    log_success "Ingress部署完成！"
    if [ -n "$INGRESS_IP" ]; then
        echo "外部IP: $INGRESS_IP"
        echo "访问地址: https://$DOMAIN"
        echo "API地址: https://api.$DOMAIN"
        echo "管理界面: https://admin.$DOMAIN"
        echo ""
        echo "请将以下DNS记录添加到您的域名提供商："
        echo "A记录: $DOMAIN -> $INGRESS_IP"
        echo "A记录: *.${DOMAIN} -> $INGRESS_IP"
    else
        log_warning "Ingress外部IP尚未分配，请稍后检查"
    fi
}

# 部署代理方式
deploy_proxy() {
    log_info "部署反向代理方式..."
    
    local proxy_type="${3:-nginx}"
    
    if [ "$proxy_type" = "nginx" ]; then
        log_info "部署Nginx反向代理..."
        kubectl apply -f "$K8S_DIR/proxy/nginx-proxy.yaml"
        
        # 等待代理服务就绪
        kubectl wait --for=condition=ready pod -l app=nginx-proxy -n $NAMESPACE --timeout=300s
        
        # 获取代理服务外部IP
        PROXY_IP=$(kubectl get service nginx-proxy-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -z "$PROXY_IP" ]; then
            PROXY_IP=$(kubectl get service nginx-proxy-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
        fi
        
        log_success "Nginx代理部署完成！"
    elif [ "$proxy_type" = "traefik" ]; then
        log_info "部署Traefik反向代理..."
        kubectl apply -f "$K8S_DIR/proxy/traefik-proxy.yaml"
        
        # 等待代理服务就绪
        kubectl wait --for=condition=ready pod -l app=traefik -n $NAMESPACE --timeout=300s
        
        # 获取代理服务外部IP
        PROXY_IP=$(kubectl get service traefik -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -z "$PROXY_IP" ]; then
            PROXY_IP=$(kubectl get service traefik -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
        fi
        
        log_success "Traefik代理部署完成！"
    fi
    
    if [ -n "$PROXY_IP" ]; then
        echo "代理IP: $PROXY_IP"
        echo "访问地址: https://$DOMAIN"
        echo ""
        echo "请将以下DNS记录添加到您的域名提供商："
        echo "A记录: $DOMAIN -> $PROXY_IP"
        echo "A记录: *.${DOMAIN} -> $PROXY_IP"
    else
        log_warning "代理外部IP尚未分配，请稍后检查"
    fi
}

# 部署安全策略
deploy_security() {
    log_info "部署安全策略..."
    
    # 部署网络策略
    kubectl apply -f "$K8S_DIR/security/network-policies.yaml"
    
    # 部署Pod安全策略
    kubectl apply -f "$K8S_DIR/security/pod-security.yaml"
    
    log_success "安全策略部署完成"
}

# 显示访问信息
show_access_info() {
    log_info "系统访问信息："
    
    echo "=== 服务状态 ==="
    kubectl get pods -n $NAMESPACE -o wide
    
    echo ""
    echo "=== 服务端点 ==="
    kubectl get services -n $NAMESPACE
    
    echo ""
    echo "=== Ingress信息 ==="
    kubectl get ingress -n $NAMESPACE
    
    echo ""
    echo "=== 访问地址 ==="
    echo "主站: https://$DOMAIN"
    echo "API: https://api.$DOMAIN"
    echo "管理: https://admin.$DOMAIN"
    
    echo ""
    echo "=== 健康检查 ==="
    echo "curl -k https://$DOMAIN/health"
    echo "curl -k https://api.$DOMAIN/api/health"
}

# 主函数
main() {
    log_info "开始部署视频会议系统在线访问..."
    
    check_prerequisites
    
    if [ "$DOMAIN" != "your-domain.com" ]; then
        update_domain_config "$DOMAIN"
    fi
    
    case "$METHOD" in
        "loadbalancer")
            deploy_loadbalancer
            ;;
        "nodeport")
            deploy_nodeport
            ;;
        "ingress")
            deploy_ingress
            ;;
        "proxy")
            deploy_proxy
            ;;
        "nginx")
            deploy_proxy "nginx"
            ;;
        "traefik")
            deploy_proxy "traefik"
            ;;
        "security")
            deploy_security
            ;;
        "info")
            show_access_info
            ;;
        "help"|"-h"|"--help")
            echo "用法: $0 [method] [domain]"
            echo "方法:"
            echo "  loadbalancer - 使用LoadBalancer服务"
            echo "  nodeport     - 使用NodePort服务"
            echo "  ingress      - 使用Ingress控制器 (推荐)"
            echo "  proxy        - 使用Nginx反向代理"
            echo "  nginx        - 使用Nginx反向代理"
            echo "  traefik      - 使用Traefik反向代理"
            echo "  security     - 部署安全策略"
            echo "  info         - 显示访问信息"
            echo ""
            echo "示例:"
            echo "  $0 ingress example.com"
            echo "  $0 loadbalancer myapp.com"
            echo "  $0 nodeport"
            ;;
        *)
            log_error "未知方法: $METHOD"
            echo "使用 '$0 help' 查看帮助信息"
            exit 1
            ;;
    esac
    
    if [ "$METHOD" != "help" ] && [ "$METHOD" != "info" ]; then
        deploy_security
        show_access_info
    fi
    
    log_success "在线访问部署完成！"
}

main "$@"
