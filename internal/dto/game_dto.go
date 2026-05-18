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

type GameSummary struct {
	ID          uint      `json:"id"`
	SteamID     int       `json:"steam_id"`
	Title       string    `json:"title"`
	Poster      string    `json:"poster"`
	ReleaseDate time.Time `json:"release_date"`
	Price       float64   `json:"price"`
	IsFree      bool      `json:"is_free"`
	AvgRating   float64   `json:"avg_rating"`
	ReviewCount int       `json:"review_count"`
}

type GenreDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PlatformDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type StudioDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type GameDetailResponse struct {
	ID              uint                     `json:"id"`
	SteamID         int                      `json:"steam_id"`
	Title           string                   `json:"title"`
	Description     string                   `json:"description"`
	ReleaseDate     time.Time                `json:"release_date"`
	Price           float64                  `json:"price"`
	IsFree          bool                     `json:"is_free"`
	AvgRating       float64                  `json:"avg_rating"`
	ReviewCount     int                      `json:"review_count"`
	LikeCount       int                      `json:"like_count"`
	AveragePlaytime float64                  `json:"average_playtime"`
	Studio          *StudioDTO               `json:"studio"`
	Genres          []GenreDTO               `json:"genres"`
	Platforms       []PlatformDTO            `json:"platforms"`
	Images          []string                 `json:"images"`
	PopularReviews  []ReviewCompactResponse `json:"popular_reviews"`
	RecentReviews   []ReviewCompactResponse `json:"recent_reviews"`
	MoreFromStudio  []GameTrendingResponse  `json:"more_from_studio"`
	SimilarGames    []GameTrendingResponse  `json:"similar_games"`
}

type PaginationDTO struct {
	TotalRecords int `json:"total_records"`
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	Limit        int `json:"limit"`
}

type CreateGameRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	ReleaseDate string   `json:"releaseDate"`
	Studio      string   `json:"studio"`
	Rating      string   `json:"rating"`
	Platforms   []string `json:"platforms"`
	Genres      []string `json:"genres"`
	Images      struct {
		Header string `json:"header"` // Banner chính
		Main   string `json:"main"`   // Ảnh bìa cover
	} `json:"images"`
	Screenshots []string `json:"screenshots"`
}
