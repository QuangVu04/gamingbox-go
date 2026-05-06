package handlers

import (
	"net/http"
	"strconv"

	_ "vault/be/internal/dto"
	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"
	"vault/be/internal/dto"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Me godoc
// @Summary      Lấy Profile Người dùng
// @Description  Lấy thông tin profile đầy đủ của người dùng (Cần đăng nhập)
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.SuccessResponse[dto.UserProfileResponse]
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	result, err := h.userService.GetUserProfile(userID)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// GetProfile godoc
// @Summary      Xem Profile người dùng khác
// @Description  Lấy thông tin profile đầy đủ của một người dùng dựa theo ID
// @Tags         Users
// @Produce      json
// @Param        id query int true "ID của Người dùng"
// @Success      200  {object}  dto.SuccessResponse[dto.UserProfileResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.ValidationError(c, "Vui lòng cung cấp ID người dùng")
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	profile, err := h.userService.GetUserProfile(uint(userID))
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, profile)
}

// FollowUser godoc
// @Summary      Follow / Unfollow Người dùng khác
// @Description  Thực hiện follow hoặc unfollow một người dùng khác (Cần đăng nhập)
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.FollowRequest true "Thông tin follow"
// @Success      200  {object}  dto.SuccessResponse[dto.FollowResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /users/follow [post]
func (h *UserHandler) FollowUser(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	res, err := h.userService.ToggleFollow(userID, req.UserID)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, res)
}