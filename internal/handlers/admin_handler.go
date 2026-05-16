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
