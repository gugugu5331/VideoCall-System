package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
)

type options struct {
	wsURL       string
	jwtSecret   string
	meetingID   uint
	users       uint
	rate        uint
	runFor      time.Duration
	joinMessage bool
}

type clientMetrics struct {
	connected int64
	failed    int64
	acks      int64
	latency   int64
}

func main() {
	var opt options
	flag.StringVar(&opt.wsURL, "ws", "ws://127.0.0.1:8081/ws/signaling", "WebSocket 完整地址")
	flag.StringVar(&opt.jwtSecret, "secret", "", "JWT Secret")
	flag.UintVar(&opt.meetingID, "meeting", 9999, "会议 ID")
	flag.UintVar(&opt.users, "users", 50, "模拟用户总数")
	flag.UintVar(&opt.rate, "rate", 10, "每秒拉起的连接数")
	flag.DurationVar(&opt.runFor, "duration", 60*time.Second, "持续时间")
	flag.BoolVar(&opt.joinMessage, "join", true, "是否发送 join 消息")
	flag.Parse()
	if opt.jwtSecret == "" {
		log.Fatal("必须指定 -secret")
	}

	ctx, cancel := context.WithTimeout(context.Background(), opt.runFor)
	defer cancel()

	metrics := &clientMetrics{}
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(time.Second / time.Duration(opt.rate))
	defer ticker.Stop()

	for i := uint(0); i < opt.users; i++ {
		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			wg.Add(1)
			go simulateClient(ctx, wg, opt, i+1, metrics)
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-stop:
		cancel()
	}
	wg.Wait()

	report(metrics, opt)
}

func simulateClient(ctx context.Context, wg *sync.WaitGroup, opt options, userID uint, metrics *clientMetrics) {
	defer wg.Done()

	peerID := randomID()
	wsURL, err := buildURL(opt.wsURL, userID, opt.meetingID, peerID)
	if err != nil {
		log.Printf("[user %d] 构建 URL 失败: %v", userID, err)
		atomic.AddInt64(&metrics.failed, 1)
		return
	}

	token, err := issueToken(opt.jwtSecret, userID)
	if err != nil {
		log.Printf("[user %d] 生成 JWT 失败: %v", userID, err)
		atomic.AddInt64(&metrics.failed, 1)
		return
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	start := time.Now()
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			log.Printf("[user %d] 建连失败: status=%d body=%q", userID, resp.StatusCode, readBody(resp.Body))
		} else {
			log.Printf("[user %d] 建连失败: %v", userID, err)
		}
		atomic.AddInt64(&metrics.failed, 1)
		return
	}
	defer conn.Close()

	atomic.AddInt64(&metrics.connected, 1)

	if opt.joinMessage {
		payload := models.WebSocketMessage{
			ID:        "join_" + peerID,
			Type:      models.MessageTypeJoinRoom,
			MeetingID: opt.meetingID,
			SessionID: "",
			PeerID:    peerID,
		}
		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("[user %d] join 发送失败: %v", userID, err)
			return
		}
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var resp models.WebSocketMessage
			if err := json.Unmarshal(msg, &resp); err != nil {
				continue
			}
			if resp.Type == models.MessageTypeRoomInfo {
				atomic.AddInt64(&metrics.acks, 1)
				latency := time.Since(start)
				atomic.AddInt64(&metrics.latency, latency.Microseconds())
				return
			}
		}
	}
}

func buildURL(base string, userID, meetingID uint, peerID string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("user_id", fmt.Sprintf("%d", userID))
	q.Set("meeting_id", fmt.Sprintf("%d", meetingID))
	q.Set("peer_id", peerID)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func issueToken(secret string, userID uint) (string, error) {
	claims := middleware.JWTClaims{
		UserID:   userID,
		Username: fmt.Sprintf("stress_user_%d", userID),
		Email:    fmt.Sprintf("stress%d@example.com", userID),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func readBody(rc io.ReadCloser) string {
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Sprintf("<read error: %v>", err)
	}
	return string(data)
}

func report(m *clientMetrics, opt options) {
	total := atomic.LoadInt64(&m.connected) + atomic.LoadInt64(&m.failed)
	success := atomic.LoadInt64(&m.connected)
	fail := atomic.LoadInt64(&m.failed)
	acks := atomic.LoadInt64(&m.acks)
	var avgLatency float64
	if acks > 0 {
		avgLatency = float64(atomic.LoadInt64(&m.latency)) / float64(acks) / 1e3
	}
	fmt.Println("====== 信令服务压测报告 ======")
	fmt.Printf("目标地址: %s\n", opt.wsURL)
	fmt.Printf("会议 ID: %d\n", opt.meetingID)
	fmt.Printf("用户总数: %d\n", opt.users)
	fmt.Printf("建连成功: %d (%.2f%%)\n", success, float64(success)/float64(total)*100)
	fmt.Printf("建连失败: %d (%.2f%%)\n", fail, float64(fail)/float64(total)*100)
	fmt.Printf("收到 RoomInfo ACK: %d\n", acks)
	if acks > 0 {
		fmt.Printf("平均 ACK 延迟: %.2f ms\n", avgLatency)
	}
	fmt.Println("=============================")
}
