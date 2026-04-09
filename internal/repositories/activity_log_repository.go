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
	err := r.db.Where("user_id = ?", userID).
		Order("created_at desc").Limit(limit).Find(&activities).Error
	return activities, err
}
