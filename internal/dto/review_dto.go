package dto

// ReviewGameInfo contains game information for review response
type ReviewGameInfo struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Thumbnail   string `json:"thumbnail"`
	ReleaseDate string `json:"release_date"`
}

// ReviewUserInfo contains user information for review response
type ReviewUserInfo struct {
	ID       uint    `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
}

// ReviewTrendingResponse represents a trending review in response
type ReviewTrendingResponse struct {
	ReviewID     uint           `json:"review_id"`
	Game         ReviewGameInfo `json:"game"`
	User         ReviewUserInfo `json:"user"`
	Rating       float64        `json:"rating,omitempty"`
	Content      string         `json:"content"`
	LikeCount    int            `json:"like_count"`
	CommentCount int            `json:"comment_count"`
	IsSpoiler    bool           `json:"is_spoiler"`
	CreatedAt    string         `json:"created_at"`
}
