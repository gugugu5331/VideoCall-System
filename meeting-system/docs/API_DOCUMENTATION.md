# Meeting System API Documentation

**Version**: 1.0  
**Base URL**: http://js1.blockelite.cn:21058  
**Last Updated**: 2025-10-05

---

## Table of Contents

1. [Overview](#overview)
2. [Monitoring Services](#monitoring-services)
3. [Microservices API](#microservices-api)
4. [Authentication](#authentication)
5. [Health & Metrics](#health--metrics)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)

---

## Overview

The Meeting System is a distributed microservices architecture providing video conferencing capabilities with AI-powered features. All services are accessible through the Nginx gateway at `http://js1.blockelite.cn:21058`.

### Architecture

- **5 Microservices**: User, Meeting, Signaling, Media, AI
- **Monitoring Stack**: Prometheus, Grafana, Jaeger, Alertmanager, Loki
- **Infrastructure**: PostgreSQL, Redis, etcd, MongoDB, MinIO

---

## Monitoring Services

### Prometheus (Metrics Collection)

**URL**: http://js1.blockelite.cn:21059  
**Purpose**: Time-series metrics collection and querying

#### Key Endpoints

**Query Metrics**
```bash
curl -s "http://js1.blockelite.cn:21059/api/v1/query?query=up"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {"job": "user-service", "instance": "user-service:8080"},
        "value": [1728134400, "1"]
      }
    ]
  }
}
```

**View Monitoring Targets**
```bash
curl -s "http://js1.blockelite.cn:21059/api/v1/targets"
```

**Common Queries**:
- `up` - Service availability (1=up, 0=down)
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `go_goroutines` - Active goroutines

---

### Grafana (Visualization)

**URL**: http://js1.blockelite.cn:21062  
**Credentials**: `admin` / `admin123`  
**Purpose**: Metrics visualization and dashboards

#### Features
- Pre-configured Prometheus datasource
- Pre-configured Loki datasource for logs
- Real-time metrics visualization
- Custom dashboard creation

#### Health Check
```bash
curl -s "http://js1.blockelite.cn:21062/api/health"
```

**Response**:
```json
{
  "commit": "161e3cac5075540918e3a39004f2364ad104d5bb",
  "database": "ok",
  "version": "10.2.2"
}
```

---

### Jaeger (Distributed Tracing)

**URL**: http://js1.blockelite.cn:21061  
**Purpose**: Distributed request tracing across microservices

#### Key Endpoints

**List Services**
```bash
curl -s "http://js1.blockelite.cn:21061/api/services"
```

**Response**:
```json
{
  "data": [
    "user-service",
    "meeting-service",
    "signaling-service",
    "media-service",
    "ai-service",
    "jaeger-all-in-one"
  ],
  "total": 6
}
```

**Search Traces**
```bash
curl -s "http://js1.blockelite.cn:21061/api/traces?service=user-service&limit=10"
```

---

### Alertmanager (Alert Management)

**URL**: http://js1.blockelite.cn:21060  
**Purpose**: Alert routing and notification

#### Key Endpoints

**View Active Alerts**
```bash
curl -s "http://js1.blockelite.cn:21060/api/v1/alerts"
```

**View Status**
```bash
curl -s "http://js1.blockelite.cn:21060/api/v1/status"
```

#### Configured Alert Rules
- **ServiceDown**: Service unavailable for >1 minute
- **HighCPUUsage**: CPU usage >80% for 5 minutes
- **HighMemoryUsage**: Memory usage >2GB for 5 minutes
- **HighErrorRate**: Error rate >5% for 2 minutes
- **SlowResponseTime**: 95th percentile >2s for 3 minutes

---

### Loki (Log Aggregation)

**URL**: http://js1.blockelite.cn:21063  
**Purpose**: Centralized log collection and querying

#### Key Endpoints

**Ready Check**
```bash
curl -s "http://js1.blockelite.cn:21063/ready"
```

**Response**: `ready` (when operational)

**Query Logs**
```bash
curl -s "http://js1.blockelite.cn:21063/loki/api/v1/query?query={container=\"meeting-user-service\"}"
```

---

## Microservices API

### User Service

**Base Path**: `/api/v1`  
**Internal Port**: 8080  
**Purpose**: User authentication and management

#### Endpoints

##### 1. Get CSRF Token
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/csrf-token"
```

**Response**:
```json
{
  "token": "csrf-token-value",
  "expires_at": "2025-10-05T13:00:00Z"
}
```

##### 2. User Registration
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Response** (201 Created):
```json
{
  "user_id": "uuid-here",
  "username": "testuser",
  "email": "test@example.com",
  "created_at": "2025-10-05T12:00:00Z"
}
```

##### 3. User Login
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePass123!"
  }'
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh-token-here",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

##### 4. User Logout
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/auth/logout" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

##### 5. Get User Profile
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/users/{user_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "user_id": "uuid-here",
  "username": "testuser",
  "email": "test@example.com",
  "avatar_url": "https://example.com/avatar.jpg",
  "created_at": "2025-10-05T12:00:00Z",
  "updated_at": "2025-10-05T12:30:00Z"
}
```

##### 6. Update User Profile
```bash
curl -X PUT "http://js1.blockelite.cn:21058/api/v1/users/{user_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newusername",
    "avatar_url": "https://example.com/new-avatar.jpg"
  }'
```

**Response** (200 OK):
```json
{
  "user_id": "uuid-here",
  "username": "newusername",
  "email": "test@example.com",
  "avatar_url": "https://example.com/new-avatar.jpg",
  "updated_at": "2025-10-05T13:00:00Z"
}
```

---

### Signaling Service

**Base Path**: `/api/v1`
**WebSocket Endpoint**: `ws://js1.blockelite.cn:21058/ws/signaling`
**Internal Port**: 8081
**Purpose**: WebRTC signaling for peer-to-peer connections and session management

---

#### WebSocket Signaling Protocol

##### Connection Establishment

```bash
# Using websocat (install: cargo install websocat)
websocat "ws://js1.blockelite.cn:21058/ws/signaling?token=YOUR_ACCESS_TOKEN"
```

**Connection Parameters**:
- `token` (query parameter, required): JWT access token for authentication

**Connection Flow**:
1. Client connects with valid JWT token
2. Server validates token and establishes WebSocket connection
3. Client sends `join` message to enter a meeting room
4. Server broadcasts `user-joined` to other participants
5. Clients exchange WebRTC signaling messages (offer, answer, ICE candidates)
6. Client sends `leave` message or disconnects to exit room

---

##### WebSocket Message Types

All messages are JSON-formatted with the following structure:

**Base Message Format**:
```json
{
  "type": "message_type",
  "room_id": "meeting-uuid",
  "user_id": "sender-user-uuid",
  "timestamp": "2025-10-05T14:00:00Z",
  "data": {}
}
```

---

##### Client → Server Messages

**1. Join Room**
```json
{
  "type": "join",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "user_info": {
    "username": "Alice",
    "avatar_url": "https://example.com/avatar.jpg"
  }
}
```

**Server Response**:
```json
{
  "type": "joined",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "participants": [
    {
      "user_id": "user-b",
      "username": "Bob",
      "joined_at": "2025-10-05T13:55:00Z"
    }
  ],
  "timestamp": "2025-10-05T14:00:00Z"
}
```

---

**2. WebRTC Offer**
```json
{
  "type": "offer",
  "room_id": "meeting-123",
  "from_user_id": "user-a",
  "target_user_id": "user-b",
  "sdp": {
    "type": "offer",
    "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\n..."
  }
}
```

**Server Action**: Forwards offer to `target_user_id`

---

**3. WebRTC Answer**
```json
{
  "type": "answer",
  "room_id": "meeting-123",
  "from_user_id": "user-b",
  "target_user_id": "user-a",
  "sdp": {
    "type": "answer",
    "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1\r\n..."
  }
}
```

**Server Action**: Forwards answer to `target_user_id`

---

**4. ICE Candidate**
```json
{
  "type": "ice-candidate",
  "room_id": "meeting-123",
  "from_user_id": "user-a",
  "target_user_id": "user-b",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

**Server Action**: Forwards ICE candidate to `target_user_id`

---

**5. Media State Update**
```json
{
  "type": "media-state",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "audio_enabled": true,
  "video_enabled": false,
  "screen_sharing": false
}
```

**Server Action**: Broadcasts media state to all participants in the room

---

**6. Chat Message**
```json
{
  "type": "chat",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "message": "Hello everyone!",
  "timestamp": "2025-10-05T14:05:00Z"
}
```

**Server Action**: Broadcasts chat message to all participants

---

**7. Leave Room**
```json
{
  "type": "leave",
  "room_id": "meeting-123",
  "user_id": "user-a"
}
```

**Server Response**:
```json
{
  "type": "left",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "timestamp": "2025-10-05T14:30:00Z"
}
```

---

##### Server → Client Messages

**1. User Joined Notification**
```json
{
  "type": "user-joined",
  "room_id": "meeting-123",
  "user_id": "user-c",
  "user_info": {
    "username": "Charlie",
    "avatar_url": "https://example.com/avatar3.jpg"
  },
  "timestamp": "2025-10-05T14:10:00Z"
}
```

---

**2. User Left Notification**
```json
{
  "type": "user-left",
  "room_id": "meeting-123",
  "user_id": "user-b",
  "timestamp": "2025-10-05T14:30:00Z"
}
```

---

**3. Error Message**
```json
{
  "type": "error",
  "code": "INVALID_ROOM",
  "message": "Meeting room does not exist",
  "timestamp": "2025-10-05T14:00:00Z"
}
```

**Error Codes**:
- `INVALID_ROOM`: Room does not exist
- `UNAUTHORIZED`: Invalid or expired token
- `ROOM_FULL`: Maximum participants reached
- `INVALID_MESSAGE`: Malformed message format
- `TARGET_NOT_FOUND`: Target user not in room

---

**4. Heartbeat (Keep-Alive)**
```json
{
  "type": "ping",
  "timestamp": "2025-10-05T14:15:00Z"
}
```

**Client Response**:
```json
{
  "type": "pong",
  "timestamp": "2025-10-05T14:15:00Z"
}
```

**Heartbeat Interval**: 30 seconds
**Connection Timeout**: 90 seconds (3 missed heartbeats)

---

#### HTTP REST API Endpoints

##### 1. Get Session Information
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/sessions/{session_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "data": {
    "session_id": "session-uuid",
    "user_id": "user-uuid",
    "meeting_id": 123,
    "connected_at": "2025-10-05T14:00:00Z",
    "last_activity": "2025-10-05T14:25:00Z",
    "status": "active",
    "peer_connections": 3,
    "media_state": {
      "audio_enabled": true,
      "video_enabled": true,
      "screen_sharing": false
    }
  }
}
```

**Error Response** (401 Unauthorized):
```json
{
  "code": 401,
  "message": "Missing authorization header",
  "timestamp": "2025-10-05T14:00:00Z"
}
```

---

##### 2. Get Room Sessions
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/sessions/room/{meeting_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "data": [
    {
      "session_id": "session-001",
      "user_id": "user-a",
      "username": "Alice",
      "connected_at": "2025-10-05T13:55:00Z",
      "status": "active",
      "media_state": {
        "audio_enabled": true,
        "video_enabled": true
      }
    },
    {
      "session_id": "session-002",
      "user_id": "user-b",
      "username": "Bob",
      "connected_at": "2025-10-05T14:00:00Z",
      "status": "active",
      "media_state": {
        "audio_enabled": true,
        "video_enabled": false
      }
    }
  ]
}
```

---

##### 3. Get Message History
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/messages/history/{meeting_id}?limit=50" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Query Parameters**:
- `limit` (integer, optional): Maximum number of messages to retrieve (default: 100, max: 1000)

**Response** (200 OK):
```json
{
  "data": [
    {
      "message_id": "msg-001",
      "type": "offer",
      "from_user_id": "user-a",
      "to_user_id": "user-b",
      "meeting_id": 123,
      "payload": {"sdp": "..."},
      "created_at": "2025-10-05T14:00:00Z"
    },
    {
      "message_id": "msg-002",
      "type": "answer",
      "from_user_id": "user-b",
      "to_user_id": "user-a",
      "meeting_id": 123,
      "payload": {"sdp": "..."},
      "created_at": "2025-10-05T14:00:05Z"
    }
  ]
}
```

---

##### 4. Get Statistics Overview
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/stats/overview" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "active_sessions": 15,
  "connected_clients": 15,
  "active_rooms": 3,
  "room_details": {
    "meeting-123": {
      "participant_count": 5,
      "created_at": "2025-10-05T13:00:00Z"
    },
    "meeting-456": {
      "participant_count": 7,
      "created_at": "2025-10-05T13:30:00Z"
    },
    "meeting-789": {
      "participant_count": 3,
      "created_at": "2025-10-05T14:00:00Z"
    }
  }
}
```

---

##### 5. Get Room Statistics
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/stats/rooms" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "data": {
    "meeting-123": {
      "room_id": "meeting-123",
      "participant_count": 5,
      "created_at": "2025-10-05T13:00:00Z",
      "duration": 5400,
      "total_messages": 1523,
      "active_connections": 5
    }
  }
}
```

---

##### 6. Cleanup Expired Sessions (Admin)
```bash
curl -X POST "http://js1.blockelite.cn:21058/admin/cleanup/sessions" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Cleanup completed",
  "cleaned_sessions": 12,
  "timestamp": "2025-10-05T14:00:00Z"
}
```

---

##### 7. List All Sessions (Admin)
```bash
curl -X GET "http://js1.blockelite.cn:21058/admin/sessions" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "data": [
    {
      "session_id": "session-001",
      "user_id": "user-a",
      "meeting_id": 123,
      "status": "active",
      "connected_at": "2025-10-05T13:55:00Z"
    }
  ],
  "total": 15,
  "active": 15,
  "inactive": 0
}
```

---

#### WebSocket Connection Example (JavaScript)

```javascript
// Establish WebSocket connection
const token = "YOUR_ACCESS_TOKEN";
const ws = new WebSocket(`ws://js1.blockelite.cn:21058/ws/signaling?token=${token}`);

// Connection opened
ws.onopen = () => {
  console.log("WebSocket connected");

  // Join meeting room
  ws.send(JSON.stringify({
    type: "join",
    room_id: "meeting-123",
    user_id: "user-a",
    user_info: {
      username: "Alice"
    }
  }));
};

// Receive messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  switch (message.type) {
    case "joined":
      console.log("Joined room:", message.room_id);
      console.log("Participants:", message.participants);
      break;

    case "user-joined":
      console.log("User joined:", message.user_id);
      break;

    case "offer":
      // Handle WebRTC offer
      handleOffer(message.sdp, message.from_user_id);
      break;

    case "answer":
      // Handle WebRTC answer
      handleAnswer(message.sdp, message.from_user_id);
      break;

    case "ice-candidate":
      // Handle ICE candidate
      handleIceCandidate(message.candidate, message.from_user_id);
      break;

    case "user-left":
      console.log("User left:", message.user_id);
      break;

    case "error":
      console.error("Error:", message.message);
      break;

    case "ping":
      // Respond to heartbeat
      ws.send(JSON.stringify({ type: "pong" }));
      break;
  }
};

