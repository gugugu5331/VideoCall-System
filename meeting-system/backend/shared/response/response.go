package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	response := Response{
		Code:      http.StatusCreated,
		Message:   "created",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusCreated, response)
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	response := Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	response := Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}
	c.JSON(code, response)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	response := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: getRequestID(c),
	}
	c.JSON(code, response)
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Page 分页响应
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	pageData := PageData{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	Success(c, pageData)
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = c.GetHeader("Request-ID")
	}
	return requestID
}

// ValidationError 参数验证错误
func ValidationError(c *gin.Context, err error) {
	BadRequest(c, "Parameter validation failed: "+err.Error())
}

// DatabaseError 数据库错误
func DatabaseError(c *gin.Context, err error) {
	InternalServerError(c, "Database operation failed")
}

// ServiceError 服务错误
func ServiceError(c *gin.Context, message string) {
	InternalServerError(c, message)
}
