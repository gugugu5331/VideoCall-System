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
)

// MeetingGRPCServer 会议服务gRPC服务器
type MeetingGRPCServer struct {
	pb.UnimplementedMeetingServiceServer
	db *gorm.DB
}

// NewMeetingGRPCServer 创建会议gRPC服务器
func NewMeetingGRPCServer() *MeetingGRPCServer {
	return &MeetingGRPCServer{
		db: database.GetDB(),
	}
}

// GetMeeting 获取会议信息
func (s *MeetingGRPCServer) GetMeeting(ctx context.Context, req *pb.GetMeetingRequest) (*pb.GetMeetingResponse, error) {
	logger.Info("gRPC GetMeeting called", logger.Uint32("meeting_id", req.MeetingId))

	var meeting models.Meeting
	if err := s.db.First(&meeting, req.MeetingId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("meeting not found: %d", req.MeetingId)
		}
		logger.Error("Failed to get meeting from database", logger.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	response := &pb.GetMeetingResponse{
		Id:              uint32(meeting.ID),
		Title:           meeting.Title,
		Description:     meeting.Description,
		CreatorId:       uint32(meeting.CreatorID),
		Status:          meeting.Status.String(),
		StartTime:       timestamppb.New(meeting.StartTime),
		EndTime:         timestamppb.New(meeting.EndTime),
		MaxParticipants: int32(meeting.MaxParticipants),
	}

	logger.Info("Meeting retrieved successfully", 
		logger.Uint32("meeting_id", uint32(meeting.ID)),
		logger.String("title", meeting.Title))

	return response, nil
}

// ValidateUserAccess 验证用户访问权限
func (s *MeetingGRPCServer) ValidateUserAccess(ctx context.Context, req *pb.ValidateUserAccessRequest) (*pb.ValidateUserAccessResponse, error) {
	logger.Info("gRPC ValidateUserAccess called", 
		logger.Uint32("user_id", req.UserId),
		logger.Uint32("meeting_id", req.MeetingId))

	// 检查会议是否存在
	var meeting models.Meeting
	if err := s.db.First(&meeting, req.MeetingId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.ValidateUserAccessResponse{
				HasAccess: false,
				Error:     "meeting not found",
			}, nil
		}
		logger.Error("Failed to get meeting from database", logger.Err(err))
		return &pb.ValidateUserAccessResponse{
			HasAccess: false,
			Error:     "database error",
		}, nil
	}

	// 检查会议状态
	statusStr := meeting.Status.String()
	if statusStr != "active" && statusStr != "scheduled" {
		return &pb.ValidateUserAccessResponse{
			HasAccess: false,
			Error:     "meeting is not active",
		}, nil
	}

	// 检查用户是否是创建者
	if uint32(meeting.CreatorID) == req.UserId {
		return &pb.ValidateUserAccessResponse{
			HasAccess: true,
			Role:      "host",
		}, nil
	}

	// 检查用户是否在参与者列表中
	var participant models.MeetingParticipant
	err := s.db.Where("meeting_id = ? AND user_id = ?", req.MeetingId, req.UserId).First(&participant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.ValidateUserAccessResponse{
				HasAccess: false,
				Error:     "user not invited to meeting",
			}, nil
		}
		logger.Error("Failed to check participant", logger.Err(err))
		return &pb.ValidateUserAccessResponse{
			HasAccess: false,
			Error:     "database error",
		}, nil
	}

	response := &pb.ValidateUserAccessResponse{
		HasAccess: true,
		Role:      participant.Role.String(),
	}

	logger.Info("User access validated",
		logger.Uint32("user_id", req.UserId),
		logger.Uint32("meeting_id", req.MeetingId),
		logger.String("role", participant.Role.String()))

	return response, nil
}

// UpdateMeetingStatus 更新会议状态
func (s *MeetingGRPCServer) UpdateMeetingStatus(ctx context.Context, req *pb.UpdateMeetingStatusRequest) (*emptypb.Empty, error) {
	logger.Info("gRPC UpdateMeetingStatus called", 
		logger.Uint32("meeting_id", req.MeetingId),
		logger.String("status", req.Status))

	// 更新会议状态
	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.ParticipantCount > 0 {
		updates["current_participants"] = req.ParticipantCount
	}

	result := s.db.Model(&models.Meeting{}).Where("id = ?", req.MeetingId).Updates(updates)
	if result.Error != nil {
		logger.Error("Failed to update meeting status", logger.Err(result.Error))
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("meeting not found: %d", req.MeetingId)
	}

	logger.Info("Meeting status updated successfully", 
		logger.Uint32("meeting_id", req.MeetingId),
		logger.String("status", req.Status))

	return &emptypb.Empty{}, nil
}

// GetActiveMeetings 获取活跃会议列表
func (s *MeetingGRPCServer) GetActiveMeetings(ctx context.Context, req *emptypb.Empty) (*pb.GetActiveMeetingsResponse, error) {
	logger.Info("gRPC GetActiveMeetings called")

	var meetings []models.Meeting
	// 使用整数状态值：1=scheduled, 2=started/active
	if err := s.db.Where("status IN ?", []models.MeetingStatus{models.MeetingStatusScheduled, models.MeetingStatusStarted}).Find(&meetings).Error; err != nil {
		logger.Error("Failed to get active meetings", logger.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	var meetingResponses []*pb.GetMeetingResponse
	for _, meeting := range meetings {
		meetingResponse := &pb.GetMeetingResponse{
			Id:              uint32(meeting.ID),
			Title:           meeting.Title,
			Description:     meeting.Description,
			CreatorId:       uint32(meeting.CreatorID),
			Status:          meeting.Status.String(),
			StartTime:       timestamppb.New(meeting.StartTime),
			EndTime:         timestamppb.New(meeting.EndTime),
			MaxParticipants: int32(meeting.MaxParticipants),
		}
		meetingResponses = append(meetingResponses, meetingResponse)
	}

	response := &pb.GetActiveMeetingsResponse{
		Meetings: meetingResponses,
	}

	logger.Info("Active meetings retrieved", logger.Int("count", len(meetingResponses)))
	return response, nil
}

// StartMeetingGRPCServer 启动会议服务gRPC服务器
func StartMeetingGRPCServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB
	)

	meetingServer := NewMeetingGRPCServer()
	pb.RegisterMeetingServiceServer(grpcServer, meetingServer)

	logger.Info("Meeting gRPC server starting", logger.Int("port", port))

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// HealthCheck 健康检查
func (s *MeetingGRPCServer) HealthCheck() error {
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
func (s *MeetingGRPCServer) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 获取会议总数
	var totalMeetings int64
	s.db.Model(&models.Meeting{}).Count(&totalMeetings)
	metrics["total_meetings"] = totalMeetings

	// 获取活跃会议数
	var activeMeetings int64
	s.db.Model(&models.Meeting{}).Where("status = ?", "active").Count(&activeMeetings)
	metrics["active_meetings"] = activeMeetings

	// 获取今日会议数
	today := time.Now().Truncate(24 * time.Hour)
	var todayMeetings int64
	s.db.Model(&models.Meeting{}).Where("created_at >= ?", today).Count(&todayMeetings)
	metrics["today_meetings"] = todayMeetings

	metrics["timestamp"] = time.Now().Unix()
	return metrics
}