// Send WebRTC offer
function sendOffer(targetUserId, sdp) {
  ws.send(JSON.stringify({
    type: "offer",
    room_id: "meeting-123",
    from_user_id: "user-a",
    target_user_id: targetUserId,
    sdp: sdp
  }));
}

// Leave room
function leaveRoom() {
  ws.send(JSON.stringify({
    type: "leave",
    room_id: "meeting-123",
    user_id: "user-a"
  }));
  ws.close();
}

// Connection closed
ws.onclose = () => {
  console.log("WebSocket disconnected");
};

// Connection error
ws.onerror = (error) => {
  console.error("WebSocket error:", error);
};
```

---

#### Signaling Service Summary

**Total Endpoints**: 7 HTTP REST + 1 WebSocket

**HTTP REST Endpoints**:
- Session Management: 2 endpoints
- Message History: 1 endpoint
- Statistics: 2 endpoints
- Admin Operations: 2 endpoints

**WebSocket Message Types**:
- Client → Server: 7 types (join, offer, answer, ice-candidate, media-state, chat, leave)
- Server → Client: 4 types (user-joined, user-left, error, ping)

**Connection Limits**:
- Max participants per room: 100 (configurable)
- Heartbeat interval: 30 seconds
- Connection timeout: 90 seconds
- Max message size: 64KB

**Rate Limits**:
- WebSocket connections: 30/minute (burst: 10)
- API endpoints: 100/minute (burst: 50)

---



---

### Media Service

**Base Path**: `/api/v1`
**Internal Port**: 8083
**Architecture**: **SFU (Selective Forwarding Unit)**
**Purpose**: Real-time media stream forwarding, WebRTC peer management, recording, and live streaming

---

#### SFU Architecture Overview

**Core Principles**:
- 🎯 **Stream Forwarding Only**: Media Service acts as a central media router that receives RTP/RTCP packets from publishers and selectively forwards them to subscribers
- ⚡ **No Transcoding**: Does NOT transcode, mix, or process media streams in real-time (ensures ultra-low latency)
- 📊 **Scalability**: Can handle hundreds of participants by avoiding CPU-intensive media processing
- 🔄 **Simulcast Support**: Supports multiple quality layers (simulcast) and SVC for adaptive bitrate streaming
- 🚀 **Low Latency**: Typical latency: 100-300ms (vs 3-10 seconds for traditional streaming)

**What SFU Does**:
- ✅ Forward RTP/RTCP packets between peers
- ✅ Manage WebRTC peer connections
- ✅ Track active media streams and participants
- ✅ Record meetings to persistent storage
- ✅ Push live streams to external RTMP/HLS endpoints

**What SFU Does NOT Do**:
- ❌ Transcode video codecs (e.g., H.264 → VP8)
- ❌ Mix multiple audio/video streams into one
- ❌ Apply real-time filters or effects
- ❌ Process media with AI (handled by AI Service)
- ❌ Store uploaded files (handled by MinIO directly)

**Note**: Some endpoints in the codebase (FFmpeg, AI processing, filters) exist for **offline/post-processing** of recordings, NOT for real-time stream manipulation.

---

#### Core SFU Endpoints (Real-Time Stream Forwarding)

---

#### WebRTC Peer Management (Core SFU Functions)

##### 1. Get Room Peers
**Purpose**: Retrieve all active WebRTC peers in a meeting room
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/webrtc/room/{roomId}/peers" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "room_id": "meeting-123",
  "peers": [
    {
      "peer_id": "peer-uuid-1",
      "user_id": "user-a",
      "media_type": "video",
      "status": "connected",
      "tracks": [
        {
          "track_id": "track-audio-1",
          "kind": "audio",
          "codec": "opus",
          "bitrate": 64000
        },
        {
          "track_id": "track-video-1",
          "kind": "video",
          "codec": "vp8",
          "resolution": "1280x720",
          "bitrate": 1500000,
          "frame_rate": 30
        }
      ],
      "connected_at": "2025-10-05T14:00:00Z",
      "last_activity": "2025-10-05T14:25:00Z"
    },
    {
      "peer_id": "peer-uuid-2",
      "user_id": "user-b",
      "media_type": "screen",
      "status": "connected",
      "tracks": [
        {
          "track_id": "track-screen-1",
          "kind": "video",
          "codec": "vp9",
          "resolution": "1920x1080",
          "bitrate": 2500000,
          "frame_rate": 15
        }
      ],
      "connected_at": "2025-10-05T14:05:00Z"
    }
  ],
  "total_peers": 2
}
```

