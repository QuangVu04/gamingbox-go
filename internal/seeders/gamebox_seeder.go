package seeders

import (
	"log"
	"math/rand"
	"time"
	"vault/be/internal/models"

	"gorm.io/gorm"
)

func SeedGameboxData(db *gorm.DB) {
	log.Println("Bắt đầu chạy Gamebox Seeder...")

	// 1. Lấy danh sách users và games đã có
	var users []models.User
	db.Find(&users)
	if len(users) == 0 {
		log.Println("Không tìm thấy user nào, vui lòng chạy seed random users trước.")
		return
	}

	var games []models.Game
	db.Find(&games)
	if len(games) == 0 {
		log.Println("Không tìm thấy game nào, vui lòng chạy seed games từ Steam trước.")
		return
	}

	// 2. Seed Ratings & Reviews
	log.Println("Đang tạo Ratings và Reviews...")
	recommends := []string{"recommend", "mixed", "not_recommend"}
	for i := 0; i < 100; i++ {
		user := users[rand.Intn(len(users))]
		game := games[rand.Intn(len(games))]

		// Create Rating
		rating := models.Rating{
			UserID:    user.ID,
			GameID:    game.ID,
			Rating:    float64(rand.Intn(9)+1) * 0.5, // 0.5 to 5.0
			CreatedAt: time.Now().AddDate(0, 0, -rand.Intn(30)),
		}
		db.FirstOrCreate(&rating, models.Rating{UserID: user.ID, GameID: game.ID})

		// Create Review
		review := models.Review{
			UserID:     user.ID,
			TargetID:   game.ID,
			TargetType: "game",
			Title:      "Review cho " + game.Title,
			Content:    "Đây là một bài review giả lập cho game này. Cảm nhận chung là khá ổn.",
			Recommend:  recommends[rand.Intn(len(recommends))],
			LikeCount:  rand.Intn(50),
			IsSpoiler:  rand.Intn(4) == 0,
		}
		review.CreatedAt = rating.CreatedAt
		db.Create(&review)

		// Create some comments
		for j := 0; j < rand.Intn(5); j++ {
			commenter := users[rand.Intn(len(users))]
			db.Create(&models.Comment{
				ReviewID: review.ID,
				UserID:   commenter.ID,
				Content:  "Bình luận thứ " + string(rune(j+1)) + " của " + commenter.Username,
			})
		}
	}

	// 3. Seed Lists
	log.Println("Đang tạo Lists...")
	listTitles := []string{"Top RPG", "Nên chơi thử", "Game tuổi thơ", "Đồ họa đỉnh", "Cày cuốc"}
	for i := 0; i < 20; i++ {
		user := users[rand.Intn(len(users))]
		list := models.List{
			UserID:      user.ID,
			Title:       listTitles[rand.Intn(len(listTitles))] + " #" + string(rune(i+1)),
			Description: "Danh sách này tổng hợp các game hay nhất mà tôi từng chơi.",
			IsPublic:    true,
			LikeCount:   rand.Intn(30),
		}
		db.Create(&list)

		// Add games to list
		for j := 0; j < rand.Intn(10)+5; j++ {
			game := games[rand.Intn(len(games))]
			db.FirstOrCreate(&models.ListEntry{}, models.ListEntry{
				ListID:   list.ID,
				GameID:   game.ID,
				Position: j + 1,
				GhiChu:   "Game cực hay",
			})
		}
		db.Model(&list).Update("game_count", db.Model(&models.ListEntry{}).Where("list_id = ?", list.ID).Count)
	}

	// 4. Seed Game Logs (Diary)
	log.Println("Đang tạo Game Logs (Diary)...")
	statuses := []string{"playing", "completed", "dropped"}
	for i := 0; i < 50; i++ {
		user := users[rand.Intn(len(users))]
		game := games[rand.Intn(len(games))]
		db.FirstOrCreate(&models.GameLog{}, models.GameLog{
			UserID:   user.ID,
			GameID:   game.ID,
			Status:   statuses[rand.Intn(len(statuses))],
			LoggedAt: time.Now().AddDate(0, 0, -rand.Intn(60)),
		})
	}

	log.Println("Gamebox Seeder đã hoàn tất!")
}
