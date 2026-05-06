package workers

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"vault/be/database"
	"vault/be/internal/dto"
	"vault/be/internal/services"
)

type NotificationWorker struct {
	service services.NotificationService
}

func NewNotificationWorker(service services.NotificationService) *NotificationWorker {
	return &NotificationWorker{service: service}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	log.Println("Notification Worker started...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Notification Worker stopping...")
			return
		default:
			// BRPop (Blocking Right Pop) để lấy task từ Redis
			// Timeout 5 giây để không treo goroutine quá lâu khi shutdown
			result, err := database.RDB.BRPop(ctx, 5*time.Second, services.NotificationQueueKey).Result()
			
			if err != nil {
				// Nếu timeout thì continue
				if err.Error() == "redis: nil" {
					continue
				}
				log.Printf("Worker error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// result[0] là key, result[1] là value
			var task dto.NotificationTask
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				log.Printf("Error unmarshaling task: %v", err)
				continue
			}

			if err := w.service.ProcessNotification(task); err != nil {
				log.Printf("Error processing notification: %v", err)
			}
		}
	}
}
