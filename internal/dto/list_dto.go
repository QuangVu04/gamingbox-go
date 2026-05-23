package dto


// ListAuthorInfo contains author information for list response
type ListAuthorInfo struct {
	ID       uint    `json:"id"`
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
	WeeklyLikesCount int64          `json:"weekly_likes_count"`
	TotalLikes       int            `json:"total_likes"`
	UserHasLiked     bool           `json:"user_has_liked"`
	CreatedAt        string         `json:"created_at"`
	CommentCount     int            `json:"comment_count"`
}

type ListEntryRequest struct {
	GameID uint   `json:"game_id"`
	Note   string `json:"note"`
}

type CreateListRequest struct {
	Title        string             `json:"title" binding:"required"`
	Description  string             `json:"description"`
	ThumbnailImg string             `json:"thumbnail_img"`
	IsPublic     bool               `json:"is_public"`
	GameIDs      []uint             `json:"game_ids"` // deprecated, use Entries instead
	Entries      []ListEntryRequest `json:"entries"`
}

type UpdateListRequest struct {
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	ThumbnailImg string             `json:"thumbnail_img"`
	IsPublic     *bool              `json:"is_public"`
	GameIDs      []uint             `json:"game_ids"` // deprecated, use Entries instead
	Entries      []ListEntryRequest `json:"entries"`
}

type ListEntryDTO struct {
	GameID uint   `json:"game_id"`
	Title  string `json:"title"`
	Poster string `json:"poster"`
	Note   string `json:"note"`
}

type ListDetailResponse struct {
	ID           uint           `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Author       ListAuthorInfo `json:"author"`
	ThumbnailImg string         `json:"thumbnail_img"`
	GameCount    int            `json:"game_count"`
	LikeCount    int            `json:"like_count"`
	UserHasLiked bool           `json:"user_has_liked"`
	CreatedAt    string         `json:"created_at"`
	IsLiked      bool           `json:"is_liked"`
	Games        []ListEntryDTO `json:"games"`
	CommentCount int            `json:"comment_count"`
}
type GameListsResponse struct {
	Lists      []ListTrendingResponse `json:"lists"`
	Pagination PaginationDTO          `json:"pagination"`
}
