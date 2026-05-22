package repositories

import (
	"fmt"
	"time"
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type GameRepository interface {
	GetTrendingGames(page, limit int) ([]models.GameTrending, int64, error)
	GetByID(id uint) (*models.Game, error)
	Search(query string, page, limit int) ([]models.Game, int64, error)
	GetPopular(page, limit int) ([]models.Game, int64, error)
	GetByStudio(studioID uint, excludeGameID uint, limit int) ([]models.Game, error)
	GetByGenres(genreIDs []uint, excludeGameID uint, limit int) ([]models.Game, error)
	Count() (int64, error)
	CountRecent(since time.Time) (int64, error)
	GetGenreStats(since time.Time) (map[string]int, error)
	SearchAdminGames(search, category, platform, minRating, startDate, endDate, sort string, page, limit int) ([]models.Game, int64, error)
	GetGenres() ([]models.Genre, error)
	GetPlatforms() ([]models.Platform, error)
	CreateGame(game *models.Game) error
	UpdateGame(game *models.Game) error
	DeleteGenreByName(name string) error
	DeletePlatformByName(name string) error
	SearchStudios(query string) ([]models.Studio, error)
	DeleteGame(id uint) error
	GetByIDs(ids []uint) ([]models.Game, error)
}


type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) GetTrendingGames(page, limit int) ([]models.GameTrending, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	offset := (page - 1) * limit
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	// Query to get games with trending score
	var games []struct {
		GameID        uint
		ReviewCount   int
		RatingCount   int
		TrendingScore int
	}

	// Raw SQL query to calculate trending scores
	query := `
	SELECT 
		g.id as game_id,
		COUNT(DISTINCT r.id) as review_count,
		COUNT(DISTINCT rat.id) as rating_count,
		(COUNT(DISTINCT r.id) * 10 + COUNT(DISTINCT rat.id) * 5) as trending_score
	FROM games g
	LEFT JOIN reviews r ON g.id = r.target_id AND r.target_type = 'game' AND r.created_at >= ?
	LEFT JOIN ratings rat ON g.id = rat.game_id AND rat.created_at >= ?
	WHERE r.id IS NOT NULL OR rat.id IS NOT NULL
	GROUP BY g.id
	ORDER BY trending_score DESC
	LIMIT ? OFFSET ?
	`

	if err := r.db.Raw(query, sevenDaysAgo, sevenDaysAgo, limit, offset).Scan(&games).Error; err != nil {
		return nil, 0, err
	}

	// Get total count
	var totalCount int64
	countQuery := `
	SELECT COUNT(DISTINCT g.id) as total
	FROM games g
	LEFT JOIN reviews r ON g.id = r.target_id AND r.target_type = 'game' AND r.created_at >= ?
	LEFT JOIN ratings rat ON g.id = rat.game_id AND rat.created_at >= ?
	WHERE r.id IS NOT NULL OR rat.id IS NOT NULL
	`
	if err := r.db.Raw(countQuery, sevenDaysAgo, sevenDaysAgo).Row().Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Fetch full game data with preloads
	var gameIDs []uint
	for _, g := range games {
		gameIDs = append(gameIDs, g.GameID)
	}

	var fullGames []models.Game
	if len(gameIDs) > 0 {
		if err := r.db.Preload("Studio").
			Preload("Images", "img_type IN ?", []string{"header", "cover"}).
			Preload("Genres").
			Preload("Platforms").
			Where("id IN ?", gameIDs).
			Find(&fullGames).Error; err != nil {
			return nil, 0, err
		}
	}

	// Map results to GameTrending with trending scores
	trendingGames := make([]models.GameTrending, 0, len(games))
	gameMap := make(map[uint]models.Game)
	for _, g := range fullGames {
		gameMap[g.ID] = g
	}

	for _, g := range games {
		if game, ok := gameMap[g.GameID]; ok {
			trendingGames = append(trendingGames, models.GameTrending{
				Game:          game,
				TrendingScore: g.TrendingScore,
				ReviewCount7d: g.ReviewCount,
				RatingCount7d: g.RatingCount,
			})
		}
	}

	return trendingGames, totalCount, nil
}

