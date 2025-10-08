package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
	"meeting-system/shared/utils"
)

// UserService 用户服务
type UserService struct {
	db    *gorm.DB
	cache *database.RedisCache
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		db:    database.GetDB(),
		cache: database.NewRedisCache(),
	}
}

// Register 用户注册
func (s *UserService) Register(req *models.UserCreateRequest) (*models.User, error) {
	// 验证用户名格式
	if !utils.IsValidUsername(req.Username) {
		return nil, errors.New("invalid username format")
	}

	// 验证邮箱格式
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// 验证密码强度（增强版）
	valid, message := utils.ValidatePasswordStrength(req.Password)
	if !valid {
		return nil, errors.New(message)
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Nickname:     req.Nickname,
		Phone:        req.Phone,
		Role:         models.UserRoleUser, // 默认为普通用户
		Status:       models.UserStatusActive,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User registered successfully")

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	// 查找用户（支持用户名或邮箱登录）
	var user models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("user account is inactive")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	s.db.Save(&user)

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 缓存用户会话
	ctx := context.Background()
	sessionKey := fmt.Sprintf("user:session:%d", user.ID)
	if err := s.cache.Set(ctx, sessionKey, token, 24*time.Hour); err != nil {
		logger.Warn("Failed to cache user session: " + err.Error())
	}

	logger.Info("User logged in successfully")

	return &models.UserLoginResponse{
		User:  user.ToProfile(),
		Token: token,
	}, nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(userID uint) (*models.UserProfile, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToProfile(), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID uint, req *models.UserUpdateRequest) (*models.UserProfile, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.AvatarURL = req.Avatar
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User profile updated")

	return user.ToProfile(), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	// 验证新密码强度（增强版）
	valid, message := utils.ValidatePasswordStrength(req.NewPassword)
	if !valid {
		return errors.New(message)
	}

	// 确保新密码与旧密码不同
	if req.OldPassword == req.NewPassword {
		return errors.New("new password must be different from old password")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info("User password changed")

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, keyword string) ([]*models.UserProfile, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	// 搜索条件
	if keyword != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR nickname ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// 转换为用户资料
	profiles := make([]*models.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = user.ToProfile()
	}

	return profiles, total, nil
}

// GetUser 获取指定用户
func (s *UserService) GetUser(userID uint) (*models.UserProfile, error) {
	return s.GetProfile(userID)
}

// UpdateUser 更新指定用户（管理员功能）
func (s *UserService) UpdateUser(userID uint, req *models.UserUpdateRequest) (*models.UserProfile, error) {
	return s.UpdateProfile(userID, req)
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint) error {
	if err := s.db.Delete(&models.User{}, userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted")
	return nil
}

// BanUser 禁用用户
func (s *UserService) BanUser(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Status = models.UserStatusBanned
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	logger.Info("User banned")

	return nil
}

// UnbanUser 解禁用户
func (s *UserService) UnbanUser(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Status = models.UserStatusActive
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	logger.Info("User unbanned")

	return nil
}

// RefreshToken 刷新token
func (s *UserService) RefreshToken(tokenString string) (string, error) {
	return utils.RefreshToken(tokenString)
}
