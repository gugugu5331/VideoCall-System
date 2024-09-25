# 会议系统架构设计

**文档版本**: v1.0  
**创建时间**: 2025-09-29  
**最后更新**: 2025-09-29  
**维护者**: 开发团队  
**文档类型**: 架构设计  

---

## 📋 文档概览

### 目的
本文档描述智能视频会议平台的整体架构设计，包括系统概述、技术架构、微服务设计和AI集成方案。

### 读者对象
- 系统架构师
- 后端开发工程师
- 项目管理人员

### 先决条件
- 熟悉微服务架构
- 了解WebRTC技术
- 具备分布式系统基础知识

---

## 📖 系统概述

基于现有Edge-LLM-Infra基础设施构建的企业级会议系统，采用SFU架构实现低延迟、高安全性的音视频通话，集成分布式AI推理服务。

## 技术架构

### 整体架构图

```
客户端层
├── Qt6 桌面客户端
├── Web 浏览器客户端
└── 移动端客户端

网关层
├── Nginx 负载均衡
└── API 网关

微服务层
├── 用户服务 (User Service)
├── 会议服务 (Meeting Service)  
├── 信令服务 (Signaling Service)
├── 媒体服务 (Media Service - SFU)
├── AI检测服务 (AI Detection Service)
└── 通知服务 (Notification Service)

AI推理层 (基于Edge-LLM-Infra)
├── Edge-Model-Infra
├── 模型管理器
└── 推理节点集群

数据层
├── PostgreSQL (主数据库)
├── Redis (缓存)
├── MongoDB (文档存储)
└── MinIO (对象存储)
```

### 技术栈

- **后端**: Go (Gin + GORM) + gRPC
- **前端**: Qt6 + Web (React/Vue)
- **通信**: HTTP + WebRTC (SFU架构)
- **数据库**: PostgreSQL + Redis + MongoDB + MinIO
- **AI推理**: 基于现有Edge-LLM-Infra
- **音视频**: FFmpeg + OpenCV + OpenGL
- **部署**: Docker + Nginx

## 项目结构

```
meeting-system/
├── backend/                    # 后端微服务
│   ├── api-gateway/           # API网关
│   ├── user-service/          # 用户服务
│   ├── meeting-service/       # 会议服务
│   ├── signaling-service/     # 信令服务
│   ├── media-service/         # 媒体服务(SFU)
│   ├── ai-service/            # AI检测服务
│   ├── notification-service/  # 通知服务
│   ├── shared/               # 共享库
│   └── scripts/              # 部署脚本
├── frontend/                  # 前端
│   ├── web-admin/            # Web管理界面
│   ├── web-client/           # Web客户端
│   └── qt-client/            # Qt桌面客户端
├── deployment/               # 部署配置
│   ├── docker/              # Docker配置
│   ├── nginx/               # Nginx配置
│   └── k8s/                 # Kubernetes配置
├── docs/                    # 文档
└── scripts/                 # 构建脚本
```

## 核心功能

### 1. 用户管理
- 用户注册/登录
- JWT认证
- 权限管理
- 用户资料管理

### 2. 会议管理
- 会议创建/删除
- 会议室管理
- 参与者管理
- 会议录制

### 3. 音视频通信
- SFU架构媒体路由
- WebRTC信令处理
- 音视频编解码
- 屏幕共享

### 4. AI功能集成
- 实时语音识别
- 情绪识别
- 音视频质量优化
- 智能降噪

### 5. 实时通信
- 文字聊天
- 文件共享
- 白板协作
- 实时通知

## 与Edge-LLM-Infra集成架构

### 集成策略

会议系统采用**混合架构**设计，充分利用Edge-LLM-Infra的分布式AI推理能力：

```
┌─────────────────────────────────────────────────────────────┐
│                    会议系统整体架构                              │
├─────────────────────────────────────────────────────────────┤
│  Go微服务层 (业务逻辑)                                          │
│  ├── API网关 ──┐                                             │
│  ├── 用户服务   │                                             │
│  ├── 会议服务   ├─── ZMQ/TCP协议网关 ──┐                        │
│  ├── 信令服务   │                     │                      │
│  ├── 媒体服务   │                     │                      │
│  └── 通知服务 ──┘                     │                      │
├─────────────────────────────────────────┼─────────────────────┤
│  Edge-LLM-Infra (AI推理基础设施)         │                      │
│  ├── unit-manager (全局节点管理)        │                      │
│  ├── 会议AI节点 (继承StackFlow)         │                      │
│  │   ├── 语音识别任务                  │                      │
│  │   ├── 情绪识别任务                  │                      │
│  │   ├── 音频降噪任务                  │                      │
│  │   └── 视频增强任务                  │                      │
│  └── 推理节点集群                      │                      │
└─────────────────────────────────────────┼─────────────────────┘
                                        │
                                   标准化JSON协议
                                   (request_id, work_id,
                                    object, data, error)
```

