package dto

import (
    "time"

    "vault/be/internal/models"
)

type UserResponse struct {
    ID                      uint            `json:"id"`
    Email                   string          `json:"email"`
    Username                string          `json:"username"`
    AvatarURL               *string         `json:"avatar_url"`
    Bio                     *string         `json:"bio"`
    Role                    models.UserRole `json:"role"`
    SteamID                 string          `json:"steam_id"`
    FollowerCount           int             `json:"follower_count"`
    FollowingCount          int             `json:"following_count"`
    ReviewCount             int             `json:"review_count"`
    ListCount               int             `json:"list_count"`
    GameLogsCount           int             `json:"game_logs_count"`
    AverageRating           float64         `json:"average_rating_count"`
    CreatedAt               time.Time       `json:"created_at"`
    RecentGames             []GameSummary   `json:"recent_games"`
}

// UserSummary contains basic information of a user for listing
type UserSummary struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
}