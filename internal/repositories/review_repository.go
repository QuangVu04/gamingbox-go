package repositories

import (
	"time"
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	GetRecentByUserID(userID uint, limit int) ([]models.Review, error)
	GetPopularByUserID(userID uint, limit int) ([]models.Review, error)
	GetCommentCounts(reviewIDs []uint) (map[uint]int, error)
	GetTrendingReviews(page, limit int) ([]models.Review, int64, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) GetRecentByUserID(userID uint, limit int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("Game.Images", "img_type = ?", "header").
		Where("user_id = ? AND target_type = ?", userID, "game").
		Order("created_at desc").Limit(limit).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetPopularByUserID(userID uint, limit int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("Game.Images", "img_type = ?", "header").
		Where("user_id = ? AND target_type = ?", userID, "game").
		Order("like_count desc").Limit(limit * 3).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetCommentCounts(reviewIDs []uint) (map[uint]int, error) {
	counts := make(map[uint]int)
	if len(reviewIDs) == 0 {
		return counts, nil
	}

	type commentCount struct {
		ReviewID uint
		Count    int
	}
	var rows []commentCount
	err := r.db.Model(&models.Comment{}).
		Select("review_id, COUNT(*) AS count").
		Where("review_id IN ?", reviewIDs).
		Group("review_id").Scan(&rows).Error

	for _, row := range rows {
		counts[row.ReviewID] = row.Count
	}
	return counts, err
}

// GetTrendingReviews retrieves trending reviews from the last 7 days, sorted by like_count
func (r *reviewRepository) GetTrendingReviews(page, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	err := r.db.Model(&models.Review{}).
		Where("target_type = ? AND created_at > ?", "game", time.Now().AddDate(0, 0, -7)).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Get trending reviews with preloads
	err = r.db.
		Preload("Game.Images", "img_type = ?", "header").
		Preload("Game.Studio").
		Preload("User").
		Where("target_type = ? AND created_at > ?", "game", time.Now().AddDate(0, 0, -7)).
		Order("like_count DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	return reviews, total, err
}
