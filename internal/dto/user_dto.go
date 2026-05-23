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
    Location                *string         `json:"location"`
    Role                    models.UserRole `json:"role"`
    SteamID                 string          `json:"steam_id"`
    FollowerCount           int             `json:"follower_count"`
    FollowingCount          int             `json:"following_count"`
    ReviewCount             int             `json:"review_count"`
    ListCount               int             `json:"list_count"`
    GameLogsCount           int             `json:"game_logs_count"`
    AverageRating           float64         `json:"average_rating_count"`
    CreatedAt               time.Time       `json:"created_at"`
}

type UpdateProfileRequest struct {
    Username  *string `json:"username"`
    Email     *string `json:"email"`
    Location  *string `json:"location"`
    Bio       *string `json:"bio"`
    AvatarURL *string `json:"avatar_url"`
}

type RequestEmailChangeRequest struct {
    NewEmail string `json:"new_email" binding:"required,email"`
}

type VerifyEmailChangeRequest struct {
    NewEmail string `json:"new_email" binding:"required,email"`
    Code     string `json:"code" binding:"required,len=6"`
}

// UserSummary contains basic information of a user for listing
type UserSummary struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
}