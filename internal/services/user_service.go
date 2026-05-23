package services

import (
	"context"
	"fmt"
	"time"
    
	mathRand "math/rand"
	"sort"
	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
	"vault/be/database"
	"vault/be/pkg/utils"


	"gorm.io/gorm"
)

type UserService interface {
	GetUserProfile(userID uint) (*dto.UserProfileResponse, error)
	ToggleFollow(followerID, followingID uint) (*dto.FollowResponse, error)
	GetFollowing(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserSummary], error)
	GetFollowers(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserSummary], error)
	GetUserStats(userID uint) (*dto.UserStatsResponse, error)
	GetDiary(userID uint, status string, page, limit int) (*dto.PaginatedResponse[[]dto.DiaryEntry], error)
	GetWatchlist(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.GameSummary], error)
	UpdateUserStatus(userID uint, status string) error
	UpdateUserRole(userID uint, role string) error
	UpdateFavoriteGames(userID uint, gameIDs []uint) error
	GetUserReviews(userID uint, page, limit int, filter string) (*dto.PaginatedResponse[[]dto.ReviewSummary], error)
	GetUserLists(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.ListSummary], error)
	GetUserActivities(userID uint, page, limit int, filterType, searchQuery string) (*dto.PaginatedResponse[[]dto.ActivitySummary], error)
	UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error
	RequestEmailChangeOTP(userID uint, req *dto.RequestEmailChangeRequest) error
	VerifyEmailChangeOTP(userID uint, req *dto.VerifyEmailChangeRequest) error
}

type userService struct {
	userRepo        repositories.UserRepository
	reviewRepo      repositories.ReviewRepository
	gameLogRepo     repositories.GameLogRepository
	listRepo        repositories.ListRepository
	activityLogRepo repositories.ActivityLogRepository
	ratingRepo      repositories.RatingRepository
	gameRepo        repositories.GameRepository
	notifService    NotificationService
	db              *gorm.DB
}

func NewUserService(
	userRepo repositories.UserRepository,
	reviewRepo repositories.ReviewRepository,
	gameLogRepo repositories.GameLogRepository,
	listRepo repositories.ListRepository,
	activityLogRepo repositories.ActivityLogRepository,
	ratingRepo repositories.RatingRepository,
	gameRepo repositories.GameRepository,
	notifService NotificationService,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepo:        userRepo,
		reviewRepo:      reviewRepo,
		gameLogRepo:     gameLogRepo,
		listRepo:        listRepo,
		activityLogRepo: activityLogRepo,
		ratingRepo:      ratingRepo,
		gameRepo:        gameRepo,
		notifService:    notifService,
		db:              db,
	}
}

func (s *userService) GetUserProfile(userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}

	averageRating, err := s.fetchAverageRating(userID)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể tính điểm trung bình")
	}

	// Fetch Favorite Games
	favoriteGames := make([]dto.GameSummary, 0)
	if len(user.FavoriteGameIDs) > 0 {
		rawFavs, err := s.gameRepo.GetByIDs(user.FavoriteGameIDs)
		if err == nil {
			favMap := make(map[uint]models.Game)
			for _, g := range rawFavs {
				favMap[g.ID] = g
			}
			for _, id := range user.FavoriteGameIDs {
				if g, ok := favMap[id]; ok {
					favoriteGames = append(favoriteGames, mapper.ToGameSummary(&g))
				}
			}
			s.populateGameInteractions(userID, favoriteGames)
		}
	}

	return mapper.ToUserProfileResponse(
		user,
		averageRating,
		[]dto.ReviewSummary{},   // recentReviews
		[]dto.ReviewSummary{},   // popularReviews
		[]dto.GameSummary{},     // backlogGames
		[]dto.DiaryEntry{},      // diary
		[]dto.ActivitySummary{}, // recentActivity
		[]dto.ListSummary{},     // lists
		favoriteGames,
	), nil
}

