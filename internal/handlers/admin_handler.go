package handlers

import (
	"net/http"
	"strconv"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService services.UserService
	adminService services.AdminService
}

func NewAdminHandler(userService services.UserService, adminService services.AdminService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		adminService: adminService,
	}
}

func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	profile, err := h.userService.GetUserProfile(uint(userID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, profile)
}

func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	err = h.userService.UpdateUserStatus(uint(userID), req.Status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Cập nhật trạng thái thành công"})
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	err = h.userService.UpdateUserRole(uint(userID), req.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Cập nhật quyền hạn thành công"})
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "7 Ngày")
	stats, err := h.adminService.GetDashboardStats(timeframe)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, stats)
}

func (h *AdminHandler) GetActivityChart(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "7 Ngày")
	chartData, err := h.adminService.GetActivityChart(timeframe)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, chartData)
}

func (h *AdminHandler) GetGames(c *gin.Context) {
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	search := c.Query("search")
	category := c.Query("category")
	platform := c.Query("platform")
	minRating := c.Query("min_rating")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	sort := c.Query("sort")

	games, total, err := h.adminService.GetAdminGames(page, limit, search, category, platform, minRating, startDate, endDate, sort)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"pagination": gin.H{
			"total_records": total,
			"current_page":  page,
			"limit":         limit,
		},
		"data": games,
	})
}

func (h *AdminHandler) GetAdminActivities(c *gin.Context) {
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 20, 1, 100)
	filterType := c.Query("filter")
	searchQuery := c.Query("search")

	res, err := h.adminService.GetUserActivities(0, page, limit, filterType, searchQuery)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	search := c.Query("search")
	role := c.Query("role")
	sort := c.Query("sort")

	users, total, err := h.adminService.GetAdminUsers(page, limit, search, role, sort)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"pagination": gin.H{
			"total_records": total,
			"current_page":  page,
			"limit":         limit,
		},
		"data": users,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	if err := h.adminService.DeleteUser(uint(userID)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Xóa người dùng thành công"})
}

func (h *AdminHandler) DeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	reviewID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID đánh giá không hợp lệ")
		return
	}

	if err := h.adminService.DeleteReview(uint(reviewID)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Xóa đánh giá thành công"})
}

func (h *AdminHandler) DeleteList(c *gin.Context) {
	idStr := c.Param("id")
	listID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID danh sách không hợp lệ")
		return
	}

	if err := h.adminService.DeleteList(uint(listID)); err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Xóa danh sách thành công"})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required"`
		Status   string `json:"status" binding:"required"`
		Bio      string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	err = h.adminService.UpdateUser(uint(userID), req.Username, req.Email, req.Role, req.Status, req.Bio)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Cập nhật người dùng thành công"})
}

func (h *AdminHandler) GetUserOverview(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}

	overview, err := h.adminService.GetUserDetailOverview(uint(userID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	utils.Success(c, http.StatusOK, overview)
}

func (h *AdminHandler) GetUserActivitiesPaginated(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)
	filterType := c.Query("filter")
	searchQuery := c.Query("search")

	res, err := h.adminService.GetUserActivities(uint(userID), page, limit, filterType, searchQuery)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUserReviewsPaginated(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)
	filter := c.DefaultQuery("filter", "recent")

	res, err := h.adminService.GetUserReviews(uint(userID), page, limit, filter)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUserListsPaginated(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	res, err := h.adminService.GetUserLists(uint(userID), page, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUserBacklogPaginated(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	res, err := h.adminService.GetUserBacklog(uint(userID), page, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) GetUserGamesPaginated(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "ID người dùng không hợp lệ")
		return
	}
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	res, err := h.adminService.GetUserLibraryGames(uint(userID), page, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}
