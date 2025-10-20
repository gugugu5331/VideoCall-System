# 🔌 API 文档

本目录包含智能视频会议系统的所有 API 接口文档。

## 📖 文档列表

### API 接口文档
- **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** - 完整的 API 接口参考

包含以下内容：
- 认证与授权
- 用户服务 API
- 会议服务 API
- 信令服务 API
- 媒体服务 API
- AI 服务 API
- 数据模型
- 错误码
- 限流规则

## 🚀 快速开始

### 基础 URL
```
http://gateway:8000
```

### 认证方式
所有 API 请求（除公开接口外）需要在请求头中携带 JWT Token：
```
Authorization: Bearer <your_jwt_token>
```

### 响应格式
所有响应均为 JSON 格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 📚 相关文档

- [客户端 API 使用指南](../CLIENT/API_USAGE_GUIDE.md) - 如何在客户端中调用 API
- [客户端-服务器通信设计](../CLIENT/COMMUNICATION_DESIGN.md) - 通信架构设计

## 🔗 相关链接

- [项目主 README](../../README.md)
- [后端系统 README](../README.md)
- [文档中心](../README.md)

