package seeders

import (
	"fmt"
	"log"
	"time"
	"vault/be/internal/models"
	"gorm.io/gorm"
)

func SeedTodayActivity(db *gorm.DB) {
	log.Println("Bắt đầu Seed dữ liệu cho ngày hôm nay (15/5)...")

	var users []models.User
	db.Limit(10).Find(&users)

	var games []models.Game
	db.Limit(10).Find(&games)

	if len(users) == 0 || len(games) == 0 {
		log.Println("Không tìm thấy User hoặc Game để seed activity!")
		return
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for hour := 0; hour < 24; hour++ {
		// Tạo thời điểm cụ thể trong giờ đó
		seedTime := today.Add(time.Duration(hour) * time.Hour).Add(15 * time.Minute)
		
		// Mỗi giờ tạo 1-3 bản ghi ngẫu nhiên để thấy biểu đồ trồi sụt
		count := (hour % 3) + 1 
		
		for i := 0; i < count; i++ {
			user := users[i%len(users)]
			game := games[(hour+i)%len(games)]

			// 1. Seed GameLog (Lượt Logs)
			logEntry := models.GameLog{
				UserID:   user.ID,
				GameID:   game.ID,
				LoggedAt: seedTime,
				Status:   "playing",
			}
			db.Save(&logEntry)

			// 2. Seed Rating
			rating := models.Rating{
				UserID:    user.ID,
				GameID:    game.ID,
				Rating:    float64(4 + (i % 2)),
				CreatedAt: seedTime,
			}
			// Xóa nếu đã tồn tại để tránh lỗi Unique Index
			db.Where("user_id = ? AND game_id = ?", user.ID, game.ID).Delete(&models.Rating{})
			db.Create(&rating)

			// 3. Seed Review (mỗi 2 giờ tạo 1 cái)
			if hour%2 == 0 && i == 0 {
				review := models.Review{
					UserID:     user.ID,
					TargetID:   game.ID,
					TargetType: "game",
					Title:      fmt.Sprintf("Review seeded at %02d:00", hour),
					Content:    "Đây là dữ liệu seed để kiểm tra biểu đồ theo giờ.",
					Recommend:  "recommend",
				}
				review.CreatedAt = seedTime
				db.Create(&review)
			}
		}
	}

	log.Println("Đã Seed xong dữ liệu cho 24 giờ ngày hôm nay!")
}
