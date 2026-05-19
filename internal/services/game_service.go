package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
	redisUtil "vault/be/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const CacheTTL = 15 * time.Minute

type GameService interface {
	GetTrendingGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	RateGame(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error)
	GetGameByID(ctx context.Context, id uint) (*dto.GameDetailResponse, error)
	SearchGames(ctx context.Context, query, category, platform, sort string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetPopularGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetGenres(ctx context.Context) ([]models.Genre, error)
	GetPlatforms(ctx context.Context) ([]models.Platform, error)
	CreateGame(ctx context.Context, req dto.CreateGameRequest) (*models.Game, error)
	UpdateGame(ctx context.Context, id uint, req dto.CreateGameRequest) (*models.Game, error)
	DeleteGenre(ctx context.Context, name string) error
	DeletePlatform(ctx context.Context, name string) error
	SearchStudios(ctx context.Context, query string) ([]models.Studio, error)
	DeleteGame(ctx context.Context, id uint) error
}


type gameService struct {
	gameRepo   repositories.GameRepository
	ratingRepo repositories.RatingRepository
	reviewRepo repositories.ReviewRepository
	db         *gorm.DB
	rdb        *redis.Client
}

func NewGameService(gameRepo repositories.GameRepository, ratingRepo repositories.RatingRepository, reviewRepo repositories.ReviewRepository, db *gorm.DB, rdb *redis.Client) GameService {
	return &gameService{
		gameRepo:   gameRepo,
		ratingRepo: ratingRepo,
		reviewRepo: reviewRepo,
		db:         db,
		rdb:        rdb,
	}
}

// CachedGameResponse is used for marshaling/unmarshaling game data with pagination
type CachedGameResponse struct {
	Data       []dto.GameTrendingResponse `json:"data"`
	Pagination *dto.PaginationDTO         `json:"pagination"`
}

func (s *gameService) GetTrendingGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("games", page, limit)
	cached, err := redisUtil.GetCached[CachedGameResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending games (page=%d, limit=%d)", page, limit)
		return cached.Data, cached.Pagination, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for trending games (page=%d, limit=%d), fetching from database", page, limit)
	trendingGames, totalCount, err := s.gameRepo.GetTrendingGames(page, limit)
	if err != nil {
		return nil, nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy dữ liệu game trending")
	}

	// Convert to DTOs
	responses := mapper.ToGameTrendingResponses(trendingGames)

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(totalCount),
		CurrentPage:  page,
		TotalPages:   totalPages,
		Limit:        limit,
	}

	// Cache the response
	cacheData := &CachedGameResponse{
		Data:       responses,
		Pagination: pagination,
	}
	err = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	if err != nil {
		log.Printf("⚠ Failed to cache trending games: %v", err)
	}

	return responses, pagination, nil
}

func (s *gameService) RateGame(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error) {
	// Validate rating range
	if rating < 0.5 || rating > 5.0 {
		return nil, dto.NewServiceError("VALIDATION_ERROR", "điểm số phải từ 0.5 đến 5.0")
	}

	// Use transaction to ensure consistency
	var result *dto.RateGameResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Upsert rating
		ratingRecord := &models.Rating{
			UserID:    userID,
			GameID:    gameID,
			Rating:    rating,
			CreatedAt: time.Now(),
		}

		// handle upsert with transaction
		if err := s.ratingRepo.UpsertRating(ratingRecord); err != nil {
			return err
		}

		// Get updated game stats
		avgRating, totalRatings, err := s.ratingRepo.GetGameStats(gameID)
		if err != nil {
			return err
		}

		// Update game with new stats
		if err := s.ratingRepo.UpdateGameRating(gameID, avgRating, totalRatings); err != nil {
			return err
		}

		result = &dto.RateGameResponse{
			MyRating:     rating,
			NewGameAvg:   avgRating,
			TotalRatings: totalRatings,
		}

		return nil
	})

	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lưu đánh giá")
	}

	return result, nil
}

