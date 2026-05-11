package handlers

import (
	"context"
	"net/http"
	"strconv"

	_ "vault/be/internal/dto"
	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	likeService services.LikeService
}

func NewLikeHandler(likeService services.LikeService) *LikeHandler {
	return &LikeHandler{
		likeService: likeService,
	}
}

// LikeGame godoc
// @Summary      Thích hoặc Bỏ thích Game
// @Description  Thích hoặc Bỏ thích một game (Cần đăng nhập)
// @Tags         Likes
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của Game"
// @Success      200  {object}  dto.SuccessResponse[dto.LikeGameResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /games/{id}/like [post]
func (h *LikeHandler) LikeGame(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "game_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(id), "game")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// LikeReview godoc
// @Summary      Thích hoặc Bỏ thích Review
// @Description  Thích hoặc Bỏ thích một review (Cần đăng nhập)
// @Tags         Likes
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của Review"
// @Success      200  {object}  dto.SuccessResponse[dto.LikeGameResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /reviews/{id}/like [post]
func (h *LikeHandler) LikeReview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "review_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(id), "review")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// LikeList godoc
// @Summary      Thích hoặc Bỏ thích List
// @Description  Thích hoặc Bỏ thích một list (Cần đăng nhập)
// @Tags         Likes
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID của List"
// @Success      200  {object}  dto.SuccessResponse[dto.LikeGameResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /lists/{id}/like [post]
func (h *LikeHandler) LikeList(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "list_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(id), "list")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}