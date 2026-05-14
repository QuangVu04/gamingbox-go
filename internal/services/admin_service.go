package services

import (
	"fmt"
	"time"
	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/repositories"
)

type AdminService interface {
	GetDashboardStats(timeframe string) (*dto.DashboardStatsResponse, error)
	GetActivityChart(timeframe string) ([]dto.ChartItem, error)
}

type adminService struct {
	userRepo        repositories.UserRepository
	gameRepo        repositories.GameRepository
	reviewRepo      repositories.ReviewRepository
	gameLogRepo     repositories.GameLogRepository
	activityLogRepo repositories.ActivityLogRepository
	ratingRepo      repositories.RatingRepository
}

func NewAdminService(
	userRepo repositories.UserRepository,
	gameRepo repositories.GameRepository,
	reviewRepo repositories.ReviewRepository,
	gameLogRepo repositories.GameLogRepository,
	activityLogRepo repositories.ActivityLogRepository,
	ratingRepo repositories.RatingRepository,
) AdminService {
	return &adminService{
		userRepo:        userRepo,
		gameRepo:        gameRepo,
		reviewRepo:      reviewRepo,
		gameLogRepo:     gameLogRepo,
		activityLogRepo: activityLogRepo,
		ratingRepo:      ratingRepo,
	}
}

func (s *adminService) GetDashboardStats(timeframe string) (*dto.DashboardStatsResponse, error) {
	now := time.Now()
	var since time.Time

	switch timeframe {
	case "Hôm nay":
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "30 Ngày":
		since = now.AddDate(0, 0, -30)
	default: // 7 Ngày
		since = now.AddDate(0, 0, -7)
	}

	totalUsers, _ := s.userRepo.Count()
	newUsers, _ := s.userRepo.CountRecent(since)
	
	totalGames, _ := s.gameRepo.Count()
	newGames, _ := s.gameRepo.CountRecent(since)

	totalReviews, _ := s.reviewRepo.Count()
	newReviews, _ := s.reviewRepo.CountRecent(since)

	totalLogs, _ := s.gameLogRepo.Count()
	newLogs, _ := s.gameLogRepo.CountRecent(since)

	// Build recent activities
	activities, _ := s.activityLogRepo.GetRecentByUserID(0, 5) // 0 for all users
	activityDTOs := make([]dto.ActivitySummary, 0)
	for _, act := range activities {
		activityDTOs = append(activityDTOs, mapper.ToActivitySummary(&act))
	}

	// 1. Genre Stats
	rawGenreStats, _ := s.gameRepo.GetGenreStats()
	colors := []string{"#3b82f6", "#10b981", "#f59e0b", "#ec4899", "#8b5cf6", "#ef4444"}
	genreStats := make([]dto.GenreStatItem, 0)
	idx := 0
	for name, count := range rawGenreStats {
		if count > 0 {
			genreStats = append(genreStats, dto.GenreStatItem{
				Name:  name,
				Value: count,
				Color: colors[idx%len(colors)],
			})
			idx++
		}
		if idx >= 6 { // Limit to 6 main genres for UI
			break
		}
	}

	// 2. Top Games
	trendingGames, _, _ := s.gameRepo.GetTrendingGames(1, 5)
	topGames := make([]dto.TopGameItem, 0)
	for _, tg := range trendingGames {
		status := "Phổ biến"
		if tg.TrendingScore > 20 {
			status = "Đang tăng"
		}
		
		genreName := "Đang cập nhật"
		if len(tg.Game.Genres) > 0 {
			genreName = tg.Game.Genres[0].Name
		}

		topGames = append(topGames, dto.TopGameItem{
			ID:      tg.Game.ID,
			Title:   tg.Game.Title,
			Genre:   genreName,
			Members: tg.Game.ReviewCount*12 + 50, // Approximation for demo
			Rating:  float64(int(tg.Game.AvgRating*10)) / 10,
			Status:  status,
		})
	}

	return &dto.DashboardStatsResponse{
		TotalUsers: dto.StatItem{
			Value:      fmt.Sprintf("%d", totalUsers),
			Change:     fmt.Sprintf("+%d", newUsers),
			IsPositive: true,
		},
		TotalGames: dto.StatItem{
			Value:      fmt.Sprintf("%d", totalGames),
			Change:     fmt.Sprintf("+%d", newGames),
			IsPositive: true,
		},
		TotalReviews: dto.StatItem{
			Value:      fmt.Sprintf("%d", totalReviews),
			Change:     fmt.Sprintf("+%d", newReviews),
			IsPositive: true,
		},
		TotalLogs: dto.StatItem{
			Value:      fmt.Sprintf("%d", totalLogs),
			Change:     fmt.Sprintf("+%d", newLogs),
			IsPositive: true,
		},
		RecentActivities: activityDTOs,
		GenreStats:       genreStats,
		TopGames:         topGames,
	}, nil
}

func (s *adminService) GetActivityChart(timeframe string) ([]dto.ChartItem, error) {
	now := time.Now()
	days := 7
	if timeframe == "30 Ngày" {
		days = 30
	}

	chartSince := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	logDaily, _ := s.gameLogRepo.GetDailyCounts(chartSince)
	reviewDaily, _ := s.reviewRepo.GetDailyCounts(chartSince)
	ratingDaily, _ := s.ratingRepo.GetDailyCounts(chartSince)

	weekdayNames := []string{"CN", "T2", "T3", "T4", "T5", "T6", "T7"}
	chartData := make([]dto.ChartItem, 0)

	for i := 0; i < days; i++ {
		date := chartSince.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		label := weekdayNames[date.Weekday()]
		if days > 7 {
			label = date.Format("02/01")
		}

		chartData = append(chartData, dto.ChartItem{
			Name:    label,
			Logs:    logDaily[dateStr],
			Reviews: reviewDaily[dateStr],
			Ratings: ratingDaily[dateStr],
		})
	}

	return chartData, nil
}

