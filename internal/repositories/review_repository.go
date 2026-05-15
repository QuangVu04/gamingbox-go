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
	FindByID(id uint) (*models.Review, error)
	Create(review *models.Review) error
	Update(review *models.Review) error
	Delete(id uint) error
	GetComments(reviewID uint) ([]models.Comment, error)
	AddComment(comment *models.Comment) error
	GetByGameID(gameID uint, orderBy string, limit int) ([]models.Review, error)
	GetGameReviews(gameID uint, page, limit int, sort string) ([]models.Review, int64, error)
	Count() (int64, error)
	CountRecent(since time.Time) (int64, error)
	GetDailyCounts(since time.Time) (map[string]int, error)
	GetHourlyCounts(since time.Time) (map[string]int, error)
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

func (r *reviewRepository) FindByID(id uint) (*models.Review, error) {
	var review models.Review
	err := r.db.First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&models.Review{}, id).Error
}

func (r *reviewRepository) GetComments(reviewID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("review_id = ?", reviewID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *reviewRepository) AddComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *reviewRepository) GetByGameID(gameID uint, orderBy string, limit int) ([]models.Review, error) {
	var reviews []models.Review
	db := r.db.Preload("User").
		Where("target_id = ? AND target_type = ?", gameID, "game")

	if orderBy == "popular" {
		db = db.Order("like_count DESC")
	} else {
		db = db.Order("created_at DESC")
	}

	err := db.Limit(limit).Find(&reviews).Error
	return reviews, err
}
func (r *reviewRepository) GetGameReviews(gameID uint, page, limit int, sort string) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64
	offset := (page - 1) * limit

	db := r.db.Table("reviews").
		Select("reviews.*, ra.rating").
		Joins("LEFT JOIN ratings ra ON reviews.user_id = ra.user_id AND reviews.target_id = ra.game_id").
		Where("reviews.target_id = ? AND reviews.target_type = ?", gameID, "game")

	// Get total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort logic
	switch sort {
	case "earliest":
		db = db.Order("reviews.created_at ASC")
	case "highest_rating":
		db = db.Order("ra.rating DESC, reviews.created_at DESC")
	case "lowest_rating":
		db = db.Order("ra.rating ASC, reviews.created_at DESC")
	default: // newest
		db = db.Order("reviews.created_at DESC")
	}

	err := db.Preload("User").Offset(offset).Limit(limit).Find(&reviews).Error
	return reviews, total, err
}

func (r *reviewRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Review{}).Count(&count).Error
	return count, err
}

func (r *reviewRepository) CountRecent(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Review{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *reviewRepository) GetDailyCounts(since time.Time) (map[string]int, error) {
	var results []struct {
		Date  string
		Count int
	}

	err := r.db.Model(&models.Review{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("date").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, res := range results {
		counts[res.Date] = res.Count
	}
	return counts, nil
}

func (r *reviewRepository) GetHourlyCounts(since time.Time) (map[string]int, error) {
	var results []struct {
		Hour  string
		Count int
	}

	err := r.db.Model(&models.Review{}).
		Select("DATE_FORMAT(created_at, '%H') as hour, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("hour").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, res := range results {
		counts[res.Hour] = res.Count
	}
	return counts, nil
}

