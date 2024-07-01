package utils

import (
	"regexp"
	"strings"
)

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUsername 验证用户名格式
func IsValidUsername(username string) bool {
	// 用户名只能包含字母、数字、下划线，长度3-20
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// IsValidPassword 验证密码强度
func IsValidPassword(password string) bool {
	// 密码至少8位，包含字母和数字
	if len(password) < 8 {
		return false
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasLetter && hasNumber
}

// IsValidMeetingTitle 验证会议标题
func IsValidMeetingTitle(title string) bool {
	title = strings.TrimSpace(title)
	return len(title) >= 1 && len(title) <= 100
}

// IsValidMeetingDescription 验证会议描述
func IsValidMeetingDescription(description string) bool {
	return len(description) <= 1000
}

// IsValidPhoneNumber 验证手机号码
func IsValidPhoneNumber(phone string) bool {
	// 简单的手机号验证，支持中国手机号
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// SanitizeString 清理字符串，移除危险字符
func SanitizeString(input string) string {
	// 移除HTML标签
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned := htmlRegex.ReplaceAllString(input, "")

	// 移除SQL注入相关字符
	sqlRegex := regexp.MustCompile(`[';\"\\]`)
	cleaned = sqlRegex.ReplaceAllString(cleaned, "")

	return strings.TrimSpace(cleaned)
}

// ValidateFileExtension 验证文件扩展名
func ValidateFileExtension(filename string, allowedExts []string) bool {
	if len(allowedExts) == 0 {
		return true
	}

	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return false
	}

	ext := strings.ToLower(parts[len(parts)-1])
	for _, allowedExt := range allowedExts {
		if ext == strings.ToLower(allowedExt) {
			return true
		}
	}

	return false
}

// ValidateFileSize 验证文件大小（字节）
func ValidateFileSize(size int64, maxSize int64) bool {
	return size > 0 && size <= maxSize
}
