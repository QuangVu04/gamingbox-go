package services

import (
	"context"
	"encoding/json"
	"vault/be/database"
	"vault/be/internal/dto"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
)

const NotificationQueueKey = "queue:notifications"

type NotificationService interface {
	// Async trigger
	TriggerNotification(task dto.NotificationTask) error
	
	// For Worker
	ProcessNotification(task dto.NotificationTask) error
	
	// For User
	GetNotifications(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.NotificationResponse], error)
	MarkAsRead(notificationID uint64, userID uint) error
	MarkAllAsRead(userID uint) error
}

type notificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) TriggerNotification(task dto.NotificationTask) error {
	// Push to Redis Queue
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	
	ctx := context.Background()
	return database.RDB.LPush(ctx, NotificationQueueKey, data).Err()
}

func (s *notificationService) ProcessNotification(task dto.NotificationTask) error {
	// Tránh tự gửi thông báo cho chính mình (ví dụ tự like bài mình)
	if task.SenderID == task.ReceiverID {
		return nil
	}

	notification := &models.Notification{
		ReceiverID: task.ReceiverID,
		SenderID:   task.SenderID,
		ActionType: task.ActionType,
		TargetID:   task.TargetID,
		TargetType: task.TargetType,
	}
	
	return s.repo.Create(notification)
}

func (s *notificationService) GetNotifications(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.NotificationResponse], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	notifications, total, err := s.repo.GetByUserID(userID, offset, limit)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi khi lấy thông báo")
	}

	var data []dto.NotificationResponse
	for _, n := range notifications {
		data = append(data, dto.NotificationResponse{
			ID:         n.ID,
			SenderID:   n.SenderID,
			ActionType: n.ActionType,
			TargetID:   n.TargetID,
			TargetType: n.TargetType,
			IsRead:     n.IsRead,
			CreatedAt:  n.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	if data == nil {
		data = []dto.NotificationResponse{}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.PaginatedResponse[[]dto.NotificationResponse]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *notificationService) MarkAsRead(notificationID uint64, userID uint) error {
	return s.repo.MarkAsRead(notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.repo.MarkAllAsRead(userID)
}
