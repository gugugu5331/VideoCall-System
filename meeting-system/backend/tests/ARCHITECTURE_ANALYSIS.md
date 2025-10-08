# 会议系统微服务架构分析

## 1. 服务注册验证

### 当前注册状态
```bash
$ docker exec meeting-etcd etcdctl get /services/ --prefix --keys-only
```

**已注册服务**:
- ✅ `user-service` - 2 个实例 (HTTP:8080 + gRPC:50051)
- ✅ `meeting-service` - 2 个实例 (HTTP:8082 + gRPC:50052)
- ✅ `signaling-service` - 1 个实例 (HTTP:8081)
- ✅ `media-service` - 1 个实例 (HTTP:8083)

**未注册服务**:
- ⚠️ `ai-service` - Python 服务，暂未实现 etcd 注册

### 服务注册示例
<augment_code_snippet path="meeting-system/backend/user-service/main.go" mode="EXCERPT">
````go
httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
    Name:     "user-service",
    Host:     advertiseHost,
    Port:     cfg.Server.Port,
    Protocol: "http",
    Metadata: metadata,
})
````
</augment_code_snippet>

---

## 2. Nginx 网关路由机制

### Upstream 配置

**负载均衡策略**:
- `user_service`: `least_conn` (最少连接)
- `meeting_service`: `least_conn`
- `signaling_service`: `ip_hash` (会话保持，用于 WebSocket)
- `media_service`: `least_conn`
- `ai_service`: `least_conn`

### 路由规则

| 路径 | 目标服务 | 认证要求 | 说明 |
|------|----------|----------|------|
| `/health` | Nginx 自身 | ❌ 无 | 网关健康检查 |
| `/api/v1/csrf-token` | user_service | ❌ 无 | 获取 CSRF token |
| `/api/v1/auth/*` | user_service | ❌ 无 | 注册/登录（需 CSRF） |
| `/api/v1/users` | user_service | ✅ JWT | 用户管理 |
| `/api/v1/meetings` | meeting_service | ✅ JWT | 会议管理 |
| `/api/v1/media/*` | media_service | ✅ JWT | 媒体服务 |
| `/api/v1/ai/*` | ai_service | ✅ JWT | AI 服务 |
| `/ws/signaling` | signaling_service | ✅ JWT | WebSocket 信令 |

### 访问流程
```
客户端 → Nginx (8800) → 内部服务 (8080/8081/8082/8083/8084)
```

---

## 3. 信令服务 (Signaling Service)

### 功能说明
信令服务负责 WebRTC 会议的信令交换，主要功能：

1. **WebSocket 连接管理**
   - 维护客户端 WebSocket 连接
   - 管理会议房间和参与者

2. **信令消息转发**
   - SDP (Session Description Protocol) 交换
   - ICE (Interactive Connectivity Establishment) 候选交换
   - 会议控制消息（加入/离开/静音等）

3. **会议状态同步**
   - 参与者列表同步
   - 媒体流状态同步
   - 房间状态管理

### 与其他服务的交互

**与 media-service 的关系**:
- 信令服务：负责 WebRTC 信令交换（控制平面）
- 媒体服务：负责媒体流处理和录制（数据平面）

**交互流程**:
```
1. 客户端 → 信令服务: 建立 WebSocket 连接
2. 信令服务 → meeting-service: 验证用户权限
3. 客户端 ↔ 信令服务: 交换 SDP/ICE
4. 客户端 ↔ 客户端: 建立 P2P 媒体流
5. 媒体服务: 录制/转码媒体流（可选）
```

### 协议和端口

**协议**: WebSocket (升级自 HTTP)  
**端口**: 8081  
**路径**: `/ws/signaling`  
**参数**: `?meeting_id=<id>&peer_id=<id>`

**Nginx 配置**:
<augment_code_snippet path="meeting-system/nginx/nginx.conf" mode="EXCERPT">
````nginx
location /ws/signaling {
    proxy_pass http://signaling_service;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    # ...
}
````
</augment_code_snippet>

---

## 4. 网关身份验证机制

### 认证流程

#### 4.1 CSRF 保护
**公开接口** (register/login) 需要 CSRF token:

```bash
# 1. 获取 CSRF token
curl http://localhost:8800/api/v1/csrf-token
# 返回: {"csrf_token":"..."}

# 2. 使用 CSRF token 注册/登录
curl -X POST http://localhost:8800/api/v1/auth/register \
  -H "X-CSRF-Token: <token>" \
  -d '{"username":"...","password":"...","email":"..."}'
```

