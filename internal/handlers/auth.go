package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"vault/be/internal/dto"
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
	Email      string `json:"email"    binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary      Đăng ký tài khoản
// @Description  Đăng ký tài khoản mới bằng email và mật khẩu
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      registerRequest  true  "Thông tin đăng ký"
// @Success      201      {object}  dto.AuthResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Router       /auth/register [post]
type verifyRegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Code     string `json:"code"     binding:"required,len=6"`
}

// RequestRegisterOTP godoc
func (h *AuthHandler) RequestRegisterOTP(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.authService.RequestRegisterOTP(dto.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Mã xác nhận đã được gửi tới email của bạn",
	})
}

// VerifyRegisterOTP godoc
func (h *AuthHandler) VerifyRegisterOTP(c *gin.Context) {
	var req verifyRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	result, err := h.authService.VerifyRegisterOTP(dto.VerifyRegisterOTPInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Code:     req.Code,
	})

	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, result)
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Đăng nhập bằng email và mật khẩu để lấy access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      loginRequest  true  "Thông tin đăng nhập"
// @Success      200      {object}  dto.AuthResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	result, err := h.authService.Login(dto.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
	})

	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// RefreshToken godoc
// @Summary      Làm mới token
// @Description  Sử dụng refresh token để lấy access token mới
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      refreshRequest  true  "Refresh token"
// @Success      200      {object}  dto.AuthResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Refresh token là bắt buộc")
		return
	}

	result, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// Logout godoc
// @Summary      Đăng xuất
// @Description  Hủy bỏ refresh token hiện tại
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      logoutRequest  true  "Refresh token để hủy"
// @Success      200      {object}  dto.MessageResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)

	h.authService.Logout(req.RefreshToken)
	utils.Success(c, http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Đăng xuất thành công",
	})
}

// Me godoc
// @Summary      Lấy thông tin bản thân
// @Description  Lấy thông tin chi tiết của người dùng đang đăng nhập
// @Tags         Authentication
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  dto.UserResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	result, err := h.authService.GetMe(userID)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Nếu email tồn tại, mã xác nhận sẽ được gửi tới bạn",
	})
}

func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req dto.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	result, err := h.authService.VerifyCode(req)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.authService.ResetPassword(req)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Đổi mật khẩu thành công",
	})
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := utils.GoogleOauthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	res, err := h.authService.HandleGoogleLogin(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handleSocialCallbackSuccess(c, res)
}

func (h *AuthHandler) FacebookLogin(c *gin.Context) {
	url := utils.FacebookOauthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) FacebookCallback(c *gin.Context) {
	code := c.Query("code")
	res, err := h.authService.HandleFacebookLogin(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handleSocialCallbackSuccess(c, res)
}

func handleSocialCallbackSuccess(c *gin.Context, res *dto.AuthResponse) {
	jsonRes, _ := json.Marshal(res)
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Authentication Successful</title></head>
		<body>
			<script>
				window.opener.postMessage(%s, "*");
				window.close();
			</script>
			<p>Đăng nhập thành công! Vui lòng chờ trong giây lát...</p>
		</body>
		</html>
	`, string(jsonRes))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}