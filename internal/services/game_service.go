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
}

type gameService struct {
	gameRepo repositories.GameRepository
	ratingRepo repositories.RatingRepository
	db         *gorm.DB
	rdb      *redis.Client
}

func NewGameService(gameRepo repositories.GameRepository, ratingRepo repositories.RatingRepository, db *gorm.DB, rdb *redis.Client) GameService {
	return &gameService{
		gameRepo: gameRepo,
		ratingRepo: ratingRepo,
		db:         db,
		rdb:      rdb,

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
