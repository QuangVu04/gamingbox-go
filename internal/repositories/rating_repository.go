package repositories

import (
	"vault/be/internal/models"

	"gorm.io/gorm"
)


type RatingRepository interface {
	GetAverageRatingByUserID(userID uint) (float64, error)
	// UpsertRating saves or updates a rating for a user on a game
	UpsertRating(rating *models.Rating) error
	// FindByUserAndGame retrieves a specific rating if exists
	FindByUserAndGame(userID, gameID uint) (*models.Rating, error)
	// GetGameStats returns avg rating and total count for a game
	GetGameStats(gameID uint) (avgRating float64, totalRatings int, err error)
	// UpdateGameRating updates the Game's average rating and review count
	UpdateGameRating(gameID uint, avgRating float64, totalRatings int) error
}

type ratingRepository struct {
	db *gorm.DB
}

func NewRatingRepository(db *gorm.DB) RatingRepository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) GetAverageRatingByUserID(userID uint) (float64, error) {
	var average float64
	err := r.db.Model(&models.Rating{}).
		Where("user_id = ?", userID).
		Select("COALESCE(AVG(rating), 0)").Scan(&average).Error
	return average, err
}

func (r *ratingRepository) UpsertRating(rating *models.Rating) error {
    // Kiểm tra rating đã tồn tại chưa
    var existing models.Rating
    err := r.db.Where("user_id = ? AND game_id = ?", rating.UserID, rating.GameID).
        First(&existing).Error
    
    if err == gorm.ErrRecordNotFound {
        // Tạo bản ghi mới
        return r.db.Create(rating).Error
    }
    
    if err != nil {
        return err
    }
    
    // Cập nhật bản ghi hiện tại
    return r.db.Model(&existing).Updates(rating).Error
}

func (r *ratingRepository) FindByUserAndGame(userID, gameID uint) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.Where("user_id = ? AND game_id = ?", userID, gameID).First(&rating).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}

func (r *ratingRepository) GetGameStats(gameID uint) (float64, int, error) {
	var avgRating float64
	var totalRatings int

	err := r.db.Model(&models.Rating{}).
		Where("game_id = ?", gameID).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as total_ratings").
		Row().Scan(&avgRating, &totalRatings)

	if err != nil {
		return 0, 0, err
	}

	return avgRating, totalRatings, nil
}

func (r *ratingRepository) UpdateGameRating(gameID uint, avgRating float64, totalRatings int) error {
	return r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(map[string]interface{}{
			"avg_rating":   avgRating,
			"review_count": totalRatings,
		}).Error
}