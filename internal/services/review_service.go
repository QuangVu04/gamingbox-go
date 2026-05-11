package services

import (
	"context"
	"log"
	"math"
	"vault/be/internal/models"

	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/repositories"
	redisUtil "vault/be/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type ReviewService interface {
	GetTrendingReviews(ctx context.Context, page, limit int) ([]dto.ReviewTrendingResponse, *dto.PaginationDTO, error)
	CreateReview(ctx context.Context, userID uint, req dto.CreateReviewRequest) (*dto.ReviewTrendingResponse, error)
	UpdateReview(ctx context.Context, userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewTrendingResponse, error)
	DeleteReview(ctx context.Context, userID, reviewID uint) error
	GetComments(ctx context.Context, reviewID uint) ([]dto.CommentResponse, error)
	AddComment(ctx context.Context, userID, reviewID uint, req dto.CommentRequest) (*dto.CommentResponse, error)
	GetReviewByID(ctx context.Context, reviewID uint) (*dto.ReviewTrendingResponse, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
	userRepo   repositories.UserRepository
	rdb        *redis.Client
}

func NewReviewService(reviewRepo repositories.ReviewRepository, userRepo repositories.UserRepository, rdb *redis.Client) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
		userRepo:   userRepo,
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

func (s *reviewService) CreateReview(ctx context.Context, userID uint, req dto.CreateReviewRequest) (*dto.ReviewTrendingResponse, error) {
	review := &models.Review{
		UserID:     userID,
		TargetID:   req.GameID,
		TargetType: "game",
		Title:      req.Title,
		Content:    req.Content,
		Img:        req.Img,
		Recommend:  req.Recommend,
		IsSpoiler:  req.IsSpoiler,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể tạo review")
	}

	// Fetch with preloads for mapping
	fullReview, err := s.reviewRepo.FindByID(review.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToReviewTrendingResponse(fullReview, 0), nil
}

func (s *reviewService) UpdateReview(ctx context.Context, userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewTrendingResponse, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy review")
	}

	if review.UserID != userID {
		return nil, dto.NewServiceError("FORBIDDEN", "không có quyền chỉnh sửa")
	}

	if req.Title != "" {
		review.Title = req.Title
	}
	if req.Content != "" {
		review.Content = req.Content
	}
	if req.Recommend != "" {
		review.Recommend = req.Recommend
	}
	review.IsSpoiler = req.IsSpoiler
	if req.Img != "" {
		review.Img = req.Img
	}

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể cập nhật review")
	}

	commentCounts, _ := s.reviewRepo.GetCommentCounts([]uint{reviewID})
	return mapper.ToReviewTrendingResponse(review, commentCounts[reviewID]), nil
}

func (s *reviewService) DeleteReview(ctx context.Context, userID, reviewID uint) error {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return dto.NewServiceError("NOT_FOUND", "không tìm thấy review")
	}

	if review.UserID != userID {
		return dto.NewServiceError("FORBIDDEN", "không có quyền xóa")
	}

	return s.reviewRepo.Delete(reviewID)
}

func (s *reviewService) GetComments(ctx context.Context, reviewID uint) ([]dto.CommentResponse, error) {
	comments, err := s.reviewRepo.GetComments(reviewID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0)
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
	}

	userMap := make(map[uint]models.User)
	for _, uid := range userIDs {
		if _, ok := userMap[uid]; !ok {
			user, err := s.userRepo.FindByID(uid)
			if err == nil {
				userMap[uid] = *user
			}
		}
	}

	return mapper.ToCommentResponses(comments, userMap), nil
}

func (s *reviewService) AddComment(ctx context.Context, userID, reviewID uint, req dto.CommentRequest) (*dto.CommentResponse, error) {
	comment := &models.Comment{
		ReviewID: reviewID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := s.reviewRepo.AddComment(comment); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể thêm bình luận")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCommentResponse(comment, user), nil
}

func (s *reviewService) GetReviewByID(ctx context.Context, reviewID uint) (*dto.ReviewTrendingResponse, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy review")
	}

	commentCounts, _ := s.reviewRepo.GetCommentCounts([]uint{reviewID})
	return mapper.ToReviewTrendingResponse(review, commentCounts[reviewID]), nil
}