**Peer Statuses**: connecting, connected, disconnected, failed

---

##### 2. Get Room Statistics
**Purpose**: Get real-time statistics for a meeting room (bandwidth, packet loss, jitter, etc.)
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/webrtc/room/{roomId}/stats" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "room_id": "meeting-123",
  "meeting_id": 123,
  "created_at": "2025-10-05T13:00:00Z",
  "duration": 5400,
  "is_recording": true,
  "recording_id": "recording-uuid",
  "statistics": {
    "total_peers": 5,
    "active_peers": 5,
    "total_tracks": 15,
    "audio_tracks": 5,
    "video_tracks": 8,
    "screen_tracks": 2,
    "total_bandwidth": 12500000,
    "inbound_bandwidth": 6250000,
    "outbound_bandwidth": 6250000,
    "packet_loss_rate": 0.02,
    "average_jitter": 15,
    "average_rtt": 45
  },
  "quality_metrics": {
    "excellent": 3,
    "good": 2,
    "fair": 0,
    "poor": 0
  }
}
```

**Metrics Explanation**:
- `total_bandwidth`: Total bandwidth usage in bits per second
- `packet_loss_rate`: Packet loss percentage (0-1)
- `average_jitter`: Average jitter in milliseconds
- `average_rtt`: Average round-trip time in milliseconds

---

##### 3. Update Peer Media State
**Purpose**: Update a peer's media state (mute/unmute audio/video, enable/disable screen sharing)
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/webrtc/peer/{peerId}/media" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "audio_enabled": true,
    "video_enabled": false,
    "screen_sharing": false
  }'
```

**Request Body**:
```json
{
  "audio_enabled": "boolean (optional)",
  "video_enabled": "boolean (optional)",
  "screen_sharing": "boolean (optional)",
  "video_quality": "string (optional): 'low', 'medium', 'high'"
}
```

**Response** (200 OK):
```json
{
  "peer_id": "peer-uuid",
  "user_id": "user-a",
  "media_state": {
    "audio_enabled": true,
    "video_enabled": false,
    "screen_sharing": false,
    "video_quality": "medium"
  },
  "updated_at": "2025-10-05T14:30:00Z"
}
```

---

#### Recording Management (Offline Processing)

**Note**: Recording captures real-time streams and saves them to persistent storage. The recording process runs **asynchronously** and does not affect real-time stream forwarding latency.

##### 4. Start Recording
**Purpose**: Start recording a meeting (captures all audio/video/screen streams)
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/recording/start" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "meeting_id": 123,
    "room_id": "meeting-123",
    "format": "mp4",
    "quality": "high",
    "record_audio": true,
    "record_video": true,
    "record_screen": true
  }'
```

**Request Body**:
```json
{
  "meeting_id": "integer (required)",
  "room_id": "string (required)",
  "format": "string (optional): 'mp4', 'webm', 'mkv' (default: 'mp4')",
  "quality": "string (optional): 'low', 'medium', 'high' (default: 'medium')",
  "record_audio": "boolean (optional, default: true)",
  "record_video": "boolean (optional, default: true)",
  "record_screen": "boolean (optional, default: true)",
  "layout": "string (optional): 'grid', 'speaker', 'gallery' (default: 'grid')"
}
```

**Response** (200 OK):
```json
{
  "recording_id": "recording-uuid",
  "meeting_id": 123,
  "room_id": "meeting-123",
  "status": "recording",
  "format": "mp4",
  "quality": "high",
  "started_at": "2025-10-05T14:00:00Z",
  "estimated_size": 0,
  "output_path": "/recordings/meeting-123/recording-uuid.mp4"
}
```

---

##### 5. Stop Recording
**Purpose**: Stop an active recording and finalize the output file
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/recording/stop" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recording_id": "recording-uuid"
  }'
```

**Response** (200 OK):
```json
{
  "recording_id": "recording-uuid",
  "status": "completed",
  "started_at": "2025-10-05T14:00:00Z",
  "stopped_at": "2025-10-05T15:30:00Z",
  "duration": 5400,
  "file_size": 524288000,
  "file_path": "/recordings/meeting-123/recording-uuid.mp4",
  "download_url": "http://js1.blockelite.cn:21058/api/v1/recording/download/recording-uuid"
}
```

---

