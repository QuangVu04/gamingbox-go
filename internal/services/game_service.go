package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
	redisUtil "vault/be/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const CacheTTL = 15 * time.Minute

type GameService interface {
	GetTrendingGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	RateGame(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error)
	RateGameDB(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error)
	GetGameByID(ctx context.Context, id uint) (*dto.GameDetailResponse, error)
	GetGameUserState(ctx context.Context, userID, gameID uint) (*dto.GameUserStateResponse, error)
	SearchGames(ctx context.Context, query, category, platform, sort string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetPopularGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error)
	GetGenres(ctx context.Context) ([]models.Genre, error)
	GetPlatforms(ctx context.Context) ([]models.Platform, error)
	CreateGame(ctx context.Context, req dto.CreateGameRequest) (*models.Game, error)
	UpdateGame(ctx context.Context, id uint, req dto.CreateGameRequest) (*models.Game, error)
	DeleteGenre(ctx context.Context, name string) error
	DeletePlatform(ctx context.Context, name string) error
	SearchStudios(ctx context.Context, query string) ([]models.Studio, error)
	GetStudioDetail(ctx context.Context, id uint) (*dto.StudioDetailResponse, error)
	DeleteGame(ctx context.Context, id uint) error
	LogGameStatus(ctx context.Context, userID, gameID uint, status string) error
	LogGameStatusDB(ctx context.Context, userID, gameID uint, status string) error
}

type gameService struct {
	gameRepo   repositories.GameRepository
	ratingRepo repositories.RatingRepository
	reviewRepo repositories.ReviewRepository
	db         *gorm.DB
	rdb        *redis.Client
}

func NewGameService(gameRepo repositories.GameRepository, ratingRepo repositories.RatingRepository, reviewRepo repositories.ReviewRepository, db *gorm.DB, rdb *redis.Client) GameService {
	return &gameService{
		gameRepo:   gameRepo,
		ratingRepo: ratingRepo,
		reviewRepo: reviewRepo,
		db:         db,
		rdb:        rdb,
	}
}

// CachedGameResponse is used for marshaling/unmarshaling game data with pagination
type CachedGameResponse struct {
	Data       []dto.GameTrendingResponse `json:"data"`
	Pagination *dto.PaginationDTO         `json:"pagination"`
}

func (s *gameService) GetTrendingGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("games", page, limit)
	cached, err := redisUtil.GetCached[CachedGameResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending games (page=%d, limit=%d)", page, limit)
		return cached.Data, cached.Pagination, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for trending games (page=%d, limit=%d), fetching from database", page, limit)
	trendingGames, totalCount, err := s.gameRepo.GetTrendingGames(page, limit)
	if err != nil {
		return nil, nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy dữ liệu game trending")
	}

	// Convert to DTOs
	responses := mapper.ToGameTrendingResponses(trendingGames)

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(totalCount),
		CurrentPage:  page,
		TotalPages:   totalPages,
		Limit:        limit,
	}

	// Cache the response
	cacheData := &CachedGameResponse{
		Data:       responses,
		Pagination: pagination,
	}
	err = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	if err != nil {
		log.Printf("⚠ Failed to cache trending games: %v", err)
	}

	return responses, pagination, nil
}

func (s *gameService) RateGame(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error) {
	// Validate rating range
	if rating < 0.5 || rating > 5.0 {
		return nil, dto.NewServiceError("VALIDATION_ERROR", "điểm số phải từ 0.5 đến 5.0")
	}

	task := dto.InteractionTask{
		UserID:   userID,
		TargetID: gameID,
		Type:     "rate",
		Rating:   rating,
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể queue đánh giá")
	}

	if s.rdb != nil {
		err = s.rdb.LPush(ctx, InteractionQueueKey, taskBytes).Err()
		if err != nil {
			return nil, dto.NewServiceError("SERVER_ERROR", "không thể queue đánh giá vào redis")
		}
	} else {
		// Fallback to direct DB write if Redis is down
		return s.RateGameDB(ctx, userID, gameID, rating)
	}

	// For the response: get current game stats to return, so we don't break frontend expectations.
	avgRating, totalRatings, err := s.ratingRepo.GetGameStats(gameID)
	if err != nil {
		avgRating = 0.0
		totalRatings = 0
	}

	return &dto.RateGameResponse{
		MyRating:     rating,
		NewGameAvg:   avgRating,
		TotalRatings: totalRatings,
	}, nil
}

