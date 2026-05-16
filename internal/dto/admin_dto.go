package dto

type DashboardStatsResponse struct {
	TotalLogs        StatItem `json:"total_logs"`
	TotalUsers       StatItem `json:"total_users"`
	TotalGames       StatItem `json:"total_games"`
	TotalReviews     StatItem `json:"total_reviews"`
	TopGames         []TopGameItem `json:"top_games"`
	RecentActivities []ActivitySummary `json:"recent_activities"`
	GenreStats       []GenreStatItem `json:"genre_stats"`
}

type StatItem struct {
	Value      string `json:"value"`
	Change     string `json:"change"`
	IsPositive bool   `json:"is_positive"`
}

type ChartItem struct {
	Name    string `json:"name"`
	Logs    int    `json:"logs"`
	Reviews int    `json:"reviews"`
	Ratings int    `json:"ratings"`
}

type TopGameItem struct {
	ID      uint    `json:"id"`
	Title   string  `json:"title"`
	Genre   string  `json:"genre"`
	Reviews int     `json:"reviews"`
	Ratings int     `json:"ratings"`
	Rating  float64 `json:"rating"`
	Image   string  `json:"image"`
}

type GenreStatItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

type GameAdminResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Studio      string    `json:"studio"`
	Genres      []string  `json:"genres"`
	Platforms   []string  `json:"platforms"`
	Rating      float64   `json:"rating"`
	Reviews     int       `json:"reviews"`
	ReleaseDate string    `json:"releaseDate"`
	Image       string    `json:"image"`
}

type UserAdminResponse struct {
	ID            uint             `json:"id"`
	Name          string           `json:"name"`
	Email         string           `json:"email"`
	AvatarURL     string           `json:"avatarUrl"`
	Role          string           `json:"role"`
	Status        string           `json:"status"`
	JoinDate      string           `json:"joinDate"`
	Influence     InfluenceDTO     `json:"influence"`
	Contributions ContributionsDTO `json:"contributions"`
}

type InfluenceDTO struct {
	Followers int    `json:"followers"`
	Likes     string `json:"likes"`
}

type ContributionsDTO struct {
	Logs    int `json:"logs"`
	Reviews int `json:"reviews"`
}