##### 6. Get Recording Status
**Purpose**: Check the status of an ongoing or completed recording
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/recording/status/{recording_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "recording_id": "recording-uuid",
  "meeting_id": 123,
  "status": "recording",
  "started_at": "2025-10-05T14:00:00Z",
  "duration": 3600,
  "file_size": 314572800,
  "participants": 5,
  "quality_metrics": {
    "average_bitrate": 2500000,
    "peak_bitrate": 3500000,
    "dropped_frames": 12,
    "encoding_errors": 0
  }
}
```

**Recording Statuses**: pending, recording, paused, completed, failed, cancelled

---

##### 7. List Recordings
**Purpose**: Retrieve a list of recordings with optional filters
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/recording/list?meeting_id=123&status=completed&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Query Parameters**:
- `meeting_id` (integer, optional): Filter by meeting ID
- `status` (string, optional): Filter by status
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)

**Actual Response** (200 OK) from running server:
```json
{
  "error": "Failed to retrieve recordings"
}
```

**Expected Response** (when recordings exist):
```json
{
  "recordings": [
    {
      "recording_id": "recording-uuid",
      "meeting_id": 123,
      "status": "completed",
      "started_at": "2025-10-05T14:00:00Z",
      "stopped_at": "2025-10-05T15:30:00Z",
      "duration": 5400,
      "file_size": 524288000,
      "format": "mp4",
      "download_url": "http://js1.blockelite.cn:21058/api/v1/recording/download/recording-uuid"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "pages": 1
  }
}
```

---

##### 8. Download Recording
**Purpose**: Download a completed recording file
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/recording/download/{recording_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -o meeting-recording.mp4
```

**Response** (200 OK):
- Binary video data
- Content-Type: video/mp4 (or appropriate format)
- Content-Disposition: attachment; filename="meeting-123-recording.mp4"
- Supports HTTP Range requests for resumable downloads

---

##### 9. Delete Recording
**Purpose**: Delete a recording file from storage
```bash
curl -X DELETE "http://js1.blockelite.cn:21058/api/v1/recording/{recording_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Recording deleted successfully",
  "recording_id": "recording-uuid"
}
```

---

#### Live Streaming (External Broadcasting)

**Note**: Live streaming pushes the meeting's media streams to external platforms (YouTube, Twitch, custom RTMP servers) or generates HLS/DASH streams for web playback.

##### 10. Start Live Stream
**Purpose**: Start broadcasting the meeting to an external RTMP/HLS endpoint
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/streaming/start" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "meeting_id": 123,
    "room_id": "meeting-123",
    "stream_type": "rtmp",
    "rtmp_url": "rtmp://live.example.com/app/stream-key",
    "quality": "high"
  }'
```

**Request Body**:
```json
{
  "meeting_id": "integer (required)",
  "room_id": "string (required)",
  "stream_type": "string (required): 'rtmp', 'hls', 'dash'",
  "rtmp_url": "string (required for RTMP)",
  "quality": "string (optional): 'low', 'medium', 'high' (default: 'medium')",
  "bitrate": "integer (optional): Target bitrate in bps",
  "resolution": "string (optional): e.g., '1920x1080'"
}
```

**Response** (200 OK):
```json
{
  "stream_id": "stream-uuid",
  "meeting_id": 123,
  "stream_type": "rtmp",
  "status": "streaming",
  "rtmp_url": "rtmp://live.example.com/app/stream-key",
  "hls_url": "http://js1.blockelite.cn:21058/api/v1/streaming/hls/stream-uuid/playlist.m3u8",
  "started_at": "2025-10-05T14:00:00Z",
  "viewers": 0
}
```

---

##### 11. Stop Live Stream
**Purpose**: Stop an active live stream
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/streaming/stop" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stream_id": "stream-uuid"
  }'
```

**Response** (200 OK):
```json
{
  "stream_id": "stream-uuid",
  "status": "stopped",
  "started_at": "2025-10-05T14:00:00Z",
  "stopped_at": "2025-10-05T15:30:00Z",
  "duration": 5400,
  "total_viewers": 152,
  "peak_viewers": 45
}
```

---

##### 12. Get Stream Status
**Purpose**: Check the status and metrics of an active live stream
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/streaming/status/{stream_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "stream_id": "stream-uuid",
  "meeting_id": 123,
  "status": "streaming",
  "stream_type": "rtmp",
  "started_at": "2025-10-05T14:00:00Z",
  "duration": 3600,
  "current_viewers": 32,
  "total_viewers": 87,
  "peak_viewers": 45,
  "quality_metrics": {
    "bitrate": 2500000,
    "frame_rate": 30,
    "resolution": "1920x1080",
    "dropped_frames": 5,
    "encoding_errors": 0
  }
}
```

---

##### 13. List Active Streams
**Purpose**: Retrieve a list of active live streams
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/streaming/list?status=streaming&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Actual Response** (200 OK) from running server:
```json
{
  "filters": {
    "status": "",
    "stream_type": "",
    "user_id": ""
  },
  "message": "Streams retrieved successfully",
  "pagination": {
    "page": 1,
    "page_size": 20,
    "pages": 0,
    "total": 0
  },
  "streams": []
}
```

---

#### Media File Management (Storage)

**Note**: These endpoints handle persistent media storage (recordings, uploads), separate from real-time stream forwarding.

##### 14. Upload Media File
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/media/upload" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@/path/to/video.mp4" \
  -F "meeting_id=123" \
  -F "file_type=video" \
  -F "description=Meeting recording"
```

**Request Parameters**:
- `file` (file, required): Media file to upload
- `meeting_id` (integer, optional): Associated meeting ID
- `file_type` (string, optional): "video", "audio", "image", "document"
- `description` (string, optional): File description

**Supported Formats**:
- **Video**: MP4, WebM, MKV, AVI, MOV, FLV
- **Audio**: MP3, WAV, OGG, FLAC, AAC, M4A
- **Image**: JPG, PNG, GIF, WebP, BMP
- **Document**: PDF, TXT

**Response** (201 Created):
```json
{
  "media_id": "media-uuid",
  "filename": "video.mp4",
  "file_type": "video",
  "size": 52428800,
  "mime_type": "video/mp4",
  "duration": 1800,
  "resolution": "1920x1080",
  "meeting_id": 123,
  "description": "Meeting recording",
  "url": "http://js1.blockelite.cn:21058/api/v1/media/download/media-uuid",
  "thumbnail_url": "http://js1.blockelite.cn:21058/api/v1/media/thumbnail/media-uuid",
  "uploaded_at": "2025-10-05T14:30:00Z",
  "uploaded_by": "user-uuid"
}
```

**Limits**:
- Max file size: 1000MB (configurable per file type)
- Rate limit: 5 uploads per minute
- Concurrent uploads: 3 per user

**Error Response** (413 Payload Too Large):
```json
{
  "error": "File size exceeds maximum limit of 1000MB"
}
```

---

##### 2. List Media Files
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/media?page=1&page_size=20&file_type=video&meeting_id=123" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Query Parameters**:
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)
- `file_type` (string, optional): Filter by file type
- `meeting_id` (integer, optional): Filter by meeting ID
- `user_id` (string, optional): Filter by uploader

**Actual Response** (200 OK) from running server:
```json
{
  "data": {
    "media_files": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "filters": {
    "file_type": "",
    "meeting_id": "",
    "user_id": ""
  },
  "message": "Media list retrieved successfully",
  "offset": 0
}
```

---

##### 3. Get Media Information
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/media/info/{media_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "media_id": "media-uuid",
  "filename": "video.mp4",
  "file_type": "video",
  "size": 52428800,
  "mime_type": "video/mp4",
  "duration": 1800,
  "resolution": "1920x1080",
  "bitrate": 2500000,
  "codec": "h264",
  "audio_codec": "aac",
  "frame_rate": 30,
  "meeting_id": 123,
  "uploaded_at": "2025-10-05T14:30:00Z",
  "uploaded_by": "user-uuid",
  "download_count": 15,
  "last_accessed": "2025-10-05T16:00:00Z"
}
```

---

##### 4. Download Media File
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/media/download/{media_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -o downloaded-file.mp4
```

**Response** (200 OK):
- Binary media data
- Content-Type: Appropriate MIME type
- Content-Disposition: attachment; filename="original-filename.mp4"
- Content-Length: File size in bytes

**Supports Range Requests**:
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/media/download/{media_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Range: bytes=0-1048575" \
  -o partial-file.mp4
```

**Response** (206 Partial Content):
- Content-Range: bytes 0-1048575/52428800
- Enables seeking and resumable downloads

---

##### 5. Delete Media File
```bash
curl -X DELETE "http://js1.blockelite.cn:21058/api/v1/media/{media_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Media file deleted successfully",
  "media_id": "media-uuid"
}
```

**Error Response** (404 Not Found):
```json
{
  "error": "Media file not found"
}
```

---

##### 6. Process Media (Generic Processing)
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/media/process" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "media-uuid",
    "operations": ["extract_audio", "generate_thumbnail"],
    "output_format": "mp3"
  }'
```