func (s *gameService) RateGameDB(ctx context.Context, userID, gameID uint, rating float64) (*dto.RateGameResponse, error) {
	// Validate rating range
	if rating < 0.5 || rating > 5.0 {
		return nil, dto.NewServiceError("VALIDATION_ERROR", "điểm số phải từ 0.5 đến 5.0")
	}

	var result *dto.RateGameResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Upsert rating
		ratingRecord := &models.Rating{
			UserID:    userID,
			GameID:    gameID,
			Rating:    rating,
			CreatedAt: time.Now(),
		}

		if err := s.ratingRepo.UpsertRating(ratingRecord); err != nil {
			return err
		}

		avgRating, totalRatings, err := s.ratingRepo.GetGameStats(gameID)
		if err != nil {
			return err
		}

		if err := s.ratingRepo.UpdateGameRating(gameID, avgRating, totalRatings); err != nil {
			return err
		}

		result = &dto.RateGameResponse{
			MyRating:     rating,
			NewGameAvg:   avgRating,
			TotalRatings: totalRatings,
		}

		return nil
	})

	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lưu đánh giá")
	}

	return result, nil
}

func (s *gameService) GetGameByID(ctx context.Context, id uint) (*dto.GameDetailResponse, error) {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy game")
	}

	// 1. Fetch Popular Reviews (top 3)
	popularReviews, _ := s.reviewRepo.GetByGameID(id, "popular", 3)
	popCommentCounts, _ := s.reviewRepo.GetCommentCounts(getReviewIDs(popularReviews))
	popularDTOs := mapper.ToReviewCompactResponses(popularReviews, popCommentCounts)

	// 2. Fetch Recent Reviews (top 3)
	recentReviews, _ := s.reviewRepo.GetByGameID(id, "recent", 3)
	recCommentCounts, _ := s.reviewRepo.GetCommentCounts(getReviewIDs(recentReviews))
	recentDTOs := mapper.ToReviewCompactResponses(recentReviews, recCommentCounts)

	// 3. More from this studio (top 6)
	moreFromStudio, _ := s.gameRepo.GetByStudio(game.StudioID, game.ID, 6)
	studioDTOs := mapper.ToSimpleGameResponses(moreFromStudio)

	// 4. Similar Games (same genres, top 6)
	genreIDs := make([]uint, 0, len(game.Genres))
	for _, g := range game.Genres {
		genreIDs = append(genreIDs, g.ID)
	}
	similarGames, _ := s.gameRepo.GetByGenres(genreIDs, game.ID, 6)
	similarDTOs := mapper.ToSimpleGameResponses(similarGames)

	// Map core response
	resp := mapper.ToGameDetailResponse(game)

	// Attach extra data
	resp.PopularReviews = popularDTOs
	resp.RecentReviews = recentDTOs
	resp.MoreFromStudio = studioDTOs
	resp.SimilarGames = similarDTOs

	// Fetch dynamic stats from database
	var playsCount int64
	var playingCount int64
	var droppedCount int64
	var backlogCount int64
	var wishlistCount int64
	var listsCount int64
	var ratingCount int64

	s.db.Model(&models.GameLog{}).Where("game_id = ? AND status = ?", id, "completed").Count(&playsCount)
	s.db.Model(&models.GameLog{}).Where("game_id = ? AND status = ?", id, "playing").Count(&playingCount)
	s.db.Model(&models.GameLog{}).Where("game_id = ? AND status = ?", id, "dropped").Count(&droppedCount)
	s.db.Model(&models.GameLog{}).Where("game_id = ? AND status = ?", id, "backlog").Count(&backlogCount)
	
	s.db.Table("list_entries").
		Joins("JOIN lists ON lists.id = list_entries.list_id").
		Where("list_entries.game_id = ? AND lists.title = ?", id, "Muốn chơi").
		Count(&wishlistCount)

	s.db.Model(&models.ListEntry{}).Where("game_id = ?", id).Count(&listsCount)
	s.db.Model(&models.Rating{}).Where("game_id = ?", id).Count(&ratingCount)

	distribution := []int{0, 0, 0, 0, 0}
	var ratingRows []struct {
		Star  int
		Count int
	}
	s.db.Model(&models.Rating{}).
		Select("ROUND(rating) as star, COUNT(*) as count").
		Where("game_id = ?", id).
		Group("star").
		Scan(&ratingRows)

	for _, row := range ratingRows {
		starIdx := row.Star - 1
		if starIdx >= 0 && starIdx < 5 {
			distribution[starIdx] = row.Count
		} else if starIdx < 0 {
			distribution[0] += row.Count
		} else if starIdx >= 5 {
			distribution[4] += row.Count
		}
	}

	resp.PlaysCount = int(playsCount)
	resp.PlayingCount = int(playingCount)
	resp.DroppedCount = int(droppedCount)
	resp.BacklogCount = int(backlogCount)
	resp.WishlistCount = int(wishlistCount)
	resp.ListsCount = int(listsCount)
	resp.RatingCount = int(ratingCount)
	resp.RatingDistribution = distribution

	return resp, nil
}

