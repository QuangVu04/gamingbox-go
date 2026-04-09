package services

import (
	"context"
	"log"
	"math"

	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/repositories"
	redisUtil "vault/be/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type ListService interface {
	GetTrendingLists(ctx context.Context, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error)
}

type listService struct {
	listRepo repositories.ListRepository
	rdb      *redis.Client
}

func NewListService(listRepo repositories.ListRepository, rdb *redis.Client) ListService {
	return &listService{
		listRepo: listRepo,
		rdb:      rdb,
	}
}

// CachedListResponse is used for marshaling/unmarshaling list data with pagination
type CachedListResponse struct {
	Data       []dto.ListTrendingResponse `json:"data"`
	Pagination *dto.PaginationDTO         `json:"pagination"`
}

func (s *listService) GetTrendingLists(ctx context.Context, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("lists", page, limit)
	cached, err := redisUtil.GetCached[CachedListResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending lists (page=%d, limit=%d)", page, limit)
		return cached.Data, cached.Pagination, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for trending lists (page=%d, limit=%d), fetching from database", page, limit)

	// Fetch trending lists from repository (already includes weekly likes count)
	listsData, total, err := s.listRepo.GetTrendingLists(page, limit)
	if err != nil {
		return nil, nil, err
	}

	// Map to response DTOs
	responses := mapper.ToTrendingListResponses(listsData)

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   totalPages,
		Limit:        limit,
	}

	// Cache the response
	cacheData := &CachedListResponse{
		Data:       responses,
		Pagination: pagination,
	}
	err = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	if err != nil {
		log.Printf("⚠ Failed to cache trending lists: %v", err)
	}

	return responses, pagination, nil
}
