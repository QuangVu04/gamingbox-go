package handlers

import (
	"net/http"

	"vault/be/internal/dto"
	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService 
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	result, err := h.authService.Register(dto.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	result, err := h.authService.Login(dto.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
	})
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	user, err := h.authService.GetMe(userID)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, user)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "thiếu refresh_token")
		return
	}

	result, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)
	h.authService.Logout(req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "đăng xuất thành công"})
}