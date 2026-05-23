package seeders

import (
	"log"
	"math/rand"

	"vault/be/internal/models"

	"github.com/go-faker/faker/v4"
	"gorm.io/gorm"
)

func SeedRandomLists(db *gorm.DB) {
	log.Println("Đang tạo seed lists...")

	var users []models.User
	db.Find(&users)

	var games []models.Game
	db.Find(&games)

	if len(users) == 0 || len(games) == 0 {
		log.Println("Không có users hoặc games để tạo lists, bỏ qua...")
		return
	}

	var existingLists int64
	db.Model(&models.List{}).Count(&existingLists)
	if existingLists > 0 {
		log.Println("Đã có lists trong database, bỏ qua seed lists...")
		return
	}

	listThemes := []string{
		"Top 10 Game Đáng Chơi Nhất Năm Nay",
		"Danh sách Game Thế Giới Mở Yêu Thích Của Tôi",
		"Những Tựa Game Hành Động Nhịp Độ Nhanh",
		"Các Game RPG Có Cốt Truyện Xuất Sắc",
		"Game Để Cày Cuốc Dịp Cuối Tuần",
		"Top Game Khó Chơi Nhất Tôi Từng Trải Nghiệm",
		"Tuyển Tập Game Đồ Họa Đỉnh Cao",
		"Các Game Co-op Chơi Cùng Hội Bạn",
		"Must Play Cho Tín Đồ Game Bắn Súng",
		"Những Tựa Game Giải Đố Hại Não",
		"Game Thể Thao Không Thể Bỏ Qua",
		"Những Game Indie Gây Bất Ngờ",
		"Top Game Soulslike Yêu Thích",
		"Khám Phá Vũ Trụ Qua Các Game Viễn Tưởng",
		"Danh Sách Game Tuổi Thơ Đáng Nhớ",
	}

	// Xáo trộn themes
	rand.Shuffle(len(listThemes), func(i, j int) {
		listThemes[i], listThemes[j] = listThemes[j], listThemes[i]
	})

	for i := 0; i < 15; i++ {
		// Chọn random 1 user làm tác giả
		user := users[rand.Intn(len(users))]

		// Chọn ngẫu nhiên 3 đến 8 game cho list
		numGames := rand.Intn(6) + 3
		
		// Trộn games để lấy ngẫu nhiên
		rand.Shuffle(len(games), func(i, j int) {
			games[i], games[j] = games[j], games[i]
		})

		selectedGames := games[:numGames]

		var entries []models.ListEntry
		for pos, game := range selectedGames {
			entries = append(entries, models.ListEntry{
				GameID:   game.ID,
				Position: pos + 1,
				GhiChu:   faker.Sentence(),
			})
		}

		thumbnail := ""
		if len(selectedGames) > 0 {
			var firstImg models.GameImg
			db.Where("game_id = ? AND img_type = ?", selectedGames[0].ID, "header").First(&firstImg)
			if firstImg.Thumb != "" {
				thumbnail = firstImg.Thumb
			}
		}

		list := models.List{
			UserID:       user.ID,
			Title:        listThemes[i%len(listThemes)],
			Description:  faker.Paragraph(),
			ThumbnailImg: thumbnail,
			IsPublic:     true,
			GameCount:    len(selectedGames),
			LikeCount:    rand.Intn(200),
			Entries:      entries,
		}

		db.Create(&list)
	}

	log.Println("Đã tạo xong seed lists!")
}
