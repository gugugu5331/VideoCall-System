# 🎥 Meeting System Backend - 后端服务文档

## 📋 目录

- [系统概述](#-系统概述)
- [微服务架构](#-微服务架构)
- [技术栈](#-技术栈)
- [快速开始](#-快速开始)
- [服务详解](#-服务详解)
- [数据库设计](#-数据库设计)
- [API 文档](#-api-文档)
- [配置说明](#-配置说明)
- [部署指南](#-部署指南)

---

## 📖 系统概述

Meeting System Backend 是一个基于 Go 语言的微服务架构视频会议系统后端，采用 SFU (Selective Forwarding Unit) 媒体转发架构，集成 Edge-LLM-Infra 分布式 AI 推理框架。

**核心特性：**
- 🏗️ **微服务架构**: 5个独立的 Go 微服务 + AI 推理服务
- 🔐 **安全认证**: JWT + CSRF 保护 + 限流
- 📡 **实时通信**: WebSocket 信令 + WebRTC 媒体传输
- 🤖 **AI 集成**: ZeroMQ 连接 Edge-LLM-Infra
- 📊 **完整监控**: Prometheus + Jaeger + Loki
- 🔄 **服务发现**: etcd 服务注册与发现
- 🐳 **容器化**: Docker Compose 一键部署

---

## 🏗️ 微服务架构

### 服务组件
<svg aria-roledescription="flowchart-v2" role="graphics-document document" viewBox="0 0 2972.3984375 1936" style="max-width: 2972.3984375px;" class="flowchart" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" width="100%" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3"><style>#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3{font-family:"trebuchet ms",verdana,arial,sans-serif;font-size:16px;fill:#ccc;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .error-icon{fill:#a44141;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .error-text{fill:#ddd;stroke:#ddd;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-thickness-normal{stroke-width:1px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-thickness-thick{stroke-width:3.5px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-pattern-solid{stroke-dasharray:0;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-thickness-invisible{stroke-width:0;fill:none;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-pattern-dashed{stroke-dasharray:3;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edge-pattern-dotted{stroke-dasharray:2;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .marker{fill:lightgrey;stroke:lightgrey;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .marker.cross{stroke:lightgrey;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 svg{font-family:"trebuchet ms",verdana,arial,sans-serif;font-size:16px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 p{margin:0;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .label{font-family:"trebuchet ms",verdana,arial,sans-serif;color:#ccc;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster-label text{fill:#F9FFFE;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster-label span{color:#F9FFFE;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster-label span p{background-color:transparent;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .label text,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 span{fill:#ccc;color:#ccc;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node rect,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node circle,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node ellipse,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node polygon,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node path{fill:#1f2020;stroke:#ccc;stroke-width:1px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .rough-node .label text,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node .label text,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .image-shape .label,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .icon-shape .label{text-anchor:middle;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node .katex path{fill:#000;stroke:#000;stroke-width:1px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .rough-node .label,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node .label,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .image-shape .label,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .icon-shape .label{text-align:center;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .node.clickable{cursor:pointer;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .root .anchor path{fill:lightgrey!important;stroke-width:0;stroke:lightgrey;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .arrowheadPath{fill:lightgrey;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edgePath .path{stroke:lightgrey;stroke-width:2.0px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .flowchart-link{stroke:lightgrey;fill:none;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edgeLabel{background-color:hsl(0, 0%, 34.4117647059%);text-align:center;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edgeLabel p{background-color:hsl(0, 0%, 34.4117647059%);}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .edgeLabel rect{opacity:0.5;background-color:hsl(0, 0%, 34.4117647059%);fill:hsl(0, 0%, 34.4117647059%);}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .labelBkg{background-color:rgba(87.75, 87.75, 87.75, 0.5);}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster rect{fill:hsl(180, 1.5873015873%, 28.3529411765%);stroke:rgba(255, 255, 255, 0.25);stroke-width:1px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster text{fill:#F9FFFE;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .cluster span{color:#F9FFFE;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 div.mermaidTooltip{position:absolute;text-align:center;max-width:200px;padding:2px;font-family:"trebuchet ms",verdana,arial,sans-serif;font-size:12px;background:hsl(20, 1.5873015873%, 12.3529411765%);border:1px solid rgba(255, 255, 255, 0.25);border-radius:2px;pointer-events:none;z-index:100;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .flowchartTitleText{text-anchor:middle;font-size:18px;fill:#ccc;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 rect.text{fill:none;stroke-width:0;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .icon-shape,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .image-shape{background-color:hsl(0, 0%, 34.4117647059%);text-align:center;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .icon-shape p,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .image-shape p{background-color:hsl(0, 0%, 34.4117647059%);padding:2px;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .icon-shape rect,#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .image-shape rect{opacity:0.5;background-color:hsl(0, 0%, 34.4117647059%);fill:hsl(0, 0%, 34.4117647059%);}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 :root{--mermaid-font-family:"trebuchet ms",verdana,arial,sans-serif;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .client&gt;*{fill:#e1f5ff!important;stroke:#01579b!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .client span{fill:#e1f5ff!important;stroke:#01579b!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .client tspan{fill:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .gateway&gt;*{fill:#fff3e0!important;stroke:#e65100!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .gateway span{fill:#fff3e0!important;stroke:#e65100!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .gateway tspan{fill:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .service&gt;*{fill:#f3e5f5!important;stroke:#4a148c!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .service span{fill:#f3e5f5!important;stroke:#4a148c!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .service tspan{fill:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .ai&gt;*{fill:#e8f5e9!important;stroke:#1b5e20!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .ai span{fill:#e8f5e9!important;stroke:#1b5e20!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .ai tspan{fill:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .data&gt;*{fill:#fce4ec!important;stroke:#880e4f!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .data span{fill:#fce4ec!important;stroke:#880e4f!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .data tspan{fill:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .obs&gt;*{fill:#f1f8e9!important;stroke:#33691e!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .obs span{fill:#f1f8e9!important;stroke:#33691e!important;stroke-width:2px!important;color:#000!important;}#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3 .obs tspan{fill:#000!important;}</style><g><marker orient="auto" markerHeight="8" markerWidth="8" markerUnits="userSpaceOnUse" refY="5" refX="5" viewBox="0 0 10 10" class="marker flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd"><path style="stroke-width: 1; stroke-dasharray: 1, 0;" class="arrowMarkerPath" d="M 0 0 L 10 5 L 0 10 z"></path></marker><marker orient="auto" markerHeight="8" markerWidth="8" markerUnits="userSpaceOnUse" refY="5" refX="4.5" viewBox="0 0 10 10" class="marker flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointStart"><path style="stroke-width: 1; stroke-dasharray: 1, 0;" class="arrowMarkerPath" d="M 0 5 L 10 10 L 10 0 z"></path></marker><marker orient="auto" markerHeight="11" markerWidth="11" markerUnits="userSpaceOnUse" refY="5" refX="11" viewBox="0 0 10 10" class="marker flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-circleEnd"><circle style="stroke-width: 1; stroke-dasharray: 1, 0;" class="arrowMarkerPath" r="5" cy="5" cx="5"></circle></marker><marker orient="auto" markerHeight="11" markerWidth="11" markerUnits="userSpaceOnUse" refY="5" refX="-1" viewBox="0 0 10 10" class="marker flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-circleStart"><circle style="stroke-width: 1; stroke-dasharray: 1, 0;" class="arrowMarkerPath" r="5" cy="5" cx="5"></circle></marker><marker orient="auto" markerHeight="11" markerWidth="11" markerUnits="userSpaceOnUse" refY="5.2" refX="12" viewBox="0 0 11 11" class="marker cross flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-crossEnd"><path style="stroke-width: 2; stroke-dasharray: 1, 0;" class="arrowMarkerPath" d="M 1,1 l 9,9 M 10,1 l -9,9"></path></marker><marker orient="auto" markerHeight="11" markerWidth="11" markerUnits="userSpaceOnUse" refY="5.2" refX="-1" viewBox="0 0 11 11" class="marker cross flowchart-v2" id="mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-crossStart"><path style="stroke-width: 2; stroke-dasharray: 1, 0;" class="arrowMarkerPath" d="M 1,1 l 9,9 M 10,1 l -9,9"></path></marker><g class="root"><g class="clusters"><g data-look="classic" id="Observability" class="cluster"><rect height="305" width="1648.13671875" y="1623" x="11.3828125" style=""></rect><g transform="translate(790.052734375, 1623)" class="cluster-label"><foreignObject height="24" width="90.796875"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>📊 可观测性</p></span></div></foreignObject></g></g><g data-look="classic" id="DataLayer" class="cluster"><rect height="378" width="545.890625" y="1397" x="1881.6015625" style=""></rect><g transform="translate(2117.1484375, 1397)" class="cluster-label"><foreignObject height="24" width="74.796875"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>💾 数据层</p></span></div></foreignObject></g></g><g data-look="classic" id="AILayer" class="cluster"><rect height="378" width="516.90625" y="1397" x="2447.4921875" style=""></rect><g transform="translate(2605.9921875, 1397)" class="cluster-label"><foreignObject height="24" width="199.90625"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>🤖 AI推理层 Edge-LLM-Infra</p></span></div></foreignObject></g></g><g data-look="classic" id="Microservices" class="cluster"><rect height="1057" width="1853.6015625" y="492" x="8" style=""></rect><g transform="translate(856.21484375, 492)" class="cluster-label"><foreignObject height="24" width="157.171875"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>🎯 微服务层 Go + Gin</p></span></div></foreignObject></g></g><g data-look="classic" id="Gateway" class="cluster"><rect height="256" width="1072.08984375" y="186" x="553.5234375" style=""></rect><g transform="translate(1052.169921875, 186)" class="cluster-label"><foreignObject height="24" width="74.796875"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>🌐 网关层</p></span></div></foreignObject></g></g><g data-look="classic" id="Client" class="cluster"><rect height="128" width="751.640625" y="8" x="664.22265625" style=""></rect><g transform="translate(994.64453125, 8)" class="cluster-label"><foreignObject height="24" width="90.796875"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" xmlns="http://www.w3.org/1999/xhtml"><span class="nodeLabel"><p>🖥️ 客户端层</p></span></div></foreignObject></g></g></g><g class="edgePaths"><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_Qt6_Nginx_0" d="M817.918,111L817.918,115.167C817.918,119.333,817.918,127.667,817.918,136C817.918,144.333,817.918,152.667,817.918,161C817.918,169.333,817.918,177.667,843.546,188.616C869.175,199.565,920.431,213.13,946.06,219.912L971.688,226.695"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_Web_Nginx_1" d="M1079.457,111L1079.457,115.167C1079.457,119.333,1079.457,127.667,1079.457,136C1079.457,144.333,1079.457,152.667,1079.457,161C1079.457,169.333,1079.457,177.667,1078.37,185.363C1077.283,193.059,1075.11,200.118,1074.023,203.648L1072.936,207.177"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_Mobile_Nginx_2" d="M1301.582,111L1301.582,115.167C1301.582,119.333,1301.582,127.667,1301.582,136C1301.582,144.333,1301.582,152.667,1301.582,161C1301.582,169.333,1301.582,177.667,1275.954,188.616C1250.325,199.565,1199.069,213.13,1173.44,219.912L1147.812,226.695"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_Nginx_APIGateway_3" d="M1059.75,289L1059.75,293.167C1059.75,297.333,1059.75,305.667,1059.75,313.333C1059.75,321,1059.75,328,1059.75,331.5L1059.75,335"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_UserSvc_4" d="M967.484,400.97L940.017,407.809C912.549,414.647,857.615,428.323,830.147,439.328C802.68,450.333,802.68,458.667,802.68,467C802.68,475.333,802.68,483.667,802.68,491.333C802.68,499,802.68,506,802.68,509.5L802.68,513"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_MeetingSvc_5" d="M986.82,417L979.029,421.167C971.237,425.333,955.654,433.667,947.862,442C940.07,450.333,940.07,458.667,940.07,467C940.07,475.333,940.07,483.667,940.07,500.5C940.07,517.333,940.07,542.667,940.07,570C940.07,597.333,940.07,626.667,941.259,646.848C942.447,667.03,944.823,678.06,946.012,683.575L947.2,689.09"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_SignalSvc_6" d="M1084.415,417L1087.051,421.167C1089.686,425.333,1094.956,433.667,1097.591,442C1100.227,450.333,1100.227,458.667,1100.227,467C1100.227,475.333,1100.227,483.667,1100.227,500.5C1100.227,517.333,1100.227,542.667,1100.227,570C1100.227,597.333,1100.227,626.667,1100.227,656C1100.227,685.333,1100.227,714.667,1100.227,744C1100.227,773.333,1100.227,802.667,1101.415,822.848C1102.603,843.03,1104.98,854.06,1106.168,859.575L1107.356,865.09"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_MediaSvc_7" d="M1152.016,407.587L1169.901,413.323C1187.786,419.058,1223.557,430.529,1241.443,440.431C1259.328,450.333,1259.328,458.667,1259.328,467C1259.328,475.333,1259.328,483.667,1259.328,500.5C1259.328,517.333,1259.328,542.667,1259.328,570C1259.328,597.333,1259.328,626.667,1259.328,656C1259.328,685.333,1259.328,714.667,1259.328,744C1259.328,773.333,1259.328,802.667,1259.328,832C1259.328,861.333,1259.328,890.667,1259.328,920C1259.328,949.333,1259.328,978.667,1260.516,998.848C1261.705,1019.03,1264.081,1030.06,1265.27,1035.575L1266.458,1041.09"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_AISvc_8" d="M1152.016,389.887L1219.435,398.572C1286.854,407.258,1421.693,424.629,1489.112,437.481C1556.531,450.333,1556.531,458.667,1556.531,467C1556.531,475.333,1556.531,483.667,1556.531,500.5C1556.531,517.333,1556.531,542.667,1556.531,570C1556.531,597.333,1556.531,626.667,1556.531,656C1556.531,685.333,1556.531,714.667,1556.531,744C1556.531,773.333,1556.531,802.667,1556.531,832C1556.531,861.333,1556.531,890.667,1556.531,920C1556.531,949.333,1556.531,978.667,1556.531,1008C1556.531,1037.333,1556.531,1066.667,1556.531,1096C1556.531,1125.333,1556.531,1154.667,1558.271,1174.864C1560.011,1195.061,1563.49,1206.123,1565.229,1211.654L1566.969,1217.184"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_APIGateway_NotifySvc_9" d="M967.484,392.97L917.118,401.141C866.753,409.313,766.021,425.657,715.655,437.995C665.289,450.333,665.289,458.667,665.289,467C665.289,475.333,665.289,483.667,665.289,500.5C665.289,517.333,665.289,542.667,665.289,570C665.289,597.333,665.289,626.667,665.289,656C665.289,685.333,665.289,714.667,665.289,744C665.289,773.333,665.289,802.667,665.289,832C665.289,861.333,665.289,890.667,665.289,920C665.289,949.333,665.289,978.667,665.289,1008C665.289,1037.333,665.289,1066.667,665.289,1096C665.289,1125.333,665.289,1154.667,665.289,1184C665.289,1213.333,665.289,1242.667,665.289,1272C665.289,1301.333,665.289,1330.667,665.289,1351.5C665.289,1372.333,665.289,1384.667,667.355,1394.422C669.42,1404.178,673.551,1411.355,675.617,1414.944L677.683,1418.533"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_UserSvc_MeetingSvc_10" d="M905.07,611.322L922.67,618.768C940.27,626.214,975.469,641.107,989.787,654.145C1004.106,667.183,997.544,678.367,994.263,683.958L990.981,689.55"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MeetingSvc_SignalSvc_11" d="M1065.227,788.234L1082.738,795.529C1100.25,802.823,1135.273,817.411,1149.538,830.296C1163.803,843.18,1157.31,854.361,1154.063,859.951L1150.817,865.541"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_SignalSvc_MediaSvc_12" d="M1224.328,964.6L1241.38,971.833C1258.432,979.067,1292.536,993.533,1306.521,1006.349C1320.506,1019.165,1314.372,1030.33,1311.304,1035.912L1308.237,1041.494"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MediaSvc_AISvc_13" d="M1377.914,1120.65L1420.586,1131.209C1463.258,1141.767,1548.602,1162.883,1588.116,1179.028C1627.631,1195.173,1621.317,1206.345,1618.16,1211.931L1615.002,1217.518"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_NotifySvc_14" d="M1481.82,1283.907L1372.76,1296.589C1263.701,1309.271,1045.581,1334.636,936.521,1353.484C827.461,1372.333,827.461,1384.667,821.529,1394.64C815.597,1404.613,803.734,1412.226,797.802,1416.033L791.87,1419.84"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_ModelMgr_15" d="M1686.602,1309.956L1709.102,1318.297C1731.602,1326.637,1776.602,1343.319,1799.102,1357.826C1821.602,1372.333,1821.602,1384.667,1931.087,1401.651C2040.572,1418.636,2259.542,1440.273,2369.027,1451.091L2478.512,1461.909"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_InferEngine_16" d="M1686.602,1307.007L1712.435,1315.839C1738.268,1324.671,1789.935,1342.336,1815.768,1357.334C1841.602,1372.333,1841.602,1384.667,1992.174,1402.372C2142.746,1420.078,2443.891,1443.156,2594.463,1454.695L2745.035,1466.234"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_InferEngine_InferCluster_17" d="M2833.328,1512L2833.328,1518.167C2833.328,1524.333,2833.328,1536.667,2833.328,1549C2833.328,1561.333,2833.328,1573.667,2833.328,1586C2833.328,1598.333,2833.328,1610.667,2833.328,1622.333C2833.328,1634,2833.328,1645,2833.328,1650.5L2833.328,1656"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_UserSvc_PostgreSQL_18" d="M905.07,577.597L1044.492,590.664C1183.914,603.731,1462.758,629.866,1602.18,657.599C1741.602,685.333,1741.602,714.667,1741.602,744C1741.602,773.333,1741.602,802.667,1741.602,832C1741.602,861.333,1741.602,890.667,1741.602,920C1741.602,949.333,1741.602,978.667,1741.602,1008C1741.602,1037.333,1741.602,1066.667,1741.602,1096C1741.602,1125.333,1741.602,1154.667,1741.602,1184C1741.602,1213.333,1741.602,1242.667,1741.602,1272C1741.602,1301.333,1741.602,1330.667,1741.602,1351.5C1741.602,1372.333,1741.602,1384.667,1771.576,1398.579C1801.55,1412.491,1861.499,1427.982,1891.473,1435.728L1921.448,1443.474"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MeetingSvc_PostgreSQL_19" d="M1065.227,755.644L1181.289,768.37C1297.352,781.096,1529.477,806.548,1645.539,833.941C1761.602,861.333,1761.602,890.667,1761.602,920C1761.602,949.333,1761.602,978.667,1761.602,1008C1761.602,1037.333,1761.602,1066.667,1761.602,1096C1761.602,1125.333,1761.602,1154.667,1761.602,1184C1761.602,1213.333,1761.602,1242.667,1761.602,1272C1761.602,1301.333,1761.602,1330.667,1761.602,1351.5C1761.602,1372.333,1761.602,1384.667,1788.246,1398.221C1814.89,1411.775,1868.178,1426.549,1894.822,1433.937L1921.466,1441.324"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_SignalSvc_Redis_20" d="M1224.328,935.359L1307.207,947.466C1390.086,959.573,1555.844,983.786,1638.723,1010.56C1721.602,1037.333,1721.602,1066.667,1721.602,1096C1721.602,1125.333,1721.602,1154.667,1721.602,1184C1721.602,1213.333,1721.602,1242.667,1721.602,1272C1721.602,1301.333,1721.602,1330.667,1721.602,1351.5C1721.602,1372.333,1721.602,1384.667,1721.602,1403.5C1721.602,1422.333,1721.602,1447.667,1721.602,1473C1721.602,1498.333,1721.602,1523.667,1758.837,1542.5C1796.073,1561.333,1870.544,1573.667,1907.78,1586C1945.016,1598.333,1945.016,1610.667,1948.54,1620.518C1952.065,1630.37,1959.114,1637.74,1962.639,1641.425L1966.164,1645.109"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MediaSvc_PostgreSQL_21" d="M1377.914,1113.419L1445.195,1125.182C1512.477,1136.946,1647.039,1160.473,1714.32,1186.903C1781.602,1213.333,1781.602,1242.667,1781.602,1272C1781.602,1301.333,1781.602,1330.667,1781.602,1351.5C1781.602,1372.333,1781.602,1384.667,1804.916,1397.806C1828.23,1410.946,1874.859,1424.892,1898.174,1431.865L1921.488,1438.838"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_MongoDB_22" d="M1686.602,1313.448L1705.768,1321.207C1724.935,1328.965,1763.268,1344.483,1782.435,1358.408C1801.602,1372.333,1801.602,1384.667,1866.693,1400.874C1931.784,1417.081,2061.966,1437.162,2127.057,1447.203L2192.148,1457.243"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_NotifySvc_Redis_23" d="M795.422,1479.065L961.452,1490.721C1127.482,1502.377,1459.542,1525.688,1660.257,1543.511C1860.971,1561.333,1930.341,1573.667,1965.026,1586C1999.711,1598.333,1999.711,1610.667,2000.544,1620.351C2001.377,1630.036,2003.044,1637.072,2003.877,1640.59L2004.71,1644.108"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_PostgreSQL_Redis_24" d="M2035.711,1524L2035.711,1528.167C2035.711,1532.333,2035.711,1540.667,2035.711,1551C2035.711,1561.333,2035.711,1573.667,2035.711,1586C2035.711,1598.333,2035.711,1610.667,2034.878,1620.351C2034.045,1630.036,2032.378,1637.072,2031.545,1640.59L2030.712,1644.108"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MongoDB_MinIO_25" d="M2294.297,1524L2294.297,1528.167C2294.297,1532.333,2294.297,1540.667,2294.297,1551C2294.297,1561.333,2294.297,1573.667,2294.297,1586C2294.297,1598.333,2294.297,1610.667,2294.297,1620.333C2294.297,1630,2294.297,1637,2294.297,1640.5L2294.297,1644"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_UserSvc_Prometheus_26" d="M700.289,580.047L592.702,592.706C485.115,605.365,269.94,630.682,162.353,658.008C54.766,685.333,54.766,714.667,54.766,744C54.766,773.333,54.766,802.667,54.766,832C54.766,861.333,54.766,890.667,54.766,920C54.766,949.333,54.766,978.667,54.766,1008C54.766,1037.333,54.766,1066.667,54.766,1096C54.766,1125.333,54.766,1154.667,54.766,1184C54.766,1213.333,54.766,1242.667,54.766,1272C54.766,1301.333,54.766,1330.667,54.766,1351.5C54.766,1372.333,54.766,1384.667,54.766,1403.5C54.766,1422.333,54.766,1447.667,54.766,1473C54.766,1498.333,54.766,1523.667,54.766,1542.5C54.766,1561.333,54.766,1573.667,54.766,1586C54.766,1598.333,54.766,1610.667,62.999,1622.617C71.233,1634.567,87.7,1646.134,95.933,1651.917L104.166,1657.701"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MeetingSvc_Prometheus_27" d="M852.836,754.898L727.618,767.748C602.401,780.599,351.966,806.299,226.749,833.816C101.531,861.333,101.531,890.667,101.531,920C101.531,949.333,101.531,978.667,101.531,1008C101.531,1037.333,101.531,1066.667,101.531,1096C101.531,1125.333,101.531,1154.667,101.531,1184C101.531,1213.333,101.531,1242.667,101.531,1272C101.531,1301.333,101.531,1330.667,101.531,1351.5C101.531,1372.333,101.531,1384.667,101.531,1403.5C101.531,1422.333,101.531,1447.667,101.531,1473C101.531,1498.333,101.531,1523.667,101.531,1542.5C101.531,1561.333,101.531,1573.667,101.531,1586C101.531,1598.333,101.531,1610.667,106.097,1622.482C110.662,1634.296,119.793,1645.593,124.358,1651.241L128.923,1656.889"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_SignalSvc_Prometheus_28" d="M1014.047,929.53L869.755,942.608C725.464,955.687,436.88,981.843,292.589,1009.588C148.297,1037.333,148.297,1066.667,148.297,1096C148.297,1125.333,148.297,1154.667,148.297,1184C148.297,1213.333,148.297,1242.667,148.297,1272C148.297,1301.333,148.297,1330.667,148.297,1351.5C148.297,1372.333,148.297,1384.667,148.297,1403.5C148.297,1422.333,148.297,1447.667,148.297,1473C148.297,1498.333,148.297,1523.667,148.297,1542.5C148.297,1561.333,148.297,1573.667,148.297,1586C148.297,1598.333,148.297,1610.667,149.36,1622.345C150.424,1634.024,152.551,1645.048,153.615,1650.56L154.678,1656.072"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MediaSvc_Prometheus_29" d="M1178.664,1104.093L1014.73,1117.411C850.797,1130.729,522.93,1157.364,358.996,1185.349C195.063,1213.333,195.063,1242.667,195.063,1272C195.063,1301.333,195.063,1330.667,195.063,1351.5C195.063,1372.333,195.063,1384.667,195.063,1403.5C195.063,1422.333,195.063,1447.667,195.063,1473C195.063,1498.333,195.063,1523.667,195.063,1542.5C195.063,1561.333,195.063,1573.667,195.063,1586C195.063,1598.333,195.063,1610.667,192.717,1622.386C190.372,1634.105,185.681,1645.21,183.336,1650.763L180.991,1656.315"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_Prometheus_30" d="M1481.82,1278.712L1275.155,1292.26C1068.49,1305.808,655.159,1332.904,448.493,1352.619C241.828,1372.333,241.828,1384.667,241.828,1403.5C241.828,1422.333,241.828,1447.667,241.828,1473C241.828,1498.333,241.828,1523.667,241.828,1542.5C241.828,1561.333,241.828,1573.667,241.828,1586C241.828,1598.333,241.828,1610.667,235.909,1622.537C229.99,1634.408,218.151,1645.816,212.232,1651.52L206.313,1657.224"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_NotifySvc_Prometheus_31" d="M622.641,1488.616L566.966,1498.68C511.292,1508.744,399.943,1528.872,344.268,1545.103C288.594,1561.333,288.594,1573.667,288.594,1586C288.594,1598.333,288.594,1610.667,278.97,1622.655C269.347,1634.643,250.1,1646.286,240.476,1652.108L230.853,1657.93"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-solid edge-thickness-normal edge-pattern-solid flowchart-link" id="L_Prometheus_Grafana_32" d="M162.961,1738L162.961,1744.167C162.961,1750.333,162.961,1762.667,162.961,1773C162.961,1783.333,162.961,1791.667,162.961,1799.333C162.961,1807,162.961,1814,162.961,1817.5L162.961,1821"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_UserSvc_Jaeger_33" d="M700.289,587.281L639.467,598.734C578.646,610.187,457.003,633.094,396.181,659.213C335.359,685.333,335.359,714.667,335.359,744C335.359,773.333,335.359,802.667,335.359,832C335.359,861.333,335.359,890.667,335.359,920C335.359,949.333,335.359,978.667,335.359,1008C335.359,1037.333,335.359,1066.667,335.359,1096C335.359,1125.333,335.359,1154.667,335.359,1184C335.359,1213.333,335.359,1242.667,335.359,1272C335.359,1301.333,335.359,1330.667,335.359,1351.5C335.359,1372.333,335.359,1384.667,335.359,1403.5C335.359,1422.333,335.359,1447.667,335.359,1473C335.359,1498.333,335.359,1523.667,335.359,1542.5C335.359,1561.333,335.359,1573.667,335.359,1586C335.359,1598.333,335.359,1610.667,342.658,1622.587C349.956,1634.508,364.552,1646.016,371.851,1651.77L379.149,1657.524"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MeetingSvc_Jaeger_34" d="M852.836,760.067L773.598,772.056C694.359,784.045,535.883,808.022,456.645,834.678C377.406,861.333,377.406,890.667,377.406,920C377.406,949.333,377.406,978.667,377.406,1008C377.406,1037.333,377.406,1066.667,377.406,1096C377.406,1125.333,377.406,1154.667,377.406,1184C377.406,1213.333,377.406,1242.667,377.406,1272C377.406,1301.333,377.406,1330.667,377.406,1351.5C377.406,1372.333,377.406,1384.667,377.406,1403.5C377.406,1422.333,377.406,1447.667,377.406,1473C377.406,1498.333,377.406,1523.667,377.406,1542.5C377.406,1561.333,377.406,1573.667,377.406,1586C377.406,1598.333,377.406,1610.667,381.429,1622.458C385.451,1634.249,393.495,1645.498,397.518,1651.122L401.54,1656.746"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_SignalSvc_Jaeger_35" d="M1014.047,933.223L914.948,945.686C815.849,958.148,617.651,983.074,518.552,1010.204C419.453,1037.333,419.453,1066.667,419.453,1096C419.453,1125.333,419.453,1154.667,419.453,1184C419.453,1213.333,419.453,1242.667,419.453,1272C419.453,1301.333,419.453,1330.667,419.453,1351.5C419.453,1372.333,419.453,1384.667,419.453,1403.5C419.453,1422.333,419.453,1447.667,419.453,1473C419.453,1498.333,419.453,1523.667,419.453,1542.5C419.453,1561.333,419.453,1573.667,419.453,1586C419.453,1598.333,419.453,1610.667,420.345,1622.342C421.237,1634.017,423.021,1645.034,423.912,1650.543L424.804,1656.051"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MediaSvc_Jaeger_36" d="M1178.664,1106.733L1059.137,1119.611C939.609,1132.489,700.555,1158.244,581.027,1185.789C461.5,1213.333,461.5,1242.667,461.5,1272C461.5,1301.333,461.5,1330.667,461.5,1351.5C461.5,1372.333,461.5,1384.667,461.5,1403.5C461.5,1422.333,461.5,1447.667,461.5,1473C461.5,1498.333,461.5,1523.667,461.5,1542.5C461.5,1561.333,461.5,1573.667,461.5,1586C461.5,1598.333,461.5,1610.667,459.33,1622.379C457.159,1634.092,452.819,1645.183,450.648,1650.729L448.478,1656.275"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_Jaeger_37" d="M1481.82,1280.338L1318.775,1293.615C1155.729,1306.892,829.638,1333.446,666.592,1352.89C503.547,1372.333,503.547,1384.667,503.547,1403.5C503.547,1422.333,503.547,1447.667,503.547,1473C503.547,1498.333,503.547,1523.667,503.547,1542.5C503.547,1561.333,503.547,1573.667,503.547,1586C503.547,1598.333,503.547,1610.667,498.18,1622.515C492.812,1634.364,482.078,1645.728,476.711,1651.41L471.344,1657.092"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_NotifySvc_Jaeger_38" d="M622.641,1513.172L609.799,1519.144C596.958,1525.115,571.276,1537.057,558.435,1549.195C545.594,1561.333,545.594,1573.667,545.594,1586C545.594,1598.333,545.594,1610.667,536.912,1622.63C528.229,1634.593,510.865,1646.186,502.183,1651.982L493.5,1657.779"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_UserSvc_Loki_39" d="M700.289,609.901L681.514,617.584C662.74,625.267,625.19,640.634,606.415,662.984C587.641,685.333,587.641,714.667,587.641,744C587.641,773.333,587.641,802.667,587.641,832C587.641,861.333,587.641,890.667,587.641,920C587.641,949.333,587.641,978.667,587.641,1008C587.641,1037.333,587.641,1066.667,587.641,1096C587.641,1125.333,587.641,1154.667,587.641,1184C587.641,1213.333,587.641,1242.667,587.641,1272C587.641,1301.333,587.641,1330.667,587.641,1351.5C587.641,1372.333,587.641,1384.667,587.641,1403.5C587.641,1422.333,587.641,1447.667,587.641,1473C587.641,1498.333,587.641,1523.667,587.641,1542.5C587.641,1561.333,587.641,1573.667,587.641,1586C587.641,1598.333,587.641,1610.667,714.582,1628.408C841.524,1646.149,1095.407,1669.299,1222.348,1680.873L1349.29,1692.448"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MeetingSvc_Loki_40" d="M959.031,795L959.031,801.167C959.031,807.333,959.031,819.667,959.031,840.5C959.031,861.333,959.031,890.667,959.031,920C959.031,949.333,959.031,978.667,959.031,1008C959.031,1037.333,959.031,1066.667,959.031,1096C959.031,1125.333,959.031,1154.667,959.031,1184C959.031,1213.333,959.031,1242.667,959.031,1272C959.031,1301.333,959.031,1330.667,959.031,1351.5C959.031,1372.333,959.031,1384.667,959.031,1403.5C959.031,1422.333,959.031,1447.667,959.031,1473C959.031,1498.333,959.031,1523.667,959.031,1542.5C959.031,1561.333,959.031,1573.667,959.031,1586C959.031,1598.333,959.031,1610.667,1024.08,1627.531C1089.13,1644.396,1219.228,1665.792,1284.277,1676.49L1349.326,1687.188"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_SignalSvc_Loki_41" d="M1224.328,951.5L1255.759,960.917C1287.19,970.333,1350.052,989.167,1381.483,1013.25C1412.914,1037.333,1412.914,1066.667,1412.914,1096C1412.914,1125.333,1412.914,1154.667,1412.914,1184C1412.914,1213.333,1412.914,1242.667,1412.914,1272C1412.914,1301.333,1412.914,1330.667,1412.914,1351.5C1412.914,1372.333,1412.914,1384.667,1412.914,1403.5C1412.914,1422.333,1412.914,1447.667,1412.914,1473C1412.914,1498.333,1412.914,1523.667,1412.914,1542.5C1412.914,1561.333,1412.914,1573.667,1412.914,1586C1412.914,1598.333,1412.914,1610.667,1413.51,1622.337C1414.107,1634.008,1415.299,1645.016,1415.896,1650.519L1416.492,1656.023"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_MediaSvc_Loki_42" d="M1375.961,1147L1387.771,1153.167C1399.58,1159.333,1423.2,1171.667,1435.01,1192.5C1446.82,1213.333,1446.82,1242.667,1446.82,1272C1446.82,1301.333,1446.82,1330.667,1446.82,1351.5C1446.82,1372.333,1446.82,1384.667,1446.82,1403.5C1446.82,1422.333,1446.82,1447.667,1446.82,1473C1446.82,1498.333,1446.82,1523.667,1446.82,1542.5C1446.82,1561.333,1446.82,1573.667,1446.82,1586C1446.82,1598.333,1446.82,1610.667,1444.951,1622.368C1443.081,1634.07,1439.342,1645.14,1437.472,1650.675L1435.602,1656.21"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_AISvc_Loki_43" d="M1584.211,1323L1584.211,1329.167C1584.211,1335.333,1584.211,1347.667,1584.211,1360C1584.211,1372.333,1584.211,1384.667,1584.211,1403.5C1584.211,1422.333,1584.211,1447.667,1584.211,1473C1584.211,1498.333,1584.211,1523.667,1584.211,1542.5C1584.211,1561.333,1584.211,1573.667,1584.211,1586C1584.211,1598.333,1584.211,1610.667,1568.951,1623.946C1553.69,1637.225,1523.17,1651.45,1507.909,1658.563L1492.649,1665.675"></path><path marker-end="url(#mermaid-d99ceb03-aba1-47cd-9bb9-25765c89a5f3_flowchart-v2-pointEnd)" style="" class="edge-thickness-normal edge-pattern-dotted edge-thickness-normal edge-pattern-solid flowchart-link" id="L_NotifySvc_Loki_44" d="M795.422,1480.222L932.538,1491.685C1069.654,1503.148,1343.885,1526.074,1481.001,1543.704C1618.117,1561.333,1618.117,1573.667,1618.117,1586C1618.117,1598.333,1618.117,1610.667,1597.224,1624.895C1576.33,1639.124,1534.543,1655.247,1513.649,1663.309L1492.755,1671.371"></path></g><g class="edgeLabels"><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g transform="translate(1010.66796875, 656)" class="edgeLabel"><g transform="translate(-17.921875, -12)" class="label"><foreignObject height="24" width="35.84375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>gRPC</p></span></div></foreignObject></g></g><g transform="translate(1170.296875, 832)" class="edgeLabel"><g transform="translate(-17.921875, -12)" class="label"><foreignObject height="24" width="35.84375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>gRPC</p></span></div></foreignObject></g></g><g transform="translate(1326.640625, 1008)" class="edgeLabel"><g transform="translate(-17.921875, -12)" class="label"><foreignObject height="24" width="35.84375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>gRPC</p></span></div></foreignObject></g></g><g transform="translate(1633.9453125, 1184)" class="edgeLabel"><g transform="translate(-17.921875, -12)" class="label"><foreignObject height="24" width="35.84375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>gRPC</p></span></div></foreignObject></g></g><g transform="translate(827.4609375, 1360)" class="edgeLabel"><g transform="translate(-17.921875, -12)" class="label"><foreignObject height="24" width="35.84375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>gRPC</p></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g transform="translate(2035.7109375, 1586)" class="edgeLabel"><g transform="translate(-16, -12)" class="label"><foreignObject height="24" width="32"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>缓存</p></span></div></foreignObject></g></g><g transform="translate(2294.296875, 1586)" class="edgeLabel"><g transform="translate(-16, -12)" class="label"><foreignObject height="24" width="32"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>存储</p></span></div></foreignObject></g></g><g transform="translate(54.765625, 1096)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g transform="translate(101.53125, 1184)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g transform="translate(148.296875, 1272)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g transform="translate(195.0625, 1360)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g transform="translate(241.828125, 1473)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g transform="translate(288.59375, 1586)" class="edgeLabel"><g transform="translate(-26.765625, -12)" class="label"><foreignObject height="24" width="53.53125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>metrics</p></span></div></foreignObject></g></g><g class="edgeLabel"><g transform="translate(0, 0)" class="label"><foreignObject height="0" width="0"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"></span></div></foreignObject></g></g><g transform="translate(335.359375, 1096)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(377.40625, 1184)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(419.453125, 1272)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(461.5, 1360)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(503.546875, 1473)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(545.59375, 1586)" class="edgeLabel"><g transform="translate(-22.046875, -12)" class="label"><foreignObject height="24" width="44.09375"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>traces</p></span></div></foreignObject></g></g><g transform="translate(587.640625, 1096)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g><g transform="translate(959.03125, 1184)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g><g transform="translate(1412.9140625, 1272)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g><g transform="translate(1446.8203125, 1360)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g><g transform="translate(1584.2109375, 1473)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g><g transform="translate(1618.1171875, 1586)" class="edgeLabel"><g transform="translate(-13.90625, -12)" class="label"><foreignObject height="24" width="27.8125"><div style="display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;" class="labelBkg" xmlns="http://www.w3.org/1999/xhtml"><span class="edgeLabel"><p>logs</p></span></div></foreignObject></g></g></g><g class="nodes"><g transform="translate(817.91796875, 72)" id="flowchart-Qt6-448" class="node default client"><rect height="78" width="237.390625" y="-39" x="-118.6953125" style="fill:#e1f5ff !important;stroke:#01579b !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-88.6953125, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="177.390625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Qt6 桌面客户端<br>(Windows/Linux/macOS)</p></span></div></foreignObject></g></g><g transform="translate(1079.45703125, 72)" id="flowchart-Web-449" class="node default client"><rect height="78" width="185.6875" y="-39" x="-92.84375" style="fill:#e1f5ff !important;stroke:#01579b !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-62.84375, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="125.6875"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>🌐 Web 浏览器<br>(Chrome/Firefox)</p></span></div></foreignObject></g></g><g transform="translate(1301.58203125, 72)" id="flowchart-Mobile-450" class="node default client"><rect height="78" width="158.5625" y="-39" x="-79.28125" style="fill:#e1f5ff !important;stroke:#01579b !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-49.28125, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="98.5625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>📱 移动端<br>(iOS/Android)</p></span></div></foreignObject></g></g><g transform="translate(1059.75, 250)" id="flowchart-Nginx-451" class="node default gateway"><rect height="78" width="168.390625" y="-39" x="-84.1953125" style="fill:#fff3e0 !important;stroke:#e65100 !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-54.1953125, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="108.390625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Nginx 负载均衡<br>(8800/8443)</p></span></div></foreignObject></g></g><g transform="translate(1059.75, 378)" id="flowchart-APIGateway-452" class="node default gateway"><rect height="78" width="184.53125" y="-39" x="-92.265625" style="fill:#fff3e0 !important;stroke:#e65100 !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-62.265625, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="124.53125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>API 网关<br>(路由/限流/认证)</p></span></div></foreignObject></g></g><g transform="translate(802.6796875, 568)" id="flowchart-UserSvc-453" class="node default service"><rect height="102" width="204.78125" y="-51" x="-102.390625" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-72.390625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="144.78125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>👤 用户服务<br>:8080<br>认证/授权/用户管理</p></span></div></foreignObject></g></g><g transform="translate(959.03125, 744)" id="flowchart-MeetingSvc-454" class="node default service"><rect height="102" width="212.390625" y="-51" x="-106.1953125" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-76.1953125, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="152.390625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>📞 会议服务<br>:8082<br>会议管理/参与者管理</p></span></div></foreignObject></g></g><g transform="translate(1119.1875, 920)" id="flowchart-SignalSvc-455" class="node default service"><rect height="102" width="210.28125" y="-51" x="-105.140625" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-75.140625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="150.28125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>📡 信令服务<br>:8081<br>WebSocket/媒体协商</p></span></div></foreignObject></g></g><g transform="translate(1278.2890625, 1096)" id="flowchart-MediaSvc-456" class="node default service"><rect height="102" width="199.25" y="-51" x="-99.625" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-69.625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="139.25"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>🎬 媒体服务<br>:8083<br>SFU转发/录制/转码</p></span></div></foreignObject></g></g><g transform="translate(1584.2109375, 1272)" id="flowchart-AISvc-457" class="node default service"><rect height="102" width="204.78125" y="-51" x="-102.390625" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-72.390625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="144.78125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>🤖 AI检测服务<br>:8084<br>情感/合成/音频处理</p></span></div></foreignObject></g></g><g transform="translate(709.03125, 1473)" id="flowchart-NotifySvc-458" class="node default service"><rect height="102" width="172.78125" y="-51" x="-86.390625" style="fill:#f3e5f5 !important;stroke:#4a148c !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-56.390625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="112.78125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>🔔 通知服务<br>:8085<br>邮件/短信/推送</p></span></div></foreignObject></g></g><g transform="translate(2590.7578125, 1473)" id="flowchart-ModelMgr-459" class="node default ai"><rect height="78" width="216.53125" y="-39" x="-108.265625" style="fill:#e8f5e9 !important;stroke:#1b5e20 !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-78.265625, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="156.53125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>模型管理器<br>(加载/卸载/版本管理)</p></span></div></foreignObject></g></g><g transform="translate(2833.328125, 1473)" id="flowchart-InferEngine-460" class="node default ai"><rect height="78" width="168.609375" y="-39" x="-84.3046875" style="fill:#e8f5e9 !important;stroke:#1b5e20 !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-54.3046875, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="108.609375"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>推理引擎<br>(C++/GPU优化)</p></span></div></foreignObject></g></g><g transform="translate(2833.328125, 1699)" id="flowchart-InferCluster-461" class="node default ai"><rect height="78" width="192.140625" y="-39" x="-96.0703125" style="fill:#e8f5e9 !important;stroke:#1b5e20 !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-66.0703125, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="132.140625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>推理节点集群<br>(分布式/负载均衡)</p></span></div></foreignObject></g></g><g transform="translate(2035.7109375, 1473)" id="flowchart-PostgreSQL-462" class="node default data"><rect height="102" width="220.78125" y="-51" x="-110.390625" style="fill:#fce4ec !important;stroke:#880e4f !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-80.390625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="160.78125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>🗄️ PostgreSQL<br>(主数据库)<br>用户/会议/参与者数据</p></span></div></foreignObject></g></g><g transform="translate(2017.7109375, 1699)" id="flowchart-Redis-463" class="node default data"><rect height="102" width="183.65625" y="-51" x="-91.828125" style="fill:#fce4ec !important;stroke:#880e4f !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-61.828125, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="123.65625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>⚡ Redis<br>(缓存/队列)<br>Session/消息队列</p></span></div></foreignObject></g></g><g transform="translate(2294.296875, 1473)" id="flowchart-MongoDB-464" class="node default data"><rect height="102" width="196.390625" y="-51" x="-98.1953125" style="fill:#fce4ec !important;stroke:#880e4f !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-68.1953125, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="136.390625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>📊 MongoDB<br>(AI数据)<br>推理结果/分析数据</p></span></div></foreignObject></g></g><g transform="translate(2294.296875, 1699)" id="flowchart-MinIO-465" class="node default data"><rect height="102" width="172.78125" y="-51" x="-86.390625" style="fill:#fce4ec !important;stroke:#880e4f !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-56.390625, -36)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="72" width="112.78125"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>📦 MinIO<br>(对象存储)<br>录制/媒体/头像</p></span></div></foreignObject></g></g><g transform="translate(162.9609375, 1699)" id="flowchart-Prometheus-466" class="node default obs"><rect height="78" width="144.015625" y="-39" x="-72.0078125" style="fill:#f1f8e9 !important;stroke:#33691e !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-42.0078125, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="84.015625"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Prometheus<br>(监控)</p></span></div></foreignObject></g></g><g transform="translate(162.9609375, 1864)" id="flowchart-Grafana-467" class="node default obs"><rect height="78" width="119.75" y="-39" x="-59.875" style="fill:#f1f8e9 !important;stroke:#33691e !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-29.875, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="59.75"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Grafana<br>(可视化)</p></span></div></foreignObject></g></g><g transform="translate(431.7578125, 1699)" id="flowchart-Jaeger-468" class="node default obs"><rect height="78" width="151.75" y="-39" x="-75.875" style="fill:#f1f8e9 !important;stroke:#33691e !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-45.875, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="91.75"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Jaeger<br>(分布式追踪)</p></span></div></foreignObject></g></g><g transform="translate(1421.1484375, 1699)" id="flowchart-Loki-469" class="node default obs"><rect height="78" width="135.75" y="-39" x="-67.875" style="fill:#f1f8e9 !important;stroke:#33691e !important;stroke-width:2px !important" class="basic label-container"></rect><g transform="translate(-37.875, -24)" style="color:#000 !important" class="label"><rect></rect><foreignObject height="48" width="75.75"><div xmlns="http://www.w3.org/1999/xhtml" style="color: rgb(0, 0, 0) !important; display: table-cell; white-space: nowrap; line-height: 1.5; max-width: 200px; text-align: center;"><span class="nodeLabel" style="color:#000 !important"><p>Loki<br>(日志聚合)</p></span></div></foreignObject></g></g></g></g></g></svg>

### 服务职责

| 服务 | 端口 | 职责 | 依赖 |
|------|------|------|------|
| **user-service** | 8080 | 用户认证、资料管理、权限控制 | PostgreSQL, Redis, etcd |
| **meeting-service** | 8082 | 会议创建、管理、参与者控制 | PostgreSQL, Redis, etcd |
| **signaling-service** | 8081 | WebSocket 信令、房间管理 | Redis, etcd |
| **media-service** | 8083 | SFU 媒体转发、录制、存储 | PostgreSQL, MinIO |
| **ai-service** | 8084 | AI 分析请求、结果管理 | MongoDB, ZMQ |
| **ai-inference-service** | 8085 | AI 模型推理、ZMQ 通信 | PostgreSQL, Redis, ZMQ |

---

## 🛠️ 技术栈

### 核心框架
| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.24.0+ | 主要开发语言 |
| **Gin** | 1.9.1 | HTTP Web 框架 |
| **GORM** | 1.31.0 | ORM 数据库框架 |
| **gRPC** | 1.75.1 | 服务间 RPC 通信 |

### 通信协议
| 技术 | 版本 | 用途 |
|------|------|------|
| **WebSocket** | gorilla/websocket 1.5.3 | 实时信令通信 |
| **ZeroMQ** | pebbe/zmq4 1.4.0 | AI 服务高性能通信 |
| **HTTP/2** | - | RESTful API |

### 数据存储
| 技术 | 版本 | 用途 |
|------|------|------|
| **PostgreSQL** | 15-alpine | 用户数据、会议数据 |
| **Redis** | 7-alpine | 缓存、消息队列、会话 |
| **MongoDB** | 6.0.14 | AI 分析结果存储 |
| **MinIO** | latest | 对象存储（录制文件） |

### 基础设施
| 技术 | 版本 | 用途 |
|------|------|------|
| **etcd** | 3.6.5 | 服务注册与发现 |
| **Nginx** | alpine | API 网关、反向代理 |
| **Docker** | 20.0+ | 容器化部署 |

### 监控与追踪
| 技术 | 版本 | 用途 |
|------|------|------|
| **Prometheus** | 2.48.0 | 指标收集 |
| **Jaeger** | 1.51 | 分布式追踪 |
| **Grafana** | 10.2.2 | 可视化面板 |
| **Loki** | 2.9.3 | 日志聚合 |

---

## 🚀 快速开始

### 环境要求

- **Docker**: 20.0+
- **Docker Compose**: 2.0+
- **Go**: 1.24.0+ (本地开发)
- **Make**: (可选)

### 一键启动（Docker Compose）

```bash
# 1. 进入项目目录
cd meeting-system

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f user-service
```

### 本地开发启动

```bash
# 1. 启动基础设施服务
docker-compose up -d postgres redis mongodb minio etcd jaeger

# 2. 编译并启动用户服务
cd backend/user-service
go build -o user-service
./user-service -config=../config/config.yaml

# 3. 启动其他服务
cd ../meeting-service
go run main.go -config=../config/meeting-service.yaml

cd ../signaling-service
go run main.go -config=../config/signaling-service.yaml

# 或使用脚本启动所有服务
cd ../scripts
./start_all_services.sh
```

### 验证服务

```bash
# 检查用户服务健康状态
curl http://localhost:8080/health

# 检查会议服务
curl http://localhost:8082/health

# 检查信令服务
curl http://localhost:8081/health

# 查看 Prometheus 指标
curl http://localhost:8080/metrics
```

---

## 🔍 服务详解

### 1. User Service (用户服务)

**端口**: 8080
**职责**: 用户认证、资料管理、权限控制

**主要功能**:
- ✅ 用户注册与登录
- ✅ JWT Token 生成与验证
- ✅ CSRF 保护
- ✅ 用户资料 CRUD
- ✅ 头像上传
- ✅ 密码修改
- ✅ 用户封禁/解封（管理员）
- ✅ 请求限流

**技术实现**:
- Gin Web 框架
- GORM ORM
- JWT 认证 (golang-jwt/jwt v5)
- PostgreSQL 用户数据存储
- Redis 会话缓存
- etcd 服务注册

**API 端点**:
```
POST   /api/v1/register          # 用户注册
POST   /api/v1/login             # 用户登录
POST   /api/v1/refresh-token     # 刷新 Token
GET    /api/v1/profile           # 获取用户资料
PUT    /api/v1/profile           # 更新用户资料
POST   /api/v1/change-password   # 修改密码
POST   /api/v1/upload-avatar     # 上传头像
DELETE /api/v1/account           # 删除账户
GET    /api/v1/admin/users       # 管理员：用户列表
```

**配置文件**: `backend/config/config.yaml`

---

### 2. Meeting Service (会议服务)

**端口**: 8082
**职责**: 会议管理、参与者控制

**主要功能**:
- ✅ 会议创建/删除
- ✅ 会议列表查询
- ✅ 参与者加入/离开
- ✅ 参与者管理（踢出、静音）
- ✅ 会议状态管理
- ✅ 会议权限控制

**技术实现**:
- Gin Web 框架
- GORM ORM
- PostgreSQL 会议数据存储
- Redis 会议状态缓存
- gRPC 服务间通信
- etcd 服务注册

**API 端点**:
```
POST   /api/v1/meetings                    # 创建会议
GET    /api/v1/meetings                    # 获取会议列表
GET    /api/v1/meetings/:id                # 获取会议详情
PUT    /api/v1/meetings/:id                # 更新会议
DELETE /api/v1/meetings/:id                # 删除会议
POST   /api/v1/meetings/:id/join           # 加入会议
POST   /api/v1/meetings/:id/leave          # 离开会议
GET    /api/v1/meetings/:id/participants   # 参与者列表
POST   /api/v1/meetings/:id/participants/:uid/kick  # 踢出参与者
```

**配置文件**: `backend/config/meeting-service.yaml`

---

### 3. Signaling Service (信令服务)

**端口**: 8081
**职责**: WebSocket 信令、房间管理

**主要功能**:
- ✅ WebSocket 连接管理
- ✅ 信令消息转发（offer/answer/candidate）
- ✅ 房间状态管理
- ✅ 客户端心跳检测
- ✅ 连接统计

**技术实现**:
- Gin Web 框架
- gorilla/websocket
- Redis Pub/Sub 消息分发
- 内存房间管理
- etcd 服务注册

**WebSocket 协议**:
```json
// 客户端 -> 服务器
{
  "type": "join",
  "room_id": "meeting-123",
  "user_id": "user-456"
}

{
  "type": "offer",
  "target": "user-789",
  "sdp": "..."
}

{
  "type": "candidate",
  "target": "user-789",
  "candidate": "..."
}

// 服务器 -> 客户端
{
  "type": "user-joined",
  "user_id": "user-789",
  "user_info": {...}
}

{
  "type": "offer",
  "from": "user-456",
  "sdp": "..."
}
```

**API 端点**:
```
GET    /ws/signaling             # WebSocket 连接
GET    /api/v1/stats             # 统计信息
GET    /api/v1/rooms/stats       # 房间统计
```

**配置文件**: `backend/config/signaling-service.yaml`

---

### 4. Media Service (媒体服务)

**端口**: 8083
**职责**: SFU 媒体转发、录制、存储

**主要功能**:
- ✅ 媒体文件上传/下载
- ✅ 会议录制
- ✅ MinIO 对象存储集成
- ✅ 录制文件管理

**技术实现**:
- Gin Web 框架
- MinIO Go SDK
- PostgreSQL 媒体元数据
- FFmpeg 媒体处理（计划）

**API 端点**:
```
POST   /api/v1/media/upload      # 上传媒体文件
GET    /api/v1/media/:id         # 获取媒体文件
DELETE /api/v1/media/:id         # 删除媒体文件
GET    /api/v1/recordings        # 录制列表
POST   /api/v1/recordings/start  # 开始录制
POST   /api/v1/recordings/stop   # 停止录制
```

**配置文件**: `backend/config/media-service.yaml`

---

### 5. AI Service (AI 服务)

**端口**: 8084
**职责**: AI 分析请求、结果管理

**主要功能**:
- ✅ AI 分析任务提交
- ✅ 分析结果查询
- ✅ MongoDB 结果存储

**技术实现**:
- Gin Web 框架
- MongoDB Go Driver
- ZMQ 通信（与 AI Inference Service）

**API 端点**:
```
POST   /api/v1/ai/analyze        # 提交分析任务
GET    /api/v1/ai/results/:id    # 获取分析结果
```

**配置文件**: `backend/config/ai-service.yaml`

---

### 6. AI Inference Service (AI 推理服务)

**端口**: 8085
**职责**: AI 模型推理、ZMQ 通信

**主要功能**:
- ✅ 推理任务调度
- ✅ ZMQ 连接 Unit Manager
- ✅ 模型列表查询
- ✅ 推理结果返回

**技术实现**:
- Gin Web 框架
- ZeroMQ (pebbe/zmq4)
- 连接宿主机 Unit Manager (:19001)

**API 端点**:
```
POST   /api/v1/inference/submit  # 提交推理任务
GET    /api/v1/inference/:id     # 获取推理结果
GET    /api/v1/models            # 可用模型列表
```

**配置文件**: `backend/config/ai-inference-service.yaml`

---

## 🗄️ 数据库设计

### PostgreSQL 表结构

#### users 表（用户表）
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### meetings 表（会议表）
```sql
CREATE TABLE meetings (
    id SERIAL PRIMARY KEY,
    meeting_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'scheduled',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    max_participants INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### participants 表（参与者表）
```sql
CREATE TABLE participants (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER REFERENCES meetings(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'participant',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    UNIQUE(meeting_id, user_id)
);
```

### Redis 数据结构

```
# 用户会话
session:{user_id} -> {token, expires_at}

# 会议状态
meeting:{meeting_id}:status -> {active|ended}
meeting:{meeting_id}:participants -> Set{user_id1, user_id2, ...}

# 在线用户
online:users -> Set{user_id1, user_id2, ...}

# 限流
ratelimit:{user_id}:{endpoint} -> counter
```

### MongoDB 集合

```javascript
// AI 分析结果
{
  _id: ObjectId,
  task_id: "task-123",
  meeting_id: "meeting-456",
  user_id: "user-789",
  type: "emotion|transcription|quality",
  result: {...},
  created_at: ISODate
}
```

## 📝 API 文档

### 通用响应格式

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {...}
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "error message",
  "error": "detailed error"
}
```

### 认证方式

所有需要认证的接口都需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <jwt_token>
```

### 用户服务 API

详见 [服务详解 - User Service](#1-user-service-用户服务)

### 会议服务 API

详见 [服务详解 - Meeting Service](#2-meeting-service-会议服务)

### 信令服务 API

详见 [服务详解 - Signaling Service](#3-signaling-service-信令服务)

---

## ⚙️ 配置说明

### 配置文件位置

所有配置文件位于 `backend/config/` 目录：

```
backend/config/
├── config.yaml                 # user-service 配置
├── meeting-service.yaml        # meeting-service 配置
├── signaling-service.yaml      # signaling-service 配置
├── media-service.yaml          # media-service 配置
├── ai-service.yaml             # ai-service 配置
└── ai-inference-service.yaml   # ai-inference-service 配置
```

### 配置文件示例 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug | release

database:
  host: "postgres"
  port: 5432
  user: "postgres"
  password: "password"
  dbname: "meeting_system"
  sslmode: "disable"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: "redis"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

etcd:
  endpoints:
    - "etcd:2379"
  dial_timeout: 5

jwt:
  secret: "your-secret-key-change-in-production"
  expire_hours: 24
  refresh_expire_hours: 168

log:
  level: "info"
  filename: "logs/user-service.log"
  max_size: 100
  max_age: 30
  max_backups: 10
  compress: true

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allowed_headers:
    - "Origin"
    - "Content-Type"
    - "Authorization"
```

### 环境变量

可以通过环境变量覆盖配置文件：

```bash
# 数据库配置
export DATABASE_HOST=postgres
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=password

# Redis 配置
export REDIS_HOST=redis
export REDIS_PORT=6379

# JWT 配置
export JWT_SECRET=your-super-secret-key

# etcd 配置
export ETCD_ENDPOINTS=etcd:2379

# ZMQ 配置（AI 服务）
export ZMQ_UNIT_MANAGER_HOST=host.docker.internal
export ZMQ_UNIT_MANAGER_PORT=19001
```

---

## 🐳 部署指南

### Docker Compose 部署

**完整部署**（推荐）:
```bash
cd meeting-system
docker-compose up -d
```

**分步部署**:
```bash
# 1. 启动基础设施
docker-compose up -d postgres redis mongodb minio etcd

# 2. 启动监控服务
docker-compose up -d prometheus grafana jaeger loki promtail

# 3. 启动业务服务
docker-compose up -d user-service meeting-service signaling-service media-service

# 4. 启动 AI 服务
docker-compose up -d ai-service ai-inference-service

# 5. 启动网关
docker-compose up -d nginx
```

### 服务健康检查

```bash
# 检查所有服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f user-service

# 检查服务健康
curl http://localhost:8800/api/v1/health
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

---

## 🔧 开发指南

### 添加新的微服务

1. **创建服务目录**:
```bash
cd backend
mkdir new-service
cd new-service
```

2. **初始化 Go 模块**:
```bash
go mod init meeting-system/new-service
```

3. **创建 main.go**:
```go
package main

import (
    "github.com/gin-gonic/gin"
    "meeting-system/shared/config"
    "meeting-system/shared/logger"
)

func main() {
    config.InitConfig("../config/new-service.yaml")
    logger.InitLogger(...)

    r := gin.Default()
    r.GET("/health", healthCheck)
    r.Run(":8086")
}
```

4. **创建 Dockerfile**:
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o new-service

FROM alpine:latest
COPY --from=builder /app/new-service /app/
CMD ["/app/new-service"]
```

5. **添加到 docker-compose.yml**:
```yaml
new-service:
  build:
    context: ./backend
    dockerfile: new-service/Dockerfile
  container_name: meeting-new-service
  ports:
    - "8086:8086"
  networks:
    - meeting-network
```

### 共享库使用

所有微服务共享 `backend/shared/` 目录下的库：

```go
import (
    "meeting-system/shared/config"      // 配置管理
    "meeting-system/shared/database"    // 数据库连接
    "meeting-system/shared/logger"      // 日志工具
    "meeting-system/shared/middleware"  // Gin 中间件
    "meeting-system/shared/models"      // 数据模型
    "meeting-system/shared/discovery"   // 服务发现
    "meeting-system/shared/metrics"     // Prometheus 指标
    "meeting-system/shared/tracing"     // Jaeger 追踪
)
```

### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 遵循 Go 官方代码规范
- 添加必要的注释和文档

---

## 🧪 测试

### 单元测试

```bash
cd backend/user-service
go test ./... -v
```

### 集成测试

```bash
cd meeting-system/scripts
./test_integration.sh
```

### E2E 测试

```bash
cd meeting-system/scripts
./run_e2e_test.sh
```

### 压力测试

```bash
cd backend/stress-test
go run main.go -config=../config/stress-test-config.yaml
```

---

## 📊 监控与日志

### Prometheus 指标

访问: http://localhost:8801

**可用指标**:
- `http_requests_total`: HTTP 请求总数
- `http_request_duration_seconds`: 请求延迟
- `grpc_server_handled_total`: gRPC 调用统计
- `db_connections`: 数据库连接数
- `active_users`: 在线用户数
- `active_meetings`: 活跃会议数

### Grafana 面板

访问: http://localhost:8804 (admin/admin123)

**预配置面板**:
1. 服务概览
2. 数据库性能
3. Redis 性能
4. 系统资源
5. 业务指标

### Jaeger 追踪

访问: http://localhost:8803

查看分布式调用链路和性能分析。

### Loki 日志

在 Grafana 中通过 Explore 查询日志：

```
{container_name="meeting-user-service"} |= "error"
```

---

## 🔗 相关链接

- [项目主页](https://github.com/gugugu5331/VideoCall-System)
- [Qt6 客户端文档](../qt6-client/README.md)
- [部署文档](docs/deployment/)
- [测试文档](docs/testing/)
- [Edge-LLM-Infra](Edge-LLM-Infra-master/)

---

## 📄 许可证

MIT License

---

**最后更新**: 2025-10-08
