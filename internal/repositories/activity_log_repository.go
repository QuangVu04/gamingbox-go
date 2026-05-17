package repositories

import (
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	GetRecentByUserID(userID uint, limit int) ([]models.ActivityLog, error)
	GetByUserIDPaginated(userID uint, page, limit int) ([]models.ActivityLog, int64, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) GetRecentByUserID(userID uint, limit int) ([]models.ActivityLog, error) {
	var activities []models.ActivityLog
	db := r.db.Preload("User").Order("created_at desc").Limit(limit)
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}
	err := db.Find(&activities).Error
	return activities, err
}

func (r *activityLogRepository) GetByUserIDPaginated(userID uint, page, limit int) ([]models.ActivityLog, int64, error) {
	var activities []models.ActivityLog
	var total int64
	offset := (page - 1) * limit
	db := r.db.Model(&models.ActivityLog{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(limit).Find(&activities).Error
	return activities, total, err
}
