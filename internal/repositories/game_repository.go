package repositories

import (
	"time"
	"vault/be/internal/models"

	"gorm.io/gorm"
)

type GameRepository interface {
	GetTrendingGames(page, limit int) ([]models.GameTrending, int64, error)
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
