package handlers

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"io"
	"meeting-system/shared/config"
	"meeting-system/shared/database"
	sharedgrpc "meeting-system/shared/grpc"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/signaling-service/services"
)

func TestWebSocketHandlerStress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config.GlobalConfig = &config.Config{
		WebSocket: config.WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     true,
			PingPeriod:      54,
			PongWait:        60,
			WriteWait:       10,
			MaxMessageSize:  512,
		},
		Redis: config.RedisConfig{
			SessionPrefix: "test:signaling:session:",
			RoomPrefix:    "test:signaling:room:",
			SessionTTL:    3600,
		},
		Signaling: config.SignalingConfig{
			Room:    config.RoomConfig{MaxParticipants: 100, CleanupInterval: 300, InactiveTimeout: 1800},
			Session: config.SessionConfig{HeartbeatInterval: 30, ConnectionTimeout: 60},
		},
		JWT: config.JWTConfig{Secret: "stress-secret", ExpireTime: 24},
	}

    originalDB := database.GetDB()
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    require.NoError(t, err)
    sqlDB, err := db.DB()
    require.NoError(t, err)
    sqlDB.SetMaxOpenConns(1)
    db.Exec("PRAGMA busy_timeout = 5000")
    require.NoError(t, db.AutoMigrate(
        &models.User{},
        &models.Meeting{},
        &models.MeetingParticipant{},
        &models.SignalingSession{},
        &models.SignalingMessage{},
    ))
    database.SetDB(db)
    defer func() {
        sqlDB.Close()
        database.SetDB(originalDB)
    }()

	meeting := models.Meeting{
		ID:              1,
		Title:           "Stress Meeting",
		CreatorID:       1,
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 200,
		Status:          models.MeetingStatusScheduled,
		MeetingType:     models.MeetingTypeVideo,
	}
	require.NoError(t, db.Create(&meeting).Error)

	totalUsers := 100
	for i := 1; i <= totalUsers; i++ {
		user := models.User{ID: uint(i), Username: fmt.Sprintf("user_%d", i), Email: fmt.Sprintf("user_%d@example.com", i), Password: "pwd", Status: models.UserStatusActive}
		require.NoError(t, db.Create(&user).Error)
		participant := models.MeetingParticipant{MeetingID: meeting.ID, UserID: user.ID, Role: models.ParticipantRoleParticipant, Status: models.ParticipantStatusJoined}
		require.NoError(t, db.Create(&participant).Error)
	}

	mockClient := &mockMeetingClient{}
	mockClient.reset()

	grpcClients := &sharedgrpc.ServiceClients{MeetingClient: mockClient}
	signalingService := services.NewSignalingService(grpcClients)
	handler := NewWebSocketHandler(signalingService)
	defer handler.Stop()

	router := gin.New()
	router.GET("/ws/signaling", handler.HandleWebSocket)
	server := httptest.NewServer(router)
	defer server.Close()

	dialer := websocket.DefaultDialer
	concurrency := 50
	var success int64
	var wg sync.WaitGroup

	for i := 1; i <= concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := uint(idx)
			token := generateStressJWT(t, userID)

			u, err := url.Parse(server.URL)
			require.NoError(t, err)
			u.Scheme = "ws"
			u.Path = "/ws/signaling"
			q := u.Query()
			q.Set("user_id", strconv.Itoa(idx))
			q.Set("meeting_id", "1")
			q.Set("peer_id", fmt.Sprintf("peer_%d", idx))
			u.RawQuery = q.Encode()

			header := map[string][]string{"Authorization": {"Bearer " + token}}
			conn, resp, err := dialer.Dial(u.String(), header)
			if err != nil {
				if resp != nil {
					bodyBytes, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					t.Logf("dial failed for user %d: %v (status %d, body %s)", userID, err, resp.StatusCode, string(bodyBytes))
				} else {
					t.Logf("dial failed for user %d: %v", userID, err)
				}
				return
			}
			defer conn.Close()

			join := models.WebSocketMessage{
				ID:         fmt.Sprintf("join_%d", idx),
				Type:       models.MessageTypeJoinRoom,
				MeetingID:  meeting.ID,
				FromUserID: userID,
				PeerID:     fmt.Sprintf("peer_%d", idx),
				SessionID:  "",
				Payload: models.JoinRoomRequest{
					MeetingID: meeting.ID,
					UserID:    userID,
					PeerID:    fmt.Sprintf("peer_%d", idx),
				},
				Timestamp: time.Now(),
			}
			payload, _ := json.Marshal(join)
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				t.Logf("write message failed for user %d: %v", userID, err)
				return
			}

			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Logf("read message error for user %d: %v", userID, err)
			}

			atomic.AddInt64(&success, 1)
		}(i)
	}

	wg.Wait()

	require.Equal(t, int64(concurrency), success, "not all clients completed successfully")
	require.Equal(t, concurrency, mockClient.countCalls(uint32(meeting.ID)))
}

func generateStressJWT(t *testing.T, userID uint) string {
	claims := middleware.JWTClaims{
		UserID:   userID,
		Username: fmt.Sprintf("user_%d", userID),
		Email:    fmt.Sprintf("user_%d@example.com", userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
	require.NoError(t, err)
	return signed
}
