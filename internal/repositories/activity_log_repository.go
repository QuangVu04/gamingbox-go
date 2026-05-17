package repositories

import (
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	GetRecentByUserID(userID uint, limit int) ([]models.ActivityLog, error)
	GetByUserIDPaginated(userID uint, page, limit int, filterType, searchQuery string) ([]models.ActivityLog, int64, error)
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

func (r *activityLogRepository) GetByUserIDPaginated(userID uint, page, limit int, filterType, searchQuery string) ([]models.ActivityLog, int64, error) {
	var activities []models.ActivityLog
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.ActivityLog{})
	if userID != 0 {
		db = db.Where("activity_logs.user_id = ?", userID)
	}

	switch filterType {
	case "Logs":
		db = db.Where("activity_logs.target_type = ?", "game")
	case "Reviews":
		db = db.Where("activity_logs.target_type = ?", "review")
	case "Đánh giá":
		db = db.Where("activity_logs.target_type = ?", "rating")
	case "Thành viên":
		db = db.Where("activity_logs.target_type = ?", "user")
	case "Danh sách":
		db = db.Where("activity_logs.target_type = ?", "list")
	}

	if searchQuery != "" {
		db = db.Joins("LEFT JOIN users ON users.id = activity_logs.user_id").
			Where("users.username LIKE ? OR activity_logs.preview LIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").Order("activity_logs.created_at desc").Offset(offset).Limit(limit).Find(&activities).Error
	return activities, total, err
}
