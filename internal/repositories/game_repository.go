package repositories

import (
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
			Preload("Images", "img_type = ?", "header").
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
		Preload("Images", "img_type = ?", "header").
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

func (r *gameRepository) GetPopular(page, limit int) ([]models.Game, int64, error) {
	var games []models.Game
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.Game{}).
		Preload("Studio").
		Preload("Genres").
		Preload("Platforms").
		Preload("Images", "img_type = ?", "header")

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
		Preload("Images", "img_type = ?", "header").
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
		Preload("Images", "img_type = ?", "header").
		Where("game_genres.genre_id IN ? AND games.id != ?", genreIDs, excludeGameID).
		Group("games.id").
		Limit(limit).
		Find(&games).Error
	return games, err
}
