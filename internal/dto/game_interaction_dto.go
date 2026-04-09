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
