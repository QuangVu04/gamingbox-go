package handlers

import (
	"context"
	"io"
	"net/http"
	"net/url"

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

// GetGameDetail godoc
// @Summary      Lấy chi tiết Game
// @Description  Lấy thông tin đầy đủ của một game dựa trên ID
// @Tags         Games
// @Produce      json
// @Param        id path int true "ID của Game"
// @Success      200  {object}  dto.SuccessResponse[dto.GameDetailResponse]
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /games/{id} [get]
func (h *GameHandler) GetGameDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseUint(idStr)
	if err != nil {
		utils.ValidationError(c, "ID không hợp lệ")
		return
	}

	ctx := context.Background()
	game, err := h.gameService.GetGameByID(ctx, uint(id))
	if err != nil {
		handleGameInteractionError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, game)
}

// SearchGames godoc
// @Summary      Tìm kiếm Game
// @Description  Tìm kiếm game theo tên có phân trang
// @Tags         Games
// @Produce      json
// @Param        q query string true "Từ khóa tìm kiếm"
// @Param        page query int false "Trang hiện tại" default(1)
// @Param        limit query int false "Số lượng mỗi trang" default(12)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.GameTrendingResponse]
// @Router       /games/search [get]
func (h *GameHandler) SearchGames(c *gin.Context) {
	query := c.Query("q")
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 12, 1, 100)

	ctx := context.Background()
	games, pagination, err := h.gameService.SearchGames(ctx, query, page, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", "lỗi tìm kiếm")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"pagination": pagination,
		"data":       games,
	})
}

// PopularGames godoc
// @Summary      Danh sách Game phổ biến
// @Description  Lấy danh sách game có đánh giá cao nhất
// @Tags         Games
// @Produce      json
// @Param        page query int false "Trang hiện tại" default(1)
// @Param        limit query int false "Số lượng mỗi trang" default(12)
// @Success      200  {object}  dto.PaginatedResponse[[]dto.GameTrendingResponse]
// @Router       /games/popular [get]
func (h *GameHandler) PopularGames(c *gin.Context) {
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 12, 1, 100)

	ctx := context.Background()
	games, pagination, err := h.gameService.GetPopularGames(ctx, page, limit)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "SERVER_ERROR", "lỗi lấy dữ liệu")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"pagination": pagination,
		"data":       games,
	})
}

// SearchSteamGames godoc
// @Summary      Tìm kiếm game trực tiếp từ Steam Store
// @Description  Proxy tìm kiếm game từ Steam API không bị lỗi CORS
// @Tags         Games
// @Produce      json
// @Param        term query string true "Từ khóa tìm kiếm"
// @Success      200  {object}  interface{}
// @Router       /games/steam-search [get]
func (h *GameHandler) SearchSteamGames(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu từ khóa tìm kiếm"})
		return
	}

	targetURL := "https://store.steampowered.com/api/storesearch/?term=" + url.QueryEscape(term) + "&l=vietnamese&cc=VN"
	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kết nối đến Steam API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc dữ liệu từ Steam"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

// GetSteamAppDetails godoc
// @Summary      Lấy chi tiết game từ Steam Store
// @Description  Proxy lấy thông tin chi tiết game từ Steam API
// @Tags         Games
// @Produce      json
// @Param        appId query string true "App ID của Steam"
// @Success      200  {object}  interface{}
// @Router       /games/steam-detail [get]
func (h *GameHandler) GetSteamAppDetails(c *gin.Context) {
	appID := c.Query("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu App ID"})
		return
	}

	targetURL := "https://store.steampowered.com/api/appdetails?appids=" + url.QueryEscape(appID) + "&l=vietnamese"
	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kết nối đến Steam API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc dữ liệu từ Steam"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

// GetGenres godoc
// @Summary      Lấy danh sách thể loại
// @Description  Lấy toàn bộ danh sách thể loại game từ database
// @Tags         Games
// @Produce      json
// @Success      200  {object}  interface{}
// @Router       /games/genres [get]
func (h *GameHandler) GetGenres(c *gin.Context) {
	ctx := context.Background()
	genres, err := h.gameService.GetGenres(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách thể loại"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": genres})
}

// GetPlatforms godoc
// @Summary      Lấy danh sách nền tảng
// @Description  Lấy toàn bộ danh sách hệ máy/nền tảng từ database
// @Tags         Games
// @Produce      json
// @Success      200  {object}  interface{}
// @Router       /games/platforms [get]
func (h *GameHandler) GetPlatforms(c *gin.Context) {
	ctx := context.Background()
	platforms, err := h.gameService.GetPlatforms(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách nền tảng"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": platforms})
}
