
# 完整端到端消息队列集成测试报告

**测试时间**: 2025-10-06 11:39:13
**测试时长**: 13.69 秒
**总测试数**: 14
**通过**: 9
**失败**: 5
**跳过**: 0
**成功率**: 64.29%

## 测试详情


### 用户注册

- ✅ e2e_user_1: ID: 81
- ✅ e2e_user_2: ID: 82
- ✅ e2e_user_3: ID: 83
- ✅ e2e_user_4: ID: 84

### 用户登录

- ✅ e2e_user_1: 获取 token 成功
- ✅ e2e_user_2: 获取 token 成功
- ✅ e2e_user_3: 获取 token 成功
- ✅ e2e_user_4: 获取 token 成功

### 获取资料

- ✅ e2e_user_1: 

### 创建会议

- ❌ e2e_user_1: 状态码: 400, 响应: {"code":400,"message":"Parameter validation failed: parsing time \"2025-10-06T11:39:19.141000\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"Z07:00\"","timestamp":"2025-10-06T11:39:19+08:0

### 情绪识别

- ❌ e2e_user_1→e2e_user_2: 状态码: 502, 响应: <html>
<head><title>502 Bad Gateway</title></head>
<body>
<center><h1>502 Bad Gateway</h1></cente
- ❌ e2e_user_2→e2e_user_1: 状态码: 502, 响应: <html>
<head><title>502 Bad Gateway</title></head>
<body>
<center><h1>502 Bad Gateway</h1></cente

### 语音识别

- ❌ e2e_user_1→e2e_user_2: 状态码: 502
- ❌ e2e_user_2→e2e_user_1: 状态码: 502

## Redis 队列最终状态

- critical_queue: 0
- high_queue: 0
- normal_queue: 0
- low_queue: 0
- dead_letter_queue: 0
- processing_queue: 0
