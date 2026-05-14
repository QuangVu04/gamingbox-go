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
	Members int     `json:"members"`
	Rating  float64 `json:"rating"`
	Status  string  `json:"status"`
}

type GenreStatItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}
