package dto

import "time"

type GameTrendingResponse struct {
	GameID        uint      `json:"game_id"`
	Title         string    `json:"title"`
	Thumbnail     string    `json:"thumbnail"`
	TrendingScore int       `json:"trending_score"`
	AvgRating     float64   `json:"avg_rating"`
	TotalReviews  int       `json:"total_reviews"`
	ReleaseDate   time.Time `json:"release_date"`
	Studios       []string  `json:"studios"`
}

type PaginationDTO struct {
	TotalRecords int `json:"total_records"`
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	Limit        int `json:"limit"`
}
