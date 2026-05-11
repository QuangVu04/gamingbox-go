package repositories

import (
	"vault/be/internal/models"
	"gorm.io/gorm"
)

type GameLogRepository interface {
	GetByUserID(userID uint, limit int) ([]models.GameLog, error)
	LogGame(log *models.GameLog) error
	RemoveLog(userID, gameID uint) error
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

func (r *gameLogRepository) LogGame(log *models.GameLog) error {
	return r.db.Save(log).Error
}

func (r *gameLogRepository) RemoveLog(userID, gameID uint) error {
	return r.db.Where("user_id = ? AND game_id = ?", userID, gameID).Delete(&models.GameLog{}).Error
}