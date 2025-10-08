package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
)

type MeetingService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewMeetingService() *MeetingService {
	return &MeetingService{
		db:    database.GetDB(),
		redis: database.GetRedis(),
	}
}

// 辅助函数：将字符串转换为MeetingType
func stringToMeetingType(s string) models.MeetingType {
	switch s {
	case "video":
		return models.MeetingTypeVideo
	case "audio":
		return models.MeetingTypeAudio
	default:
		return models.MeetingTypeVideo
	}
}

// generateRoomID 生成高熵的会议室ID，减少并发场景下的碰撞概率
func generateRoomID(meetingID uint) string {
	seed := time.Now().UTC().UnixNano()
	r := rand.New(rand.NewSource(seed))
	return fmt.Sprintf("room_%d_%d_%06d", meetingID, seed, r.Intn(1_000_000))
}

// 辅助函数：将MeetingSettings序列化为JSON字符串
func settingsToJSON(settings models.MeetingSettings) string {
	data, _ := json.Marshal(settings)
	return string(data)
}

// 辅助函数：将MeetingType转换为字符串
func meetingTypeToString(mt models.MeetingType) string {
	switch mt {
	case models.MeetingTypeVideo:
		return "video"
	case models.MeetingTypeAudio:
		return "audio"
	case models.MeetingTypePublic:
		return "public"
	case models.MeetingTypePrivate:
		return "private"
	default:
		return "video"
	}
}

// 辅助函数：将JSON字符串解析为MeetingSettings
func jsonToSettings(jsonStr string) models.MeetingSettings {
	var settings models.MeetingSettings
	json.Unmarshal([]byte(jsonStr), &settings)
	return settings
}

// CreateMeeting 创建会议
func (s *MeetingService) CreateMeeting(req *models.CreateMeetingRequest) (*models.Meeting, error) {
	var meeting *models.Meeting

	// 使用事务确保数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		meeting = &models.Meeting{
			Title:           req.Title,
			Description:     req.Description,
			CreatorID:       req.CreatorID,
			StartTime:       req.StartTime,
			EndTime:         req.EndTime,
			MaxParticipants: req.MaxParticipants,
			MeetingType:     stringToMeetingType(req.MeetingType),
			Password:        req.Password,
			Settings:        settingsToJSON(req.Settings),
			Status:          models.MeetingStatusScheduled,
		}

		// 创建会议
		if err := tx.Create(meeting).Error; err != nil {
			logger.Error("Failed to create meeting", logger.Err(err))
			return err
		}

		// 添加创建者为主持人
		participant := &models.MeetingParticipant{
			MeetingID: meeting.ID,
			UserID:    req.CreatorID,
			Role:      models.ParticipantRoleHost,
			Status:    models.ParticipantStatusInvited,
		}

		if err := tx.Create(participant).Error; err != nil {
			logger.Error("Failed to add creator as participant", logger.Err(err))
			return err // 事务会自动回滚
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 缓存会议信息
	s.cacheMeeting(meeting)

	logger.Info("Meeting created successfully",
		logger.Uint("meeting_id", meeting.ID),
		logger.Uint("creator_id", req.CreatorID))

	return meeting, nil
}

// GetMeeting 获取会议信息
func (s *MeetingService) GetMeeting(meetingID uint, userID uint) (*models.MeetingResponse, error) {
	// 先从缓存获取
	if meeting := s.getMeetingFromCache(meetingID); meeting != nil {
		// 检查用户权限
		if !s.canAccessMeeting(meetingID, userID) {
			return nil, fmt.Errorf("access denied")
		}
		return s.buildMeetingResponse(meeting)
	}

	// 从数据库获取
	var meeting models.Meeting
	if err := s.db.Preload("Participants").Preload("Room").First(&meeting, meetingID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("meeting not found")
		}
		return nil, err
	}

	// 检查用户权限
	if !s.canAccessMeeting(meetingID, userID) {
		return nil, fmt.Errorf("access denied")
	}

	// 缓存会议信息
	s.cacheMeeting(&meeting)

	return s.buildMeetingResponse(&meeting)
}

// UpdateMeeting 更新会议
func (s *MeetingService) UpdateMeeting(meetingID uint, userID uint, req *models.UpdateMeetingRequest) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return nil, err
	}

	// 检查权限（只有创建者或主持人可以修改）
	if !s.canModifyMeeting(meetingID, userID) {
		return nil, fmt.Errorf("permission denied")
	}

	// 更新字段
	if req.Title != nil && *req.Title != "" {
		meeting.Title = *req.Title
	}
	if req.Description != nil && *req.Description != "" {
		meeting.Description = *req.Description
	}
	if req.StartTime != nil && !req.StartTime.IsZero() {
		meeting.StartTime = *req.StartTime
	}
	if req.EndTime != nil && !req.EndTime.IsZero() {
		meeting.EndTime = *req.EndTime
	}
	if req.MaxParticipants != nil && *req.MaxParticipants > 0 {
		meeting.MaxParticipants = *req.MaxParticipants
	}
	if req.Settings != nil {
		meeting.Settings = settingsToJSON(*req.Settings)
	}

	if err := s.db.Save(&meeting).Error; err != nil {
		return nil, err
	}

	// 更新缓存
	s.cacheMeeting(&meeting)

	return &meeting, nil
}

