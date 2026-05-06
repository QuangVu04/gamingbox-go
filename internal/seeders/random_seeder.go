package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"vault/be/internal/models"
	"vault/be/pkg/utils"

	"github.com/go-faker/faker/v4"
	"gorm.io/gorm"
)

// Một danh sách các Game nổi tiếng trên Steam (App ID)
var steamAppIDs = []int{
	730, 570, 578080, 271590, 1091500, 1245620, 1086940, 
	1172470, 413150, 252490, 105600, 292030, 359550, 289070,
}

type SteamAppDetails struct {
	Success bool `json:"success"`
	Data    struct {
		Name             string `json:"name"`
		ShortDescription string `json:"short_description"`
		IsFree           bool   `json:"is_free"`
		PriceOverview    struct {
			Initial int `json:"initial"` // Tính bằng cent
		} `json:"price_overview"`
		ReleaseDate struct {
			Date string `json:"date"`
		} `json:"release_date"`
		HeaderImage string `json:"header_image"`
		Developers  []string `json:"developers"`
		Genres      []struct {
			Description string `json:"description"`
		} `json:"genres"`
		Platforms struct {
			Windows bool `json:"windows"`
			Mac     bool `json:"mac"`
			Linux   bool `json:"linux"`
		} `json:"platforms"`
		Screenshots []struct {
			PathThumbnail string `json:"path_thumbnail"`
			PathFull      string `json:"path_full"`
		} `json:"screenshots"`
	} `json:"data"`
}

func SeedRandomData(db *gorm.DB) {
	log.Println("Bắt đầu chạy Random Seeder...")

	seedGamesFromSteam(db)
	seedRandomUsers(db, 50) // Tạo ngẫu nhiên 50 người dùng
	
	log.Println("Random Seeder đã hoàn tất!")
}

func seedGamesFromSteam(db *gorm.DB) {
	var count int64
	db.Model(&models.Game{}).Count(&count)
	if count > 0 {
		log.Println("Games đã có sẵn trong database, bỏ qua việc cào dữ liệu từ Steam...")
		return
	}

	log.Println("Đang lấy dữ liệu game thật từ Steam API...")
	for _, appID := range steamAppIDs {
		url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d", appID)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Lỗi khi fetch game %d: %v", appID, err)
			continue
		}

		var result map[string]SteamAppDetails
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Lỗi decode json %d: %v", appID, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		appIDStr := strconv.Itoa(appID)
		appData, ok := result[appIDStr]
		if !ok || !appData.Success {
			continue
		}

		data := appData.Data

		// Xử lý Studio (Lấy Studio đầu tiên)
		studioName := "Unknown Studio"
		if len(data.Developers) > 0 {
			studioName = data.Developers[0]
		}
		var studio models.Studio
		db.FirstOrCreate(&studio, models.Studio{Name: studioName})

		// Xử lý Ngày phát hành
		parsedDate, _ := time.Parse("2 Jan, 2006", data.ReleaseDate.Date)

		game := models.Game{
			SteamID:     appID,
			Title:       data.Name,
			Description: data.ShortDescription,
			IsFree:      data.IsFree,
			Price:       float64(data.PriceOverview.Initial) / 100.0,
			ReleaseDate: parsedDate,
			StudioID:    studio.ID,
		}
		db.Create(&game)

		// Thể loại (Genres)
		for _, g := range data.Genres {
			var genre models.Genre
			db.FirstOrCreate(&genre, models.Genre{Name: g.Description})
			db.Model(&game).Association("Genres").Append(&genre)
		}

		// Nền tảng (Platforms)
		if data.Platforms.Windows {
			p := models.Platform{}
			db.FirstOrCreate(&p, models.Platform{Name: "Windows"})
			db.Model(&game).Association("Platforms").Append(&p)
		}
		if data.Platforms.Mac {
			p := models.Platform{}
			db.FirstOrCreate(&p, models.Platform{Name: "Mac"})
			db.Model(&game).Association("Platforms").Append(&p)
		}

		// Hình ảnh (Images)
		headerImg := models.GameImg{
			OgURL:   data.HeaderImage,
			Thumb:   data.HeaderImage,
			ImgType: "header",
			GameID:  game.ID,
		}
		db.Create(&headerImg)
		
		for _, s := range data.Screenshots {
			screenshot := models.GameImg{
				OgURL:   s.PathFull,
				Thumb:   s.PathThumbnail,
				ImgType: "screenshot",
				GameID:  game.ID,
			}
			db.Create(&screenshot)
		}

		log.Printf("Đã cào xong game: %s", data.Name)
		// Ngủ 1.5 giây để tránh bị Steam block IP vì Spam Request
		time.Sleep(1500 * time.Millisecond)
	}
}

func seedRandomUsers(db *gorm.DB, count int) {
	log.Printf("Đang tạo %d tài khoản người dùng giả lập...", count)
	
	// Tất cả user fake sẽ có chung mật khẩu là 123456
	password, _ := utils.HashPassword("123456")

	for i := 0; i < count; i++ {
		user := models.User{
			Email:    faker.Email(),
			Username: faker.Username(),
			Password: password,
			Role:     models.RoleUser,
		}
		db.Create(&user)
	}
}
