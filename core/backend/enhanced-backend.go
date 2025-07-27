package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 用户结构
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Status   string `json:"status"`
}

// 通话结构
type Call struct {
	ID         int        `json:"id"`
	CallerID   int        `json:"caller_id"`
	ReceiverID int        `json:"receiver_id"`
	Status     string     `json:"status"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Duration   int        `json:"duration"`
}

// 模拟数据存储
var (
	users = map[string]User{
		"webuser": {
			ID:       1,
			Username: "webuser",
			Email:    "webuser@example.com",
			FullName: "Test User",
			Status:   "online",
		},
		"admin": {
			ID:       2,
			Username: "admin",
			Email:    "admin@example.com",
			FullName: "Administrator",
			Status:   "online",
		},
	}

	calls = []Call{
		{
			ID:         1,
			CallerID:   1,
			ReceiverID: 2,
			Status:     "completed",
			StartTime:  time.Now().Add(-time.Hour),
			EndTime:    &[]time.Time{time.Now().Add(-time.Hour + time.Minute*30)}[0],
			Duration:   1800,
		},
	}

	callCounter = 3
)

// JWT密钥
var jwtSecret = []byte("your-secret-key")

// 生成JWT token
func generateToken(userID int, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString(jwtSecret)
}

// 验证JWT token
func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// 认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 提取Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Next()
	}
}

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.New()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "VideoCall Backend is running",
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				var req struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
					return
				}

				// 用户验证
				user, exists := users[req.Username]
				if !exists || req.Password != "test123" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
					return
				}

				// 生成JWT token
				token, err := generateToken(user.ID, user.Username)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "Login successful",
					"token":   token,
					"user":    user,
				})
			})

			auth.POST("/register", func(c *gin.Context) {
				var req struct {
					Username string `json:"username"`
					Email    string `json:"email"`
					Password string `json:"password"`
					FullName string `json:"full_name"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
					return
				}

				// 检查用户名是否已存在
				if _, exists := users[req.Username]; exists {
					c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
					return
				}

				// 创建新用户
				newUser := User{
					ID:       len(users) + 1,
					Username: req.Username,
					Email:    req.Email,
					FullName: req.FullName,
					Status:   "online",
				}
				users[req.Username] = newUser

				c.JSON(http.StatusCreated, gin.H{
					"message": "User registered successfully",
					"user":    newUser,
				})
			})

			auth.GET("/validate", authMiddleware(), func(c *gin.Context) {
				username := c.GetString("username")

				user, exists := users[username]
				if !exists {
					c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"user": user,
				})
			})
		}

		// 用户管理路由
		usersGroup := api.Group("/users")
		usersGroup.Use(authMiddleware())
		{
			usersGroup.GET("/profile", func(c *gin.Context) {
				username := c.GetString("username")
				user, exists := users[username]
				if !exists {
					c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"user": user,
				})
			})

			usersGroup.PUT("/profile", func(c *gin.Context) {
				username := c.GetString("username")
				user, exists := users[username]
				if !exists {
					c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
					return
				}

				var req struct {
					FullName string `json:"full_name"`
					Email    string `json:"email"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
					return
				}

				// 更新用户信息
				user.FullName = req.FullName
				user.Email = req.Email
				users[username] = user

				c.JSON(http.StatusOK, gin.H{
					"message": "Profile updated successfully",
					"user":    user,
				})
			})

			usersGroup.GET("/list", func(c *gin.Context) {
				userList := make([]User, 0, len(users))
				for _, user := range users {
					userList = append(userList, user)
				}

				c.JSON(http.StatusOK, gin.H{
					"users": userList,
				})
			})
		}

		// 通话管理路由
		callsGroup := api.Group("/calls")
		callsGroup.Use(authMiddleware())
		{
			callsGroup.POST("/start", func(c *gin.Context) {
				callerID := c.GetInt("user_id")

				var req struct {
					ReceiverUsername string `json:"receiver_username"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
					return
				}

				// 查找接收者
				receiver, exists := users[req.ReceiverUsername]
				if !exists {
					c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
					return
				}

				// 创建新通话
				newCall := Call{
					ID:         callCounter,
					CallerID:   callerID,
					ReceiverID: receiver.ID,
					Status:     "ringing",
					StartTime:  time.Now(),
				}
				calls = append(calls, newCall)
				callCounter++

				c.JSON(http.StatusOK, gin.H{
					"message": "Call started",
					"call":    newCall,
				})
			})

			callsGroup.POST("/:id/answer", func(c *gin.Context) {
				callID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid call ID"})
					return
				}

				// 查找通话
				for i, call := range calls {
					if call.ID == callID {
						calls[i].Status = "active"
						c.JSON(http.StatusOK, gin.H{
							"message": "Call answered",
							"call":    calls[i],
						})
						return
					}
				}

				c.JSON(http.StatusNotFound, gin.H{"error": "Call not found"})
			})

			callsGroup.POST("/:id/end", func(c *gin.Context) {
				callID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid call ID"})
					return
				}

				// 查找并结束通话
				for i, call := range calls {
					if call.ID == callID {
						endTime := time.Now()
						duration := int(endTime.Sub(call.StartTime).Seconds())
						calls[i].Status = "completed"
						calls[i].EndTime = &endTime
						calls[i].Duration = duration

						c.JSON(http.StatusOK, gin.H{
							"message": "Call ended",
							"call":    calls[i],
						})
						return
					}
				}

				c.JSON(http.StatusNotFound, gin.H{"error": "Call not found"})
			})

			callsGroup.GET("/history", func(c *gin.Context) {
				userID := c.GetInt("user_id")

				userCalls := make([]Call, 0)
				for _, call := range calls {
					if call.CallerID == userID || call.ReceiverID == userID {
						userCalls = append(userCalls, call)
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"calls": userCalls,
				})
			})
		}

		// 系统状态路由
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "running",
				"users": gin.H{
					"total":  len(users),
					"online": len(users), // 简化：假设所有用户都在线
				},
				"calls": gin.H{
					"total":     len(calls),
					"active":    0, // 简化：假设没有活跃通话
					"completed": len(calls),
				},
				"uptime": time.Now().Format(time.RFC3339),
			})
		})
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Enhanced backend server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  - GET  /health")
	log.Printf("  - POST /api/v1/auth/login")
	log.Printf("  - POST /api/v1/auth/register")
	log.Printf("  - GET  /api/v1/auth/validate")
	log.Printf("  - GET  /api/v1/users/profile")
	log.Printf("  - PUT  /api/v1/users/profile")
	log.Printf("  - GET  /api/v1/users/list")
	log.Printf("  - POST /api/v1/calls/start")
	log.Printf("  - POST /api/v1/calls/:id/answer")
	log.Printf("  - POST /api/v1/calls/:id/end")
	log.Printf("  - GET  /api/v1/calls/history")
	log.Printf("  - GET  /api/v1/status")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
