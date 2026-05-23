package services

import (
	"context"
	"encoding/json"
	"math"

	"vault/be/internal/dto"
	"vault/be/internal/repositories"

	"github.com/redis/go-redis/v9"
)

const InteractionQueueKey = "queue:game_interactions"

type LikeService interface {
	// ToggleLike handles polymorphic like/unlike with atomic counter update
	ToggleLike(ctx context.Context, userID, targetID uint, targetType string) (*dto.LikeGameResponse, error)
	// ToggleLikeDB performs like/unlike directly in DB
	ToggleLikeDB(ctx context.Context, userID, targetID uint, targetType string) (bool, error)
	// CheckLike checks if user has liked a target
	CheckLike(ctx context.Context, userID, targetID uint, targetType string) (bool, error)
	// GetGameLikes retrieves users who liked a game with pagination and sorting
	GetGameLikes(ctx context.Context, gameID uint, page, limit int, sort string) (*dto.GameLikesResponse, error)
}

type likeService struct {
	likeRepo     repositories.LikeRepository
	reviewRepo   repositories.ReviewRepository
	listRepo     repositories.ListRepository
	rdb          *redis.Client
	notifService NotificationService
}

func NewLikeService(
	likeRepo repositories.LikeRepository,
	rdb *redis.Client,
	notifService NotificationService,
	reviewRepo repositories.ReviewRepository,
	listRepo repositories.ListRepository,
) LikeService {
	return &likeService{
		likeRepo:     likeRepo,
		rdb:          rdb,
		notifService: notifService,
		reviewRepo:   reviewRepo,
		listRepo:     listRepo,
	}
}

// ToggleLike queues the like/unlike request to Redis for async execution
func (s *likeService) ToggleLike(ctx context.Context, userID, targetID uint, targetType string) (*dto.LikeGameResponse, error) {
	// Validate targetType
	if targetType != "game" && targetType != "review" && targetType != "list" && targetType != "comment" {
		return nil, dto.NewServiceError("VALIDATION_ERROR", "loại target không hợp lệ")
	}

	// Check current state to determine optimistic response
	currentLiked, err := s.CheckLike(ctx, userID, targetID, targetType)
	if err != nil {
		currentLiked = false
	}
	expectedLikedState := !currentLiked

	// Queue the task
	task := dto.InteractionTask{
		UserID:     userID,
		TargetID:   targetID,
		Type:       "like",
		TargetType: targetType,
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể queue yêu cầu")
	}

	if s.rdb != nil {
		err = s.rdb.LPush(ctx, InteractionQueueKey, taskBytes).Err()
		if err != nil {
			return nil, dto.NewServiceError("SERVER_ERROR", "không thể queue yêu cầu vào redis")
		}
	} else {
		// Fallback to direct DB write if Redis is down
		_, err = s.ToggleLikeDB(ctx, userID, targetID, targetType)
		if err != nil {
			return nil, err
		}
	}

	return &dto.LikeGameResponse{IsLiked: expectedLikedState}, nil
}

// ToggleLikeDB performs like/unlike with atomic counter update in database
func (s *likeService) ToggleLikeDB(ctx context.Context, userID, targetID uint, targetType string) (bool, error) {
	// Validate targetType
	if targetType != "game" && targetType != "review" && targetType != "list" && targetType != "comment" {
		return false, dto.NewServiceError("VALIDATION_ERROR", "loại target không hợp lệ")
	}

	// Use ToggleLike which handles atomic counter update in transaction
	isLiked, err := s.likeRepo.ToggleLike(userID, targetID, targetType)
	if err != nil {
		return false, dto.NewServiceError("SERVER_ERROR", "không thể xử lý like")
	}

	// Invalidate trending cache if it's a game
	if targetType == "game" && s.rdb != nil {
		s.invalidateTrendingCache(ctx)
	}

	// Trigger notification if liked
	if isLiked {
		go s.handleLikeNotification(userID, targetID, targetType)
	}

	return isLiked, nil
}

func (s *likeService) handleLikeNotification(senderID, targetID uint, targetType string) {
	var receiverID uint

	switch targetType {
	case "review":
		review, err := s.reviewRepo.FindByID(targetID)
		if err == nil {
			receiverID = review.UserID
		}
	case "list":
		list, err := s.listRepo.FindByID(targetID)
		if err == nil {
			receiverID = list.UserID
		}
	}

	if receiverID > 0 {
		s.notifService.TriggerNotification(dto.NotificationTask{
			ReceiverID: receiverID,
			SenderID:   senderID,
			ActionType: "like",
			TargetID:   targetID,
			TargetType: targetType,
		})
	}
}

// CheckLike checks if user has already liked the target
func (s *likeService) CheckLike(ctx context.Context, userID, targetID uint, targetType string) (bool, error) {
	if targetType != "game" && targetType != "review" && targetType != "list" && targetType != "comment" {
		return false, dto.NewServiceError("VALIDATION_ERROR", "loại target không hợp lệ")
	}

	return s.likeRepo.CheckLike(userID, targetID, targetType)
}

// invalidateTrendingCache xóa cache trending games
func (s *likeService) invalidateTrendingCache(ctx context.Context) {
	if s.rdb == nil {
		return
	}

	pattern := "trending:games:*"
	iter := s.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		_ = s.rdb.Del(ctx, iter.Val())
	}
}

func (s *likeService) GetGameLikes(ctx context.Context, gameID uint, page, limit int, sort string) (*dto.GameLikesResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	likes, total, err := s.likeRepo.GetLikesWithUser(gameID, "game", page, limit, sort)
	if err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy danh sách like")
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &dto.GameLikesResponse{
		Likes: likes,
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
	}, nil
}
