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

// Register godoc
// @Summary      Đăng ký tài khoản
// @Description  Đăng ký tài khoản mới bằng email và mật khẩu
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body registerRequest true "Thông tin đăng ký"
// @Success      201  {object}  dto.SuccessResponse[dto.AuthResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Đăng nhập
// @Description  Đăng nhập bằng email và mật khẩu để nhận token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Thông tin đăng nhập"
// @Success      200  {object}  dto.SuccessResponse[dto.AuthResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /auth/login [post]
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

// Me godoc
// @Summary      Lấy thông tin cá nhân
// @Description  Lấy thông tin của người dùng đang đăng nhập
// @Tags         Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.SuccessResponse[dto.UserProfileResponse]
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /auth/me [get]
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

// RefreshToken godoc
// @Summary      Làm mới token
// @Description  Lấy access token mới bằng refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body refreshRequest true "Refresh Token"
// @Success      200  {object}  dto.SuccessResponse[dto.AuthResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Đăng xuất
// @Description  Xóa refresh token để đăng xuất
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body logoutRequest true "Refresh Token"
// @Success      200  {object}  dto.MessageResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)
	h.authService.Logout(req.RefreshToken)
	c.JSON(http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "đăng xuất thành công",
	})
}

// ForgotPassword godoc
// @Summary      Quên mật khẩu
// @Description  Gửi yêu cầu khôi phục mật khẩu. (Luôn trả về thành công để bảo mật)
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email đăng ký"
// @Success      200  {object}  dto.MessageResponse
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Email không hợp lệ")
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Nếu email tồn tại, một liên kết khôi phục đã được gửi đi.",
	})
}

// VerifyCode godoc
// @Summary      Xác thực mã khôi phục
// @Description  Xác thực mã 6 số gửi qua email để lấy token đổi mật khẩu
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyCodeRequest true "Mã xác thực"
// @Success      200  {object}  dto.SuccessResponse[dto.VerifyCodeResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /auth/verify-code [post]
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req dto.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	res, err := h.authService.VerifyCode(req)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, res)
}

// ResetPassword godoc
// @Summary      Đặt lại mật khẩu
// @Description  Đặt lại mật khẩu bằng token được cấp từ bước xác thực
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Thông tin đổi mật khẩu"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
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