// DeleteMeeting 删除会议
func (s *MeetingService) DeleteMeeting(meetingID uint, userID uint) error {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return err
	}

	// 检查权限（只有创建者可以删除）
	if meeting.CreatorID != userID {
		return fmt.Errorf("permission denied")
	}

	// 软删除
	if err := s.db.Delete(&meeting).Error; err != nil {
		return err
	}

	// 删除缓存
	s.deleteMeetingFromCache(meetingID)

	return nil
}

// JoinMeeting 加入会议
func (s *MeetingService) JoinMeeting(meetingID uint, userID uint, password string) (*models.JoinMeetingResponse, error) {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return nil, err
	}

	// 检查会议状态
	if meeting.Status != models.MeetingStatusOngoing && meeting.Status != models.MeetingStatusScheduled {
		return nil, fmt.Errorf("meeting is not available")
	}

	// 检查密码
	if meeting.Password != "" && meeting.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	// 检查参与者是否已存在
	var participant models.MeetingParticipant
	err := s.db.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&participant).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新参与者
		participant = models.MeetingParticipant{
			MeetingID: meetingID,
			UserID:    userID,
			Role:      models.ParticipantRoleParticipant,
			Status:    models.ParticipantStatusJoined,
			JoinedAt:  &time.Time{},
		}
		now := time.Now()
		participant.JoinedAt = &now

		if err := s.db.Create(&participant).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// 更新参与者状态
		now := time.Now()
		participant.Status = models.ParticipantStatusJoined
		participant.JoinedAt = &now
		participant.LeftAt = nil

		if err := s.db.Save(&participant).Error; err != nil {
			return nil, err
		}
	}

	// 如果会议还未开始，自动开始
	if meeting.Status == models.MeetingStatusScheduled {
		meeting.Status = models.MeetingStatusOngoing
		s.db.Save(&meeting)
		s.cacheMeeting(&meeting)
	}

	// 获取或创建会议室
	room, err := s.getOrCreateMeetingRoom(meetingID)
	if err != nil {
		return nil, err
	}

	response := &models.JoinMeetingResponse{
		MeetingID:     meetingID,
		RoomID:        room.RoomID,
		ParticipantID: participant.ID,
		Role:          participant.Role,
		SFUNode:       room.SFUNode,
	}

	return response, nil
}

// LeaveMeeting 离开会议
func (s *MeetingService) LeaveMeeting(meetingID uint, userID uint) error {
	var participant models.MeetingParticipant
	if err := s.db.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&participant).Error; err != nil {
		return err
	}

	// 更新参与者状态
	now := time.Now()
	participant.Status = models.ParticipantStatusLeft
	participant.LeftAt = &now

	if err := s.db.Save(&participant).Error; err != nil {
		return err
	}

	return nil
}

// GetParticipants 获取会议参与者
func (s *MeetingService) GetParticipants(meetingID uint, userID uint) ([]*models.ParticipantResponse, error) {
	// 检查权限
	if !s.canAccessMeeting(meetingID, userID) {
		return nil, fmt.Errorf("access denied")
	}

	var participants []models.MeetingParticipant
	if err := s.db.Where("meeting_id = ?", meetingID).Find(&participants).Error; err != nil {
		return nil, err
	}

	var responses []*models.ParticipantResponse
	for _, p := range participants {
		responses = append(responses, &models.ParticipantResponse{
			ID:       p.ID,
			UserID:   p.UserID,
			Role:     p.Role,
			Status:   p.Status,
			JoinedAt: p.JoinedAt,
			LeftAt:   p.LeftAt,
		})
	}

	return responses, nil
}