func (s *gameService) GetGameByID(ctx context.Context, id uint) (*dto.GameDetailResponse, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy game")
	}

	// 1. Fetch Popular Reviews (top 3)
	popularReviews, _ := s.reviewRepo.GetByGameID(id, "popular", 3)
	popCommentCounts, _ := s.reviewRepo.GetCommentCounts(getReviewIDs(popularReviews))
	popularDTOs := mapper.ToReviewCompactResponses(popularReviews, popCommentCounts)

	// 2. Fetch Recent Reviews (top 3)
	recentReviews, _ := s.reviewRepo.GetByGameID(id, "recent", 3)
	recCommentCounts, _ := s.reviewRepo.GetCommentCounts(getReviewIDs(recentReviews))
	recentDTOs := mapper.ToReviewCompactResponses(recentReviews, recCommentCounts)

	// 3. More from this studio (top 6)
	moreFromStudio, _ := s.gameRepo.GetByStudio(game.StudioID, game.ID, 6)
	studioDTOs := mapper.ToSimpleGameResponses(moreFromStudio)

	// 4. Similar Games (same genres, top 6)
	genreIDs := make([]uint, 0, len(game.Genres))
	for _, g := range game.Genres {
		genreIDs = append(genreIDs, g.ID)
	}
	similarGames, _ := s.gameRepo.GetByGenres(genreIDs, game.ID, 6)
	similarDTOs := mapper.ToSimpleGameResponses(similarGames)

	// Map core response
	resp := mapper.ToGameDetailResponse(game)

	// Attach extra data
	resp.PopularReviews = popularDTOs
	resp.RecentReviews = recentDTOs
	resp.MoreFromStudio = studioDTOs
	resp.SimilarGames = similarDTOs

	return resp, nil
}

func getReviewIDs(reviews []models.Review) []uint {
	ids := make([]uint, 0, len(reviews))
	for _, r := range reviews {
		ids = append(ids, r.ID)
	}
	return ids
}

func (s *gameService) SearchGames(ctx context.Context, query, category, platform, sort string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	games, total, err := s.gameRepo.SearchAdminGames(query, category, platform, "", "", "", sort, page, limit)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.GameTrendingResponse, 0, len(games))
	for _, g := range games {
		thumbnail := ""
		for _, img := range g.Images {
			if img.ImgType == "header" {
				thumbnail = img.OgURL
				break
			}
		}
		if thumbnail == "" && len(g.Images) > 0 {
			thumbnail = g.Images[0].OgURL
		}

		studios := make([]string, 0)
		if g.Studio.ID > 0 {
			studios = append(studios, g.Studio.Name)
		}

		responses = append(responses, dto.GameTrendingResponse{
			GameID:       g.ID,
			Title:        g.Title,
			Thumbnail:    thumbnail,
			AvgRating:    g.AvgRating,
			TotalReviews: g.ReviewCount,
			ReleaseDate:  g.ReleaseDate,
			Studios:      studios,
		})
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   int(math.Ceil(float64(total) / float64(limit))),
		Limit:        limit,
	}

	return responses, pagination, nil
}

func (s *gameService) GetPopularGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	games, total, err := s.gameRepo.GetPopular(page, limit)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.GameTrendingResponse, 0, len(games))
	for _, g := range games {
		thumbnail := ""
		for _, img := range g.Images {
			if img.ImgType == "header" {
				thumbnail = img.OgURL
				break
			}
		}
		responses = append(responses, dto.GameTrendingResponse{
			GameID:       g.ID,
			Title:        g.Title,
			Thumbnail:    thumbnail,
			AvgRating:    g.AvgRating,
			TotalReviews: g.ReviewCount,
			ReleaseDate:  g.ReleaseDate,
		})
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   int(math.Ceil(float64(total) / float64(limit))),
		Limit:        limit,
	}

	return responses, pagination, nil
}

func (s *gameService) GetGenres(ctx context.Context) ([]models.Genre, error) {
	return s.gameRepo.GetGenres()
}

func (s *gameService) GetPlatforms(ctx context.Context) ([]models.Platform, error) {
	return s.gameRepo.GetPlatforms()
}

