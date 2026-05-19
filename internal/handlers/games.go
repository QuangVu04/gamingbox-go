package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

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

// SearchStudios godoc
// @Summary      Tìm kiếm nhà phát triển (Studio)
// @Description  Tìm kiếm danh sách nhà phát triển/studio trong database theo từ khóa
// @Tags         Games
// @Produce      json
// @Param        q query string false "Từ khóa tìm kiếm"
// @Success      200  {object}  interface{}
// @Router       /games/studios [get]
func (h *GameHandler) SearchStudios(c *gin.Context) {
	query := c.Query("q")
	ctx := context.Background()
	studios, err := h.gameService.SearchStudios(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tìm kiếm nhà phát triển"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": studios})
}

// SearchSteamStudios godoc
// @Summary      Tìm kiếm tên studio/nhà phát triển từ Steam Store
// @Description  Tra cứu tên nhà phát triển/studio chính xác từ Steam thông qua storesearch và store web search
// @Tags         Games
// @Produce      json
// @Param        q query string true "Từ khóa tên studio"
// @Success      200  {object}  interface{}
// @Router       /games/steam-studios [get]
func (h *GameHandler) SearchSteamStudios(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu từ khóa tìm kiếm studio"})
		return
	}

	var appIDs []int
	seenApp := make(map[int]bool)

	// 1. Thử tìm kiếm qua API storesearch (phù hợp khi gõ tên game hoặc từ khóa chung)
	searchURL := "https://store.steampowered.com/api/storesearch/?term=" + url.QueryEscape(query) + "&l=vietnamese&cc=VN"
	if resp, err := http.Get(searchURL); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			var searchResult struct {
				Items []struct {
					ID int `json:"id"`
				} `json:"items"`
			}
			if json.Unmarshal(body, &searchResult) == nil {
				for _, item := range searchResult.Items {
					if !seenApp[item.ID] {
						seenApp[item.ID] = true
						appIDs = append(appIDs, item.ID)
						if len(appIDs) >= 4 {
							break
						}
					}
				}
			}
		}
	}

	// 2. Nếu API storesearch không ra kết quả (ví dụ gõ tên riêng studio như Atlus, EA, SEGA),
	// tự động cào trang tìm kiếm web của Steam (nơi có index toàn bộ tên developer/publisher!)
	if len(appIDs) == 0 {
		webSearchURL := "https://store.steampowered.com/search/?term=" + url.QueryEscape(query)
		if webResp, err := http.Get(webSearchURL); err == nil {
			defer webResp.Body.Close()
			if webBody, err := io.ReadAll(webResp.Body); err == nil {
				re := regexp.MustCompile(`https://store\.steampowered\.com/app/(\d+)`)
				matches := re.FindAllStringSubmatch(string(webBody), -1)
				for _, m := range matches {
					if id, err := strconv.Atoi(m[1]); err == nil && !seenApp[id] {
						seenApp[id] = true
						appIDs = append(appIDs, id)
						if len(appIDs) >= 4 {
							break
						}
					}
				}
			}
		}
	}

	if len(appIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": []string{}})
		return
	}

	// 3. Lấy chi tiết các app để trích xuất developers và publishers chính chủ
	studioMap := make(map[string]bool)
	var studios []string

	for _, appID := range appIDs {
		detailURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&l=vietnamese", appID)
		dResp, err := http.Get(detailURL)
		if err != nil {
			continue
		}
		dBody, err := io.ReadAll(dResp.Body)
		dResp.Body.Close()
		if err != nil {
			continue
		}

		var detailResult map[string]struct {
			Success bool `json:"success"`
			Data    struct {
				Developers []string `json:"developers"`
				Publishers []string `json:"publishers"`
			} `json:"data"`
		}

		if err := json.Unmarshal(dBody, &detailResult); err == nil {
			appKey := fmt.Sprintf("%d", appID)
			if appObj, exists := detailResult[appKey]; exists && appObj.Success {
				for _, dev := range appObj.Data.Developers {
					if dev != "" && !studioMap[dev] {
						studioMap[dev] = true
						studios = append(studios, dev)
					}
				}
				for _, pub := range appObj.Data.Publishers {
					if pub != "" && !studioMap[pub] {
						studioMap[pub] = true
						studios = append(studios, pub)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": studios})
}


// CreateGame godoc
// @Summary      Thêm Game mới (Admin)
// @Description  Thêm tựa game mới vào hệ thống kèm hình ảnh, thể loại, nền tảng
// @Tags         Games
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateGameRequest true "Thông tin game"
// @Success      201  {object}  interface{}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req dto.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Dữ liệu gửi lên không hợp lệ: " + err.Error()})
		return
	}

	ctx := context.Background()
	game, err := h.gameService.CreateGame(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Không thể thêm game: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"id":           game.ID,
			"title":        game.Title,
			"release_date": game.ReleaseDate,
		},
	})
}

// DeleteGenre godoc
// @Summary      Xóa Thể loại (Admin)
// @Description  Xóa vĩnh viễn một thể loại khỏi hệ thống
// @Tags         Games
// @Security     BearerAuth
// @Produce      json
// @Param        name path string true "Tên Thể loại"
// @Success      200  {object}  interface{}
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /games/genres/{name} [delete]
func (h *GameHandler) DeleteGenre(c *gin.Context) {
	name := c.Param("name")
	ctx := context.Background()
	if err := h.gameService.DeleteGenre(ctx, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Không thể xóa thể loại: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Đã xóa thể loại thành công"})
}

// DeletePlatform godoc
// @Summary      Xóa Nền tảng (Admin)
// @Description  Xóa vĩnh viễn một nền tảng khỏi hệ thống
// @Tags         Games
// @Security     BearerAuth
// @Produce      json
// @Param        name path string true "Tên Nền tảng"
// @Success      200  {object}  interface{}
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /games/platforms/{name} [delete]
func (h *GameHandler) DeletePlatform(c *gin.Context) {
	name := c.Param("name")
	ctx := context.Background()
	if err := h.gameService.DeletePlatform(ctx, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Không thể xóa nền tảng: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Đã xóa nền tảng thành công"})
}
