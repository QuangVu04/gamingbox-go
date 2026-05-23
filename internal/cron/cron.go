package cron

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"vault/be/internal/dto"
	"vault/be/internal/services"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultPageLimit = 12
	CacheTTL         = 15 * time.Minute
)

type CronManager struct {
	scheduler     gocron.Scheduler
	rdb           *redis.Client
	gameService   services.GameService
	reviewService services.ReviewService
	listService   services.ListService
}

func NewCronManager(rdb *redis.Client, gameService services.GameService, reviewService services.ReviewService, listService services.ListService) (*CronManager, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &CronManager{
		scheduler:     scheduler,
		rdb:           rdb,
		gameService:   gameService,
		reviewService: reviewService,
		listService:   listService,
	}, nil
}

// Start initializes and starts the cron jobs
func (cm *CronManager) Start() error {
	log.Println("Starting cron jobs...")

	// Schedule trending games cache update every 10 minutes
	_, err := cm.scheduler.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(cm.updateTrendingGamesCache),
	)
	if err != nil {
		return err
	}

	// Schedule trending reviews cache update every 10 minutes
	_, err = cm.scheduler.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(cm.updateTrendingReviewsCache),
	)
	if err != nil {
		return err
	}

	// Schedule trending lists cache update every 10 minutes
	_, err = cm.scheduler.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(cm.updateTrendingListsCache),
	)
	if err != nil {
		return err
	}

	// Start the scheduler
	cm.scheduler.Start()
	log.Println("✓ Cron jobs started successfully")

	return nil
}

// Stop gracefully stops the cron jobs
func (cm *CronManager) Stop() error {
	if cm.scheduler != nil {
		return cm.scheduler.Shutdown()
	}
	return nil
}

// updateTrendingGamesCache updates the trending games cache
func (cm *CronManager) updateTrendingGamesCache() {
	log.Println("🔄 Updating trending games cache...")
	ctx := context.Background()

	games, _, err := cm.gameService.GetTrendingGames(ctx, 1, DefaultPageLimit)
	if err != nil {
		log.Printf("❌ Error updating trending games cache: %v", err)
		return
	}

	log.Printf("✓ Trending games cache updated successfully (%d games)", len(games))
}

// updateTrendingReviewsCache updates the trending reviews cache
func (cm *CronManager) updateTrendingReviewsCache() {
	log.Println("🔄 Updating trending reviews cache...")
	ctx := context.Background()

	reviews, _, err := cm.reviewService.GetTrendingReviews(ctx, 0, 1, DefaultPageLimit)
	if err != nil {
		log.Printf("❌ Error updating trending reviews cache: %v", err)
		return
	}

	log.Printf("✓ Trending reviews cache updated successfully (%d reviews)", len(reviews))
}

// updateTrendingListsCache updates the trending lists cache
func (cm *CronManager) updateTrendingListsCache() {
	log.Println("🔄 Updating trending lists cache...")
	ctx := context.Background()

	lists, _, err := cm.listService.GetTrendingLists(ctx, 1, DefaultPageLimit)
	if err != nil {
		log.Printf("❌ Error updating trending lists cache: %v", err)
		return
	}

	log.Printf("✓ Trending lists cache updated successfully (%d lists)", len(lists))
}

// CacheResponseWithPagination is a helper struct for caching responses with pagination
type CacheResponseWithPagination[T any] struct {
	Data       []T               `json:"data"`
	Pagination dto.PaginationDTO `json:"pagination"`
}

// MarshalCacheData converts response data and pagination to JSON for caching
func MarshalCacheData[T any](data []T, pagination *dto.PaginationDTO) ([]byte, error) {
	cacheData := CacheResponseWithPagination[T]{
		Data:       data,
		Pagination: *pagination,
	}
	return json.Marshal(cacheData)
}

// UnmarshalCacheData converts cached JSON back to response data and pagination
func UnmarshalCacheData[T any](data []byte) ([]T, *dto.PaginationDTO, error) {
	var cacheData CacheResponseWithPagination[T]
	err := json.Unmarshal(data, &cacheData)
	if err != nil {
		return nil, nil, err
	}
	return cacheData.Data, &cacheData.Pagination, nil
}
