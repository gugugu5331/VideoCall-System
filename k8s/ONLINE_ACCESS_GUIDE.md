# 🌐 视频会议系统在线访问配置指南

## 📋 概述

本指南详细说明如何将视频会议系统配置为可通过互联网访问的在线服务。我们提供了多种部署方式，适用于不同的环境和需求。

## 🚀 快速开始

### 一键部署在线访问

```bash
# 使用Ingress方式部署（推荐）
./k8s/scripts/deploy-online.sh ingress your-domain.com

# 使用LoadBalancer方式部署
./k8s/scripts/deploy-online.sh loadbalancer your-domain.com

# 使用NodePort方式部署
./k8s/scripts/deploy-online.sh nodeport
```

## 🔧 部署方式对比

| 方式 | 适用场景 | 优点 | 缺点 | 成本 |
|------|----------|------|------|------|
| **Ingress** | 生产环境 | 功能完整、SSL终止、域名路由 | 需要Ingress Controller | 中等 |
| **LoadBalancer** | 云环境 | 简单直接、高可用 | 每个服务需要独立LB | 较高 |
| **NodePort** | 开发/测试 | 简单快速 | 端口管理复杂、安全性低 | 最低 |
| **反向代理** | 混合环境 | 灵活配置、高性能 | 配置复杂 | 中等 |

## 📊 详细部署方案

### 方案1: Ingress + Let's Encrypt (推荐)

**适用场景**: 生产环境，需要HTTPS和多域名支持

**部署步骤**:

1. **安装Ingress Controller**
   ```bash
   # 安装Nginx Ingress Controller
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml
   
   # 等待安装完成
   kubectl wait --namespace ingress-nginx \
     --for=condition=ready pod \
     --selector=app.kubernetes.io/component=controller \
     --timeout=300s
   ```

2. **安装cert-manager (自动SSL证书)**
   ```bash
   # 安装cert-manager
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   
   # 等待安装完成
   kubectl wait --namespace cert-manager \
     --for=condition=ready pod \
     --selector=app.kubernetes.io/component=controller \
     --timeout=300s
   ```

3. **部署在线访问**
   ```bash
   ./k8s/scripts/deploy-online.sh ingress your-domain.com
   ```

4. **配置DNS记录**
   ```
   A记录: your-domain.com -> [Ingress外部IP]
   A记录: *.your-domain.com -> [Ingress外部IP]
   ```

**访问地址**:
- 主站: https://your-domain.com
- API: https://api.your-domain.com
- 管理: https://admin.your-domain.com

### 方案2: 云服务商LoadBalancer

**适用场景**: 云环境（AWS/Azure/GCP），需要高可用性

#### AWS EKS部署

```bash
# 安装AWS Load Balancer Controller
kubectl apply -f k8s/cloud/aws-integration.yaml

# 部署LoadBalancer服务
./k8s/scripts/deploy-online.sh loadbalancer your-domain.com
```

**特性**:
- Application Load Balancer (ALB) 支持
- Network Load Balancer (NLB) 支持WebSocket
- CloudFront CDN集成
- Route 53 DNS集成

#### Azure AKS部署

```bash
# 安装Application Gateway Ingress Controller
kubectl apply -f k8s/cloud/azure-integration.yaml

# 部署LoadBalancer服务
./k8s/scripts/deploy-online.sh loadbalancer your-domain.com
```

**特性**:
- Application Gateway集成
- Azure Front Door CDN
- Azure DNS集成

#### Google GKE部署

```bash
# 部署GCP集成配置
kubectl apply -f k8s/cloud/gcp-integration.yaml

# 部署LoadBalancer服务
./k8s/scripts/deploy-online.sh loadbalancer your-domain.com
```

**特性**:
- Google Cloud Load Balancer
- Cloud CDN集成
- Cloud DNS集成

### 方案3: 反向代理

**适用场景**: 需要高度自定义配置或混合云环境

#### Nginx反向代理

```bash
# 部署Nginx代理
./k8s/scripts/deploy-online.sh nginx your-domain.com
```

**特性**:
- 高性能HTTP/HTTPS代理
- WebSocket支持
- SSL终止
- 限流和安全防护

#### Traefik反向代理

```bash
# 部署Traefik代理
./k8s/scripts/deploy-online.sh traefik your-domain.com
```

**特性**:
- 自动服务发现
- 动态配置更新
- 内置监控面板
- 多种负载均衡算法

### 方案4: NodePort (开发/测试)

**适用场景**: 开发环境或快速测试

```bash
# 部署NodePort服务
./k8s/scripts/deploy-online.sh nodeport
```

**访问方式**:
- Web界面: http://[节点IP]:30081
- API接口: http://[节点IP]:30800
- WebSocket: ws://[节点IP]:30083

## 🔒 安全配置

