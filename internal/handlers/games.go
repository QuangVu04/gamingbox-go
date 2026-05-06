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

type GameHandler struct {
	gameService            services.GameService
}

func NewGameHandler(gameService services.GameService) *GameHandler {
	return &GameHandler{
		gameService:            gameService,
	}
}

// TrendingGames godoc
// @Summary      Lấy danh sách game thịnh hành
// @Description  Lấy danh sách game phổ biến có phân trang
// @Tags         Games
// @Produce      json
// @Param        page query int false "Trang hiện tại (Mặc định 1)"
// @Param        limit query int false "Số lượng mỗi trang (Mặc định 12)"
// @Success      200  {object}  dto.PaginatedResponse[[]dto.GameTrendingResponse]
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /games/trending [get]
func (h *GameHandler) TrendingGames(c *gin.Context) {
	// Parse pagination parameters using request utility
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 12, 1, 100)

	// Call service with caching
	ctx := context.Background()
	games, pagination, err := h.gameService.GetTrendingGames(ctx, page, limit)
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
		"data":       games,
	})
}

// RateGame godoc
// @Summary      Đánh giá Game
// @Description  Đánh giá sao cho một game (Cần đăng nhập)
// @Tags         Games
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.RateGameRequest true "Thông tin đánh giá"
// @Success      200  {object}  dto.SuccessResponse[dto.RateGameResponse]
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /games/rate [post]
func (h *GameHandler) RateGame(c *gin.Context) {
	// Get user from context
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Vui lòng đăng nhập")
		return
	}

	// Bind request
	var req dto.RateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "Yêu cầu không hợp lệ")
		return
	}

	// Call service
	ctx := context.Background()
	result, err := h.gameService.RateGame(ctx, userID, req.GameID, req.Rating)
	if err != nil {
		handleGameInteractionError(c, err)
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}
