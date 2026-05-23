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
	WeeklyLikesCount int64          `json:"weekly_likes_count"`
	TotalLikes       int            `json:"total_likes"`
	UserHasLiked     bool           `json:"user_has_liked"`
	CreatedAt        string         `json:"created_at"`
}

type CreateListRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	ThumbnailImg string `json:"thumbnail_img"`
	IsPublic     bool   `json:"is_public"`
	GameIDs      []uint `json:"game_ids"`
}

type UpdateListRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailImg string `json:"thumbnail_img"`
	IsPublic     *bool  `json:"is_public"`
	GameIDs      []uint `json:"game_ids"`
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
	Games        []ListEntryDTO `json:"games"`
}
type GameListsResponse struct {
	Lists      []ListTrendingResponse `json:"lists"`
	Pagination PaginationDTO          `json:"pagination"`
}