**Request Body**:
```json
{
  "media_id": "string (required)",
  "operations": ["array of operations (required)"],
  "output_format": "string (optional)",
  "quality": "string (optional): 'low', 'medium', 'high'"
}
```

**Available Operations**:
- `extract_audio`: Extract audio track from video
- `extract_video`: Extract video track (remove audio)
- `generate_thumbnail`: Generate thumbnail image
- `compress`: Compress file size
- `convert_format`: Convert to different format

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "processing",
  "estimated_time": 45,
  "message": "Processing started"
}
```

---

#### FFmpeg Transcoding

##### 7. Transcode Media
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ffmpeg/transcode" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input_media_id": "media-uuid",
    "output_format": "webm",
    "video_codec": "vp9",
    "audio_codec": "opus",
    "resolution": "1280x720",
    "bitrate": "2000k",
    "preset": "medium"
  }'
```

**Request Body**:
```json
{
  "input_media_id": "string (required)",
  "output_format": "string (required): 'mp4', 'webm', 'mkv', 'avi'",
  "video_codec": "string (optional): 'h264', 'h265', 'vp8', 'vp9'",
  "audio_codec": "string (optional): 'aac', 'mp3', 'opus', 'vorbis'",
  "resolution": "string (optional): 'WIDTHxHEIGHT' or preset like '720p', '1080p'",
  "bitrate": "string (optional): e.g., '2000k', '5M'",
  "frame_rate": "integer (optional): e.g., 24, 30, 60",
  "preset": "string (optional): 'ultrafast', 'fast', 'medium', 'slow' (default: 'medium')"
}
```

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "queued",
  "input_media_id": "media-uuid",
  "output_format": "webm",
  "estimated_time": 120,
  "created_at": "2025-10-05T14:00:00Z"
}
```

---

##### 8. Extract Audio from Video
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ffmpeg/extract-audio" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "video_media_id": "media-uuid",
    "audio_format": "mp3",
    "audio_bitrate": "192k"
  }'
```

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "processing",
  "output_media_id": "audio-media-uuid",
  "estimated_time": 30
}
```

---

##### 9. Extract Video (Remove Audio)
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ffmpeg/extract-video" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "media-uuid"
  }'
```

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "processing",
  "output_media_id": "video-only-uuid"
}
```

---

##### 10. Merge Media Files
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ffmpeg/merge" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "media_ids": ["media-uuid-1", "media-uuid-2", "media-uuid-3"],
    "output_format": "mp4",
    "transition": "fade"
  }'
```

**Request Body**:
```json
{
  "media_ids": ["array of media IDs (required)"],
  "output_format": "string (optional, default: 'mp4')",
  "transition": "string (optional): 'none', 'fade', 'dissolve' (default: 'none')",
  "transition_duration": "number (optional, default: 1.0)"
}
```

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "processing",
  "output_media_id": "merged-media-uuid",
  "estimated_time": 90
}
```

---

##### 11. Generate Thumbnail
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ffmpeg/thumbnail" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "media-uuid",
    "timestamp": 10.5,
    "width": 320,
    "height": 180,
    "format": "jpg"
  }'
```

**Request Body**:
```json
{
  "media_id": "string (required)",
  "timestamp": "number (optional, default: 0): Time in seconds",
  "width": "integer (optional, default: 320)",
  "height": "integer (optional, default: 180)",
  "format": "string (optional): 'jpg', 'png', 'webp' (default: 'jpg')"
}
```

**Response** (200 OK):
```json
{
  "thumbnail_id": "thumbnail-uuid",
  "url": "http://js1.blockelite.cn:21058/api/v1/media/download/thumbnail-uuid",
  "width": 320,
  "height": 180,
  "format": "jpg",
  "size": 15360
}
```

---

##### 12. Get FFmpeg Job Status
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/ffmpeg/job/{job_id}/status" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "job_id": "job-uuid",
  "status": "completed",
  "progress": 100,
  "input_media_id": "media-uuid",
  "output_media_id": "output-media-uuid",
  "started_at": "2025-10-05T14:00:00Z",
  "completed_at": "2025-10-05T14:02:15Z",
  "processing_time": 135,
  "error": null
}
```

**Job Statuses**: queued, processing, completed, failed, cancelled

---

---

### AI Service

**Base Path**: `/api/v1`
**Internal Port**: 8084
**Purpose**: AI-powered speech recognition, emotion detection, translation, summarization, and media enhancement

---

#### Speech Processing Endpoints

##### 1. Speech Recognition
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/speech/recognition" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "audio=@/path/to/audio.wav" \
  -F "language=en-US"
```

**Request Parameters**:
- `audio` (file, required): Audio file (WAV, MP3, OGG, FLAC)
- `language` (string, optional): Language code (default: "en-US")
- `model` (string, optional): Model to use (default: "whisper-base")

**Response** (200 OK):
```json
{
  "text": "Hello, this is a test of the speech recognition system.",
  "confidence": 0.95,
  "language": "en-US",
  "duration": 5.2,
  "words": [
    {
      "word": "Hello",
      "start_time": 0.0,
      "end_time": 0.5,
      "confidence": 0.98
    }
  ],
  "processing_time": 1.2
}
```

**Error Response** (400 Bad Request):
```json
{
  "error": "Unsupported audio format. Supported formats: wav, mp3, ogg, flac"
}
```

---

##### 2. Emotion Detection
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/speech/emotion" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "audio=@/path/to/audio.wav"
```

**Request Parameters**:
- `audio` (file, required): Audio file containing speech
- `model` (string, optional): Emotion detection model (default: "hubert-emotion")

**Response** (200 OK):
```json
{
  "emotions": [
    {
      "emotion": "happy",
      "confidence": 0.85,
      "start_time": 0.0,
      "end_time": 2.5
    },
    {
      "emotion": "neutral",
      "confidence": 0.92,
      "start_time": 2.5,
      "end_time": 5.0
    }
  ],
  "dominant_emotion": "neutral",
  "processing_time": 0.8
}
```

**Supported Emotions**: happy, sad, angry, neutral, surprised, fearful, disgusted

---

##### 3. Synthesis Detection (Deepfake Detection)
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/speech/synthesis-detection" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "audio=@/path/to/audio.wav"
```

**Response** (200 OK):
```json
{
  "is_synthetic": false,
  "confidence": 0.93,
  "authenticity_score": 0.95,
  "processing_time": 0.6
}
```

---

#### AI General Capabilities

##### 4. Transcribe Audio
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ai/transcribe" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "audio=@/path/to/meeting-audio.wav" \
  -F "language=en" \
  -F "format=srt"
```

**Request Parameters**:
- `audio` (file, required): Audio file to transcribe
- `language` (string, optional): Source language (default: auto-detect)
- `format` (string, optional): Output format: "text", "srt", "vtt", "json" (default: "text")
- `timestamps` (boolean, optional): Include word-level timestamps (default: false)

**Response** (200 OK):
```json
{
  "transcription": "Welcome to the meeting. Today we will discuss...",
  "language": "en",
  "duration": 120.5,
  "word_count": 245,
  "processing_time": 15.3,
  "segments": [
    {
      "text": "Welcome to the meeting.",
      "start": 0.0,
      "end": 2.5,
      "confidence": 0.96
    }
  ]
}
```

---

##### 5. Translate Text
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ai/translate" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello, how are you?",
    "source_language": "en",
    "target_language": "zh"
  }'
```

**Request Body**:
```json
{
  "text": "string (required)",
  "source_language": "string (optional, auto-detect if not provided)",
  "target_language": "string (required)"
}
```

**Response** (200 OK):
```json
{
  "translated_text": "你好，你好吗？",
  "source_language": "en",
  "target_language": "zh",
  "confidence": 0.94,
  "processing_time": 0.3
}
```

**Supported Languages**: en, zh, es, fr, de, ja, ko, ru, ar, pt, it

---

##### 6. Summarize Text
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ai/summarize" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Long meeting transcript...",
    "max_length": 200,
    "format": "bullet_points"
  }'
