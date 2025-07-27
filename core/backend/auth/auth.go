package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"videocall-backend/config"
	"videocall-backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	config *config.Config
}

// NewAuthService 创建认证服务实例
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// HashPassword 密码加密
func (as *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func (as *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT 生成JWT token
func (as *AuthService) GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(as.config.JWT.ExpireTime) * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID,
		UserUUID: user.UUID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "videocall-backend",
			Subject:   user.UUID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(as.config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT 验证JWT token
func (as *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(as.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRefreshToken 生成刷新token
func (as *AuthService) GenerateRefreshToken() (string, error) {
	// 生成一个随机的刷新token
	refreshToken := generateRandomString(64)
	return refreshToken, nil
}

// ValidateUserCredentials 验证用户凭据
func (as *AuthService) ValidateUserCredentials(username, password string, user *models.User) bool {
	// 检查用户名是否匹配
	if user.Username != username {
		return false
	}

	// 检查密码
	return as.CheckPassword(password, user.PasswordHash)
}

// 辅助函数
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
