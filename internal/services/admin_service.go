package services

import (
	"fmt"
	"time"
	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
)

type AdminService interface {
	GetDashboardStats(timeframe string) (*dto.DashboardStatsResponse, error)
	GetActivityChart(timeframe string) ([]dto.ChartItem, error)
	GetAdminGames(page, limit int, search, category, platform, minRating, startDate, endDate, sort string) ([]dto.GameAdminResponse, int64, error)
	GetAdminUsers(page, limit int, search, role, sort string) ([]dto.UserAdminResponse, int64, error)
	DeleteUser(id uint) error
	UpdateUser(id uint, username, email, role, status, bio string) error
	DeleteReview(id uint) error
	DeleteList(id uint) error
	GetUserDetailOverview(userID uint) (*dto.UserDetailOverviewResponse, error)
	GetUserActivities(userID uint, page, limit int, filterType, searchQuery string) (*dto.PaginatedResponse[[]dto.ActivitySummary], error)
	GetUserReviews(userID uint, page, limit int, filter string) (*dto.PaginatedResponse[[]dto.ReviewSummary], error)
	GetUserLists(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.ListSummary], error)
	GetUserBacklog(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.GameSummary], error)
	GetUserLibraryGames(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserLibraryGameDTO], error)
}

type adminService struct {
	userRepo        repositories.UserRepository
	gameRepo        repositories.GameRepository
	reviewRepo      repositories.ReviewRepository
	gameLogRepo     repositories.GameLogRepository
	activityLogRepo repositories.ActivityLogRepository
	ratingRepo      repositories.RatingRepository
	listRepo        repositories.ListRepository
}