func (s *gameService) GetGameUserState(ctx context.Context, userID, gameID uint) (*dto.GameUserStateResponse, error) {
	var state dto.GameUserStateResponse

	// 1. Rating
	var ratingVal float64
	if err := s.db.Model(&models.Rating{}).Where("user_id = ? AND game_id = ?", userID, gameID).Select("rating").Scan(&ratingVal).Error; err == nil {
		state.Rating = ratingVal
	}

	// 2. Like
	var likeCount int64
	if err := s.db.Model(&models.Like{}).Where("user_id = ? AND target_id = ? AND target_type = ?", userID, gameID, "game").Count(&likeCount).Error; err == nil {
		state.Liked = likeCount > 0
	}

	// 3. Log Status
	var logStatus string
	if err := s.db.Model(&models.GameLog{}).Where("user_id = ? AND game_id = ?", userID, gameID).Select("status").Scan(&logStatus).Error; err == nil {
		if logStatus == "completed" {
			state.LogStatus = "played"
		} else {
			state.LogStatus = logStatus
		}
	} else {
		state.LogStatus = "none"
	}

	// 4. Review
	var review models.Review
	if err := s.db.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, gameID, "game").First(&review).Error; err == nil {
		state.Review = &dto.UserReviewDTO{
			ReviewID:  review.ID,
			Content:   review.Content,
			Recommend: review.Recommend,
			IsSpoiler: review.IsSpoiler,
		}
	}

	return &state, nil
}

func (s *gameService) LogGameStatus(ctx context.Context, userID, gameID uint, status string) error {
	// Validate input
	mappedStatus := status
	if status == "played" {
		mappedStatus = "completed"
	}
	if status != "" && status != "none" {
		validStatuses := map[string]bool{
			"playing":   true,
			"completed": true,
			"dropped":   true,
			"backlog":   true,
		}
		if !validStatuses[mappedStatus] {
			return dto.NewServiceError("VALIDATION_ERROR", "trạng thái không hợp lệ")
		}
	}

	task := dto.InteractionTask{
		UserID:   userID,
		TargetID: gameID,
		Type:     "log",
		Status:   status,
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return dto.NewServiceError("SERVER_ERROR", "không thể queue log trạng thái")
	}

	if s.rdb != nil {
		err = s.rdb.LPush(ctx, InteractionQueueKey, taskBytes).Err()
		if err != nil {
			return dto.NewServiceError("SERVER_ERROR", "không thể queue log trạng thái vào redis")
		}
	} else {
		// Fallback to direct DB write if Redis is down
		return s.LogGameStatusDB(ctx, userID, gameID, status)
	}

	return nil
}