### SSL/TLS证书配置

#### 自动证书 (Let's Encrypt)

```yaml
# 在Ingress中添加注解
annotations:
  cert-manager.io/cluster-issuer: "letsencrypt-prod"
  cert-manager.io/acme-challenge-type: http01
```

#### 手动证书

```bash
# 创建证书Secret
kubectl create secret tls video-conference-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n video-conference
```

### 网络安全策略

```bash
# 部署网络安全策略
kubectl apply -f k8s/security/network-policies.yaml

# 部署Pod安全策略
kubectl apply -f k8s/security/pod-security.yaml
```

### 防火墙配置

**云环境安全组规则**:
```
入站规则:
- HTTP (80): 0.0.0.0/0
- HTTPS (443): 0.0.0.0/0
- WebSocket (8083): 0.0.0.0/0

出站规则:
- 所有流量: 0.0.0.0/0
```

**本地防火墙规则**:
```bash
# Ubuntu/Debian
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8083/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8083/tcp
sudo firewall-cmd --reload
```

## 🌍 域名配置

### DNS记录配置

**主域名记录**:
```
类型    名称                    值
A       your-domain.com        [外部IP]
A       www.your-domain.com    [外部IP]
A       api.your-domain.com    [外部IP]
A       admin.your-domain.com  [外部IP]
```

**通配符记录** (推荐):
```
类型    名称                值
A       your-domain.com    [外部IP]
A       *.your-domain.com  [外部IP]
```

### CDN配置

#### CloudFlare配置

1. 添加域名到CloudFlare
2. 配置DNS记录指向Kubernetes外部IP
3. 启用代理模式
4. 配置SSL/TLS设置为"完全(严格)"

#### AWS CloudFront配置

```bash
# 使用提供的配置模板
kubectl apply -f k8s/cloud/aws-integration.yaml
```

## 📊 监控和维护

### 健康检查

```bash
# 检查服务状态
./k8s/scripts/deploy-online.sh info

# 手动健康检查
curl -k https://your-domain.com/health
curl -k https://api.your-domain.com/api/health
```

### 日志监控

```bash
# 查看Ingress日志
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# 查看应用日志
kubectl logs -n video-conference deployment/gateway-service
```

### 性能监控

```bash
# 查看资源使用情况
kubectl top pods -n video-conference
kubectl top nodes

# 查看服务响应时间
curl -w "@curl-format.txt" -o /dev/null -s https://your-domain.com/
```

## 🚨 故障排除

### 常见问题

1. **无法访问服务**
   ```bash
   # 检查服务状态
   kubectl get pods -n video-conference
   kubectl get services -n video-conference
   kubectl get ingress -n video-conference
   
   # 检查DNS解析
   nslookup your-domain.com
   dig your-domain.com
   ```

2. **SSL证书问题**
   ```bash
   # 检查证书状态
   kubectl get certificates -n video-conference
   kubectl describe certificate video-conference-tls -n video-conference
   
   # 检查cert-manager日志
   kubectl logs -n cert-manager deployment/cert-manager
   ```

3. **WebSocket连接失败**
   ```bash
   # 检查信令服务状态
   kubectl get pods -l app=signaling-service -n video-conference
   kubectl logs deployment/signaling-service -n video-conference
   
   # 测试WebSocket连接
   wscat -c wss://api.your-domain.com/ws
   ```

### 性能优化

1. **启用HTTP/2**
   ```yaml
   # 在Ingress中添加注解
   annotations:
     nginx.ingress.kubernetes.io/http2-push-preload: "true"
   ```

2. **启用Gzip压缩**
   ```yaml
   # 在Ingress中添加注解
   annotations:
     nginx.ingress.kubernetes.io/enable-gzip: "true"
   ```

3. **配置缓存**
   ```yaml
   # 在Ingress中添加注解
   annotations:
     nginx.ingress.kubernetes.io/proxy-cache-valid: "200 302 10m"
   ```

## 📞 技术支持

### 获取帮助

```bash
# 查看部署脚本帮助
./k8s/scripts/deploy-online.sh help

# 查看系统状态
./k8s/scripts/deploy.sh status

# 查看服务日志
./k8s/scripts/deploy.sh logs [service-name]
```

### 联系方式

- 项目文档: [README.md](../README.md)
- 部署指南: [k8s/README.md](README.md)
- 问题反馈: GitHub Issues

---

## 🎉 总结

通过本指南，您可以选择最适合您环境的在线访问方案：

- **生产环境**: 推荐使用 Ingress + Let's Encrypt
- **云环境**: 推荐使用云服务商的LoadBalancer
- **开发测试**: 可以使用NodePort方式
- **高级用户**: 可以使用反向代理方式

所有方案都包含了完整的安全配置和监控支持，确保您的视频会议系统能够安全、稳定地提供在线服务。
