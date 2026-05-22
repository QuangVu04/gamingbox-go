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

type InteractionWorker struct {
	likeService services.LikeService
	gameService services.GameService
}

func NewInteractionWorker(likeService services.LikeService, gameService services.GameService) *InteractionWorker {
	return &InteractionWorker{
		likeService: likeService,
		gameService: gameService,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) {
	log.Println("Interaction Worker started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Interaction Worker stopping...")
			return
		default:
			// BRPop (Blocking Right Pop) to pop from Redis
			// Timeout 5 seconds to not block shutdown loop
			result, err := database.RDB.BRPop(ctx, 5*time.Second, services.InteractionQueueKey).Result()

			if err != nil {
				if err.Error() == "redis: nil" {
					continue
				}
				log.Printf("Interaction Worker error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// result[0] is the key, result[1] is the value JSON
			var task dto.InteractionTask
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				log.Printf("Error unmarshaling interaction task: %v", err)
				continue
			}

			log.Printf("[InteractionWorker] Processing task type=%s target=%d user=%d", task.Type, task.TargetID, task.UserID)

			var procErr error
			switch task.Type {
			case "like":
				_, procErr = w.likeService.ToggleLikeDB(ctx, task.UserID, task.TargetID, task.TargetType)
			case "log":
				procErr = w.gameService.LogGameStatusDB(ctx, task.UserID, task.TargetID, task.Status)
			case "rate":
				_, procErr = w.gameService.RateGameDB(ctx, task.UserID, task.TargetID, task.Rating)
			default:
				log.Printf("Unknown interaction task type: %s", task.Type)
			}

			if procErr != nil {
				log.Printf("Error processing interaction task: %v", procErr)
			}
		}
	}
}
