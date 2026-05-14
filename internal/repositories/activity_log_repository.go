package repositories

import (
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	GetRecentByUserID(userID uint, limit int) ([]models.ActivityLog, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) GetRecentByUserID(userID uint, limit int) ([]models.ActivityLog, error) {
	var activities []models.ActivityLog
	db := r.db.Order("created_at desc").Limit(limit)
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}
	err := db.Find(&activities).Error
	return activities, err
}