### 核心集成组件

#### 1. 会议AI推理节点 (C++)
基于StackFlow框架实现的专用AI推理节点：

```cpp
class MeetingAINode : public StackFlow {
public:
    // 继承StackFlow的标准接口
    int setup(const std::string& work_id, const std::string& object, const std::string& data) override;
    int exit(const std::string& work_id, const std::string& object, const std::string& data) override;

    // 会议专用AI任务
    void processSpeechRecognition(const std::string& audio_data);
    void processEmotionDetection(const std::string& video_frame);
    void processAudioDenoising(const std::string& audio_stream);
    void processVideoEnhancement(const std::string& video_stream);
};
```

#### 2. Go-ZMQ桥接库
实现Go语言与ZMQ的通信桥接：

```go
package zmqbridge

type ZMQClient struct {
    endpoint string
    timeout  time.Duration
}

// 调用Edge-LLM-Infra的AI推理服务
func (c *ZMQClient) CallAIService(request AIRequest) (*AIResponse, error)

// 标准化消息格式
type AIRequest struct {
    RequestID string `json:"request_id"`
    WorkID    string `json:"work_id"`
    Object    string `json:"object"`
    Data      string `json:"data"`
}
```

#### 3. 协议转换网关
利用unit-manager的多协议网关能力：

- **TCP接口**: Go微服务通过TCP连接
- **ZMQ转换**: 自动转换为内部ZMQ协议
- **负载均衡**: 自动分发到可用的AI推理节点

### 集成通信流程

1. **服务注册**:
   ```
   会议AI节点 → unit-manager → 服务注册表
   ```

2. **AI推理请求**:
   ```
   Go微服务 → TCP网关 → ZMQ协议转换 → 会议AI节点 → 推理结果返回
   ```

3. **实时流处理**:
   ```
   媒体流 → Go媒体服务 → AI推理请求 → Edge-LLM-Infra → 处理结果 → 客户端
   ```

### 消息协议标准

复用Edge-LLM-Infra的标准化JSON协议：

```json
{
  "request_id": "req_123456",
  "work_id": "meeting_ai_001",
  "object": "speech_recognition",
  "data": {
    "audio_format": "pcm",
    "sample_rate": 16000,
    "audio_data": "base64_encoded_audio"
  },
  "error": null
}
```

### AI功能映射

| 会议功能 | Edge-LLM-Infra任务类型 | 处理节点 |
|---------|----------------------|---------|
| 实时语音识别 | speech_recognition | 语音识别节点 |
| 情绪识别 | emotion_detection | 视觉分析节点 |
| 智能降噪 | audio_denoising | 音频处理节点 |
| 视频增强 | video_enhancement | 视频处理节点 |
| 语音合成 | text_to_speech | 语音合成节点 |

## 部署方案

### 开发环境
```bash
# 启动基础服务
docker-compose -f deployment/docker/dev.yml up -d

# 启动微服务
./scripts/start-dev.sh
```

### 生产环境
```bash
# 使用Kubernetes部署
kubectl apply -f deployment/k8s/
```

## 开发计划

1. ✅ 架构设计与规划
2. 🔄 后端微服务基础框架
3. ⏳ 数据库层设计与实现
4. ⏳ 核心微服务实现
5. ⏳ SFU媒体服务实现
6. ⏳ AI服务集成
7. ⏳ 前端开发
8. ⏳ 系统集成测试
9. ⏳ 容器化部署

## 性能指标

- **延迟**: < 100ms (音视频)
- **并发**: 支持1000+并发会议
- **可用性**: 99.9%
- **扩展性**: 水平扩展支持

---

## 📝 相关资源

### 相关文档
- [项目状态分析报告](../progress-reports/2025-09-29-project-status-analysis.md)
- [文档管理规范](../development/documentation-standards.md)

### 外部资源
- [WebRTC官方文档](https://webrtc.org/)
- [SFU架构指南](https://webrtcglossary.com/sfu/)

---

## 🔄 变更历史

| 版本 | 日期 | 变更内容 | 变更者 |
|------|------|----------|--------|
| v1.0 | 2025-09-29 | 从meeting-system/README.md迁移并规范化格式 | 开发团队 |

---

## 📞 联系信息

### 文档维护者
- **姓名**: 开发团队
- **职责**: 会议系统架构设计和维护

### 技术支持
如有关于架构设计的问题，请通过以下方式联系：
1. 创建GitHub Issue (标签: architecture)
2. 在团队协作平台讨论
3. 参加架构评审会议

---

**注意**: 本架构设计为当前版本，随着项目发展可能进行调整和优化。