func (s *gameService) LogGameStatusDB(ctx context.Context, userID, gameID uint, status string) error {
	if status == "" || status == "none" {
		return s.db.WithContext(ctx).Where("user_id = ? AND game_id = ?", userID, gameID).Delete(&models.GameLog{}).Error
	}

	mappedStatus := status
	if status == "played" {
		mappedStatus = "completed"
	}

	validStatuses := map[string]bool{
		"playing":   true,
		"completed": true,
		"dropped":   true,
		"backlog":   true,
	}
	if !validStatuses[mappedStatus] {
		return dto.NewServiceError("VALIDATION_ERROR", "trạng thái không hợp lệ")
	}

	logRecord := &models.GameLog{
		UserID:   userID,
		GameID:   gameID,
		Status:   mappedStatus,
		LoggedAt: time.Now(),
	}

	return s.db.WithContext(ctx).Save(logRecord).Error
}

func getReviewIDs(reviews []models.Review) []uint {
	ids := make([]uint, 0, len(reviews))
	for _, r := range reviews {
		ids = append(ids, r.ID)
	}
	return ids
}

func (s *gameService) SearchGames(ctx context.Context, query, category, platform, sort string, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	games, total, err := s.gameRepo.SearchAdminGames(query, category, platform, "", "", "", sort, page, limit)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToSimpleGameResponses(games)

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   int(math.Ceil(float64(total) / float64(limit))),
		Limit:        limit,
	}

	return responses, pagination, nil
}

func (s *gameService) GetPopularGames(ctx context.Context, page, limit int) ([]dto.GameTrendingResponse, *dto.PaginationDTO, error) {
	games, total, err := s.gameRepo.GetPopular(page, limit)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToSimpleGameResponses(games)

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   int(math.Ceil(float64(total) / float64(limit))),
		Limit:        limit,
	}

	return responses, pagination, nil
}

func (s *gameService) GetGenres(ctx context.Context) ([]models.Genre, error) {
	return s.gameRepo.GetGenres()
}

func (s *gameService) GetPlatforms(ctx context.Context) ([]models.Platform, error) {
	return s.gameRepo.GetPlatforms()
}

