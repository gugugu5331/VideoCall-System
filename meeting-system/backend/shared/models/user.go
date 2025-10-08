package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string         `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string         `json:"-" gorm:"column:password_hash;size:255;not null"` // 不在JSON中显示密码
	Nickname     string         `json:"nickname" gorm:"size:50"`
	AvatarURL    string         `json:"avatar_url" gorm:"column:avatar_url;size:255"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Role         UserRole       `json:"role" gorm:"default:1"` // 用户角色
	Status       UserStatus     `json:"status" gorm:"default:1"`
	LastLogin    *time.Time     `json:"last_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	CreatedMeetings []Meeting            `json:"created_meetings,omitempty" gorm:"foreignKey:CreatorID"`
	Participations  []MeetingParticipant `json:"participations,omitempty" gorm:"foreignKey:UserID"`
}

// UserRole 用户角色
type UserRole int

const (
	UserRoleGuest UserRole = 0 // 访客
	UserRoleUser  UserRole = 1 // 普通用户
	UserRoleMod   UserRole = 2 // 版主
	UserRoleAdmin UserRole = 3 // 管理员
	UserRoleSuper UserRole = 4 // 超级管理员
)

// String 返回角色字符串
func (r UserRole) String() string {
	switch r {
	case UserRoleGuest:
		return "guest"
	case UserRoleUser:
		return "user"
	case UserRoleMod:
		return "moderator"
	case UserRoleAdmin:
		return "admin"
	case UserRoleSuper:
		return "super_admin"
	default:
		return "unknown"
	}
}

// IsAdmin 检查是否为管理员
func (r UserRole) IsAdmin() bool {
	return r >= UserRoleAdmin
}

// IsModerator 检查是否为版主或更高权限
func (r UserRole) IsModerator() bool {
	return r >= UserRoleMod
}

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusInactive UserStatus = 0 // 未激活
	UserStatusActive   UserStatus = 1 // 激活
	UserStatusBanned   UserStatus = 2 // 禁用
)

// String 返回状态字符串
func (s UserStatus) String() string {
	switch s {
	case UserStatusInactive:
		return "inactive"
	case UserStatusActive:
		return "online"
	case UserStatusBanned:
		return "banned"
	default:
		return "unknown"
	}
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role.IsAdmin()
}

// IsModerator 检查用户是否为版主或更高权限
func (u *User) IsModerator() bool {
	return u.Role.IsModerator()
}

// UserProfile 用户资料（用于返回给前端，不包含敏感信息）
type UserProfile struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Phone     string     `json:"phone"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToProfile 转换为用户资料
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.AvatarURL,
		Phone:     u.Phone,
		Role:      u.Role,
		Status:    u.Status,
		LastLogin: u.LastLogin,
		CreatedAt: u.CreatedAt,
	}
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Phone    string `json:"phone" binding:"max=20"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	User  *UserProfile `json:"user"`
	Token string       `json:"token"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}