func (s *gameService) CreateGame(ctx context.Context, req dto.CreateGameRequest) (*models.Game, error) {
	// 1. Xử lý Studio: Tìm hoặc tạo mới studio
	var studio models.Studio
	if req.Studio == "" {
		req.Studio = "Independent Studio"
	}
	if err := s.db.Where("name = ?", req.Studio).FirstOrCreate(&studio, models.Studio{Name: req.Studio, Description: "Studio phát triển tựa game " + req.Title}).Error; err != nil {
		return nil, err
	}

	// 2. Xử lý Genres: Tìm hoặc tạo mới các thể loại
	var genres []models.Genre
	for _, genreName := range req.Genres {
		if genreName == "" {
			continue
		}
		var g models.Genre
		if err := s.db.Where("name = ?", genreName).FirstOrCreate(&g, models.Genre{Name: genreName}).Error; err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	// 3. Xử lý Platforms: Tìm hoặc tạo mới các nền tảng
	var platforms []models.Platform
	for _, platName := range req.Platforms {
		if platName == "" {
			continue
		}
		var p models.Platform
		if err := s.db.Where("name = ?", platName).FirstOrCreate(&p, models.Platform{Name: platName}).Error; err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}

	// 4. Xử lý Ngày ra mắt (ReleaseDate)
	releaseDate := time.Now()
	if req.ReleaseDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ReleaseDate); err == nil {
			releaseDate = parsed
		} else if parsedISO, errISO := time.Parse(time.RFC3339, req.ReleaseDate); errISO == nil {
			releaseDate = parsedISO
		}
	}

	// 5. Xử lý Rating
	var ratingVal float64
	if req.Rating != "" {
		fmt.Sscanf(req.Rating, "%f", &ratingVal)
	}
	if ratingVal == 0 {
		ratingVal = 4.5
	}

	// Parsing playtimes
	var averagePlaytime float64
	var playtimeStory float64
	var playtimeMaster float64
	if req.AveragePlaytime != "" {
		fmt.Sscanf(req.AveragePlaytime, "%f", &averagePlaytime)
	}
	if req.PlaytimeStory != "" {
		fmt.Sscanf(req.PlaytimeStory, "%f", &playtimeStory)
	}
	if req.PlaytimeMaster != "" {
		fmt.Sscanf(req.PlaytimeMaster, "%f", &playtimeMaster)
	}

	// 6. Khởi tạo đối tượng Game
	game := &models.Game{
		Title:                 req.Title,
		Description:           req.Description,
		ReleaseDate:           releaseDate,
		StudioID:              studio.ID,
		AvgRating:             ratingVal,
		ReviewCount:           1, // Để hiển thị đẹp
		LikeCount:             10,
		AveragePlaytime:       averagePlaytime,
		PlaytimeStory:         playtimeStory,
		PlaytimeCompletionist: playtimeMaster,
		Genres:                genres,
		Platforms:             platforms,
	}

	// 7. Lưu Game vào DB
	if err := s.gameRepo.CreateGame(game); err != nil {
		return nil, err
	}

	// 8. Xử lý Hình ảnh (GameImg)
	// Banner chính là header
	if req.Images.Header != "" {
		headerImg := models.GameImg{
			OgURL:   req.Images.Header,
			Thumb:   req.Images.Header,
			ImgType: "header",
			GameID:  game.ID,
		}
		s.db.Create(&headerImg)
	}
	// Ảnh bìa là cover
	if req.Images.Main != "" {
		coverImg := models.GameImg{
			OgURL:   req.Images.Main,
			Thumb:   req.Images.Main,
			ImgType: "cover",
			GameID:  game.ID,
		}
		s.db.Create(&coverImg)
	}
	// Các ảnh screenshot
	for _, ss := range req.Screenshots {
		if ss == "" {
			continue
		}
		ssImg := models.GameImg{
			OgURL:   ss,
			Thumb:   ss,
			ImgType: "screenshot",
			GameID:  game.ID,
		}
		s.db.Create(&ssImg)
	}

	// Invalidate cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")

	return game, nil
}

