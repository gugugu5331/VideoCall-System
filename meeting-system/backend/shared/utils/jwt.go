package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"meeting-system/shared/config"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username, email string) (string, error) {
	// 使用默认值，避免依赖全局配置
	expireTime := 24                              // 默认24小时
	secretKey := "default-secret-key-for-testing" // 默认密钥

	if config.GlobalConfig != nil && config.GlobalConfig.JWT.ExpireTime > 0 {
		expireTime = config.GlobalConfig.JWT.ExpireTime
	}
	if config.GlobalConfig != nil && config.GlobalConfig.JWT.Secret != "" {
		secretKey = config.GlobalConfig.JWT.Secret
	}

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "meeting-system",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*JWTClaims, error) {
	secretKey := "default-secret-key-for-testing" // 默认密钥
	if config.GlobalConfig != nil && config.GlobalConfig.JWT.Secret != "" {
		secretKey = config.GlobalConfig.JWT.Secret
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RefreshToken 刷新JWT token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查token是否即将过期（1小时内）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return tokenString, nil // 不需要刷新
	}

	// 生成新token
	return GenerateToken(claims.UserID, claims.Username, claims.Email)
}
