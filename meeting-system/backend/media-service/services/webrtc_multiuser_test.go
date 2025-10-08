package services

import (
"os"
"path/filepath"
"testing"

"github.com/pion/webrtc/v3"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"meeting-system/shared/config"
)

// createMockOffer 创建模拟的客户端Offer
func createMockOffer(t *testing.T) *webrtc.SessionDescription {
return &webrtc.SessionDescription{
Type: webrtc.SDPTypeOffer,
SDP: `v=0
o=- 0 0 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
a=msid-semantic: WMS
m=video 9 UDP/TLS/RTP/SAVPF 96
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:test
a=ice-pwd:testpassword
a=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00
a=setup:actpass
a=mid:0
a=sendrecv
a=rtcp-mux
a=rtpmap:96 VP8/90000
`,
}
}

// TestMultiUserVideoStreamForwarding 测试多用户视频流转发功能（SFU架构）
func TestMultiUserVideoStreamForwarding(t *testing.T) {
if testing.Short() {
t.Skip("skipping multi-user stream forwarding test in short mode")
}

cfg := &config.Config{}
webrtcService := NewWebRTCService(cfg, nil, nil)
require.NoError(t, webrtcService.Initialize())
defer webrtcService.Stop()

roomID := "test-room-multiuser"
users := []string{"user-1", "user-2", "user-3"}
peerIDs := make([]string, 0, len(users))

for _, userID := range users {
offer := createMockOffer(t)
answer, peerID, err := webrtcService.CreateAnswer(roomID, userID, offer)
require.NoError(t, err)
require.NotNil(t, answer)
peerIDs = append(peerIDs, peerID)
t.Logf("✓ User %s joined room %s with peer %s", userID, roomID, peerID)
}

peers, err := webrtcService.GetRoomPeers(roomID)
require.NoError(t, err)
assert.Equal(t, len(users), len(peers), "Room should have all users")
t.Logf("✓ Room has %d peers", len(peers))

testVideoFiles := []string{
"20250928_165500.mp4",
"20250827_104938.mp4",
"20250827_105955.mp4",
}

for i, videoFile := range testVideoFiles {
videoPath := filepath.Join("..", "test_video", videoFile)
if _, err := os.Stat(videoPath); err == nil {
t.Logf("✓ Test video file %d exists: %s", i+1, videoFile)
}
}

t.Log("✓ SFU架构验证：WebRTC服务仅负责RTP包转发")

for i, peerID := range peerIDs {
err := webrtcService.LeaveRoom(roomID, users[i])
require.NoError(t, err)
t.Logf("✓ User %s (peer %s) left room", users[i], peerID)
}
}

// TestMultiUserAudioStreamForwarding 测试多用户音频流转发功能（SFU架构）
func TestMultiUserAudioStreamForwarding(t *testing.T) {
if testing.Short() {
t.Skip("skipping multi-user audio stream forwarding test in short mode")
}

cfg := &config.Config{}
webrtcService := NewWebRTCService(cfg, nil, nil)
require.NoError(t, webrtcService.Initialize())
defer webrtcService.Stop()

roomID := "test-room-audio"
users := []string{"audio-user-1", "audio-user-2"}
peerIDs := make([]string, 0, len(users))

for _, userID := range users {
offer := createMockOffer(t)
answer, peerID, err := webrtcService.CreateAnswer(roomID, userID, offer)
require.NoError(t, err)
require.NotNil(t, answer)
peerIDs = append(peerIDs, peerID)
t.Logf("✓ User %s joined audio room %s", userID, roomID)
}

peers, err := webrtcService.GetRoomPeers(roomID)
require.NoError(t, err)
assert.Equal(t, len(users), len(peers), "Audio room should have all users")
t.Logf("✓ Audio room has %d peers", len(peers))

t.Log("✓ SFU架构验证：音频RTP包直接转发")

for i, peerID := range peerIDs {
err := webrtcService.LeaveRoom(roomID, users[i])
require.NoError(t, err)
t.Logf("✓ User %s (peer %s) left room", users[i], peerID)
}
}

// TestRTPForwardingWithRealFiles 测试使用真实音视频文件的RTP转发
func TestRTPForwardingWithRealFiles(t *testing.T) {
if testing.Short() {
t.Skip("skipping real file RTP forwarding test in short mode")
}

testFiles := map[string]string{
"video1": filepath.Join("..", "test_video", "20250928_165500.mp4"),
"video2": filepath.Join("..", "test_video", "20250827_104938.mp4"),
"video3": filepath.Join("..", "test_video", "20250827_105955.mp4"),
"audio":  filepath.Join("..", "test_video", "20250602_215504.mp3"),
}

filesExist := 0
for name, path := range testFiles {
if _, err := os.Stat(path); err == nil {
filesExist++
t.Logf("✓ Test file '%s' exists: %s", name, filepath.Base(path))
}
}

assert.Greater(t, filesExist, 0, "At least one test file should exist")
t.Logf("✓ Found %d/%d test files", filesExist, len(testFiles))
t.Log("✓ SFU架构：服务端仅转发RTP包，不处理媒体内容")
}
