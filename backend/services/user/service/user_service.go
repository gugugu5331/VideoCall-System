package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"video-conference-system/services/user/repository"
	"video-conference-system/shared/auth"
	"video-conference-system/shared/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
type UserService interface {
	Register(req *models.UserCreateRequest) (*models.LoginResponse, error)
	Login(req *models.UserLoginRequest) (*models.LoginResponse, error)
	RefreshToken(refreshToken string) (*models.LoginResponse, error)
	GetProfile(userID string) (*models.UserResponse, error)
	GetUser(userID string) (*models.UserResponse, error)
	UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.UserResponse, error)
	Logout(userID string) error
	ValidateToken(token string) (*auth.Claims, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
	redis      *redis.Client
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager, redis *redis.Client) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		redis:      redis,
	}
}

// Register 用户注册
func (s *userService) Register(req *models.UserCreateRequest) (*models.LoginResponse, error) {
	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 生成令牌
	return s.generateTokenResponse(user)
}

// Login 用户登录
func (s *userService) Login(req *models.UserLoginRequest) (*models.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("user account is not active")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		// 记录错误但不影响登录
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// 生成令牌
	return s.generateTokenResponse(user)
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 检查是否是刷新令牌
	if len(claims.Audience) == 0 || claims.Audience[0] != "video-conference-refresh" {
		return nil, errors.New("invalid refresh token")
	}

	// 获取用户信息
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("user account is not active")
	}

	// 生成新的令牌
	return s.generateTokenResponse(user)
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(userID string) (*models.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// GetUser 获取用户信息
func (s *userService) GetUser(userID string) (*models.UserResponse, error) {
	return s.GetProfile(userID)
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(userID string, req *models.UserUpdateRequest) (*models.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user.ToResponse(), nil
}

// Logout 用户登出
func (s *userService) Logout(userID string) error {
	// 从Redis中删除用户会话
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", userID)
	return s.redis.Del(ctx, sessionKey).Err()
}

// ValidateToken 验证令牌
func (s *userService) ValidateToken(token string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

// generateTokenResponse 生成令牌响应
func (s *userService) generateTokenResponse(user *models.User) (*models.LoginResponse, error) {
	// 生成访问令牌
	token, err := s.jwtManager.GenerateToken(
		user.ID.String(),
		user.Username,
		user.Email,
		"user", // 默认角色
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 将会话信息存储到Redis
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", user.ID.String())
	sessionData := map[string]interface{}{
		"user_id":       user.ID.String(),
		"username":      user.Username,
		"email":         user.Email,
		"last_activity": time.Now(),
	}

	sessionJSON, _ := json.Marshal(sessionData)
	s.redis.Set(ctx, sessionKey, sessionJSON, 24*time.Hour)

	return &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}
