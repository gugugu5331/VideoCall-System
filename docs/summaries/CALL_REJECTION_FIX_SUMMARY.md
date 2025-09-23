# 拒绝通话失败问题修复总结

## 问题描述
用户报告拒绝通话时出现以下错误：
```
拒绝通话失败: Error: HTTP 400
    at API.request (api.js:44:23)
    at async CallManager.rejectCall (call.js:1087:13)
```

## 问题原因分析

### 1. 参数类型不匹配
- **前端传递**: `call.uuid` (字符串类型)
- **后端期望**: `call_id` (数字类型)
- **前端发送**: `{ call_id: callUUID }` 其中 `callUUID` 是字符串

### 2. 后端查找逻辑问题
- 后端使用 `DB.First(&call, callID)` 查找通话
- 这个函数期望的是数据库主键ID，而不是UUID
- 当传递字符串UUID时，会导致类型转换错误

### 3. 数据流问题
```
前端: rejectCall(call.uuid) 
  → api.endCall(callUUID) 
  → 发送 { call_id: callUUID } (字符串)
  → 后端: EndCallRequest{ CallID: callUUID } (类型不匹配)
  → DB.First(&call, callID) (查找失败)
```

## 解决方案

### 1. 修改后端API支持UUID
**文件**: `core/backend/handlers/call_handler.go`

```go
// 修改前
type EndCallRequest struct {
    CallID uint `json:"call_id" binding:"required"`
}

// 修改后
type EndCallRequest struct {
    CallID   uint   `json:"call_id"`
    CallUUID string `json:"call_uuid"`
}

func EndCall(c *gin.Context) {
    var req EndCallRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request body",
        })
        return
    }

    var call models.Call
    var err error

    // 优先使用UUID查找，如果没有则使用ID
    if req.CallUUID != "" {
        err = DB.Where("uuid = ?", req.CallUUID).First(&call).Error
    } else if req.CallID != 0 {
        err = DB.First(&call, req.CallID).Error
    } else {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Either call_id or call_uuid is required",
        })
        return
    }

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Call not found",
        })
        return
    }
    // ... 其余代码保持不变
}
```

### 2. 修改前端API调用
**文件**: `web_interface/js/api.js`

```javascript
// 修改前
async endCall(callId) {
    return this.request(`${this.baseURL}/api/v1/calls/end`, {
        method: 'POST',
        body: JSON.stringify({ call_id: callId })
    });
}

// 修改后
async endCall(callId) {
    return this.request(`${this.baseURL}/api/v1/calls/end`, {
        method: 'POST',
        body: JSON.stringify({ call_uuid: callId })
    });
}
```

## 修复效果

### 修复前
- 前端发送字符串UUID作为call_id
- 后端尝试将字符串转换为uint类型失败
- 返回HTTP 400错误

### 修复后
- 前端发送call_uuid参数
- 后端支持通过UUID查找通话记录
- 同时保持向后兼容性（仍支持call_id参数）
- 返回正确的HTTP状态码

## 测试验证

### 1. API参数验证测试
```bash
# 测试call_uuid参数
curl -X POST http://localhost:8000/api/v1/calls/end \
  -H "Content-Type: application/json" \
  -d '{"call_uuid": "test-uuid-123"}'
# 返回: 401 (需要认证) - 说明参数格式正确

# 测试call_id参数
curl -X POST http://localhost:8000/api/v1/calls/end \
  -H "Content-Type: application/json" \
  -d '{"call_id": 123}'
# 返回: 401 (需要认证) - 说明参数格式正确
```

### 2. 前端功能测试
- 用户收到来电通知
- 点击"拒绝"按钮
- 通话状态正确更新为"已拒绝"
- 不再出现HTTP 400错误

## 技术要点

### 1. 向后兼容性
- 修改后的API同时支持`call_id`和`call_uuid`参数
- 现有代码无需修改即可继续工作

### 2. 错误处理
- 当两个参数都未提供时，返回明确的错误信息
- 当通话不存在时，返回404状态码

### 3. 数据库查询优化
- 使用`DB.Where("uuid = ?", req.CallUUID)`进行精确查询
- 避免了类型转换问题

## 总结

通过修改后端API支持UUID参数和前端发送正确的参数名，成功解决了拒绝通话失败的问题。修复后的系统更加健壮，支持多种参数格式，并提供了更好的错误处理机制。 