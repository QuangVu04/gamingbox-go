package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetValidationMsg chuyển đổi lỗi validator sang tiếng Việt
func GetValidationMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Trường %s là bắt buộc", fe.Field())
	case "email":
		return "Email không hợp lệ"
	case "min":
		return fmt.Sprintf("Trường %s phải có ít nhất %s ký tự", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Trường %s không được vượt quá %s ký tự", fe.Field(), fe.Param())
	}
	return fmt.Sprintf("Trường %s không hợp lệ", fe.Field())
}

// FormatValidationError xử lý lỗi binding và trả về phản hồi chuẩn
func FormatValidationError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	// Lấy lỗi đầu tiên để trả về
	errMsg := GetValidationMsg(errs[0])
	
	c.JSON(http.StatusBadRequest, gin.H{
		"error": errMsg,
		"code":  "VALIDATION_ERROR",
		"field": errs[0].Field(),
	})
}

// IsValidUsername kiểm tra username chỉ chứa chữ cái, số và dấu gạch dưới
func IsValidUsername(username string) bool {
	if len(username) == 0 {
		return false
	}
	for _, r := range username {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}