```

**Request Body**:
```json
{
  "text": "string (required)",
  "max_length": "integer (optional, default: 150)",
  "format": "string (optional): 'paragraph' or 'bullet_points' (default: 'paragraph')"
}
```

**Response** (200 OK):
```json
{
  "summary": "• Discussed Q4 goals\n• Reviewed budget allocation\n• Assigned action items",
  "original_length": 1250,
  "summary_length": 85,
  "compression_ratio": 0.068,
  "processing_time": 2.1
}
```

---

##### 7. Sentiment Analysis
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/ai/analyze" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "This meeting was very productive and everyone was engaged."
  }'
```

**Response** (200 OK):
```json
{
  "sentiment": "positive",
  "score": 0.89,
  "confidence": 0.92,
  "emotions": {
    "joy": 0.75,
    "trust": 0.68,
    "anticipation": 0.45
  },
  "processing_time": 0.2
}
```

**Sentiment Values**: positive, negative, neutral

---

#### Media Enhancement Endpoints

##### 8. Audio Denoising
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/audio/denoising" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "audio=@/path/to/noisy-audio.wav" \
  -F "noise_reduction_level=medium"
```

**Request Parameters**:
- `audio` (file, required): Noisy audio file
- `noise_reduction_level` (string, optional): "low", "medium", "high" (default: "medium")
- `preserve_speech` (boolean, optional): Preserve speech quality (default: true)

**Response** (200 OK):
- Binary audio data (denoised WAV format)
- Content-Type: audio/wav
- Headers:
  - `X-Processing-Time`: Processing duration in seconds
  - `X-Noise-Reduction-DB`: Noise reduction in decibels

**Performance**: ~2-5 seconds for 1 minute of audio

---

##### 9. Video Enhancement
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/video/enhancement" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "video=@/path/to/video.mp4" \
  -F "enhancement_type=super_resolution" \
  -F "target_resolution=1920x1080"
```

**Request Parameters**:
- `video` (file, required): Video file (MP4, WebM, MKV)
- `enhancement_type` (string, required): "super_resolution", "denoising", "stabilization", "color_correction"
- `target_resolution` (string, optional): Target resolution (e.g., "1920x1080")
- `quality` (string, optional): "low", "medium", "high" (default: "medium")

**Response** (200 OK):
```json
{
  "enhanced_video_id": "enhanced-media-uuid",
  "url": "http://js1.blockelite.cn:21058/api/v1/media/enhanced-media-uuid",
  "processing_time": 45.3,
  "enhancement_type": "super_resolution",
  "original_resolution": "1280x720",
  "enhanced_resolution": "1920x1080",
  "file_size": 52428800,
  "duration": 120.5
}
```

**Performance**: ~30-60 seconds per minute of video (depends on resolution and enhancement type)

---

#### Model Management Endpoints

##### 10. List Available Models
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/models" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Actual Response** (200 OK) from running server:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "count": 0,
    "models": []
  },
  "timestamp": "2025-10-05T20:35:55+08:00"
}
```

**Response Fields**:
- `code`: HTTP status code
- `message`: Response message
- `data.count`: Total number of models
- `data.models`: Array of model objects
  - `model_id`: Unique model identifier
  - `name`: Human-readable model name
  - `type`: Model type (speech_recognition, emotion_detection, translation, etc.)
  - `status`: "loaded", "unloaded", "loading", "error"
  - `version`: Model version
  - `memory_usage`: Memory consumed by the model
  - `node_id`: Node where the model is loaded

---

##### 11. Get Model Status
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/models/{model_id}/status" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "model_id": "whisper-base",
  "name": "Whisper Base",
  "type": "speech_recognition",
  "status": "loaded",
  "version": "1.0",
  "memory_usage": "512MB",
  "node_id": "node-001",
  "load_time": "2025-10-05T14:00:00Z",
  "request_count": 1523,
  "avg_latency": 1.2,
  "error_rate": 0.02
}
```

---

##### 12. Load Model
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/models/{model_id}/load" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "model_id": "whisper-base",
  "status": "loaded",
  "load_time": 3.5,
  "memory_usage": "512MB",
  "node_id": "node-001"
}
```

**Error Response** (500 Internal Server Error):
```json
{
  "error": "Failed to load model: insufficient memory"
}
```

---

##### 13. Unload Model
```bash
curl -X DELETE "http://js1.blockelite.cn:21058/api/v1/models/{model_id}/unload" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "model_id": "whisper-base",
  "status": "unloaded",
  "freed_memory": "512MB"
}
```

---

##### 14. Load Model on Specific Node
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/models/{model_id}/load-on/{node_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Model loaded successfully",
  "model_id": "whisper-base",
  "node_id": "node-002",
  "load_time": 4.2
}
```

---

##### 15. Unload Model from Specific Node
```bash
curl -X DELETE "http://js1.blockelite.cn:21058/api/v1/models/{model_id}/unload-from/{node_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Model unloaded successfully",
  "model_id": "whisper-base",
  "node_id": "node-002"
}
```

---

##### 16. Rebalance Models Across Nodes
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/models/rebalance" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Models rebalanced successfully",
  "rebalanced_models": 5,
  "node_distribution": {
    "node-001": 3,
    "node-002": 2
  },
  "rebalance_time": 12.5
}
```

---

#### Inference Node Management

##### 17. List Inference Nodes
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/nodes" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Actual Response** (200 OK) from running server:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "count": 0,
    "nodes": []
  },
  "timestamp": "2025-10-05T20:35:55+08:00"
}
```

**Response Fields**:
- `data.nodes`: Array of node objects
  - `node_id`: Unique node identifier
  - `address`: Node IP address and port
  - `status`: "online", "offline", "busy"
  - `cpu_usage`: CPU utilization percentage
  - `memory_usage`: Memory utilization percentage
  - `gpu_usage`: GPU utilization percentage (if available)
  - `loaded_models`: Number of models loaded on this node
  - `active_requests`: Current number of active inference requests
  - `total_requests`: Total requests processed
  - `avg_latency`: Average response time in seconds

---

##### 18. Get Node Status
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/nodes/{node_id}/status" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "node_id": "node-001",
  "address": "192.168.1.100:8085",
  "status": "online",
  "uptime": 86400,
  "cpu_usage": 45.2,
  "memory_usage": 62.8,
  "gpu_usage": 78.5,
  "loaded_models": ["whisper-base", "hubert-emotion"],
  "active_requests": 3,
  "total_requests": 15234,
  "success_rate": 0.98,
  "avg_latency": 1.5,
  "last_heartbeat": "2025-10-05T14:30:00Z"
}
```

---

##### 19. Node Health Check
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/nodes/{node_id}/health-check" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "node_id": "node-001",
  "status": "healthy",
  "response_time": 0.05,
  "checks": {
    "connectivity": "pass",
    "memory": "pass",
    "disk": "pass",
    "gpu": "pass"
  }
}
```

---

#### Load Balancing and Monitoring

##### 20. Get Load Balancer Statistics
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/load-balancer/stats" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Actual Response** (200 OK) from running server:
```json
{
  "stats": {}
}
```

**Expected Response** (when nodes are active):
```json
{
  "stats": {
    "total_requests": 50234,
    "requests_per_node": {
      "node-001": 25120,
      "node-002": 25114
    },
    "avg_latency_per_node": {
      "node-001": 1.2,
      "node-002": 1.3
    },
    "load_distribution": {
      "node-001": 0.50,
      "node-002": 0.50
    },
    "algorithm": "round_robin"
  }
}
```

---

##### 21. Get Monitoring Metrics
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/monitoring/metrics" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Actual Response** (200 OK) from running server:
```json
{
  "metrics": {
    "total_nodes": 0,
    "online_nodes": 0,
    "offline_nodes": 0,
    "avg_cpu_usage": 0,
    "avg_memory_usage": 0,
    "avg_gpu_usage": 0,
    "total_requests": 0,
    "success_requests": 0,
    "failed_requests": 0,
    "avg_response_time": 0,
    "requests_per_sec": 0,
    "total_models": 0,
    "loaded_models": 0,
    "active_models": 0,
    "error_rate": 0,
    "timeout_rate": 0,
    "timestamp": "2025-10-05T20:35:55.362417385+08:00"
  }
}
```

