package utils

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, data any) {
    c.JSON(status, gin.H{"data": data})
}

func Error(c *gin.Context, status int, code, message string) {
    c.JSON(status, gin.H{
        "error": message,
        "code":  code,
    })
}

func ErrorWithField(c *gin.Context, status int, code, message, field string) {
    c.JSON(status, gin.H{
        "error": message,
        "code":  code,
        "field": field,
    })
}

func ValidationError(c *gin.Context, message string) {
    Error(c, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func Unauthorized(c *gin.Context, message string) {
    Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
    Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
    Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func Conflict(c *gin.Context, code, message, field string) {
    ErrorWithField(c, http.StatusConflict, code, message, field)
}

func InternalError(c *gin.Context) {
    Error(c, http.StatusInternalServerError, "SERVER_ERROR", "đã xảy ra lỗi")
}