func NewAdminService(
	userRepo repositories.UserRepository,
	gameRepo repositories.GameRepository,
	reviewRepo repositories.ReviewRepository,
	gameLogRepo repositories.GameLogRepository,
	activityLogRepo repositories.ActivityLogRepository,
	ratingRepo repositories.RatingRepository,
	listRepo repositories.ListRepository,
) AdminService {
	return &adminService{
		userRepo:        userRepo,
		gameRepo:        gameRepo,
		reviewRepo:      reviewRepo,
		gameLogRepo:     gameLogRepo,
		activityLogRepo: activityLogRepo,
		ratingRepo:      ratingRepo,
		listRepo:        listRepo,
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
	rawGenreStats, _ := s.gameRepo.GetGenreStats(since)
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
		genreName := "Đang cập nhật"
		if len(tg.Game.Genres) > 0 {
			genreName = tg.Game.Genres[0].Name
		}

		imageURL := ""
		for _, img := range tg.Game.Images {
			if img.ImgType == "cover" {
				imageURL = img.OgURL
				break
			}
		}
		if imageURL == "" && len(tg.Game.Images) > 0 {
			imageURL = tg.Game.Images[0].OgURL
		}

		topGames = append(topGames, dto.TopGameItem{
			ID:      tg.Game.ID,
			Title:   tg.Game.Title,
			Genre:   genreName,
			Reviews: tg.ReviewCount7d,
			Ratings: tg.RatingCount7d,
			Rating:  float64(int(tg.Game.AvgRating*10)) / 10,
			Image:   imageURL,
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
	chartData := make([]dto.ChartItem, 0)

	if timeframe == "Hôm nay" {
		since := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		logHourly, err1 := s.gameLogRepo.GetHourlyCounts(since)
		reviewHourly, err2 := s.reviewRepo.GetHourlyCounts(since)
		ratingHourly, err3 := s.ratingRepo.GetHourlyCounts(since)

		// Debug logging
		fmt.Printf("DEBUG: ActivityChart Today Since: %v\n", since)
		fmt.Printf("DEBUG: Log Counts: %v (err: %v)\n", logHourly, err1)
		fmt.Printf("DEBUG: Review Counts: %v (err: %v)\n", reviewHourly, err2)
		fmt.Printf("DEBUG: Rating Counts: %v (err: %v)\n", ratingHourly, err3)

		for i := 0; i < 24; i++ {
			hourStr02 := fmt.Sprintf("%02d", i) // "09"
			hourStrRaw := fmt.Sprintf("%d", i)   // "9"
			label := fmt.Sprintf("%02d:00", i)

			// Try both formats just in case
			logs := logHourly[hourStr02]
			if logs == 0 { logs = logHourly[hourStrRaw] }

			reviews := reviewHourly[hourStr02]
			if reviews == 0 { reviews = reviewHourly[hourStrRaw] }

			ratings := ratingHourly[hourStr02]
			if ratings == 0 { ratings = ratingHourly[hourStrRaw] }

			chartData = append(chartData, dto.ChartItem{
				Name:    label,
				Logs:    logs,
				Reviews: reviews,
				Ratings: ratings,
			})
		}
		return chartData, nil
	}

	days := 7
	if timeframe == "30 Ngày" {
		days = 30
	}

	chartSince := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	logDaily, _ := s.gameLogRepo.GetDailyCounts(chartSince)
	reviewDaily, _ := s.reviewRepo.GetDailyCounts(chartSince)
	ratingDaily, _ := s.ratingRepo.GetDailyCounts(chartSince)

	weekdayNames := []string{"CN", "T2", "T3", "T4", "T5", "T6", "T7"}

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

func (s *adminService) GetAdminGames(page, limit int, search, category, platform, minRating, startDate, endDate, sort string) ([]dto.GameAdminResponse, int64, error) {
	games, total, err := s.gameRepo.SearchAdminGames(search, category, platform, minRating, startDate, endDate, sort, page, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.GameAdminResponse, 0, len(games))
	for _, g := range games {
		var studioName string
		if g.Studio.ID != 0 {
			studioName = g.Studio.Name
		} else {
			studioName = "Unknown Studio"
		}

		genres := make([]string, 0, len(g.Genres))
		for _, gen := range g.Genres {
			genres = append(genres, gen.Name)
		}

		platforms := make([]string, 0, len(g.Platforms))
		for _, plat := range g.Platforms {
			platforms = append(platforms, plat.Name)
		}

		var img string
		for _, image := range g.Images {
			if image.ImgType == "cover" {
				img = image.OgURL
				break
			}
		}
		if img == "" && len(g.Images) > 0 {
			img = g.Images[0].OgURL
		}

		result = append(result, dto.GameAdminResponse{
			ID:          g.ID,
			Title:       g.Title,
			Studio:      studioName,
			Genres:      genres,
			Platforms:   platforms,
			Rating:      g.AvgRating,
			Reviews:     g.ReviewCount,
			ReleaseDate: g.ReleaseDate.Format("2006-01-02"),
			Image:       img,
		})
	}

	return result, total, nil
}

func (s *adminService) GetAdminUsers(page, limit int, search, role, sort string) ([]dto.UserAdminResponse, int64, error) {
	users, total, err := s.userRepo.GetAdminUsers(page, limit, search, role, sort)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserAdminResponse, 0, len(users))
	for _, u := range users {
		roleStr := "User"
		if u.Role == "admin" {
			roleStr = "Super Admin"
		} else if u.Role == "influencer" {
			roleStr = "Moderator"
		}

		statusStr := "Active"
		if u.Status == "banned" {
			statusStr = "Banned"
		}

		// Calculate mock/derived likes if needed
		likesCount := u.FollowerCount * 15
		likesStr := fmt.Sprintf("%d", likesCount)
		if likesCount >= 1000 {
			likesStr = fmt.Sprintf("%.1fk", float64(likesCount)/1000.0)
		}

		var avatar string
		if u.AvatarURL != nil {
			avatar = *u.AvatarURL
		}

		result = append(result, dto.UserAdminResponse{
			ID:        u.ID,
			Name:      u.Username,
			Email:     u.Email,
			AvatarURL: avatar,
			Role:      roleStr,
			Status:    statusStr,
			JoinDate:  u.CreatedAt.Format("02/01/2006"),
			Influence: dto.InfluenceDTO{
				Followers: u.FollowerCount,
				Likes:     likesStr,
			},
			Contributions: dto.ContributionsDTO{
				Logs:    u.GameLogsCount,
				Reviews: u.ReviewCount,
			},
		})
	}

	return result, total, nil
}

func (s *adminService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

func (s *adminService) GetUserDetailOverview(userID uint) (*dto.UserDetailOverviewResponse, error) {
	u, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	avgRating, _ := s.ratingRepo.GetAverageRatingByUserID(userID)

	var avatar string
	if u.AvatarURL != nil {
		avatar = *u.AvatarURL
	}

	var bio string
	if u.Bio != nil {
		bio = *u.Bio
	}

	return &dto.UserDetailOverviewResponse{
		ID:             u.ID,
		Email:          u.Email,
		Username:       u.Username,
		AvatarURL:      &avatar,
		Bio:            &bio,
		Role:           string(u.Role),
		Status:         u.Status,
		Location:       "Hồ Chí Minh, Việt Nam",
		SteamID:        u.SteamID,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		ReviewCount:    u.ReviewCount,
		ListCount:      u.ListCount,
		GameLogsCount:  u.GameLogsCount,
		AverageRating:  avgRating,
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	}, nil
}

func (s *adminService) GetUserActivities(userID uint, page, limit int, filterType, searchQuery string) (*dto.PaginatedResponse[[]dto.ActivitySummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	activities, total, err := s.activityLogRepo.GetByUserIDPaginated(userID, page, limit, filterType, searchQuery)
	if err != nil {
		return nil, err
	}
	data := make([]dto.ActivitySummary, 0, len(activities))
	for _, act := range activities {
		summary := mapper.ToActivitySummary(&act)

		var gameID uint
		if act.TargetType == "game" || act.TargetType == "rating" {
			gameID = act.TargetID
		} else if act.TargetType == "review" {
			if r, err := s.reviewRepo.FindByID(act.TargetID); err == nil && r != nil {
				gameID = r.TargetID
			}
		} else if act.TargetType == "list" {
			if l, err := s.listRepo.FindDetailByID(act.TargetID); err == nil && l != nil && len(l.Entries) > 0 {
				gameID = l.Entries[0].GameID
			}
		}

		if gameID != 0 {
			if g, err := s.gameRepo.GetByID(gameID); err == nil && g != nil {
				imgURL := ""
				for _, img := range g.Images {
					if img.ImgType == "cover" {
						imgURL = img.OgURL
						break
					}
				}
				if imgURL == "" && len(g.Images) > 0 {
					imgURL = g.Images[0].OgURL
				}
				summary.TargetImage = imgURL
			}
		}

		data = append(data, summary)
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.ActivitySummary]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *adminService) GetUserReviews(userID uint, page, limit int, filter string) (*dto.PaginatedResponse[[]dto.ReviewSummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	reviews, total, err := s.reviewRepo.GetByUserIDPaginated(userID, page, limit, filter)
	if err != nil {
		return nil, err
	}

	reviewIDs := make([]uint, 0, len(reviews))
	for _, r := range reviews {
		reviewIDs = append(reviewIDs, r.ID)
	}
	commentCounts, _ := s.reviewRepo.GetCommentCounts(reviewIDs)

	data := make([]dto.ReviewSummary, 0, len(reviews))
	for _, r := range reviews {
		data = append(data, mapper.ToReviewSummary(&r, commentCounts[r.ID]))
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.ReviewSummary]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *adminService) GetUserLists(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.ListSummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	lists, total, err := s.listRepo.GetByUserIDPaginated(userID, page, limit)
	if err != nil {
		return nil, err
	}
	data := make([]dto.ListSummary, 0, len(lists))
	for _, l := range lists {
		data = append(data, mapper.ToListSummary(&l))
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.ListSummary]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *adminService) GetUserBacklog(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.GameSummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	logs, total, err := s.gameLogRepo.GetBacklogByUserIDPaginated(userID, page, limit)
	if err != nil {
		return nil, err
	}
	data := make([]dto.GameSummary, 0, len(logs))
	for _, lg := range logs {
		data = append(data, mapper.ToGameSummary(&lg.Game))
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.GameSummary]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *adminService) GetUserLibraryGames(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserLibraryGameDTO], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	logs, total, err := s.gameLogRepo.GetByUserIDPaginated(userID, page, limit)
	if err != nil {
		return nil, err
	}
	data := make([]dto.UserLibraryGameDTO, 0, len(logs))
	for _, lg := range logs {
		year := "2024"
		if !lg.Game.ReleaseDate.IsZero() {
			year = lg.Game.ReleaseDate.Format("2006")
		}

		img := ""
		for _, image := range lg.Game.Images {
			if image.ImgType == "cover" {
				img = image.OgURL
				break
			}
		}
		if img == "" && len(lg.Game.Images) > 0 {
			img = lg.Game.Images[0].OgURL
		}
		if img == "" {
			img = "https://cdn.akamai.steamstatic.com/steam/apps/1091500/library_600x900.jpg"
		}

		statusStr := "Completed"
		if lg.Status == "playing" {
			statusStr = "Playing"
		} else if lg.Status == "dropped" {
			statusStr = "Dropped"
		}

		data = append(data, dto.UserLibraryGameDTO{
			ID:         lg.Game.ID,
			SteamID:    lg.Game.SteamID,
			Title:      lg.Game.Title,
			Year:       year,
			Status:     statusStr,
			Rating:     lg.Game.AvgRating,
			HasReview:  true,
			LastPlayed: lg.LoggedAt.Format("02/01/2006"),
			Image:      img,
		})
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.UserLibraryGameDTO]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data: data,
	}, nil
}

func (s *adminService) UpdateUser(id uint, username, email, role, status, bio string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("người dùng không tồn tại: %w", err)
	}

	user.Username = username
	user.Email = email
	user.Role = models.UserRole(role)
	user.Status = status
	user.Bio = &bio

	return s.userRepo.Update(user)
}

func (s *adminService) DeleteReview(id uint) error {
	return s.reviewRepo.Delete(id)
}

func (s *adminService) DeleteList(id uint) error {
	return s.listRepo.Delete(id)
}

