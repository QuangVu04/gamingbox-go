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
	GetTrendingReviews(ctx context.Context, currentUserID uint, page, limit int) ([]dto.ReviewTrendingResponse, *dto.PaginationDTO, error)
	CreateReview(ctx context.Context, userID uint, req dto.CreateReviewRequest) (*dto.ReviewTrendingResponse, error)
	UpdateReview(ctx context.Context, userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewTrendingResponse, error)
	DeleteReview(ctx context.Context, userID, reviewID uint) error
	GetComments(ctx context.Context, currentUserID, reviewID uint) ([]dto.CommentResponse, error)
	AddComment(ctx context.Context, userID, reviewID uint, req dto.CommentRequest) (*dto.CommentResponse, error)
	GetReviewByID(ctx context.Context, currentUserID, reviewID uint) (*dto.ReviewTrendingResponse, error)
	GetGameReviews(ctx context.Context, currentUserID, gameID uint, page, limit int, sort string) (*dto.GameReviewsResponse, error)
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

func (s *reviewService) GetTrendingReviews(ctx context.Context, currentUserID uint, page, limit int) ([]dto.ReviewTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var responses []dto.ReviewTrendingResponse
	var pagination *dto.PaginationDTO

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("reviews", page, limit)
	cached, err := redisUtil.GetCached[CachedReviewResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending reviews (page=%d, limit=%d)", page, limit)
		responses = cached.Data
		pagination = cached.Pagination
	} else {
		// Cache miss - get from database
		log.Printf("Cache miss for trending reviews (page=%d, limit=%d), fetching from database", page, limit)

		reviews, total, err := s.reviewRepo.GetTrendingReviews(page, limit)
		if err != nil {
			return nil, nil, err
		}

		reviewIDs := make([]uint, 0, len(reviews))
		for _, review := range reviews {
			reviewIDs = append(reviewIDs, review.ID)
		}

		commentCounts, err := s.reviewRepo.GetCommentCounts(reviewIDs)
		if err != nil {
			commentCounts = make(map[uint]int)
		}

		responses = mapper.ToReviewTrendingResponses(reviews, commentCounts, nil)

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		if totalPages < 1 {
			totalPages = 1
		}

		pagination = &dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		}

		cacheData := &CachedReviewResponse{
			Data:       responses,
			Pagination: pagination,
		}
		_ = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	}

	// Populate UserHasLiked for the current user dynamically
	if currentUserID > 0 {
		reviewIDs := make([]uint, 0, len(responses))
		for _, r := range responses {
			reviewIDs = append(reviewIDs, r.ReviewID)
		}
		likedReviews, _ := s.reviewRepo.GetUserLikedReviews(currentUserID, reviewIDs)
		for i := range responses {
			responses[i].UserHasLiked = likedReviews[responses[i].ReviewID]
		}
	}

	return responses, pagination, nil
}

func (s *reviewService) CreateReview(ctx context.Context, userID uint, req dto.CreateReviewRequest) (*dto.ReviewTrendingResponse, error) {
	// Removed the check for existing review so users can have multiple reviews for the same game

	review := &models.Review{
		UserID:     userID,
		TargetID:   req.GameID,
		TargetType: "game",
		Content:    req.Content,
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

	return mapper.ToReviewTrendingResponse(fullReview, 0, false), nil
}

func (s *reviewService) UpdateReview(ctx context.Context, userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewTrendingResponse, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy review")
	}

	if review.UserID != userID {
		return nil, dto.NewServiceError("FORBIDDEN", "không có quyền chỉnh sửa")
	}

	if req.Content != "" {
		review.Content = req.Content
	}
	if req.Recommend != "" {
		review.Recommend = req.Recommend
	}
	review.IsSpoiler = req.IsSpoiler


	if err := s.reviewRepo.Update(review); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể cập nhật review")
	}

	commentCounts, _ := s.reviewRepo.GetCommentCounts([]uint{reviewID})
	likedReviews, _ := s.reviewRepo.GetUserLikedReviews(userID, []uint{reviewID})
	return mapper.ToReviewTrendingResponse(review, commentCounts[reviewID], likedReviews[reviewID]), nil
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

func (s *reviewService) GetComments(ctx context.Context, currentUserID, reviewID uint) ([]dto.CommentResponse, error) {
	comments, err := s.reviewRepo.GetComments(reviewID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0)
	commentIDs := make([]uint, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
		commentIDs = append(commentIDs, c.ID)
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

	likedComments, _ := s.reviewRepo.GetUserLikedComments(currentUserID, commentIDs)

	return mapper.ToCommentResponses(comments, userMap, likedComments), nil
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

	return mapper.ToCommentResponse(comment, user, false), nil
}

func (s *reviewService) GetReviewByID(ctx context.Context, currentUserID, reviewID uint) (*dto.ReviewTrendingResponse, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy review")
	}

	commentCounts, _ := s.reviewRepo.GetCommentCounts([]uint{reviewID})
	likedReviews, _ := s.reviewRepo.GetUserLikedReviews(currentUserID, []uint{reviewID})
	return mapper.ToReviewTrendingResponse(review, commentCounts[reviewID], likedReviews[reviewID]), nil
}
func (s *reviewService) GetGameReviews(ctx context.Context, currentUserID, gameID uint, page, limit int, sort string) (*dto.GameReviewsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	reviews, total, err := s.reviewRepo.GetGameReviews(gameID, page, limit, sort)
	if err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy danh sách review")
	}

	reviewIDs := make([]uint, 0, len(reviews))
	for _, r := range reviews {
		reviewIDs = append(reviewIDs, r.ID)
	}

	commentCounts, _ := s.reviewRepo.GetCommentCounts(reviewIDs)
	likedReviews, _ := s.reviewRepo.GetUserLikedReviews(currentUserID, reviewIDs)

	responses := mapper.ToReviewCompactResponses(reviews, commentCounts, likedReviews)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &dto.GameReviewsResponse{
		Reviews: responses,
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
	}, nil
}
