package handlers

import (
	"context"
	"net/http"

	"vault/be/internal/dto"
	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ListHandler struct {
	listService services.ListService
}

func NewListHandler(listService services.ListService) *ListHandler {
	return &ListHandler{
		listService: listService,
	}
}

// TrendingLists godoc
// @Summary      Lấy danh sách List thịnh hành
// @Description  Lấy danh sách list phổ biến có phân trang
// @Tags         Lists
// @Produce      json
// @Param        page query int false "Trang hiện tại (Mặc định 1)"
// @Param        limit query int false "Số lượng mỗi trang (Mặc định 10)"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.ListTrendingResponse]
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /lists/trending [get]
func (h *ListHandler) TrendingLists(c *gin.Context) {
	// Parse pagination parameters using request utility
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	// Call service with caching
	ctx := context.Background()
	lists, pagination, err := h.listService.GetTrendingLists(ctx, page, limit)
	if err != nil {
		if serviceErr, ok := err.(*dto.ServiceError); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"code":   serviceErr.Code,
				"error":  serviceErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"code":   "SERVER_ERROR",
			"error":  "đã xảy ra lỗi",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"pagination": pagination,
		"data":       lists,
	})
}

// CreateList godoc
// @Summary      Tạo List mới
// @Description  Tạo danh sách game cá nhân (Cần đăng nhập)
// @Tags         Lists
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateListRequest true "Thông tin list"
// @Success      201  {object}  dto.SuccessResponse[dto.ListDetailResponse]
// @Router       /lists [post]
func (h *ListHandler) CreateList(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	var req dto.CreateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	ctx := context.Background()
	list, err := h.listService.CreateList(ctx, userID, req)
	if err != nil {
		handleListError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, list)
}

// GetListDetail godoc
// @Summary      Xem chi tiết List
// @Description  Lấy thông tin và danh sách game trong một list
// @Tags         Lists
// @Produce      json
// @Param        id path int true "ID của List"
// @Success      200  {object}  dto.SuccessResponse[dto.ListDetailResponse]
// @Router       /lists/{id} [get]
func (h *ListHandler) GetListDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	ctx := context.Background()
	list, err := h.listService.GetListDetail(ctx, uint(id))
	if err != nil {
		handleListError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, list)
}

// UpdateList godoc
// @Summary      Cập nhật List
// @Description  Chỉnh sửa thông tin hoặc danh sách game trong list (Chỉ chủ sở hữu)
// @Tags         Lists
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID của List"
// @Param        request body dto.UpdateListRequest true "Thông tin cập nhật"
// @Success      200  {object}  dto.SuccessResponse[dto.ListDetailResponse]
// @Router       /lists/{id} [put]
func (h *ListHandler) UpdateList(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	var req dto.UpdateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	ctx := context.Background()
	list, err := h.listService.UpdateList(ctx, userID, uint(id), req)
	if err != nil {
		handleListError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, list)
}

// DeleteList godoc
// @Summary      Xóa List
// @Description  Xóa một danh sách game (Chỉ chủ sở hữu)
// @Tags         Lists
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của List"
// @Success      200  {object}  dto.SuccessResponse[string]
// @Router       /lists/{id} [delete]
func (h *ListHandler) DeleteList(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	ctx := context.Background()
	if err := h.listService.DeleteList(ctx, userID, uint(id)); err != nil {
		handleListError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, "Đã xóa danh sách")
}

// GetGameLists godoc
// @Summary      Lấy danh sách List chứa Game
// @Description  Lấy các danh sách game công khai có chứa game này (Phân trang 25)
// @Tags         Lists
// @Produce      json
// @Param        id path int true "ID của Game"
// @Param        page query int false "Trang hiện tại (mặc định 1)"
// @Param        limit query int false "Số lượng mỗi trang (mặc định 25)"
// @Param        sort query string false "Sắp xếp: list_name, popularity, recently_updated, newest, oldest (mặc định newest)"
// @Success      200  {object}  dto.SuccessResponse[dto.GameListsResponse]
// @Router       /games/{id}/lists [get]
func (h *ListHandler) GetGameLists(c *gin.Context) {
	idStr := c.Param("id")
	gameID, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 25, 1, 100)
	sort := c.DefaultQuery("sort", "newest")

	ctx := context.Background()
	result, err := h.listService.GetGameLists(ctx, uint(gameID), page, limit, sort)
	if err != nil {
		handleListError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}
