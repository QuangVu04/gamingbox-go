package handlers

import (
	"context"
	"net/http"
	"strconv"

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

// LikeGame handles like/unlike for games
// POST /api/v1/games/:game_id/like
func (h *LikeHandler) LikeGame(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	gameIDStr := c.Param("game_id")
	gameID, err := strconv.ParseUint(gameIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "game_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(gameID), "game")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// LikeReview handles like/unlike for reviews
// POST /api/v1/reviews/:review_id/like
func (h *LikeHandler) LikeReview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "review_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(reviewID), "review")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}

// LikeList handles like/unlike for lists
// POST /api/v1/lists/:list_id/like
func (h *LikeHandler) LikeList(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	listIDStr := c.Param("list_id")
	listID, err := strconv.ParseUint(listIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(c, "list_id không hợp lệ")
		return
	}

	ctx := context.Background()
	result, err := h.likeService.ToggleLike(ctx, userID, uint(listID), "list")
	if err != nil {
		handleLikeError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}