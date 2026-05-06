package repositories

import (
	"vault/be/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUserID(userID uint, offset, limit int) ([]models.Notification, int64, error)
	MarkAsRead(notificationID uint64, userID uint) error
	MarkAllAsRead(userID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByUserID(userID uint, offset, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	err := r.db.Model(&models.Notification{}).Where("receiver_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("receiver_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error

	return notifications, total, err
}

func (r *notificationRepository) MarkAsRead(notificationID uint64, userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND receiver_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