func (s *gameService) UpdateGame(ctx context.Context, id uint, req dto.CreateGameRequest) (*models.Game, error) {
	// 1. Tìm game cũ
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy game")
	}

	// 2. Xử lý Studio
	var studio models.Studio
	if req.Studio == "" {
		req.Studio = "Independent Studio"
	}
	if err := s.db.Where("name = ?", req.Studio).FirstOrCreate(&studio, models.Studio{Name: req.Studio, Description: "Studio phát triển tựa game " + req.Title}).Error; err != nil {
		return nil, err
	}

	// 3. Xử lý Genres
	var genres []models.Genre
	for _, genreName := range req.Genres {
		if genreName == "" {
			continue
		}
		var g models.Genre
		if err := s.db.Where("name = ?", genreName).FirstOrCreate(&g, models.Genre{Name: genreName}).Error; err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	// 4. Xử lý Platforms
	var platforms []models.Platform
	for _, platName := range req.Platforms {
		if platName == "" {
			continue
		}
		var p models.Platform
		if err := s.db.Where("name = ?", platName).FirstOrCreate(&p, models.Platform{Name: platName}).Error; err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}

	// 5. Xử lý ReleaseDate
	releaseDate := game.ReleaseDate
	if req.ReleaseDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ReleaseDate); err == nil {
			releaseDate = parsed
		} else if parsedISO, errISO := time.Parse(time.RFC3339, req.ReleaseDate); errISO == nil {
			releaseDate = parsedISO
		}
	}

	// 6. Xử lý Rating
	var ratingVal float64
	if req.Rating != "" {
		fmt.Sscanf(req.Rating, "%f", &ratingVal)
	}
	if ratingVal == 0 {
		ratingVal = game.AvgRating
	}

	// Parsing playtimes
	var averagePlaytime float64
	var playtimeStory float64
	var playtimeMaster float64
	if req.AveragePlaytime != "" {
		fmt.Sscanf(req.AveragePlaytime, "%f", &averagePlaytime)
	}
	if req.PlaytimeStory != "" {
		fmt.Sscanf(req.PlaytimeStory, "%f", &playtimeStory)
	}
	if req.PlaytimeMaster != "" {
		fmt.Sscanf(req.PlaytimeMaster, "%f", &playtimeMaster)
	}

	// Cập nhật trường dữ liệu
	game.Title = req.Title
	game.Description = req.Description
	game.ReleaseDate = releaseDate
	game.StudioID = studio.ID
	game.AvgRating = ratingVal
	game.AveragePlaytime = averagePlaytime
	game.PlaytimeStory = playtimeStory
	game.PlaytimeCompletionist = playtimeMaster
	game.Genres = genres
	game.Platforms = platforms

	// Xử lý Hình ảnh
	var gameImages []models.GameImg
	if req.Images.Header != "" {
		gameImages = append(gameImages, models.GameImg{
			OgURL:   req.Images.Header,
			Thumb:   req.Images.Header,
			ImgType: "header",
			GameID:  game.ID,
		})
	}
	if req.Images.Main != "" {
		gameImages = append(gameImages, models.GameImg{
			OgURL:   req.Images.Main,
			Thumb:   req.Images.Main,
			ImgType: "cover",
			GameID:  game.ID,
		})
	}
	for _, ss := range req.Screenshots {
		if ss == "" {
			continue
		}
		gameImages = append(gameImages, models.GameImg{
			OgURL:   ss,
			Thumb:   ss,
			ImgType: "screenshot",
			GameID:  game.ID,
		})
	}
	game.Images = gameImages

	// 7. Lưu thay đổi
	if err := s.gameRepo.UpdateGame(game); err != nil {
		return nil, err
	}

	// 8. Xóa cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")

	return game, nil
}

func (s *gameService) DeleteGenre(ctx context.Context, name string) error {
	return s.gameRepo.DeleteGenreByName(name)
}

func (s *gameService) DeletePlatform(ctx context.Context, name string) error {
	return s.gameRepo.DeletePlatformByName(name)
}

func (s *gameService) SearchStudios(ctx context.Context, query string) ([]models.Studio, error) {
	return s.gameRepo.SearchStudios(query)
}

func (s *gameService) DeleteGame(ctx context.Context, id uint) error {
	// Call repository to perform deletion inside transaction
	if err := s.gameRepo.DeleteGame(id); err != nil {
		return err
	}

	// Invalidate cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")
	return nil
}

