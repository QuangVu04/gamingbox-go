package handlers

import (
	"net/http"
	"strconv"
	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service services.NotificationService
}

func NewNotificationHandler(service services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// GetNotifications godoc
// @Summary      Lấy danh sách thông báo
// @Description  Lấy danh sách thông báo của người dùng hiện tại (Cần đăng nhập)
// @Tags         Notifications
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Trang hiện tại" default(1)
// @Param        limit query int false "Số lượng hiển thị trên mỗi trang" default(20)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.NotificationResponse]
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.service.GetNotifications(userID, page, limit)
	if err != nil {
		utils.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, res)
}

// MarkAsRead godoc
// @Summary      Đánh dấu đã đọc một thông báo
// @Description  Đánh dấu một thông báo cụ thể là đã đọc
// @Tags         Notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của thông báo"
// @Success      200  {object}  dto.MessageResponse
// @Router       /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	if err := h.service.MarkAsRead(id, userID); err != nil {
		utils.InternalError(c)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Đã đánh dấu đã đọc"})
}

// MarkAllAsRead godoc
// @Summary      Đánh dấu đã đọc tất cả thông báo
// @Description  Đánh dấu tất cả thông báo của người dùng hiện tại là đã đọc
// @Tags         Notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dto.MessageResponse
// @Router       /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	if err := h.service.MarkAllAsRead(userID); err != nil {
		utils.InternalError(c)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "Đã đánh dấu tất cả đã đọc"})
}
