package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
	"meeting-system/shared/utils"
)

// UserGRPCServer 用户服务gRPC服务器
type UserGRPCServer struct {
	pb.UnimplementedUserServiceServer
	db *gorm.DB
}

// NewUserGRPCServer 创建用户gRPC服务器
func NewUserGRPCServer() *UserGRPCServer {
	return &UserGRPCServer{
		db: database.GetDB(),
	}
}

// GetUser 获取用户信息
func (s *UserGRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	logger.Info("gRPC GetUser called", logger.Uint32("user_id", req.UserId))

	var user models.User
	if err := s.db.First(&user, req.UserId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %d", req.UserId)
		}
		logger.Error("Failed to get user from database", logger.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	response := &pb.GetUserResponse{
		Id:        uint32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.Nickname, // 使用Nickname代替FullName
		Status:    user.Status.String(),
		CreatedAt: timestamppb.New(user.CreatedAt),
	}

	logger.Info("User retrieved successfully", 
		logger.Uint32("user_id", uint32(user.ID)),
		logger.String("username", user.Username))

	return response, nil
}

// ValidateToken 验证JWT令牌
func (s *UserGRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	logger.Info("gRPC ValidateToken called")

	// 验证JWT令牌
	claims, err := utils.ValidateJWT(req.Token)
	if err != nil {
		logger.Error("Token validation failed", logger.Err(err))
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.ValidateTokenResponse{
				Valid: false,
				Error: "user not found",
			}, nil
		}
		logger.Error("Failed to get user from database", logger.Err(err))
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: "database error",
		}, nil
	}

	response := &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   uint32(user.ID),
		Username: user.Username,
	}

	logger.Info("Token validated successfully", 
		logger.Uint32("user_id", uint32(user.ID)),
		logger.String("username", user.Username))

	return response, nil
}

// GetUsersByIds 批量获取用户信息
func (s *UserGRPCServer) GetUsersByIds(ctx context.Context, req *pb.GetUsersByIdsRequest) (*pb.GetUsersByIdsResponse, error) {
	logger.Info("gRPC GetUsersByIds called", logger.Int("user_count", len(req.UserIds)))

	var users []models.User
	if err := s.db.Where("id IN ?", req.UserIds).Find(&users).Error; err != nil {
		logger.Error("Failed to get users from database", logger.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	var userResponses []*pb.GetUserResponse
	for _, user := range users {
		userResponse := &pb.GetUserResponse{
			Id:        uint32(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.Nickname, // 使用Nickname代替FullName
			Status:    user.Status.String(),
			CreatedAt: timestamppb.New(user.CreatedAt),
		}
		userResponses = append(userResponses, userResponse)
	}

	response := &pb.GetUsersByIdsResponse{
		Users: userResponses,
	}

	logger.Info("Users retrieved successfully", logger.Int("count", len(userResponses)))
	return response, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserGRPCServer) UpdateUserStatus(ctx context.Context, req *pb.UpdateUserStatusRequest) (*emptypb.Empty, error) {
	logger.Info("gRPC UpdateUserStatus called", 
		logger.Uint32("user_id", req.UserId),
		logger.String("status", req.Status))

	// 更新用户状态
	result := s.db.Model(&models.User{}).Where("id = ?", req.UserId).Update("status", req.Status)
	if result.Error != nil {
		logger.Error("Failed to update user status", logger.Err(result.Error))
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found: %d", req.UserId)
	}

	logger.Info("User status updated successfully", 
		logger.Uint32("user_id", req.UserId),
		logger.String("status", req.Status))

	return &emptypb.Empty{}, nil
}

// StartUserGRPCServer 启动用户服务gRPC服务器
func StartUserGRPCServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
	)

	userServer := NewUserGRPCServer()
	pb.RegisterUserServiceServer(grpcServer, userServer)

	logger.Info("User gRPC server starting", logger.Int("port", port))

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// HealthCheck 健康检查
func (s *UserGRPCServer) HealthCheck() error {
	// 检查数据库连接
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetMetrics 获取服务指标
func (s *UserGRPCServer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 获取用户总数
	var userCount int64
	s.db.Model(&models.User{}).Count(&userCount)
	metrics["total_users"] = userCount

	// 获取活跃用户数
	var activeUserCount int64
	s.db.Model(&models.User{}).Where("status = ?", "online").Count(&activeUserCount)
	metrics["active_users"] = activeUserCount

	// 获取数据库连接状态
	sqlDB, err := s.db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics["db_open_connections"] = stats.OpenConnections
		metrics["db_in_use"] = stats.InUse
		metrics["db_idle"] = stats.Idle
	}

	metrics["timestamp"] = time.Now().Unix()
	return metrics
}
