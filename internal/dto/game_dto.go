package dto

import "time"

type GameTrendingResponse struct {
	GameID        uint      `json:"game_id"`
	Title         string    `json:"title"`
	Thumbnail     string    `json:"thumbnail"`
	TrendingScore int       `json:"trending_score"`
	AvgRating     float64   `json:"avg_rating"`
	TotalReviews  int       `json:"total_reviews"`
	LikeCount     int       `json:"like_count"`
	ReleaseDate   time.Time `json:"release_date"`
	HeaderImage   string    `json:"header_image,omitempty"`
	CoverImage    string    `json:"cover_image,omitempty"`
	Studios       []string  `json:"studios"`
	Genres        []string  `json:"genres"`
}

type GameSummary struct {
	ID            uint      `json:"id"`
	SteamID       int       `json:"steam_id"`
	Title         string    `json:"title"`
	Poster        string    `json:"poster"`
	ReleaseDate   time.Time `json:"release_date"`
	Price         float64   `json:"price"`
	IsFree        bool      `json:"is_free"`
	AvgRating     float64   `json:"avg_rating"`
	ReviewCount   int       `json:"review_count"`
	LikeCount     int       `json:"like_count"`
	Studios       []string  `json:"studios,omitempty"`
	Genres        []string  `json:"genres,omitempty"`
	UserRating    *float64  `json:"user_rating,omitempty"`
	UserLiked     *bool     `json:"user_liked,omitempty"`
	UserHasReview *bool     `json:"user_has_review,omitempty"`
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
	ID                    uint                     `json:"id"`
	SteamID               int                      `json:"steam_id"`
	Title                 string                   `json:"title"`
	Description           string                   `json:"description"`
	ReleaseDate           time.Time                `json:"release_date"`
	Price                 float64                  `json:"price"`
	IsFree                bool                     `json:"is_free"`
	AvgRating             float64                  `json:"avg_rating"`
	ReviewCount           int                      `json:"review_count"`
	LikeCount             int                      `json:"like_count"`
	AveragePlaytime       float64                  `json:"average_playtime"`
	PlaytimeStory         float64                  `json:"playtime_story"`
	PlaytimeCompletionist float64                  `json:"playtime_completionist"`
	Studio                *StudioDTO               `json:"studio"`
	Genres                []GenreDTO               `json:"genres"`
	Platforms             []PlatformDTO            `json:"platforms"`
	Images                []string                 `json:"images"`
	HeaderImage           string                   `json:"header_image,omitempty"`
	CoverImage            string                   `json:"cover_image,omitempty"`
	Screenshots           []string                 `json:"screenshots,omitempty"`
	PopularReviews        []ReviewCompactResponse `json:"popular_reviews"`
	RecentReviews         []ReviewCompactResponse `json:"recent_reviews"`
	MoreFromStudio        []GameTrendingResponse  `json:"more_from_studio"`
	SimilarGames          []GameTrendingResponse  `json:"similar_games"`
	PlaysCount            int                     `json:"plays_count"`
	PlayingCount          int                     `json:"playing_count"`
	DroppedCount          int                     `json:"dropped_count"`
	BacklogCount          int                     `json:"backlog_count"`
	WishlistCount         int                     `json:"wishlist_count"`
	RatingCount           int                     `json:"rating_count"`
	ListsCount            int                     `json:"lists_count"`
	RatingDistribution    []int                   `json:"rating_distribution"`
}

type StudioDetailResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Games       []GameTrendingResponse `json:"games"`
}

type PaginationDTO struct {
	TotalRecords int `json:"total_records"`
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	Limit        int `json:"limit"`
}

type CreateGameRequest struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description"`
	ReleaseDate     string   `json:"releaseDate"`
	Studio          string   `json:"studio"`
	Rating          string   `json:"rating"`
	Platforms       []string `json:"platforms"`
	Genres          []string `json:"genres"`
	AveragePlaytime string   `json:"averagePlaytime"`
	PlaytimeStory   string   `json:"playtimeStory"`
	PlaytimeMaster  string   `json:"playtimeMaster"`
	Images          struct {
		Header string `json:"header"` // Banner chính
		Main   string `json:"main"`   // Ảnh bìa cover
	} `json:"images"`
	Screenshots []string `json:"screenshots"`
}

type LogGameStatusRequest struct {
	GameID uint   `json:"game_id" binding:"required"`
	Status string `json:"status"`
}

type UserReviewDTO struct {
	ReviewID  uint   `json:"review_id"`
	Content   string `json:"content"`
	Recommend string `json:"recommend"`
	IsSpoiler bool   `json:"is_spoiler"`
}

type GameUserStateResponse struct {
	Rating    float64        `json:"rating"`
	Liked     bool           `json:"liked"`
	LogStatus string         `json:"log_status"`
	Review    *UserReviewDTO `json:"review"`
}