func (r *gameRepository) GetByID(id uint) (*models.Game, error) {
	var game models.Game
	err := r.db.Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images").
		First(&game, id).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) Search(query string, page, limit int) ([]models.Game, int64, error) {
	var games []models.Game
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.Game{}).
		Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images", "img_type IN ?", []string{"header", "cover"}).
		Where("title LIKE ?", "%"+query+"%")

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(limit).Offset(offset).Find(&games).Error
	if err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *gameRepository) SearchAdminGames(search, category, platform, minRating, startDate, endDate, sort string, page, limit int) ([]models.Game, int64, error) {
	var games []models.Game
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.Game{}).
		Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images", "img_type IN ?", []string{"header", "cover"})

	if search != "" {
		db = db.Joins("LEFT JOIN studios ON studios.id = games.studio_id").
			Where("games.title LIKE ? OR studios.name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if category != "" && category != "All" {
		db = db.Joins("JOIN game_genres ON game_genres.game_id = games.id").
			Joins("JOIN genres ON genres.id = game_genres.genre_id").
			Where("genres.name = ?", category)
	}

	if platform != "" && platform != "All" {
		db = db.Joins("JOIN game_platforms ON game_platforms.game_id = games.id").
			Joins("JOIN platforms ON platforms.id = game_platforms.platform_id").
			Where("platforms.name = ?", platform)
	}

	if minRating != "" && minRating != "All" {
		var minVal float64
		fmt.Sscanf(minRating, "%f", &minVal)
		db = db.Where("games.avg_rating >= ?", minVal)
	}

	if startDate != "" {
		db = db.Where("games.release_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("games.release_date <= ?", endDate)
	}

	err := db.Session(&gorm.Session{}).Distinct("games.id").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	orderClause := "games.release_date DESC"
	if sort == "earliest" {
		orderClause = "games.release_date ASC"
	} else if sort == "highest_rating" {
		orderClause = "games.avg_rating DESC"
	} else if sort == "lowest_rating" {
		orderClause = "games.avg_rating ASC"
	} else if sort == "popular_week" {
		orderClause = "games.review_count DESC"
	} else if sort == "popular_all_time" {
		orderClause = "games.review_count DESC"
	}

	err = db.Select("games.*").Group("games.id").Order(orderClause).Limit(limit).Offset(offset).Find(&games).Error
	if err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *gameRepository) GetPopular(page, limit int) ([]models.Game, int64, error) {
	var games []models.Game
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.Game{}).
		Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images", "img_type IN ?", []string{"header", "cover"})

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Order("avg_rating DESC, review_count DESC").
		Limit(limit).Offset(offset).Find(&games).Error
	if err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *gameRepository) GetByStudio(studioID uint, excludeGameID uint, limit int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Model(&models.Game{}).
		Preload("Studio").
		Preload("Images", "img_type IN ?", []string{"header", "cover"}).
		Where("studio_id = ? AND id != ?", studioID, excludeGameID).
		Limit(limit).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) GetByGenres(genreIDs []uint, excludeGameID uint, limit int) ([]models.Game, error) {
	var games []models.Game
	if len(genreIDs) == 0 {
		return games, nil
	}

	err := r.db.Model(&models.Game{}).
		Joins("JOIN game_genres ON game_genres.game_id = games.id").
		Preload("Studio").
		Preload("Images", "img_type IN ?", []string{"header", "cover"}).
		Where("game_genres.genre_id IN ? AND games.id != ?", genreIDs, excludeGameID).
		Group("games.id").
		Limit(limit).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Game{}).Count(&count).Error
	return count, err
}

func (r *gameRepository) CountRecent(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Game{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *gameRepository) GetGenreStats(since time.Time) (map[string]int, error) {
	var results []struct {
		Name  string
		Count int
	}

	// Đếm số lượt tương tác (reviews + game_logs + ratings) theo từng thể loại kể từ mốc since
	query := `
	SELECT g.name, COUNT(DISTINCT act.act_id) as count
	FROM genres g
	LEFT JOIN game_genres gg ON g.id = gg.genre_id
	LEFT JOIN (
		SELECT CONCAT('rev_', id) as act_id, target_id as game_id FROM reviews WHERE target_type = 'game' AND created_at >= ?
		UNION ALL
		SELECT CONCAT('log_', user_id, '_', game_id) as act_id, game_id FROM game_logs WHERE logged_at >= ?
		UNION ALL
		SELECT CONCAT('rat_', id) as act_id, game_id FROM ratings WHERE created_at >= ?
	) act ON gg.game_id = act.game_id
	GROUP BY g.id
	`

	err := r.db.Raw(query, since, since, since).Scan(&results).Error
	if err != nil {
		fmt.Printf("ERROR GetGenreStats SQL: %v\n", err)
		return nil, err
	}

	stats := make(map[string]int)
	totalInteractions := 0
	for _, res := range results {
		stats[res.Name] = res.Count
		totalInteractions += res.Count
	}

	fmt.Printf("DEBUG: GetGenreStats since: %v | Total Interactions found: %d\n", since, totalInteractions)

	// Fallback thông minh: Nếu trong mốc since chưa có tương tác nào phát sinh (totalInteractions == 0),
	// tự động lấy cơ cấu toàn bộ thư viện game để biểu đồ không bị rỗng!
	if totalInteractions == 0 {
		var fallbackResults []struct {
			Name  string
			Count int
		}
		errFallback := r.db.Table("genres").
			Select("genres.name, COUNT(game_genres.game_id) as count").
			Joins("LEFT JOIN game_genres ON genres.id = game_genres.genre_id").
			Group("genres.id").
			Scan(&fallbackResults).Error

		if errFallback == nil {
			for _, res := range fallbackResults {
				stats[res.Name] = res.Count
			}
		}
	}

	return stats, nil
}

func (r *gameRepository) GetGenres() ([]models.Genre, error) {
	var genres []models.Genre
	err := r.db.Order("name ASC").Find(&genres).Error
	return genres, err
}

func (r *gameRepository) GetPlatforms() ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.Order("name ASC").Find(&platforms).Error
	return platforms, err
}