func (s *userService) populateGameInteractions(userID uint, summaries []dto.GameSummary) {
	if len(summaries) == 0 {
		return
	}

	gameIDs := make([]uint, 0, len(summaries))
	for _, s := range summaries {
		gameIDs = append(gameIDs, s.ID)
	}

	ratingsMap := make(map[uint]float64)
	likesMap := make(map[uint]bool)
	reviewsMap := make(map[uint]bool)

	// Get user ratings
	var ratings []models.Rating
	if err := s.db.Where("user_id = ? AND game_id IN ?", userID, gameIDs).Find(&ratings).Error; err == nil {
		for _, r := range ratings {
			ratingsMap[r.GameID] = r.Rating
		}
	}

	// Get user likes
	var likes []models.Like
	if err := s.db.Where("user_id = ? AND target_type = 'game' AND target_id IN ?", userID, gameIDs).Find(&likes).Error; err == nil {
		for _, l := range likes {
			likesMap[l.TargetID] = true
		}
	}

	// Get user reviews
	var reviews []models.Review
	if err := s.db.Where("user_id = ? AND target_type = 'game' AND target_id IN ?", userID, gameIDs).Find(&reviews).Error; err == nil {
		for _, rev := range reviews {
			reviewsMap[rev.TargetID] = true
		}
	}

	// Populate fields
	for i := range summaries {
		id := summaries[i].ID
		if rVal, exists := ratingsMap[id]; exists {
			rValCopy := rVal
			summaries[i].UserRating = &rValCopy
		}
		liked := likesMap[id]
		summaries[i].UserLiked = &liked
		hasReview := reviewsMap[id]
		summaries[i].UserHasReview = &hasReview
	}
}

func (s *userService) ToggleFollow(followerID, followingID uint) (*dto.FollowResponse, error) {
	if followerID == followingID {
		return nil, dto.NewServiceError("INVALID_ACTION", "Không thể tự follow chính mình")
	}

	// Kiểm tra xem user có tồn tại không
	_, err := s.userRepo.FindByID(followingID)
	if err != nil {
		return nil, dto.NewServiceError("USER_NOT_FOUND", "Tài khoản không tồn tại")
	}

	isFollowing, err := s.userRepo.ToggleFollow(followerID, followingID)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "Không thể thực hiện chức năng follow")
	}

	// Trigger notification if followed
	if isFollowing {
		s.notifService.TriggerNotification(dto.NotificationTask{
			ReceiverID: followingID,
			SenderID:   followerID,
			ActionType: "follow",
			TargetID:   followingID,
			TargetType: "user",
		})
	}

	return &dto.FollowResponse{IsFollowing: isFollowing}, nil
}

func (s *userService) GetFollowing(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserSummary], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := s.userRepo.GetFollowing(userID, offset, limit)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi khi lấy danh sách đang theo dõi")
	}

	var data []dto.UserSummary
	for _, u := range users {
		data = append(data, dto.UserSummary{
			UserID:   u.ID,
			Username: u.Username,
			Avatar:   u.AvatarURL,
			Bio:      u.Bio,
		})
	}
	if data == nil {
		data = []dto.UserSummary{}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.PaginatedResponse[[]dto.UserSummary]{
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

func (s *userService) GetFollowers(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.UserSummary], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := s.userRepo.GetFollowers(userID, offset, limit)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi khi lấy danh sách người theo dõi")
	}

	var data []dto.UserSummary
	for _, u := range users {
		data = append(data, dto.UserSummary{
			UserID:   u.ID,
			Username: u.Username,
			Avatar:   u.AvatarURL,
			Bio:      u.Bio,
		})
	}
	if data == nil {
		data = []dto.UserSummary{}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.PaginatedResponse[[]dto.UserSummary]{
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

func (s *userService) fetchRecentReviews(userID uint, limit int) ([]dto.ReviewSummary, error) {
	reviews, err := s.reviewRepo.GetRecentByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	commentCounts, err := s.fetchReviewCommentCounts(reviews)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ReviewSummary, 0, len(reviews))
	for _, review := range reviews {
		result = append(result, mapper.ToReviewSummary(&review, commentCounts[review.ID]))
	}

	return result, nil
}

func (s *userService) fetchPopularReviews(userID uint, limit int) ([]dto.ReviewSummary, error) {
	reviews, err := s.reviewRepo.GetPopularByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	commentCounts, err := s.fetchReviewCommentCounts(reviews)
	if err != nil {
		return nil, err
	}

	type rankedReview struct {
		review       models.Review
		interactions int
	}

	ranked := make([]rankedReview, 0, len(reviews))
	for _, review := range reviews {
		ranked = append(ranked, rankedReview{
			review:       review,
			interactions: review.LikeCount + commentCounts[review.ID],
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].interactions > ranked[j].interactions
	})

	limitCount := limit
	if len(ranked) < limitCount {
		limitCount = len(ranked)
	}

	result := make([]dto.ReviewSummary, 0, limitCount)
	for i := 0; i < limitCount; i++ {
		review := ranked[i].review
		result = append(result, mapper.ToReviewSummary(&review, commentCounts[review.ID]))
	}

	return result, nil
}

func (s *userService) fetchReviewCommentCounts(reviews []models.Review) (map[uint]int, error) {
	counts := make(map[uint]int)
	if len(reviews) == 0 {
		return counts, nil
	}

	ids := make([]uint, 0, len(reviews))
	for _, review := range reviews {
		ids = append(ids, review.ID)
	}

	return s.reviewRepo.GetCommentCounts(ids)
}

func (s *userService) fetchBacklogGames(userID uint, limit int) ([]dto.GameSummary, error) {
	logs, err := s.gameLogRepo.GetBacklogByUserID(userID, limit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []dto.GameSummary{}, nil
		}
		return nil, err
	}

	result := make([]dto.GameSummary, 0, len(logs))
	for _, log := range logs {
		result = append(result, mapper.ToGameSummary(&log.Game))
	}

	return result, nil
}

func (s *userService) fetchDiaryEntries(userID uint, limit int) ([]dto.DiaryEntry, error) {
	logs, err := s.gameLogRepo.GetByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.DiaryEntry, 0, len(logs))
	for _, log := range logs {
		result = append(result, mapper.ToDiaryEntry(&log))
	}

	// Populate user interactions on games in diary
	if len(result) > 0 {
		summaries := make([]dto.GameSummary, len(result))
		for i := range result {
			summaries[i] = result[i].Game
		}
		s.populateGameInteractions(userID, summaries)
		for i := range result {
			result[i].Game = summaries[i]
		}
	}

	return result, nil
}

