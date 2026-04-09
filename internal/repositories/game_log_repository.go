package repositories

import (
	"vault/be/internal/models"
	"gorm.io/gorm"
)

type GameLogRepository interface {
	GetByUserID(userID uint, limit int) ([]models.GameLog, error)
}

type gameLogRepository struct {
	db *gorm.DB
}

func NewGameLogRepository(db *gorm.DB) GameLogRepository {
	return &gameLogRepository{db: db}
}

func (r *gameLogRepository) GetByUserID(userID uint, limit int) ([]models.GameLog, error) {
	var logs []models.GameLog
	err := r.db.Preload("Game.Images", "img_type = ?", "header").
		Where("user_id = ?", userID).
		Order("logged_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}