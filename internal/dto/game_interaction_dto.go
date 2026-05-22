package dto

type LikeGameRequest struct {
	GameID uint `json:"game_id" binding:"required"`
}

type LikeGameResponse struct {
	IsLiked bool `json:"is_liked"`
}

type RateGameRequest struct {
	GameID uint    `json:"game_id" binding:"required"`
	Rating float64 `json:"rating" binding:"required,min=0.5,max=5"`
}

type RateGameResponse struct {
	MyRating     float64 `json:"my_rating"`
	NewGameAvg   float64 `json:"new_game_avg"`
	TotalRatings int     `json:"total_ratings"`
}

type InteractionTask struct {
	UserID     uint    `json:"user_id"`
	TargetID   uint    `json:"target_id"`   // game_id, review_id, list_id, comment_id
	Type       string  `json:"type"`        // "like", "log", "rate"
	Status     string  `json:"status"`      // for log
	Rating     float64 `json:"rating"`      // for rate
	TargetType string  `json:"target_type"` // for like target ("game", "review", "list", "comment")
}

