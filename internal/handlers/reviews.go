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

type ReviewHandler struct {
	reviewService services.ReviewService
}

func NewReviewHandler(reviewService services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// TrendingReviews godoc
// @Summary      Lấy danh sách Review thịnh hành
// @Description  Lấy danh sách review phổ biến có phân trang
// @Tags         Reviews
// @Produce      json
// @Param        page query int false "Trang hiện tại (Mặc định 1)"
// @Param        limit query int false "Số lượng mỗi trang (Mặc định 10)"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.ReviewTrendingResponse]
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /reviews/trending [get]
func (h *ReviewHandler) TrendingReviews(c *gin.Context) {
	// Parse pagination parameters using request utility
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	// Call service with caching
	ctx := context.Background()
	reviews, pagination, err := h.reviewService.GetTrendingReviews(ctx, page, limit)
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
		"data":       reviews,
	})
}

// CreateReview godoc
// @Summary      Tạo Review mới
// @Description  Tạo một review cho game (Cần đăng nhập)
// @Tags         Reviews
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateReviewRequest true "Thông tin review"
// @Success      201  {object}  dto.SuccessResponse[dto.ReviewTrendingResponse]
// @Router       /reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	ctx := context.Background()
	review, err := h.reviewService.CreateReview(ctx, userID, req)
	if err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, review)
}

// UpdateReview godoc
// @Summary      Cập nhật Review
// @Description  Chỉnh sửa review đã viết (Chỉ chủ sở hữu mới có quyền)
// @Tags         Reviews
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID của Review"
// @Param        request body dto.UpdateReviewRequest true "Thông tin cập nhật"
// @Success      200  {object}  dto.SuccessResponse[dto.ReviewTrendingResponse]
// @Router       /reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
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

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	ctx := context.Background()
	review, err := h.reviewService.UpdateReview(ctx, userID, uint(id), req)
	if err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, review)
}

// DeleteReview godoc
// @Summary      Xóa Review
// @Description  Xóa một review (Chỉ chủ sở hữu)
// @Tags         Reviews
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của Review"
// @Success      200  {object}  dto.SuccessResponse[string]
// @Router       /reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
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
	if err := h.reviewService.DeleteReview(ctx, userID, uint(id)); err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, "Đã xóa review")
}

// GetComments godoc
// @Summary      Lấy danh sách bình luận
// @Description  Lấy tất cả bình luận của một review
// @Tags         Reviews
// @Produce      json
// @Param        id path int true "ID của Review"
// @Success      200  {object}  dto.SuccessResponse[[]dto.CommentResponse]
// @Router       /reviews/{id}/comments [get]
func (h *ReviewHandler) GetComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	ctx := context.Background()
	comments, err := h.reviewService.GetComments(ctx, uint(id))
	if err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, comments)
}

// AddComment godoc
// @Summary      Thêm bình luận
// @Description  Bình luận vào một review (Cần đăng nhập)
// @Tags         Reviews
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID của Review"
// @Param        request body dto.CommentRequest true "Nội dung bình luận"
// @Success      201  {object}  dto.SuccessResponse[dto.CommentResponse]
// @Router       /reviews/{id}/comments [post]
func (h *ReviewHandler) AddComment(c *gin.Context) {
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

	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Dữ liệu không hợp lệ")
		return
	}

	ctx := context.Background()
	comment, err := h.reviewService.AddComment(ctx, userID, uint(id), req)
	if err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, comment)
}

// GetReviewDetail godoc
// @Summary      Xem chi tiết Review
// @Description  Lấy thông tin chi tiết một bài đánh giá
// @Tags         Reviews
// @Produce      json
// @Param        id path int true "ID của Review"
// @Success      200  {object}  dto.SuccessResponse[dto.ReviewTrendingResponse]
// @Router       /reviews/{id} [get]
func (h *ReviewHandler) GetReviewDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	ctx := context.Background()
	review, err := h.reviewService.GetReviewByID(ctx, uint(id))
	if err != nil {
		handleReviewError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, review)
}
