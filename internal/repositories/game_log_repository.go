package repositories

import (
	"time"
	"vault/be/internal/models"
	"gorm.io/gorm"
)

type GameLogRepository interface {
	GetByUserID(userID uint, limit int) ([]models.GameLog, error)
	GetBacklogByUserID(userID uint, limit int) ([]models.GameLog, error)
	GetByUserIDPaginated(userID uint, page, limit int) ([]models.GameLog, int64, error)
	GetBacklogByUserIDPaginated(userID uint, page, limit int) ([]models.GameLog, int64, error)
	LogGame(log *models.GameLog) error
	RemoveLog(userID, gameID uint) error
	Count() (int64, error)
	CountRecent(since time.Time) (int64, error)
	GetDailyCounts(since time.Time) (map[string]int, error)
	GetHourlyCounts(since time.Time) (map[string]int, error)
}

type gameLogRepository struct {
	db *gorm.DB
}

func NewGameLogRepository(db *gorm.DB) GameLogRepository {
	return &gameLogRepository{db: db}
}

func (r *gameLogRepository) GetByUserID(userID uint, limit int) ([]models.GameLog, error) {
	var logs []models.GameLog
	err := r.db.Preload("Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND status != ?", userID, "backlog").
		Order("logged_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gameLogRepository) GetBacklogByUserID(userID uint, limit int) ([]models.GameLog, error) {
	var logs []models.GameLog
	err := r.db.Preload("Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND status = ?", userID, "backlog").
		Order("logged_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gameLogRepository) GetByUserIDPaginated(userID uint, page, limit int) ([]models.GameLog, int64, error) {
	var logs []models.GameLog
	var total int64
	offset := (page - 1) * limit
	db := r.db.Model(&models.GameLog{}).Where("user_id = ? AND status != ?", userID, "backlog")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND status != ?", userID, "backlog").
		Order("logged_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gameLogRepository) GetBacklogByUserIDPaginated(userID uint, page, limit int) ([]models.GameLog, int64, error) {
	var logs []models.GameLog
	var total int64
	offset := (page - 1) * limit
	db := r.db.Model(&models.GameLog{}).Where("user_id = ? AND status = ?", userID, "backlog")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND status = ?", userID, "backlog").
		Order("logged_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gameLogRepository) LogGame(log *models.GameLog) error {
	return r.db.Save(log).Error
}

func (r *gameLogRepository) RemoveLog(userID, gameID uint) error {
	return r.db.Where("user_id = ? AND game_id = ?", userID, gameID).Delete(&models.GameLog{}).Error
}

func (r *gameLogRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.GameLog{}).Count(&count).Error
	return count, err
}

func (r *gameLogRepository) CountRecent(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.GameLog{}).Where("logged_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *gameLogRepository) GetDailyCounts(since time.Time) (map[string]int, error) {
	var results []struct {
		Date  string
		Count int
	}

	err := r.db.Model(&models.GameLog{}).
		Select("DATE_FORMAT(logged_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("logged_at >= ?", since).
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

func (r *gameLogRepository) GetHourlyCounts(since time.Time) (map[string]int, error) {
	var results []struct {
		Hour  string
		Count int
	}

	err := r.db.Model(&models.GameLog{}).
		Select("DATE_FORMAT(logged_at, '%H') as hour, COUNT(*) as count").
		Where("logged_at >= ?", since).
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