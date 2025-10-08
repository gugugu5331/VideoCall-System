package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/logger"
	"meeting-system/shared/response"
)

const (
	// CSRFTokenLength CSRF token 长度
	CSRFTokenLength = 32
	// CSRFTokenHeader CSRF token 请求头名称
	CSRFTokenHeader = "X-CSRF-Token"
	// CSRFCookieName CSRF cookie 名称
	CSRFCookieName = "csrf_token"
	// CSRFTokenTTL CSRF token 有效期
	CSRFTokenTTL = 24 * time.Hour
)

// csrfToken CSRF token 结构
type csrfToken struct {
	Token     string
	ExpiresAt time.Time
}

// csrfStore CSRF token 存储
type csrfStore struct {
	tokens map[string]*csrfToken
	mu     sync.RWMutex
}

// newCSRFStore 创建 CSRF token 存储
func newCSRFStore() *csrfStore {
	store := &csrfStore{
		tokens: make(map[string]*csrfToken),
	}
	// 启动清理过期 token 的 goroutine
	go store.cleanupExpiredTokens()
	return store
}

// set 存储 token
func (s *csrfStore) set(token string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = &csrfToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// get 获取 token
func (s *csrfStore) get(token string) (*csrfToken, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, exists := s.tokens[token]
	if !exists {
		return nil, false
	}
	// 检查是否过期
	if time.Now().After(t.ExpiresAt) {
		return nil, false
	}
	return t, true
}

// delete 删除 token
func (s *csrfStore) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
}

// cleanupExpiredTokens 清理过期的 token
func (s *csrfStore) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, t := range s.tokens {
			if now.After(t.ExpiresAt) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

// 全局 CSRF token 存储
var globalCSRFStore = newCSRFStore()

// generateCSRFToken 生成 CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CSRFProtection CSRF 保护中间件
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于安全方法（GET, HEAD, OPTIONS），不需要验证 CSRF token
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 从请求头获取 CSRF token
		token := c.GetHeader(CSRFTokenHeader)
		if token == "" {
			// 尝试从表单获取
			token = c.PostForm("csrf_token")
		}

		if token == "" {
			logger.Warn("CSRF token missing",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()))
			response.Error(c, http.StatusForbidden, "CSRF token missing")
			c.Abort()
			return
		}

		// 验证 token
		if _, exists := globalCSRFStore.get(token); !exists {
			logger.Warn("Invalid or expired CSRF token",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()))
			response.Error(c, http.StatusForbidden, "Invalid or expired CSRF token")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CSRFTokenGenerator CSRF token 生成中间件
// 为每个会话生成并返回 CSRF token
func CSRFTokenGenerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有有效的 CSRF token
		existingToken, err := c.Cookie(CSRFCookieName)
		if err == nil && existingToken != "" {
			if _, exists := globalCSRFStore.get(existingToken); exists {
				// token 仍然有效，继续使用
				c.Header(CSRFTokenHeader, existingToken)
				c.Next()
				return
			}
		}

		// 生成新的 CSRF token
		token, err := generateCSRFToken()
		if err != nil {
			logger.Error("Failed to generate CSRF token", logger.Err(err))
			response.Error(c, http.StatusInternalServerError, "Failed to generate CSRF token")
			c.Abort()
			return
		}

		// 存储 token
		expiresAt := time.Now().Add(CSRFTokenTTL)
		globalCSRFStore.set(token, expiresAt)

		// 设置 cookie
		c.SetCookie(
			CSRFCookieName,
			token,
			int(CSRFTokenTTL.Seconds()),
			"/",
			"",
			false, // 开发环境使用 HTTP，生产环境应设置为 true
			true,  // HttpOnly
		)

		// 在响应头中返回 token（方便前端获取）
		c.Header(CSRFTokenHeader, token)

		c.Next()
	}
}

// GetCSRFToken 获取 CSRF token 的处理函数
// 用于前端获取 CSRF token
func GetCSRFToken(c *gin.Context) {
	// 检查是否已有有效的 CSRF token
	existingToken, err := c.Cookie(CSRFCookieName)
	if err == nil && existingToken != "" {
		if _, exists := globalCSRFStore.get(existingToken); exists {
			response.Success(c, gin.H{
				"csrf_token": existingToken,
			})
			return
		}
	}

	// 生成新的 CSRF token
	token, err := generateCSRFToken()
	if err != nil {
		logger.Error("Failed to generate CSRF token", logger.Err(err))
		response.Error(c, http.StatusInternalServerError, "Failed to generate CSRF token")
		return
	}

	// 存储 token
	expiresAt := time.Now().Add(CSRFTokenTTL)
	globalCSRFStore.set(token, expiresAt)

	// 设置 cookie
	c.SetCookie(
		CSRFCookieName,
		token,
		int(CSRFTokenTTL.Seconds()),
		"/",
		"",
		false, // 开发环境使用 HTTP，生产环境应设置为 true
		true,  // HttpOnly
	)

	response.Success(c, gin.H{
		"csrf_token": token,
	})
}

// CSRFExempt 豁免 CSRF 保护的中间件
// 用于某些不需要 CSRF 保护的端点（如公开 API）
func CSRFExempt() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("csrf_exempt", true)
		c.Next()
	}
}

// SmartCSRFProtection 智能 CSRF 保护中间件
// 根据认证方式智能选择是否需要 CSRF 保护：
// - 使用 JWT Token (Authorization: Bearer xxx)：跳过 CSRF 检查（Token 不会自动携带）
// - 使用 Cookie/Session 认证：需要 CSRF 检查（Cookie 会自动携带）
// - 无认证信息：需要 CSRF 检查（防止公开接口被 CSRF 攻击）
func SmartCSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于安全方法（GET, HEAD, OPTIONS），不需要验证 CSRF token
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 检查是否使用了 JWT Token 认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// 如果使用了 Bearer Token，跳过 CSRF 检查
			// 因为 JWT Token 不会被浏览器自动携带，不存在 CSRF 风险
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				logger.Debug("Skipping CSRF check for JWT Token request",
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path))
				c.Next()
				return
			}
		}

		// 如果没有使用 JWT Token，执行 CSRF 检查
		// 从请求头获取 CSRF token
		token := c.GetHeader(CSRFTokenHeader)
		if token == "" {
			// 尝试从表单获取
			token = c.PostForm("csrf_token")
		}

		if token == "" {
			logger.Warn("CSRF token missing (non-JWT request)",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()))
			response.Error(c, http.StatusForbidden, "CSRF token missing")
			c.Abort()
			return
		}

		// 验证 token
		if _, exists := globalCSRFStore.get(token); !exists {
			logger.Warn("Invalid or expired CSRF token",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()))
			response.Error(c, http.StatusForbidden, "Invalid or expired CSRF token")
			c.Abort()
			return
		}

		c.Next()
	}
}
