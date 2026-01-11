# 使用 H264 SEI 传输美颜/滤镜参数（客户端本地渲染）

示例方案：在 H264 码流中写入 `user_data_unregistered` SEI，携带轻量 JSON 美颜/滤镜参数，接收端在解码前解析并在本地渲染层应用。SFU 仅转发 RTP，不改动码流。

## 发送端（注入 SEI）

- NAL 类型：`SEI` (`nal_unit_type = 6`)
- payloadType：`user_data_unregistered` (`payloadType = 5`)
- UUID：用于区分自定义 SEI（示例：`b0f7b0a1-6a3d-4c53-9b2e-6a7d3e9f1c20`）
- user_data：UTF-8 JSON，示例：

```json
{ "v": 1, "b": 20, "f": "warm", "t": 1730000000000 }
```

字段含义：版本 `v`、美颜强度 `b`、滤镜 `f`（`none|warm|cool|gray|vivid` 可扩展）、时间戳 `t`(ms)。

推荐插入：每个 Access Unit 的首个 VCL NAL 之前（如 SPS/PPS 后、首个 slice 前）；参数变化时连续插入若干帧，并在每个关键帧携带最新参数。

## 接收端（解析与渲染）

1. 扫描 NAL，找到 `nal_unit_type=6`
2. 解析 `payloadType=5`，匹配指定 UUID
3. 读取 JSON 更新渲染参数状态
4. 在解码后或渲染阶段应用滤镜/美颜（本地 CSS/GL/Shader 均可）

## Web Demo 提示

- `frontend/dist/app.js` 中演示了使用 `RTCRtpSender/Receiver.createEncodedStreams()` 注入/解析 SEI，并用 CSS filter 在本地渲染。
- 浏览器需支持 Encoded Insertable Streams；不支持时仅关闭 SEI 功能，通话不受影响。
- 当前“美颜”仅为滤镜级别示例；真正的磨皮/瘦脸等需在本地视频处理链路实现。
