# 视频会议系统 Kubernetes 部署指南

## 📋 概述

本目录包含了视频会议系统在Kubernetes上的完整部署配置，支持微服务架构的容器化部署。

## 🏗️ 架构概览

### 微服务组件
- **gateway-service** (8080) - API网关服务
- **user-service** (8081) - 用户管理服务
- **meeting-service** (8082) - 会议管理服务
- **signaling-service** (8083) - WebRTC信令服务
- **media-service** (8084) - 媒体处理服务
- **detection-service** (8085) - 检测服务
- **record-service** (8086) - 记录服务
- **notification-service** (8087) - 通知服务
- **ai-detection-service** (8501) - AI检测服务

### 数据存储
- **PostgreSQL** - 主数据库
- **MongoDB** - 日志和记录数据库
- **Redis** - 缓存和会话存储
- **RabbitMQ** - 消息队列

## 📁 目录结构

```
k8s/
├── base/                   # 基础配置
│   ├── namespace.yaml      # 命名空间
│   ├── configmap.yaml      # 配置映射
│   ├── secrets.yaml        # 密钥配置
│   └── ingress.yaml        # 入口配置
├── databases/              # 数据库服务
│   ├── postgres.yaml       # PostgreSQL
│   ├── mongodb.yaml        # MongoDB
│   ├── redis.yaml          # Redis
│   └── rabbitmq.yaml       # RabbitMQ
├── services/               # 微服务
│   ├── gateway-service.yaml
│   ├── user-service.yaml
│   ├── meeting-service.yaml
│   ├── signaling-service.yaml
│   ├── ai-detection-service.yaml
│   └── web-client-service.yaml
├── storage/                # 存储配置
│   ├── persistent-volumes.yaml
│   └── persistent-volume-claims.yaml
├── scripts/                # 部署脚本
│   ├── deploy.sh           # 部署脚本
│   └── build-images.sh     # 镜像构建脚本
└── README.md               # 本文档
```

## 🚀 快速开始

### 前置要求

1. **Kubernetes集群** (v1.20+)
2. **kubectl** 命令行工具
3. **Docker** (用于构建镜像)
4. **NGINX Ingress Controller** (可选，用于外部访问)

### 1. 构建Docker镜像

```bash
# 构建所有微服务镜像
./k8s/scripts/build-images.sh build

# 查看构建的镜像
docker images | grep video-conference
```

### 2. 部署到Kubernetes

```bash
# 一键部署
./k8s/scripts/deploy.sh deploy

# 查看部署状态
./k8s/scripts/deploy.sh status
```

### 3. 访问系统

部署完成后，可以通过以下方式访问：

- **Web界面**: http://video-conference.local
- **API文档**: http://video-conference.local/api/docs
- **管理界面**: http://admin.video-conference.local

> 注意：需要在本地hosts文件中添加域名映射，或使用LoadBalancer类型的Service

## 🔧 详细部署步骤

### 1. 准备环境

```bash
# 检查Kubernetes集群状态
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 安装NGINX Ingress Controller (如果未安装)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml
```

### 2. 创建存储

```bash
# 创建持久化存储卷
kubectl apply -f k8s/storage/persistent-volumes.yaml

# 创建存储声明
kubectl apply -f k8s/storage/persistent-volume-claims.yaml
```

### 3. 创建基础配置

```bash
# 创建命名空间
kubectl apply -f k8s/base/namespace.yaml

# 创建配置映射
kubectl apply -f k8s/base/configmap.yaml

# 创建密钥 (注意：生产环境需要修改默认密码)
kubectl apply -f k8s/base/secrets.yaml
```

### 4. 部署数据库服务

```bash
# 部署所有数据库服务
kubectl apply -f k8s/databases/

# 等待数据库服务就绪
kubectl wait --for=condition=ready pod -l app=postgres -n video-conference --timeout=300s
kubectl wait --for=condition=ready pod -l app=mongodb -n video-conference --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n video-conference --timeout=300s
kubectl wait --for=condition=ready pod -l app=rabbitmq -n video-conference --timeout=300s
```

### 5. 部署微服务

```bash
# 部署所有微服务
kubectl apply -f k8s/services/

# 检查服务状态
kubectl get pods -n video-conference
kubectl get services -n video-conference
```

### 6. 创建Ingress

```bash
# 创建入口规则
kubectl apply -f k8s/base/ingress.yaml

# 检查Ingress状态
kubectl get ingress -n video-conference
```

