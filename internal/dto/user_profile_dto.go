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
}

//user backlog game
type GameSummary struct {
	ID          uint      `json:"id"`
	SteamID     int       `json:"steam_id,omitempty"`
	Title       string    `json:"title"`
	Poster      string    `json:"poster"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Price       float64   `json:"price,omitempty"`
	IsFree      bool      `json:"is_free"`
	AvgRating   float64   `json:"avg_rating"`
	ReviewCount int       `json:"review_count"`
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
	ID         uint64    `json:"id"`
	ActionType string    `json:"action_type"`
	TargetType string    `json:"target_type"`
	TargetID   uint      `json:"target_id"`
	Preview    string    `json:"preview"`
	CreatedAt  time.Time `json:"created_at"`
}


func firstPoster(game *models.Game) string {
	if len(game.Images) == 0 {
		return ""
	}

	for _, img := range game.Images {
		if img.ImgType == "header" {
			return img.OgURL
		}
	}

	return game.Images[0].OgURL
}
