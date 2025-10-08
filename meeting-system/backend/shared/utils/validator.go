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

// PasswordStrength 密码强度等级
type PasswordStrength int

const (
	PasswordWeak   PasswordStrength = 0 // 弱密码
	PasswordMedium PasswordStrength = 1 // 中等密码
	PasswordStrong PasswordStrength = 2 // 强密码
)

// IsValidPassword 验证密码强度（增强版）
// 要求：至少8位，包含大写字母、小写字母、数字和特殊字符
func IsValidPassword(password string) bool {
	// 密码至少8位
	if len(password) < 8 {
		return false
	}

	// 检查是否包含小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// 检查是否包含大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// 检查是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含特殊字符
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)

	// 至少包含3种类型的字符
	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// GetPasswordStrength 获取密码强度等级
func GetPasswordStrength(password string) PasswordStrength {
	if len(password) < 8 {
		return PasswordWeak
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)

	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	// 根据包含的字符类型数量判断强度
	if count >= 4 && len(password) >= 12 {
		return PasswordStrong
	} else if count >= 3 && len(password) >= 8 {
		return PasswordMedium
	}
	return PasswordWeak
}

// ValidatePasswordStrength 验证密码强度并返回详细信息
func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	if len(password) > 128 {
		return false, "Password must not exceed 128 characters"
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)

	missing := []string{}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasNumber {
		missing = append(missing, "number")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}

	// 至少包含3种类型
	if len(missing) > 1 {
		return false, "Password must contain at least 3 of the following: uppercase letter, lowercase letter, number, special character"
	}

	// 检查常见弱密码
	commonPasswords := []string{
		"password", "12345678", "qwerty", "abc123", "password123",
		"admin123", "letmein", "welcome", "monkey", "dragon",
	}
	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common || strings.Contains(lowerPassword, common) {
			return false, "Password is too common, please choose a stronger password"
		}
	}

	return true, "Password is valid"
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

// SanitizeHTML 清理HTML标签，防止XSS攻击
// 注意：SQL注入防护应该依赖GORM的参数化查询，而不是字符串过滤
func SanitizeHTML(input string) string {
	// 移除HTML标签
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned := htmlRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(cleaned)
}

// SanitizeString 已废弃：请使用 SanitizeHTML
// SQL注入防护应该依赖GORM的参数化查询，而不是字符串过滤
// Deprecated: Use SanitizeHTML instead. SQL injection protection should rely on GORM parameterized queries.
func SanitizeString(input string) string {
	return SanitizeHTML(input)
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