func (s *gameService) CreateGame(ctx context.Context, req dto.CreateGameRequest) (*models.Game, error) {
	// 1. Xử lý Studio: Tìm hoặc tạo mới studio
	var studio models.Studio
	if req.Studio == "" {
		req.Studio = "Independent Studio"
	}
	if err := s.db.Where("name = ?", req.Studio).FirstOrCreate(&studio, models.Studio{Name: req.Studio, Description: "Studio phát triển tựa game " + req.Title}).Error; err != nil {
		return nil, err
	}

	// 2. Xử lý Genres: Tìm hoặc tạo mới các thể loại
	var genres []models.Genre
	for _, genreName := range req.Genres {
		if genreName == "" {
			continue
		}
		var g models.Genre
		if err := s.db.Where("name = ?", genreName).FirstOrCreate(&g, models.Genre{Name: genreName}).Error; err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	// 3. Xử lý Platforms: Tìm hoặc tạo mới các nền tảng
	var platforms []models.Platform
	for _, platName := range req.Platforms {
		if platName == "" {
			continue
		}
		var p models.Platform
		if err := s.db.Where("name = ?", platName).FirstOrCreate(&p, models.Platform{Name: platName}).Error; err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}

	// 4. Xử lý Ngày ra mắt (ReleaseDate)
	releaseDate := time.Now()
	if req.ReleaseDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ReleaseDate); err == nil {
			releaseDate = parsed
		} else if parsedISO, errISO := time.Parse(time.RFC3339, req.ReleaseDate); errISO == nil {
			releaseDate = parsedISO
		}
	}

	// 5. Xử lý Rating
	var ratingVal float64
	if req.Rating != "" {
		fmt.Sscanf(req.Rating, "%f", &ratingVal)
	}
	if ratingVal == 0 {
		ratingVal = 4.5
	}

	// Parsing playtimes
	var averagePlaytime float64
	var playtimeStory float64
	var playtimeMaster float64
	if req.AveragePlaytime != "" {
		fmt.Sscanf(req.AveragePlaytime, "%f", &averagePlaytime)
	}
	if req.PlaytimeStory != "" {
		fmt.Sscanf(req.PlaytimeStory, "%f", &playtimeStory)
	}
	if req.PlaytimeMaster != "" {
		fmt.Sscanf(req.PlaytimeMaster, "%f", &playtimeMaster)
	}

	// 6. Khởi tạo đối tượng Game
	game := &models.Game{
		Title:                 req.Title,
		Description:           req.Description,
		ReleaseDate:           releaseDate,
		StudioID:              studio.ID,
		AvgRating:             ratingVal,
		ReviewCount:           0,
		LikeCount:             0,
		AveragePlaytime:       averagePlaytime,
		PlaytimeStory:         playtimeStory,
		PlaytimeCompletionist: playtimeMaster,
		Genres:                genres,
		Platforms:             platforms,
	}

	// 7. Lưu Game vào DB
	if err := s.gameRepo.CreateGame(game); err != nil {
		return nil, err
	}

	// 8. Xử lý Hình ảnh (GameImg)
	// Banner chính là header
	if req.Images.Header != "" {
		headerImg := models.GameImg{
			OgURL:   req.Images.Header,
			Thumb:   req.Images.Header,
			ImgType: "header",
			GameID:  game.ID,
		}
		s.db.Create(&headerImg)
	}
	// Ảnh bìa là cover
	if req.Images.Main != "" {
		coverImg := models.GameImg{
			OgURL:   req.Images.Main,
			Thumb:   req.Images.Main,
			ImgType: "cover",
			GameID:  game.ID,
		}
		s.db.Create(&coverImg)
	}
	// Các ảnh screenshot
	for _, ss := range req.Screenshots {
		if ss == "" {
			continue
		}
		ssImg := models.GameImg{
			OgURL:   ss,
			Thumb:   ss,
			ImgType: "screenshot",
			GameID:  game.ID,
		}
		s.db.Create(&ssImg)
	}

	// Invalidate cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")

	return game, nil
}

