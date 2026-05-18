package services

import (
	"context"
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
	SearchGames(ctx context.Context, query string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetPopularGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetGenres(ctx context.Context) ([]models.Genre, error)
	GetPlatforms(ctx context.Context) ([]models.Platform, error)
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

func (s *gameService) SearchGames(ctx context.Context, query string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	games, total, err := s.gameRepo.Search(query, page, limit)
	if err != nil {
		return nil, nil, err
	}

	// Reuse GameTrendingResponse for search results, mapping manually or updating mapper
	// For simplicity, let's map them here or use a helper
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
