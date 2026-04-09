package dto


// ListAuthorInfo contains author information for list response
type ListAuthorInfo struct {
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
}

// ListTrendingResponse represents a trending list in response
type ListTrendingResponse struct {
	ListID           uint            `json:"list_id"`
	Title            string          `json:"title"`
	Author           ListAuthorInfo  `json:"author"`
	GameCount        int             `json:"game_count"`
	Thumbnails       []string        `json:"thumbnails"`
	WeeklyLikesCount int64             `json:"weekly_likes_count"`
	TotalLikes       int             `json:"total_likes"`
	CreatedAt        string          `json:"created_at"`
}