func (s *gameService) UpdateGame(ctx context.Context, id uint, req dto.CreateGameRequest) (*models.Game, error) {
	// 1. Tìm game cũ
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy game")
	}

	// 2. Xử lý Studio
	var studio models.Studio
	if req.Studio == "" {
		req.Studio = "Independent Studio"
	}
	if err := s.db.Where("name = ?", req.Studio).FirstOrCreate(&studio, models.Studio{Name: req.Studio, Description: "Studio phát triển tựa game " + req.Title}).Error; err != nil {
		return nil, err
	}

	// 3. Xử lý Genres
	var genres []models.Genre
	for _, genreName := range req.Genres {
		if genreName == "" {
			continue
		}
		var g models.Genre
		if err := s.db.Where("name = ?", genreName).FirstOrCreate(&g, models.Genre{Name: genreName}).Error; err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	// 4. Xử lý Platforms
	var platforms []models.Platform
	for _, platName := range req.Platforms {
		if platName == "" {
			continue
		}
		var p models.Platform
		if err := s.db.Where("name = ?", platName).FirstOrCreate(&p, models.Platform{Name: platName}).Error; err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}

	// 5. Xử lý ReleaseDate
	releaseDate := game.ReleaseDate
	if req.ReleaseDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ReleaseDate); err == nil {
			releaseDate = parsed
		} else if parsedISO, errISO := time.Parse(time.RFC3339, req.ReleaseDate); errISO == nil {
			releaseDate = parsedISO
		}
	}

	// 6. Xử lý Rating
	var ratingVal float64
	if req.Rating != "" {
		fmt.Sscanf(req.Rating, "%f", &ratingVal)
	}
	if ratingVal == 0 {
		ratingVal = game.AvgRating
	}

	// Parsing playtimes
	var averagePlaytime float64
	var playtimeStory float64
	var playtimeMaster float64
	if req.AveragePlaytime != "" {
		fmt.Sscanf(req.AveragePlaytime, "%f", &averagePlaytime)
	}
	if req.PlaytimeStory != "" {
		fmt.Sscanf(req.PlaytimeStory, "%f", &playtimeStory)
	}
	if req.PlaytimeMaster != "" {
		fmt.Sscanf(req.PlaytimeMaster, "%f", &playtimeMaster)
	}

	// Cập nhật trường dữ liệu
	game.Title = req.Title
	game.Description = req.Description
	game.ReleaseDate = releaseDate
	game.StudioID = studio.ID
	game.AvgRating = ratingVal
	game.AveragePlaytime = averagePlaytime
	game.PlaytimeStory = playtimeStory
	game.PlaytimeCompletionist = playtimeMaster
	game.Genres = genres
	game.Platforms = platforms

	// Xử lý Hình ảnh
	var gameImages []models.GameImg
	if req.Images.Header != "" {
		gameImages = append(gameImages, models.GameImg{
			OgURL:   req.Images.Header,
			Thumb:   req.Images.Header,
			ImgType: "header",
			GameID:  game.ID,
		})
	}
	if req.Images.Main != "" {
		gameImages = append(gameImages, models.GameImg{
			OgURL:   req.Images.Main,
			Thumb:   req.Images.Main,
			ImgType: "cover",
			GameID:  game.ID,
		})
	}
	for _, ss := range req.Screenshots {
		if ss == "" {
			continue
		}
		gameImages = append(gameImages, models.GameImg{
			OgURL:   ss,
			Thumb:   ss,
			ImgType: "screenshot",
			GameID:  game.ID,
		})
	}
	game.Images = gameImages

	// 7. Lưu thay đổi
	if err := s.gameRepo.UpdateGame(game); err != nil {
		return nil, err
	}

	// 8. Xóa cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")

	return game, nil
}

func (s *gameService) DeleteGenre(ctx context.Context, name string) error {
	return s.gameRepo.DeleteGenreByName(name)
}

func (s *gameService) DeletePlatform(ctx context.Context, name string) error {
	return s.gameRepo.DeletePlatformByName(name)
}

func (s *gameService) SearchStudios(ctx context.Context, query string) ([]models.Studio, error) {
	return s.gameRepo.SearchStudios(query)
}

func (s *gameService) DeleteGame(ctx context.Context, id uint) error {
	// Call repository to perform deletion inside transaction
	if err := s.gameRepo.DeleteGame(id); err != nil {
		return err
	}

	// Invalidate cache
	s.rdb.Del(ctx, "trending_games:1:12", "trending_games:1:10")
	return nil
}

func (s *gameService) GetStudioDetail(ctx context.Context, id uint) (*dto.StudioDetailResponse, error) {
	var studio models.Studio
	if err := s.db.WithContext(ctx).First(&studio, id).Error; err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy studio")
	}

	var games []models.Game
	if err := s.db.WithContext(ctx).
		Preload("Images", "img_type IN ?", []string{"header", "cover"}).
		Preload("Studio").
		Joins("JOIN studios ON studios.id = games.studio_id").
		Where("studios.name = ?", studio.Name).
		Find(&games).Error; err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "lỗi lấy danh sách game")
	}

	gameDTOs := mapper.ToSimpleGameResponses(games)

	return &dto.StudioDetailResponse{
		ID:          studio.ID,
		Name:        studio.Name,
		Description: studio.Description,
		Games:       gameDTOs,
	}, nil
}

