package middleware

import (
    "strings"

    "vault/be/pkg/utils"

    "github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            utils.Unauthorized(c, "hãy đăng nhập để thực hiện")
            c.Abort()
            return
        }

        claims, err := utils.ParseToken(token)
        if err != nil {
            utils.Unauthorized(c, "token không hợp lệ hoặc đã hết hạn")
            c.Abort()
            return
        }

        if claims.Type != utils.AccessToken {
            utils.Unauthorized(c, "token không đúng loại")
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}

func OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token != "" {
            claims, err := utils.ParseToken(token)
            if err == nil && claims.Type == utils.AccessToken {
                c.Set("userID", claims.UserID)
                c.Set("username", claims.Username)
                c.Set("role", claims.Role)
            }
        }
        c.Next()
    }
}

func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Get("role")
        if role != "admin" {
            utils.Forbidden(c, "chỉ admin mới có quyền thực hiện thao tác này")
            c.Abort()
            return
        }
        c.Next()
    }
}

func GetCurrentUserID(c *gin.Context) (uint, bool) {
    val, exists := c.Get("userID")
    if !exists {
        return 0, false
    }
    id, ok := val.(uint)
    return id, ok
}

func GetOptionalUserID(c *gin.Context) (uint, bool) {
    if id, ok := GetCurrentUserID(c); ok {
        return id, true
    }

    token := extractToken(c)
    if token == "" {
        return 0, false
    }

    claims, err := utils.ParseToken(token)
    if err != nil {
        return 0, false
    }

    if claims.Type != utils.AccessToken {
        return 0, false
    }

    return claims.UserID, true
}

func extractToken(c *gin.Context) string {
    header := c.GetHeader("Authorization")
    if header == "" {
        return ""
    }
    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
        return ""
    }
    return strings.TrimSpace(parts[1])
}