**Metrics Explanation**:
- `total_nodes`: Total number of registered inference nodes
- `online_nodes`: Number of nodes currently online
- `offline_nodes`: Number of nodes currently offline
- `avg_cpu_usage`: Average CPU usage across all nodes (%)
- `avg_memory_usage`: Average memory usage across all nodes (%)
- `avg_gpu_usage`: Average GPU usage across all nodes (%)
- `total_requests`: Total inference requests processed
- `success_requests`: Number of successful requests
- `failed_requests`: Number of failed requests
- `avg_response_time`: Average response time in seconds
- `requests_per_sec`: Current request rate
- `error_rate`: Error rate (0-1)
- `timeout_rate`: Timeout rate (0-1)

---

##### 22. Get Active Alerts
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/monitoring/alerts" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "alerts": [
    {
      "alert_id": "alert-001",
      "severity": "warning",
      "type": "high_memory_usage",
      "node_id": "node-001",
      "message": "Memory usage exceeded 80%",
      "value": 85.2,
      "threshold": 80.0,
      "timestamp": "2025-10-05T14:25:00Z",
      "status": "active"
    }
  ]
}
```

**Alert Severities**: info, warning, error, critical

---

##### 23. Get Failover Events
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/failover/events" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "events": [
    {
      "event_id": "event-001",
      "type": "node_failure",
      "failed_node": "node-002",
      "backup_node": "node-001",
      "timestamp": "2025-10-05T13:45:00Z",
      "recovery_time": 2.5,
      "affected_requests": 12
    }
  ]
}
```

---

##### 24. Get Discovered Nodes
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/discovery/nodes" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "nodes": [
    {
      "node_id": "node-001",
      "address": "192.168.1.100:8085",
      "discovered_at": "2025-10-05T10:00:00Z",
      "last_seen": "2025-10-05T14:30:00Z",
      "capabilities": ["speech_recognition", "emotion_detection"],
      "status": "online"
    }
  ]
}
```

---

#### AI Service Summary

**Total Endpoints**: 24

**Categories**:
- Speech Processing: 3 endpoints
- AI General Capabilities: 4 endpoints
- Media Enhancement: 2 endpoints
- Model Management: 7 endpoints
- Node Management: 3 endpoints
- Monitoring & Load Balancing: 5 endpoints

**Performance Characteristics**:
- Speech Recognition: 1-3 seconds per minute of audio
- Emotion Detection: 0.5-1 second per minute of audio
- Translation: 0.2-0.5 seconds per sentence
- Summarization: 1-3 seconds per 1000 words
- Audio Denoising: 2-5 seconds per minute of audio
- Video Enhancement: 30-60 seconds per minute of video

**Supported File Formats**:
- Audio: WAV, MP3, OGG, FLAC
- Video: MP4, WebM, MKV

**Rate Limits**:
- AI endpoints: 20 requests/minute (burst: 10)
- Model management: No limit (admin operations)

---

---

## Authentication

### JWT Token-Based Authentication

All API requests (except public endpoints like `/health`) require a valid JWT access token.

#### Obtaining a Token

1. **Register** or **Login** via User Service
2. Receive `access_token` and `refresh_token`
3. Include token in `Authorization` header

#### Using the Token

```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/users/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Token Expiration

- **Access Token**: Expires in 1 hour (3600 seconds)
- **Refresh Token**: Expires in 7 days

#### Refreshing Tokens

```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token-here"
  }'
```

**Response** (200 OK):
```json
{
  "access_token": "new-access-token",
  "refresh_token": "new-refresh-token",
  "expires_in": 3600
}
```

---

## Health & Metrics

### Health Check Endpoints

All microservices expose a `/health` endpoint for readiness checks.

#### Gateway Health
```bash
curl -s "http://js1.blockelite.cn:21058/health"
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-05T12:21:28+00:00"
}
```

#### User Service Health
```bash
curl -s "http://js1.blockelite.cn:21058/api/v1/users/health"
```

**Actual Response** (from running server):
```json
{
  "service": "user-service",
  "status": "ok",
  "time": "2025-10-05T20:21:31+08:00"
}
```

#### Meeting Service Health
```bash
curl -s "http://js1.blockelite.cn:21058/api/v1/meetings/health"
```

**Actual Response**:
```json
{
  "service": "meeting-service",
  "status": "ok",
  "time": "2025-10-05T20:21:31+08:00"
}
```

#### Signaling Service Health
```bash
curl -s "http://js1.blockelite.cn:21058/ws/signaling/health"
```

**Actual Response**:
```json
{
  "active_rooms": 0,
  "active_sessions": 0,
  "connected_clients": 0,
  "service": "signaling-service",
  "status": "ok",
  "time": "2025-10-05T20:21:31+08:00"
}
```

#### Media Service Health
```bash
curl -s "http://js1.blockelite.cn:21058/api/v1/media/health"
```

**Actual Response**:
```json
{
  "service": "media-service",
  "status": "healthy",
  "timestamp": 1759666891
}
```

#### AI Service Health
```bash
curl -s "http://js1.blockelite.cn:21058/api/v1/ai/health"
```

**Actual Response**:
```json
{
  "status": "ok"
}
```

---

### Prometheus Metrics Endpoints

All services expose Prometheus-compatible metrics at `/metrics`.

#### Example Metrics (User Service)

```bash
curl -s "http://js1.blockelite.cn:21058/api/v1/users/metrics" | head -20
```

**Actual Response**:
```
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.24.7"} 1

# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/health",method="GET",service="user-service",status_code="200"} 57
http_requests_total{endpoint="/metrics",method="GET",service="user-service",status_code="200"} 108

# HELP http_request_duration_seconds HTTP request latency
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{endpoint="/health",method="GET",service="user-service",le="0.005"} 45
http_request_duration_seconds_bucket{endpoint="/health",method="GET",service="user-service",le="0.01"} 52
http_request_duration_seconds_bucket{endpoint="/health",method="GET",service="user-service",le="0.025"} 57
```

#### Key Metrics

- `http_requests_total` - Total HTTP requests by endpoint, method, status
- `http_request_duration_seconds` - Request latency histogram
- `go_goroutines` - Number of active goroutines
- `go_memstats_alloc_bytes` - Memory allocated
- `db_connections_active` - Active database connections
- `active_users` - Currently active users
- `active_meetings` - Currently active meetings
- `webrtc_connections` - Active WebRTC connections

---

## Error Handling

### Standard Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details (optional)",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

### HTTP Status Codes

| Status Code | Meaning | Example |
|-------------|---------|---------|
| 200 | OK | Successful GET, PUT, POST |
| 201 | Created | Resource created successfully |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Missing or invalid authentication token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Resource already exists |
| 413 | Payload Too Large | File upload exceeds size limit |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Common Error Codes

#### Authentication Errors

