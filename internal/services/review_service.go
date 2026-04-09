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

type ReviewService interface {
	GetTrendingReviews(ctx context.Context, page, limit int) ([]dto.ReviewTrendingResponse, *dto.PaginationDTO, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
	rdb        *redis.Client
}

func NewReviewService(reviewRepo repositories.ReviewRepository, rdb *redis.Client) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
		rdb:        rdb,
	}
}

// CachedReviewResponse is used for marshaling/unmarshaling review data with pagination
type CachedReviewResponse struct {
	Data       []dto.ReviewTrendingResponse `json:"data"`
	Pagination *dto.PaginationDTO           `json:"pagination"`
}

func (s *reviewService) GetTrendingReviews(ctx context.Context, page, limit int) ([]dto.ReviewTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("reviews", page, limit)
	cached, err := redisUtil.GetCached[CachedReviewResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending reviews (page=%d, limit=%d)", page, limit)
		return cached.Data, cached.Pagination, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for trending reviews (page=%d, limit=%d), fetching from database", page, limit)

	// Fetch trending reviews from repository
	reviews, total, err := s.reviewRepo.GetTrendingReviews(page, limit)
	if err != nil {
		return nil, nil, err
	}

	// Get comment counts for all reviews
	reviewIDs := make([]uint, 0, len(reviews))
	for _, review := range reviews {
		reviewIDs = append(reviewIDs, review.ID)
	}

	commentCounts, err := s.reviewRepo.GetCommentCounts(reviewIDs)
	if err != nil {
		// Continue anyway if we can't get comment counts
		commentCounts = make(map[uint]int)
	}

	// Map to response DTOs
	responses := mapper.ToReviewTrendingResponses(reviews, commentCounts)

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
	cacheData := &CachedReviewResponse{
		Data:       responses,
		Pagination: pagination,
	}
	err = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	if err != nil {
		log.Printf("⚠ Failed to cache trending reviews: %v", err)
	}

	return responses, pagination, nil
}