func (s *userService) fetchAverageRating(userID uint) (float64, error) {
	return s.ratingRepo.GetAverageRatingByUserID(userID)
}

func (s *userService) fetchRecentActivity(userID uint, limit int) ([]dto.ActivitySummary, error) {
	activities, err := s.activityLogRepo.GetRecentByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ActivitySummary, 0, len(activities))
	for _, activity := range activities {
		result = append(result, mapper.ToActivitySummary(&activity))
	}

	return result, nil
}

func (s *userService) GetUserStats(userID uint) (*dto.UserStatsResponse, error) {
	totalPlayed, _ := s.ratingRepo.GetTotalRatedByUserID(userID)
	avgRating, _ := s.fetchAverageRating(userID)
	
	return &dto.UserStatsResponse{
		TotalPlayed:     int(totalPlayed),
		TotalReviews:    0, 
		AverageRating:   avgRating,
		GenreDistribution: make(map[string]int),
		RatingDistribution: make(map[int]int),
	}, nil
}

func (s *userService) GetDiary(userID uint, status string, page, limit int) (*dto.PaginatedResponse[[]dto.DiaryEntry], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	
	logs, total, err := s.gameLogRepo.GetByUserIDPaginated(userID, status, page, limit)
	if err != nil {
		return nil, err
	}

	diary := make([]dto.DiaryEntry, 0, len(logs))
	for _, log := range logs {
		diary = append(diary, mapper.ToDiaryEntry(&log))
	}

	// Populate user interactions on games in diary
	if len(diary) > 0 {
		summaries := make([]dto.GameSummary, len(diary))
		for i := range diary {
			summaries[i] = diary[i].Game
		}
		s.populateGameInteractions(userID, summaries)
		for i := range diary {
			diary[i].Game = summaries[i]
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 { totalPages++ }

	return &dto.PaginatedResponse[[]dto.DiaryEntry]{
		Status: "success",
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
		Data:   diary,
	}, nil
}

func (s *userService) GetWatchlist(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.GameSummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	logs, total, err := s.gameLogRepo.GetBacklogByUserIDPaginated(userID, page, limit)
	if err != nil {
		return nil, err
	}

	games := make([]dto.GameSummary, 0, len(logs))
	for _, log := range logs {
		games = append(games, mapper.ToGameSummary(&log.Game))
	}

	if len(games) > 0 {
		s.populateGameInteractions(userID, games)
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
		Data:   games,
	}, nil
}

func (s *userService) UpdateUserStatus(userID uint, status string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}

	user.Status = status
	return s.userRepo.Update(user)
}

func (s *userService) UpdateUserRole(userID uint, role string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}

	user.Role = models.UserRole(role)
	return s.userRepo.Update(user)
}

func (s *userService) UpdateFavoriteGames(userID uint, gameIDs []uint) error {
	if len(gameIDs) > 4 {
		return dto.NewServiceError("INVALID_INPUT", "Chỉ được chọn tối đa 4 game yêu thích")
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}
	user.FavoriteGameIDs = gameIDs
	return s.userRepo.Update(user)
}

