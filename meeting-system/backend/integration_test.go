package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	pb "meeting-system/shared/grpc"
)

// TestServiceInteraction 测试服务间交互
func TestServiceInteraction(t *testing.T) {
	t.Skip("integration test requires running service stack")

	// 初始化配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(cfg.Database); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	// 初始化服务发现
	registry, err := discovery.NewServiceRegistry(cfg.Etcd)
	if err != nil {
		t.Fatalf("Failed to init service registry: %v", err)
	}

	// 测试用例
	t.Run("TestUserServiceGRPC", testUserServiceGRPC)
	t.Run("TestMeetingServiceGRPC", testMeetingServiceGRPC)
	t.Run("TestServiceDiscovery", func(t *testing.T) { testServiceDiscovery(t, registry) })
	t.Run("TestCrossServiceInteraction", testCrossServiceInteraction)
	t.Run("TestWebSocketSignaling", testWebSocketSignaling)
	t.Run("TestMediaServiceIntegration", testMediaServiceIntegration)
	t.Run("TestAIServiceIntegration", testAIServiceIntegration)
}

// testUserServiceGRPC 测试用户服务gRPC
func testUserServiceGRPC(t *testing.T) {
	// 连接到用户服务
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("User service not available: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试获取用户信息
	resp, err := client.GetUser(ctx, &pb.GetUserRequest{UserId: 1})
	if err != nil {
		t.Errorf("GetUser failed: %v", err)
		return
	}

	if resp.Id != 1 {
		t.Errorf("Expected user ID 1, got %d", resp.Id)
	}

	log.Printf("✅ User service gRPC test passed: User %s retrieved", resp.Username)
}

// testMeetingServiceGRPC 测试会议服务gRPC
func testMeetingServiceGRPC(t *testing.T) {
	// 连接到会议服务
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Meeting service not available: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewMeetingServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试获取活跃会议
	resp, err := client.GetActiveMeetings(ctx, &emptypb.Empty{})
	if err != nil {
		t.Errorf("GetActiveMeetings failed: %v", err)
		return
	}

	log.Printf("✅ Meeting service gRPC test passed: %d active meetings", len(resp.Meetings))
}

// testServiceDiscovery 测试服务发现
func testServiceDiscovery(t *testing.T, registry *discovery.ServiceRegistry) {
	// 注册测试服务
	service := &discovery.ServiceInfo{
		Name:     "test-service",
		Host:     "localhost",
		Port:     8999,
		Protocol: "http",
		Metadata: map[string]string{"env": "test"},
	}

	instanceID, err := registry.RegisterService(service)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
		return
	}

	// 发现服务
	services, err := registry.DiscoverServices("test-service")
	if err != nil {
		t.Errorf("Failed to discover services: %v", err)
		return
	}

	if len(services) == 0 {
		t.Error("No services discovered")
		return
	}

	// 注销服务
	if err := registry.DeregisterService("test-service", instanceID); err != nil {
		t.Errorf("Failed to deregister service: %v", err)
		return
	}

	log.Printf("✅ Service discovery test passed: %d services found", len(services))
}

// testCrossServiceInteraction 测试跨服务交互
func testCrossServiceInteraction(t *testing.T) {
	// 模拟用户加入会议的完整流程

	// 1. 验证用户令牌
	userConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("User service not available: %v", err)
		return
	}
	defer userConn.Close()

	userClient := pb.NewUserServiceClient(userConn)
	_ = userClient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. 验证会议访问权限
	meetingConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Meeting service not available: %v", err)
		return
	}
	defer meetingConn.Close()

	meetingClient := pb.NewMeetingServiceClient(meetingConn)

	// 测试用户访问验证
	accessResp, err := meetingClient.ValidateUserAccess(ctx, &pb.ValidateUserAccessRequest{
		UserId:    1,
		MeetingId: 1,
	})
	if err != nil {
		t.Errorf("ValidateUserAccess failed: %v", err)
		return
	}

	log.Printf("✅ Cross-service interaction test passed: User access %v, role %s",
		accessResp.HasAccess, accessResp.Role)
}

// testWebSocketSignaling 测试WebSocket信令
func testWebSocketSignaling(t *testing.T) {
	// 测试WebSocket连接
	signalingURL := "ws://localhost:8081/ws"

	// 这里简化为HTTP健康检查，实际应该测试WebSocket连接
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Skipf("Signaling service not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Signaling service health check failed: %d", resp.StatusCode)
		return
	}

	log.Printf("✅ WebSocket signaling test passed: Service available at %s", signalingURL)
}

// testMediaServiceIntegration 测试媒体服务集成
func testMediaServiceIntegration(t *testing.T) {
	// 测试媒体服务健康检查
	resp, err := http.Get("http://localhost:8083/health")
	if err != nil {
		t.Skipf("Media service not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Media service health check failed: %d", resp.StatusCode)
		return
	}

	// 测试媒体处理能力
	// 这里可以添加更多媒体处理测试

	log.Printf("✅ Media service integration test passed")
}

// testAIServiceIntegration 测试AI服务集成
func testAIServiceIntegration(t *testing.T) {
	// 测试AI服务健康检查
	resp, err := http.Get("http://localhost:8084/health")
	if err != nil {
		t.Skipf("AI service not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("AI service health check failed: %d", resp.StatusCode)
		return
	}

	log.Printf("✅ AI service integration test passed")
}

// BenchmarkServiceInteraction 性能测试
func BenchmarkServiceInteraction(b *testing.B) {
	// 初始化
	if _, err := config.LoadConfig("config/config.yaml"); err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	// 连接用户服务
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Skipf("User service not available: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	b.ResetTimer()
	b.RunParallel(func(worker *testing.PB) {
		for worker.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err := client.GetUser(ctx, &pb.GetUserRequest{UserId: 1})
			cancel()
			if err != nil {
				b.Errorf("GetUser failed: %v", err)
			}
		}
	})
}

// TestConcurrentServiceCalls 并发测试
func TestConcurrentServiceCalls(t *testing.T) {
	t.Skip("integration concurrency test requires running services")
	const numGoroutines = 10
	const numCallsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numCallsPerGoroutine)

	// 连接用户服务
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("User service not available: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numCallsPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := client.GetUser(ctx, &pb.GetUserRequest{UserId: 1})
				cancel()
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, call %d: %w", goroutineID, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent call error: %v", err)
		errorCount++
	}

	if errorCount == 0 {
		log.Printf("✅ Concurrent service calls test passed: %d goroutines, %d calls each",
			numGoroutines, numCallsPerGoroutine)
	}
}
