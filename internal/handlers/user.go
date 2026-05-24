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
// @Param        status query string false "Trạng thái (playing, played)"
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

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.userService.GetDiary(userID, status, page, limit)
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

// UpdateFavoriteGames godoc
// @Summary      Cập nhật danh sách game yêu thích
// @Description  Cập nhật danh sách game yêu thích (Tối đa 4 item)
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body updateFavRequest true "Danh sách ID game yêu thích"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /users/favorite-games [put]
func (h *UserHandler) UpdateFavoriteGames(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	type updateFavRequest struct {
		GameIDs []uint `json:"game_ids"`
	}

	var req updateFavRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	if len(req.GameIDs) > 4 {
		utils.ValidationError(c, "Tối đa chỉ được chọn 4 game yêu thích")
		return
	}

	err := h.userService.UpdateFavoriteGames(userID, req.GameIDs)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Cập nhật danh sách game yêu thích thành công",
	})
}

// GetReviews godoc
// @Summary      Lấy danh sách review của người dùng
// @Description  Lấy danh sách review của người dùng (có phân trang)
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng (Nếu không có sẽ lấy me)"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(10)
// @Param        filter query string false "Bộ lọc (newest, popular...)"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.ReviewSummary]
// @Router       /users/reviews [get]
func (h *UserHandler) GetReviews(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.Query("filter")

	res, err := h.userService.GetUserReviews(userID, page, limit, filter)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetLists godoc
// @Summary      Lấy danh sách List game của người dùng
// @Description  Lấy danh sách List game của người dùng (có phân trang)
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng (Nếu không có sẽ lấy me)"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(10)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.ListSummary]
// @Router       /users/lists [get]
func (h *UserHandler) GetLists(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	res, err := h.userService.GetUserLists(userID, page, limit)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetActivities godoc
// @Summary      Lấy danh sách hoạt động của người dùng
// @Description  Lấy danh sách hoạt động gần đây của người dùng (có phân trang)
// @Tags         Users
// @Produce      json
// @Param        id query int false "ID người dùng (Nếu không có sẽ lấy me)"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(10)
// @Param        filterType query string false "Loại hoạt động (review, rating, game...)"
// @Param        search query string false "Từ khóa tìm kiếm"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.ActivitySummary]
// @Router       /users/activities [get]
func (h *UserHandler) GetActivities(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filterType := c.Query("filterType")
	search := c.Query("search")

	res, err := h.userService.GetUserActivities(userID, page, limit, filterType, search)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// UpdateProfile godoc
// @Summary      Cập nhật Profile Người dùng
// @Description  Cập nhật các thông tin cá nhân như Bio, Location, Avatar
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateProfileRequest true "Thông tin cập nhật"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Cập nhật hồ sơ thành công",
	})
}

// RequestEmailChangeOTP godoc
// @Summary      Yêu cầu đổi email
// @Description  Gửi mã OTP đến email mới
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        request  body      dto.RequestEmailChangeRequest  true  "Email mới"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /users/me/email/request-otp [post]
func (h *UserHandler) RequestEmailChangeOTP(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.RequestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.userService.RequestEmailChangeOTP(userID, &req)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Mã xác nhận đã được gửi tới email mới",
	})
}

// VerifyEmailChangeOTP godoc
// @Summary      Xác nhận đổi email
// @Description  Xác nhận mã OTP để đổi email
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        request  body      dto.VerifyEmailChangeRequest  true  "Mã xác nhận"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /users/me/email/verify [put]
func (h *UserHandler) VerifyEmailChangeOTP(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.VerifyEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FormatValidationError(c, err)
		return
	}

	err := h.userService.VerifyEmailChangeOTP(userID, &req)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, dto.MessageResponse{
		Status:  "success",
		Message: "Đổi email thành công",
	})
}

// SearchUsers godoc
// @Summary      Tìm kiếm người dùng
// @Description  Tìm kiếm người dùng công khai có phân trang
// @Tags         Users
// @Produce      json
// @Param        search query string false "Từ khóa tìm kiếm"
// @Param        page query int false "Trang" default(1)
// @Param        limit query int false "Giới hạn" default(10)
// @Param        sort query string false "Sắp xếp: active, followers, newest"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.UserResponse]
// @Router       /users [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sort := c.DefaultQuery("sort", "active")

	res, err := h.userService.SearchUsers(search, page, limit, sort)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