func (s *userService) GetUserReviews(userID uint, page, limit int, filter string) (*dto.PaginatedResponse[[]dto.ReviewSummary], error) {
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

func (s *userService) GetUserLists(userID uint, page, limit int) (*dto.PaginatedResponse[[]dto.ListSummary], error) {
	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	lists, total, err := s.listRepo.GetByUserIDPaginated(userID, page, limit)
	if err != nil {
		return nil, err
	}
	data := make([]dto.ListSummary, 0, len(lists))
	for _, l := range lists {
		summary := mapper.ToListSummary(&l)
		summary.IsLiked = s.listRepo.IsListLiked(userID, l.ID)
		data = append(data, summary)
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

func (s *userService) GetUserActivities(userID uint, page, limit int, filterType, searchQuery string) (*dto.PaginatedResponse[[]dto.ActivitySummary], error) {
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
func (s *userService) UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}

	if req.Username != nil && *req.Username != "" && *req.Username != user.Username {
		existingUser, _ := s.userRepo.FindByUsername(*req.Username)
		if existingUser != nil && existingUser.ID != userID {
			return dto.NewServiceError("USERNAME_EXISTS", "Username đã có người sử dụng")
		}
		user.Username = *req.Username
	}

	if req.Email != nil && *req.Email != "" && *req.Email != user.Email {
		existingUser, _ := s.userRepo.FindByEmail(*req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return dto.NewServiceError("EMAIL_EXISTS", "Email đã có người sử dụng")
		}
		user.Email = *req.Email
	}

	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Location != nil {
		user.Location = req.Location
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	return s.userRepo.Update(user)
}

func (s *userService) RequestEmailChangeOTP(userID uint, req *dto.RequestEmailChangeRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "Tài khoản không tồn tại")
	}

	if req.NewEmail == user.Email {
		return dto.NewServiceError("INVALID_ACTION", "Email mới phải khác email hiện tại")
	}

	existingUser, _ := s.userRepo.FindByEmail(req.NewEmail)
	if existingUser != nil && existingUser.ID != userID {
		return dto.NewServiceError("EMAIL_EXISTS", "Email này đã có người sử dụng")
	}

	// Generate 6-digit code
	code := fmt.Sprintf("%06d", mathRand.Intn(1000000))
	
	// Store in Redis with 5 mins expiration
	ctx := context.Background()
	cacheKey := fmt.Sprintf("update_email_otp:%d", userID)
	err = database.RDB.Set(ctx, cacheKey, req.NewEmail+":"+code, 5*time.Minute).Err()
	if err != nil {
		return dto.NewServiceError("SERVER_ERROR", "Không thể tạo mã xác nhận")
	}

	fmt.Printf("====[TESTING] Mã xác nhận đổi email cho %s là: %s ====\n", req.NewEmail, code)

	// Send email
	go func() {
		body := utils.GenerateOTPEmailTemplate(
			"Xác nhận đổi Email",
			"Bạn vừa yêu cầu thay đổi địa chỉ email cho tài khoản GamingBox của mình. Vui lòng nhập mã xác nhận gồm 6 chữ số bên dưới để hoàn tất:",
			code,
		)
		err := utils.SendEmail(req.NewEmail, "Mã xác nhận đổi email - GamingBox", body)
		if err != nil {
			fmt.Printf("====[LỖI GỬI EMAIL] Không thể gửi email tới %s: %v ====\n", req.NewEmail, err)
		} else {
			fmt.Printf("====[THÀNH CÔNG] Đã gửi email tới %s ====\n", req.NewEmail)
		}
	}()

	return nil
}

func (s *userService) VerifyEmailChangeOTP(userID uint, req *dto.VerifyEmailChangeRequest) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("update_email_otp:%d", userID)

	// Verify code from Redis
	storedData, err := database.RDB.Get(ctx, cacheKey).Result()
	if err != nil {
		return dto.NewFieldError("INVALID_CODE", "Mã xác nhận không đúng hoặc đã hết hạn", "code")
	}

	// Data is stored as new_email:code
	expectedData := req.NewEmail + ":" + req.Code
	if storedData != expectedData {
		return dto.NewFieldError("INVALID_CODE", "Mã xác nhận không đúng", "code")
	}

	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return dto.NewServiceError("USER_NOT_FOUND", "Tài khoản không tồn tại")
	}

	// Double check email uniqueness just in case
	existingUser, _ := s.userRepo.FindByEmail(req.NewEmail)
	if existingUser != nil && existingUser.ID != userID {
		return dto.NewServiceError("EMAIL_EXISTS", "Email này đã có người sử dụng")
	}

	// Update email
	user.Email = req.NewEmail
	if err := s.userRepo.Update(user); err != nil {
		return dto.NewServiceError("SERVER_ERROR", "Không thể cập nhật email")
	}

	// Delete the OTP code
	database.RDB.Del(ctx, cacheKey)

	return nil
}
