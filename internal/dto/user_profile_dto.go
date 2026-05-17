package dto

import (
	"time"

	"vault/be/internal/models"
)

type UserProfileResponse struct {
	ID             uint              `json:"id"`
	Email          string            `json:"email"`
	Username       string            `json:"username"`
	AvatarURL      *string           `json:"avatar_url"`
	Bio            *string           `json:"bio"`
	Role           models.UserRole   `json:"role"`
	Status         string            `json:"status"`
	Location       string            `json:"location"`
	SteamID        string            `json:"steam_id"`
	FollowerCount  int               `json:"follower_count"`
	FollowingCount int               `json:"following_count"`
	ReviewCount    int               `json:"review_count"`
	ListCount      int               `json:"list_count"`
	GameLogsCount  int               `json:"game_logs_count"`
	AverageRating  float64           `json:"average_rating"`
	CreatedAt      time.Time         `json:"created_at"`
	RecentReviews  []ReviewSummary   `json:"recent_reviews"`
	PopularReviews []ReviewSummary   `json:"popular_reviews"`
	BacklogGames   []GameSummary     `json:"backlog_games"`
	Diary          []DiaryEntry      `json:"diary"`
	RecentActivity []ActivitySummary `json:"recent_activity"`
	Lists          []ListSummary     `json:"lists"`
}

type ListSummary struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	GameCount int       `json:"game_count"`
	LikeCount int       `json:"like_count"`
	UpdatedAt time.Time `json:"updated_at"`
	Thumbnail string    `json:"thumbnail"`
}

type FollowRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

type FollowResponse struct {
	IsFollowing bool `json:"is_following"`
}

//review summary for user profile
type ReviewSummary struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	Img          string      `json:"img"`
	LikeCount    int         `json:"like_count"`
	CommentCount int         `json:"comment_count"`
	Recommend    string      `json:"recommend"`
	IsSpoiler    bool        `json:"is_spoiler"`
	CreatedAt    time.Time   `json:"created_at"`
	Game         GameSummary `json:"game"`
}


type DiaryEntry struct {
	Game     GameSummary `json:"game"`
	Status   string      `json:"status"`
	LoggedAt time.Time   `json:"logged_at"`
}

type ActivitySummary struct {
	ID         uint64      `json:"id"`
	ActionType string      `json:"action_type"`
	TargetType string      `json:"target_type"`
	TargetID   uint        `json:"target_id"`
	Preview    string      `json:"preview"`
	CreatedAt  time.Time   `json:"created_at"`
	User       UserSummary `json:"user"`
}

type UserStatsResponse struct {
	TotalPlayed     int            `json:"total_played"`
	TotalReviews    int            `json:"total_reviews"`
	AverageRating   float64        `json:"average_rating"`
	GenreDistribution map[string]int `json:"genre_distribution"`
	RatingDistribution map[int]int    `json:"rating_distribution"` // 1 to 5 stars
}