// ListMeetings 获取用户相关会议列表
func (s *MeetingService) ListMeetings(userID uint, req *models.MeetingListRequest) ([]*models.MeetingResponse, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	filtered := s.db.Model(&models.Meeting{}).
		Joins("LEFT JOIN meeting_participants mp ON mp.meeting_id = meetings.id AND mp.deleted_at IS NULL").
		Where("meetings.creator_id = ? OR mp.user_id = ?", userID, userID)

	if req.Status != 0 {
		filtered = filtered.Where("meetings.status = ?", req.Status)
	}

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		filtered = filtered.Where("(meetings.title ILIKE ? OR meetings.description ILIKE ?)", like, like)
	}

	countQuery := filtered.Session(&gorm.Session{}).Distinct("meetings.id")
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.MeetingResponse{}, 0, nil
	}

	offset := (page - 1) * pageSize

	// 直接查询会议，不使用 Pluck
	var meetings []models.Meeting
	if err := filtered.Session(&gorm.Session{}).
		Distinct("meetings.*").
		Order("meetings.start_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&meetings).Error; err != nil {
		return nil, 0, err
	}

	if len(meetings) == 0 {
		return []*models.MeetingResponse{}, total, nil
	}

	responses := make([]*models.MeetingResponse, 0, len(meetings))
	for i := range meetings {
		resp, err := s.buildMeetingResponse(&meetings[i])
		if err != nil {
			return nil, 0, err
		}
		responses = append(responses, resp)
	}

	return responses, total, nil
}

// 辅助方法

// canAccessMeeting 检查用户是否可以访问会议
func (s *MeetingService) canAccessMeeting(meetingID uint, userID uint) bool {
	var count int64
	s.db.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count)

	if count > 0 {
		return true
	}

	// 检查是否是会议创建者
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return false
	}

	return meeting.CreatorID == userID
}

// canModifyMeeting 检查用户是否可以修改会议
func (s *MeetingService) canModifyMeeting(meetingID uint, userID uint) bool {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return false
	}

	// 创建者可以修改
	if meeting.CreatorID == userID {
		return true
	}

	// 主持人可以修改
	var participant models.MeetingParticipant
	if err := s.db.Where("meeting_id = ? AND user_id = ? AND role = ?",
		meetingID, userID, models.ParticipantRoleHost).First(&participant).Error; err != nil {
		return false
	}

	return true
}

// cacheMeeting 缓存会议信息
func (s *MeetingService) cacheMeeting(meeting *models.Meeting) {
	key := fmt.Sprintf("meeting:%d", meeting.ID)
	data, _ := json.Marshal(meeting)
	s.redis.Set(context.Background(), key, string(data), 30*time.Minute)
}

// getMeetingFromCache 从缓存获取会议信息
func (s *MeetingService) getMeetingFromCache(meetingID uint) *models.Meeting {
	key := fmt.Sprintf("meeting:%d", meetingID)
	data, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	var meeting models.Meeting
	if err := json.Unmarshal([]byte(data), &meeting); err != nil {
		return nil
	}

	return &meeting
}

// deleteMeetingFromCache 从缓存删除会议信息
func (s *MeetingService) deleteMeetingFromCache(meetingID uint) {
	key := fmt.Sprintf("meeting:%d", meetingID)
	s.redis.Del(context.Background(), key)
}

// getOrCreateMeetingRoom 获取或创建会议室
func (s *MeetingService) getOrCreateMeetingRoom(meetingID uint) (*models.MeetingRoom, error) {
	var room models.MeetingRoom

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 优先尝试在行级锁下读取既有会议室，避免重复创建
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("meeting_id = ?", meetingID).
			First(&room).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 锁定会议记录以串行化后续创建流程
		var meeting models.Meeting
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&meeting, meetingID).Error; err != nil {
			return err
		}

		// 再次检查是否已有其它事务创建
		if err := tx.Where("meeting_id = ?", meetingID).First(&room).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		room = models.MeetingRoom{
			MeetingID:        meetingID,
			RoomID:           generateRoomID(meetingID),
			SFUNode:          "sfu-node-1", // TODO: 实现SFU节点选择逻辑
			Status:           models.RoomStatusActive,
			ParticipantCount: 0,
			MaxBitrate:       1000000,
		}

		if err := tx.Create(&room).Error; err != nil {
			if fetchErr := tx.Where("meeting_id = ?", meetingID).First(&room).Error; fetchErr == nil {
				return nil
			}
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &room, nil
}

// buildMeetingResponse 构建会议响应
func (s *MeetingService) buildMeetingResponse(meeting *models.Meeting) (*models.MeetingResponse, error) {
	response := &models.MeetingResponse{
		ID:              meeting.ID,
		Title:           meeting.Title,
		Description:     meeting.Description,
		CreatorID:       meeting.CreatorID,
		StartTime:       meeting.StartTime,
		EndTime:         meeting.EndTime,
		MaxParticipants: meeting.MaxParticipants,
		Status:          meeting.Status,
		MeetingType:     meetingTypeToString(meeting.MeetingType),
		Settings:        jsonToSettings(meeting.Settings),
		CreatedAt:       meeting.CreatedAt,
		UpdatedAt:       meeting.UpdatedAt,
	}

	return response, nil
}
