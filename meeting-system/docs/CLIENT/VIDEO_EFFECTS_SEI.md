# 使用 H264 SEI 传输美颜/滤镜参数（客户端本地渲染）

本方案通过 **H264 SEI(user_data_unregistered)** 在视频码流中携带「美颜/滤镜」参数，使参数与视频帧天然同步；SFU 仅做 RTP 转发，无需增加额外信令通道即可让接收端在本地渲染阶段应用特效。

## 1. 发送端：在编码码流中插入 SEI

- **NAL 类型**：`SEI`（`nal_unit_type = 6`）
- **SEI payloadType**：`user_data_unregistered`（`payloadType = 5`）
- **UUID**：用于区分自定义 SEI（本项目 Web Demo 使用固定 UUID：`b0f7b0a1-6a3d-4c53-9b2e-6a7d3e9f1c20`）
- **user_data**：UTF-8 JSON（建议保持小且稳定）

建议 JSON 结构：

```json
{ "v": 1, "b": 20, "f": "warm", "t": 1730000000000 }
```

- `v`：版本号
- `b`：美颜强度（0~100）
- `f`：滤镜（`none|warm|cool|gray|vivid`，可扩展）
- `t`：时间戳（毫秒）

**插入位置**：建议插在每个 Access Unit 内 **第一段 VCL NAL（slice）之前**（例如在 SPS/PPS 后、slice 前）。  
**插入频率**：推荐：
- 参数变化时立即插入（至少连续几帧/直到下一个关键帧）；
- 每个关键帧都插入一次，保证新加入的订阅者/丢包后能尽快拿到最新参数。

## 2. 接收端：解析 SEI 并更新渲染参数

接收端在解码前（或解码器可见的码流阶段）扫描 NAL：

- 找到 `nal_unit_type=6` 的 SEI；
- 解析 `payloadType=5`，匹配 UUID；
- 读取 JSON 并更新当前「渲染参数状态」；
- 对后续视频帧在本地渲染阶段应用美颜/滤镜。

## 3. Web Demo（本仓库）说明

Web 端已在 `meeting-system/frontend/dist/app.js` 内做了参考实现：

- **发送端**：通过 `RTCRtpSender.createEncodedStreams()` 对 H264 编码帧注入 SEI（浏览器需支持 Encoded Insertable Streams）。
- **接收端**：通过 `RTCRtpReceiver.createEncodedStreams()` 解析 SEI，并对对应视频 tile 使用 CSS filter 做本地渲染效果。

注意：

- 浏览器是否能访问编码帧/是否协商到 H264 取决于运行环境；不支持时会显示 `SEI：不支持`，但通话不受影响。
- 本 Demo 的“美颜”是渲染层效果（模糊/亮度/色彩），真实美颜（磨皮/瘦脸/五官点位）应在客户端视频处理链路中实现（OpenGL/OpenCV/模型推理等）。