**401 Unauthorized - Missing Token**
```json
{
  "error": {
    "code": "MISSING_TOKEN",
    "message": "Authorization token is required",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

**401 Unauthorized - Invalid Token**
```json
{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired authentication token",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

**403 Forbidden**
```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "You do not have permission to access this resource",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

#### Validation Errors

**400 Bad Request - Invalid Input**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "email": "Invalid email format",
      "password": "Password must be at least 8 characters"
    },
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

#### Resource Errors

**404 Not Found**
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Meeting with ID 'meeting-uuid' not found",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

**409 Conflict**
```json
{
  "error": {
    "code": "RESOURCE_CONFLICT",
    "message": "User with email 'test@example.com' already exists",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

#### Rate Limiting Errors

**429 Too Many Requests**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "retry_after": 60,
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

#### Server Errors

**500 Internal Server Error**
```json
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred. Please try again later.",
    "request_id": "req-uuid-for-debugging",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

**503 Service Unavailable**
```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Service is temporarily unavailable. Please try again later.",
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

---

## Rate Limiting

The API implements rate limiting to prevent abuse and ensure fair usage.

### Rate Limit Configuration

| Endpoint Type | Requests per Minute | Burst |
|---------------|---------------------|-------|
| Authentication (`/api/v1/auth`) | 10 | 5 |
| General API (`/api/v1/*`) | 100 | 50 |
| File Upload (`/api/v1/media/upload`) | 5 | 5 |
| AI Processing (`/api/v1/ai/*`) | 20 | 10 |
| WebSocket (`/ws/signaling`) | 30 | 10 |

### Rate Limit Headers

Every API response includes rate limit information in headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1728134460
```

- `X-RateLimit-Limit`: Maximum requests allowed per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets

### Example: Rate Limit Exceeded

**Request**:
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test"}'
```

**Response** (429 Too Many Requests):
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many login attempts. Please try again in 60 seconds.",
    "retry_after": 60,
    "timestamp": "2025-10-05T12:00:00Z"
  }
}
```

**Headers**:
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1728134460
Retry-After: 60
```

### Best Practices

1. **Respect Rate Limits**: Monitor `X-RateLimit-Remaining` header
2. **Implement Backoff**: Use exponential backoff when rate limited
3. **Cache Responses**: Cache GET requests to reduce API calls
4. **Batch Operations**: Use batch endpoints when available
5. **Handle 429 Errors**: Implement retry logic with `Retry-After` header

---

## WebSocket Protocol

### Signaling Service WebSocket

**Connection URL**: `ws://js1.blockelite.cn:21058/ws/signaling?token=YOUR_ACCESS_TOKEN`

#### Connection Flow

1. **Connect** with JWT token in query parameter
2. **Send** join message with room_id
3. **Exchange** WebRTC signaling messages (offer, answer, ICE candidates)
4. **Receive** notifications about other participants
5. **Disconnect** or send leave message

#### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `join` | Client → Server | Join a meeting room |
| `leave` | Client → Server | Leave a meeting room |
| `offer` | Client → Server | WebRTC offer SDP |
| `answer` | Client → Server | WebRTC answer SDP |
| `ice-candidate` | Client → Server | ICE candidate |
| `user-joined` | Server → Client | Notification: user joined |
| `user-left` | Server → Client | Notification: user left |
| `offer` | Server → Client | Forward offer to peer |
| `answer` | Server → Client | Forward answer to peer |
| `ice-candidate` | Server → Client | Forward ICE candidate to peer |
| `error` | Server → Client | Error notification |

#### Example: Complete WebRTC Handshake

**1. Client A joins room**
```json
{
  "type": "join",
  "room_id": "meeting-123",
  "user_id": "user-a"
}
```

**2. Server notifies Client B**
```json
{
  "type": "user-joined",
  "room_id": "meeting-123",
  "user_id": "user-a",
  "timestamp": "2025-10-05T14:00:00Z"
}
```

**3. Client A sends offer to Client B**
```json
{
  "type": "offer",
  "room_id": "meeting-123",
  "target_user_id": "user-b",
  "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
}
```

**4. Server forwards offer to Client B**
```json
{
  "type": "offer",
  "room_id": "meeting-123",
  "from_user_id": "user-a",
  "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
}
```

**5. Client B sends answer**
```json
{
  "type": "answer",
  "room_id": "meeting-123",
  "target_user_id": "user-a",
  "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
}
```

**6. ICE candidates exchanged**
```json
{
  "type": "ice-candidate",
  "room_id": "meeting-123",
  "target_user_id": "user-b",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

---

## Quick Start Examples

### Complete User Registration and Meeting Creation Flow

```bash
#!/bin/bash

BASE_URL="http://js1.blockelite.cn:21058"

# 1. Register a new user
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }')
echo "$REGISTER_RESPONSE"

# 2. Login to get access token
echo -e "\n2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePass123!"
  }')
echo "$LOGIN_RESPONSE"

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')

# 3. Create a meeting
echo -e "\n3. Creating meeting..."
MEETING_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/meetings" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Standup",
    "description": "Daily standup meeting",
    "start_time": "2025-10-05T14:00:00Z",
    "duration": 30,
    "max_participants": 10
  }')
echo "$MEETING_RESPONSE"

MEETING_ID=$(echo "$MEETING_RESPONSE" | jq -r '.meeting_id')

# 4. Join the meeting
echo -e "\n4. Joining meeting..."
JOIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/meetings/$MEETING_ID/join" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "$JOIN_RESPONSE"

# 5. Get meeting details
echo -e "\n5. Getting meeting details..."
DETAILS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/meetings/$MEETING_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "$DETAILS_RESPONSE"
```

---

## Monitoring and Observability

### Prometheus Queries

**Check service availability**
```bash
curl -s "http://js1.blockelite.cn:21059/api/v1/query?query=up"
```

**Get HTTP request rate**
```bash
curl -s "http://js1.blockelite.cn:21059/api/v1/query?query=rate(http_requests_total[5m])"
```

**Get 95th percentile latency**
```bash
curl -s "http://js1.blockelite.cn:21059/api/v1/query?query=histogram_quantile(0.95,rate(http_request_duration_seconds_bucket[5m]))"
```

### Jaeger Trace Search

**Find traces for a specific service**
```bash
curl -s "http://js1.blockelite.cn:21061/api/traces?service=user-service&limit=20&lookback=1h"
```

### Grafana Dashboards

Access Grafana at http://js1.blockelite.cn:21062 (admin/admin123) to view:
- Service health overview
- Request rate and latency
- Error rates
- Resource utilization (CPU, memory)
- Active users and meetings

---

## Support and Contact

For API support, issues, or feature requests:
- **Documentation**: This file
- **Monitoring**: http://js1.blockelite.cn:21062 (Grafana)
- **Tracing**: http://js1.blockelite.cn:21061 (Jaeger)
- **Metrics**: http://js1.blockelite.cn:21059 (Prometheus)

---

**Last Updated**: 2025-10-05
**API Version**: 1.0
**Server**: js1.blockelite.cn:21012
### Meeting Service

**Base Path**: `/api/v1/meetings`  
**Internal Port**: 8082  
**Purpose**: Meeting room management

#### Endpoints

##### 1. Create Meeting
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/meetings" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Standup",
    "description": "Daily standup meeting",
    "start_time": "2025-10-05T14:00:00Z",
    "duration": 30,
    "max_participants": 10
  }'
```

**Response** (201 Created):
```json
{
  "meeting_id": "meeting-uuid",
  "title": "Team Standup",
  "description": "Daily standup meeting",
  "creator_id": "user-uuid",
  "start_time": "2025-10-05T14:00:00Z",
  "duration": 30,
  "max_participants": 10,
  "status": "scheduled",
  "join_url": "http://js1.blockelite.cn:21058/meetings/meeting-uuid/join",
  "created_at": "2025-10-05T12:00:00Z"
}
```

##### 2. Get Meeting Details
```bash
curl -X GET "http://js1.blockelite.cn:21058/api/v1/meetings/{meeting_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "meeting_id": "meeting-uuid",
  "title": "Team Standup",
  "description": "Daily standup meeting",
  "creator_id": "user-uuid",
  "start_time": "2025-10-05T14:00:00Z",
  "duration": 30,
  "max_participants": 10,
  "current_participants": 3,
  "status": "in_progress",
  "created_at": "2025-10-05T12:00:00Z"
}
```

##### 3. Update Meeting
```bash
curl -X PUT "http://js1.blockelite.cn:21058/api/v1/meetings/{meeting_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Team Standup",
    "duration": 45
  }'
```

**Response** (200 OK):
```json
{
  "meeting_id": "meeting-uuid",
  "title": "Updated Team Standup",
  "duration": 45,
  "updated_at": "2025-10-05T13:00:00Z"
}
```

##### 4. Delete Meeting
```bash
curl -X DELETE "http://js1.blockelite.cn:21058/api/v1/meetings/{meeting_id}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (204 No Content)

##### 5. Join Meeting
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/meetings/{meeting_id}/join" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "meeting_id": "meeting-uuid",
  "participant_id": "participant-uuid",
  "signaling_url": "ws://js1.blockelite.cn:21058/ws/signaling",
  "ice_servers": [
    {
      "urls": ["stun:stun.l.google.com:19302"]
    }
  ],
  "joined_at": "2025-10-05T14:00:00Z"
}
```

##### 6. Leave Meeting
```bash
curl -X POST "http://js1.blockelite.cn:21058/api/v1/meetings/{meeting_id}/leave" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Left meeting successfully",
  "duration": 1800
}
```

---