func (r *gameRepository) CreateGame(game *models.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) UpdateGame(game *models.Game) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update core game fields
		if err := tx.Save(game).Error; err != nil {
			return err
		}
		
		// Replace associations (Genres and Platforms)
		if err := tx.Model(game).Association("Genres").Replace(game.Genres); err != nil {
			return err
		}
		if err := tx.Model(game).Association("Platforms").Replace(game.Platforms); err != nil {
			return err
		}

		// Delete old images
		if err := tx.Where("game_id = ?", game.ID).Delete(&models.GameImg{}).Error; err != nil {
			return err
		}

		// Create new images
		if len(game.Images) > 0 {
			for i := range game.Images {
				game.Images[i].GameID = game.ID
			}
			if err := tx.Create(&game.Images).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gameRepository) DeleteGenreByName(name string) error {
	r.db.Exec("DELETE FROM game_genres WHERE genre_id IN (SELECT id FROM genres WHERE name = ?)", name)
	return r.db.Where("name = ?", name).Delete(&models.Genre{}).Error
}

func (r *gameRepository) DeletePlatformByName(name string) error {
	r.db.Exec("DELETE FROM game_platforms WHERE platform_id IN (SELECT id FROM platforms WHERE name = ?)", name)
	return r.db.Where("name = ?", name).Delete(&models.Platform{}).Error
}

func (r *gameRepository) SearchStudios(query string) ([]models.Studio, error) {
	var studios []models.Studio
	tx := r.db.Order("name ASC").Limit(20)
	if query != "" {
		tx = tx.Where("name LIKE ?", "%"+query+"%")
	}
	err := tx.Find(&studios).Error
	return studios, err
}

func (r *gameRepository) DeleteGame(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find all reviews for this game
		var reviewIDs []uint
		if err := tx.Model(&models.Review{}).Where("target_id = ? AND target_type = ?", id, "game").Pluck("id", &reviewIDs).Error; err != nil {
			return err
		}

		// 2. Delete comments belonging to these reviews
		if len(reviewIDs) > 0 {
			if err := tx.Unscoped().Where("review_id IN ?", reviewIDs).Delete(&models.Comment{}).Error; err != nil {
				return err
			}
			// Delete likes on these reviews
			if err := tx.Where("target_id IN ? AND target_type = ?", reviewIDs, "review").Delete(&models.Like{}).Error; err != nil {
				return err
			}
			// Delete activity logs referencing these reviews
			if err := tx.Where("target_id IN ? AND target_type = ?", reviewIDs, "review").Delete(&models.ActivityLog{}).Error; err != nil {
				return err
			}
			// Delete notifications referencing these reviews
			if err := tx.Where("target_id IN ? AND target_type = ?", reviewIDs, "review").Delete(&models.Notification{}).Error; err != nil {
				return err
			}
			// Delete the reviews themselves
			if err := tx.Unscoped().Where("id IN ?", reviewIDs).Delete(&models.Review{}).Error; err != nil {
				return err
			}
		}

		// 3. Delete ratings of this game
		if err := tx.Where("game_id = ?", id).Delete(&models.Rating{}).Error; err != nil {
			return err
		}

		// 4. Delete game logs of this game
		if err := tx.Where("game_id = ?", id).Delete(&models.GameLog{}).Error; err != nil {
			return err
		}

		// 5. Delete list entries of this game
		if err := tx.Where("game_id = ?", id).Delete(&models.ListEntry{}).Error; err != nil {
			return err
		}

		// 6. Delete likes on this game
		if err := tx.Where("target_id = ? AND target_type = ?", id, "game").Delete(&models.Like{}).Error; err != nil {
			return err
		}

		// 7. Delete activity logs referencing this game
		if err := tx.Where("target_id = ? AND target_type = ?", id, "game").Delete(&models.ActivityLog{}).Error; err != nil {
			return err
		}

		// 8. Delete notifications referencing this game
		if err := tx.Where("target_id = ? AND target_type = ?", id, "game").Delete(&models.Notification{}).Error; err != nil {
			return err
		}

		// 9. Delete game images
		if err := tx.Where("game_id = ?", id).Delete(&models.GameImg{}).Error; err != nil {
			return err
		}

		// 10. Delete game genres association
		if err := tx.Exec("DELETE FROM game_genres WHERE game_id = ?", id).Error; err != nil {
			return err
		}

		// 11. Delete game platforms association
		if err := tx.Exec("DELETE FROM game_platforms WHERE game_id = ?", id).Error; err != nil {
			return err
		}

		// 12. Delete the game itself
		if err := tx.Unscoped().Delete(&models.Game{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *gameRepository) GetByIDs(ids []uint) ([]models.Game, error) {
	var games []models.Game
	if len(ids) == 0 {
		return games, nil
	}
	err := r.db.Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images", "img_type IN ?", []string{"header", "cover"}).
		Where("id IN ?", ids).
		Find(&games).Error
	return games, err
}
