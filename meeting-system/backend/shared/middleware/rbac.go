package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
	"meeting-system/shared/response"
)

// RequireRole 要求特定角色的中间件
func RequireRole(minRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID（由JWTAuth中间件设置）
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		// 从数据库获取用户信息
		db := database.GetDB()
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Error(c, http.StatusUnauthorized, "User not found")
			} else {
				logger.Error("Failed to get user from database", logger.Err(err))
				response.Error(c, http.StatusInternalServerError, "Database error")
			}
			c.Abort()
			return
		}

		// 检查用户角色
		if user.Role < minRole {
			logger.Warn("Access denied: insufficient permissions",
				logger.Uint("user_id", user.ID),
				logger.String("user_role", user.Role.String()),
				logger.String("required_role", minRole.String()))
			response.Error(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		// 将用户角色存储到上下文
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleAdmin)
}

// RequireModerator 要求版主权限的中间件
func RequireModerator() gin.HandlerFunc {
	return RequireRole(models.UserRoleMod)
}

// RequireSuperAdmin 要求超级管理员权限的中间件
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleSuper)
}

// CheckResourceOwnership 检查资源所有权的中间件工厂
// 用于确保用户只能操作自己的资源，或者管理员可以操作所有资源
func CheckResourceOwnership(getResourceOwnerID func(*gin.Context) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		currentUserID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		// 获取用户角色
		userRole, roleExists := c.Get("user_role")
		
		// 如果是管理员，直接放行
		if roleExists {
			if role, ok := userRole.(models.UserRole); ok && role.IsAdmin() {
				c.Next()
				return
			}
		}

		// 获取资源所有者ID
		resourceOwnerID, err := getResourceOwnerID(c)
		if err != nil {
			logger.Error("Failed to get resource owner ID", logger.Err(err))
			response.Error(c, http.StatusInternalServerError, "Failed to verify ownership")
			c.Abort()
			return
		}

		// 检查是否为资源所有者
		if currentUserID.(uint) != resourceOwnerID {
			response.Error(c, http.StatusForbidden, "Access denied: not the resource owner")
			c.Abort()
			return
		}

		c.Next()
	}
}