#### 4.2 JWT 认证
**受保护接口** 需要 JWT token:

```bash
# 1. 登录获取 JWT
curl -X POST http://localhost:8800/api/v1/auth/login \
  -H "X-CSRF-Token: <csrf_token>" \
  -d '{"username":"...","password":"..."}'
# 返回: {"token":"eyJhbGci..."}

# 2. 使用 JWT 访问受保护资源
curl http://localhost:8800/api/v1/meetings?page=1&page_size=10 \
  -H "Authorization: Bearer <jwt_token>"
```

### 认证实现位置

**Nginx 层**: 无认证（仅转发）  
**应用层**: JWT 中间件验证

<augment_code_snippet path="meeting-system/backend/shared/middleware/auth.go" mode="EXCERPT">
````go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        // 验证 Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, ...)
        // 将用户信息存储到上下文
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
````
</augment_code_snippet>

### 路由认证要求

**无需认证**:
- `/health` - 健康检查
- `/api/v1/csrf-token` - CSRF token
- `/api/v1/auth/register` - 注册（需 CSRF）
- `/api/v1/auth/login` - 登录（需 CSRF）

**需要 JWT 认证**:
- `/api/v1/users` - 用户管理
- `/api/v1/meetings` - 会议管理
- `/api/v1/media/*` - 媒体服务
- `/api/v1/ai/*` - AI 服务
- `/ws/signaling` - WebSocket 信令

### WebSocket 认证

信令服务的 WebSocket 连接需要 JWT 认证:

<augment_code_snippet path="meeting-system/backend/signaling-service/handlers/websocket_handler.go" mode="EXCERPT">
````go
func (h *WebSocketHandler) authorizeConnection(c *gin.Context, userID, meetingID uint) error {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return fmt.Errorf("%w: missing authorization header", errUnauthorized)
    }
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        return fmt.Errorf("%w: invalid authorization header", errUnauthorized)
    }
    // 验证 JWT token...
}
````
</augment_code_snippet>

---

## 5. 实际验证结果

### 完整认证流程测试

```bash
$ ./test_auth_flow.sh
```

**测试结果**:
1. ✅ CSRF Token: 成功获取
2. ✅ 用户注册: 成功（需 CSRF token）
3. ✅ 用户登录: 成功（返回 JWT token）
4. ✅ 受保护资源: 需要 JWT token
5. ✅ WebSocket: 需要 JWT token

### 认证流程图

```
┌─────────┐
│ 客户端  │
└────┬────┘
     │
     │ 1. GET /api/v1/csrf-token
     ├──────────────────────────────────────────┐
     │                                          │
     │ 2. POST /api/v1/auth/register            │
     │    Header: X-CSRF-Token                  │
     ├──────────────────────────────────────────┤
     │                                          │
     │ 3. POST /api/v1/auth/login               │
     │    Header: X-CSRF-Token                  │
     │    Response: JWT token                   │
     ├──────────────────────────────────────────┤
     │                                          │
     │ 4. GET /api/v1/meetings                  │
     │    Header: Authorization: Bearer <JWT>   │
     ├──────────────────────────────────────────┤
     │                                          │
     │ 5. WS /ws/signaling                      │
     │    Header: Authorization: Bearer <JWT>   │
     └──────────────────────────────────────────┘
                      │
                      ▼
              ┌──────────────┐
              │ Nginx Gateway│
              │  (Port 8800) │
              └──────┬───────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
  ┌──────────┐ ┌──────────┐ ┌──────────┐
  │  User    │ │ Meeting  │ │Signaling │
  │ Service  │ │ Service  │ │ Service  │
  │ (8080)   │ │ (8082)   │ │ (8081)   │
  └──────────┘ └──────────┘ └──────────┘
```

---

## 总结

### 架构特点
1. ✅ **服务注册**: 使用 etcd 实现服务发现
2. ✅ **API 网关**: Nginx 提供统一入口和负载均衡
3. ✅ **认证机制**: CSRF + JWT 双重保护
4. ✅ **WebSocket**: 支持实时信令通信
5. ✅ **微服务隔离**: 每个服务独立部署和扩展

### 安全措施
- CSRF token 保护公开接口
- JWT token 保护受保护资源
- WebSocket 连接需要 JWT 认证
- Nginx 限流和超时配置
- 服务间通过内网通信

---

**文档生成时间**: 2025-10-05 01:05  
**测试脚本**: `test_auth_flow.sh`