## 🛠️ 管理操作

### 查看系统状态

```bash
# 查看所有资源状态
./k8s/scripts/deploy.sh status

# 查看特定服务的Pod
kubectl get pods -l app=gateway-service -n video-conference

# 查看服务日志
./k8s/scripts/deploy.sh logs gateway-service
```

### 扩缩容操作

```bash
# 扩展网关服务到3个副本
kubectl scale deployment gateway-service --replicas=3 -n video-conference

# 扩展信令服务到5个副本 (处理更多并发连接)
kubectl scale deployment signaling-service --replicas=5 -n video-conference
```

### 更新服务

```bash
# 重新构建镜像
./k8s/scripts/build-images.sh build

# 重启服务以使用新镜像
./k8s/scripts/deploy.sh restart gateway-service

# 或者使用滚动更新
kubectl set image deployment/gateway-service gateway-service=video-conference/gateway-service:v2.0 -n video-conference
```

### 监控和调试

```bash
# 查看资源使用情况
kubectl top pods -n video-conference
kubectl top nodes

# 查看事件
kubectl get events -n video-conference --sort-by='.lastTimestamp'

# 进入Pod进行调试
kubectl exec -it deployment/gateway-service -n video-conference -- /bin/sh
```

## 🔒 安全配置

### 1. 更新默认密码

编辑 `k8s/base/secrets.yaml` 文件，更新以下密码：

```bash
# 生成新的base64编码密码
echo -n "your-new-password" | base64

# 更新secrets.yaml中的密码
# 然后重新应用配置
kubectl apply -f k8s/base/secrets.yaml
```

### 2. 网络策略

```yaml
# 创建网络策略限制Pod间通信
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: video-conference-network-policy
  namespace: video-conference
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: video-conference
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: video-conference
```

### 3. RBAC配置

```yaml
# 创建服务账户和角色绑定
apiVersion: v1
kind: ServiceAccount
metadata:
  name: video-conference-sa
  namespace: video-conference
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: video-conference-role
  namespace: video-conference
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: video-conference-rolebinding
  namespace: video-conference
subjects:
- kind: ServiceAccount
  name: video-conference-sa
  namespace: video-conference
roleRef:
  kind: Role
  name: video-conference-role
  apiGroup: rbac.authorization.k8s.io
```

## 📊 监控和日志

### Prometheus监控

```yaml
# 添加Prometheus监控注解到服务
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

### 日志收集

```yaml
# 使用Fluentd收集日志
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: fluentd
  template:
    metadata:
      labels:
        name: fluentd
    spec:
      containers:
      - name: fluentd
        image: fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: "elasticsearch.logging.svc.cluster.local"
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
```

## 🚨 故障排除

### 常见问题

1. **Pod无法启动**
   ```bash
   kubectl describe pod <pod-name> -n video-conference
   kubectl logs <pod-name> -n video-conference
   ```

2. **服务无法访问**
   ```bash
   kubectl get endpoints -n video-conference
   kubectl port-forward service/gateway-service 8080:8080 -n video-conference
   ```

3. **存储问题**
   ```bash
   kubectl get pv,pvc -n video-conference
   kubectl describe pvc <pvc-name> -n video-conference
   ```

4. **网络连接问题**
   ```bash
   kubectl exec -it <pod-name> -n video-conference -- nslookup postgres-service
   kubectl exec -it <pod-name> -n video-conference -- telnet postgres-service 5432
   ```

### 清理和重置

```bash
# 完全卸载系统
./k8s/scripts/deploy.sh undeploy

# 清理Docker镜像
./k8s/scripts/build-images.sh clean

# 清理持久化数据 (谨慎操作)
kubectl delete pvc --all -n video-conference
```

## 📝 生产环境建议

1. **高可用性**
   - 使用多副本部署关键服务
   - 配置Pod反亲和性规则
   - 使用多个可用区

2. **资源管理**
   - 设置合适的资源请求和限制
   - 使用HPA进行自动扩缩容
   - 配置资源配额

3. **安全性**
   - 使用非root用户运行容器
   - 启用Pod安全策略
   - 定期更新镜像和依赖

4. **备份策略**
   - 定期备份数据库
   - 备份配置文件
   - 测试恢复流程

## 📞 支持

如有问题，请查看：
- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [项目Issue页面](https://github.com/your-repo/issues)
- [部署日志和监控数据]
