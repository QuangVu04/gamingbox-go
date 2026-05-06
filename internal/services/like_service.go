package services

import (
	"context"

	"vault/be/internal/dto"
	"vault/be/internal/repositories"

	"github.com/redis/go-redis/v9"
)

type LikeService interface {
	// ToggleLike handles polymorphic like/unlike with atomic counter update
	ToggleLike(ctx context.Context, userID, targetID uint, targetType string) (*dto.LikeGameResponse, error)
	// CheckLike checks if user has liked a target
	CheckLike(ctx context.Context, userID, targetID uint, targetType string) (bool, error)
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

// ToggleLike performs like/unlike with atomic counter update
// Supports game, review, list target types
func (s *likeService) ToggleLike(ctx context.Context, userID, targetID uint, targetType string) (*dto.LikeGameResponse, error) {
	// Validate targetType
	if targetType != "game" && targetType != "review" && targetType != "list" {
		return nil, dto.NewServiceError("VALIDATION_ERROR", "loại target không hợp lệ")
	}

	// Use ToggleLike which handles atomic counter update in transaction
	isLiked, err := s.likeRepo.ToggleLike(userID, targetID, targetType)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể xử lý like")
	}

	// Invalidate trending cache if it's a game
	if targetType == "game" && s.rdb != nil {
		s.invalidateTrendingCache(ctx)
	}

	// Trigger notification if liked
	if isLiked {
		go s.handleLikeNotification(userID, targetID, targetType)
	}

	return &dto.LikeGameResponse{IsLiked: isLiked}, nil
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
	if targetType != "game" && targetType != "review" && targetType != "list" {
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
