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

// GetFollowing godoc
// @Summary      Lấy danh sách đang theo dõi
// @Description  Lấy danh sách những người dùng mà current user đang follow
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Trang hiện tại" default(1)
// @Param        limit query int false "Số lượng hiển thị trên mỗi trang" default(20)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.UserSummary]
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/me/following [get]
func (h *UserHandler) GetFollowing(c *gin.Context) {
	var userID uint
	idStr := c.Query("id")
	if idStr == "" {
		uid, ok := middleware.GetCurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "Vui lòng đăng nhập")
			return
		}
		userID = uid
	} else {
		id, _ := strconv.ParseUint(idStr, 10, 32)
		userID = uint(id)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.userService.GetFollowing(userID, page, limit)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetStats godoc
// @Summary      Lấy thống kê người dùng
// @Description  Xem thống kê về game đã chơi, đánh giá trung bình...
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng (Nếu không có sẽ lấy me)"
// @Success      200  {object}  dto.SuccessResponse[dto.UserStatsResponse]
// @Router       /users/stats [get]
func (h *UserHandler) GetStats(c *gin.Context) {
	var userID uint
	idStr := c.Query("id")
	if idStr == "" {
		uid, ok := middleware.GetCurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "Vui lòng đăng nhập")
			return
		}
		userID = uid
	} else {
		id, _ := strconv.ParseUint(idStr, 10, 32)
		userID = uint(id)
	}

	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, stats)
}

// GetDiary godoc
// @Summary      Xem Nhật ký chơi game
// @Description  Lấy danh sách nhật ký chơi game của người dùng
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(20)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.DiaryEntry]
// @Router       /users/diary [get]
func (h *UserHandler) GetDiary(c *gin.Context) {
	var userID uint
	idStr := c.Query("id")
	if idStr == "" {
		uid, ok := middleware.GetCurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "Vui lòng đăng nhập")
			return
		}
		userID = uid
	} else {
		id, _ := strconv.ParseUint(idStr, 10, 32)
		userID = uint(id)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.userService.GetDiary(userID, page, limit)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetWatchlist godoc
// @Summary      Xem Watchlist (Backlog)
// @Description  Lấy danh sách game muốn chơi của người dùng
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(20)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.GameSummary]
// @Router       /users/watchlist [get]
func (h *UserHandler) GetWatchlist(c *gin.Context) {
	var userID uint
	idStr := c.Query("id")
	if idStr == "" {
		uid, ok := middleware.GetCurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "Vui lòng đăng nhập")
			return
		}
		userID = uid
	} else {
		id, _ := strconv.ParseUint(idStr, 10, 32)
		userID = uint(id)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.userService.GetWatchlist(userID, page, limit)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}


// GetFollowers godoc
// @Summary      Lấy danh sách người theo dõi
// @Description  Lấy danh sách những người dùng đang follow current user
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Trang hiện tại" default(1)
// @Param        limit query int false "Số lượng hiển thị trên mỗi trang" default(20)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.UserSummary]
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/me/followers [get]
func (h *UserHandler) GetFollowers(c *gin.Context) {
	var userID uint
	idStr := c.Query("id")
	if idStr == "" {
		uid, ok := middleware.GetCurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "Vui lòng đăng nhập")
			return
		}
		userID = uid
	} else {
		id, _ := strconv.ParseUint(idStr, 10, 32)
		userID = uint(id)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.userService.GetFollowers(userID, page, limit)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}