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

type ReviewCompactResponse struct {
	ReviewID     uint           `json:"review_id"`
	User         ReviewUserInfo `json:"user"`
	Rating       float64        `json:"rating,omitempty"`
	Content      string         `json:"content"`
	LikeCount    int            `json:"like_count"`
	CommentCount int            `json:"comment_count"`
	IsSpoiler    bool           `json:"is_spoiler"`
	CreatedAt    string         `json:"created_at"`
}

type CreateReviewRequest struct {
	GameID    uint    `json:"game_id" binding:"required"`
	Title     string  `json:"title"`
	Content   string  `json:"content" binding:"required"`
	Recommend string  `json:"recommend" binding:"required,oneof=recommend mixed not_recommend"`
	IsSpoiler bool    `json:"is_spoiler"`
	Img       string  `json:"img"`
}

type UpdateReviewRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Recommend string `json:"recommend" binding:"omitempty,oneof=recommend mixed not_recommend"`
	IsSpoiler bool   `json:"is_spoiler"`
	Img       string `json:"img"`
}

type CommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parent_id"`
}

type CommentResponse struct {
	ID        uint           `json:"id"`
	User      ReviewUserInfo `json:"user"`
	Content   string         `json:"content"`
	ParentID  *uint          `json:"parent_id"`
	CreatedAt string         `json:"created_at"`
}
