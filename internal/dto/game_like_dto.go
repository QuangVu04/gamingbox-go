package dto

import "time"

type GameLikeUserDTO struct {
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	AvatarURL     *string   `json:"avatar_url"`
	Rating        *float64  `json:"rating"`
	HasReview     bool      `json:"has_review"`
	LikedAt       time.Time `json:"liked_at"`
	FollowerCount int       `json:"follower_count"`
}

type GameLikesResponse struct {
	Likes      []GameLikeUserDTO `json:"likes"`
	Pagination PaginationDTO     `json:"pagination"`
}
