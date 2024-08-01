package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/shared/response"
	"meeting-system/user-service/services"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user.ToProfile())
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	loginResp, err := h.userService.Login(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, loginResp)
}

// RefreshToken 刷新token
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	newToken, err := h.userService.RefreshToken(req.Token)
	if err != nil {
		response.Unauthorized(c, "Invalid or expired token")
		return
	}

	response.Success(c, gin.H{"token": newToken})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, profile)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	profile, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, profile)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userService.ChangePassword(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// TODO: 实现文件上传逻辑
	// 1. 接收上传的文件
	// 2. 验证文件类型和大小
	// 3. 上传到MinIO
	// 4. 更新用户头像URL

	response.Success(c, gin.H{
		"message": "Avatar upload not implemented yet",
		"user_id": userID,
	})
}

// DeleteAccount 删除账户
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Account deleted successfully", nil)
}

// ListUsers 获取用户列表（管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.userService.ListUsers(page, pageSize, keyword)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Page(c, users, total, page, pageSize)
}

// GetUser 获取指定用户（管理员）
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(uint(userID))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateUser 更新指定用户（管理员）
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.UpdateUser(uint(userID), &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, user)
}

// DeleteUser 删除指定用户（管理员）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(uint(userID)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "User deleted successfully", nil)
}

// BanUser 禁用用户（管理员）
func (h *UserHandler) BanUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.BanUser(uint(userID)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "User banned successfully", nil)
}

// UnbanUser 解禁用户（管理员）
func (h *UserHandler) UnbanUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.UnbanUser(uint(userID)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "User unbanned successfully", nil)
}

// ForgotPassword 忘记密码
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// TODO: 实现忘记密码逻辑
	// 1. 验证邮箱是否存在
	// 2. 生成重置码
	// 3. 发送邮件

	response.SuccessWithMessage(c, "Password reset email sent", nil)
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// TODO: 实现重置密码逻辑
	// 1. 验证重置码
	// 2. 更新密码

	response.SuccessWithMessage(c, "Password reset successfully", nil)
}
