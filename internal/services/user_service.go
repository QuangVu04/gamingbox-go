package services

import (
	"sort"
	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"

	"gorm.io/gorm"
)

type UserService interface {
	GetUserProfile(userID uint) (*dto.UserProfileResponse, error)
}

type userService struct {
	userRepo        repositories.UserRepository
	reviewRepo      repositories.ReviewRepository
	gameLogRepo     repositories.GameLogRepository
	listRepo        repositories.ListRepository
	activityLogRepo repositories.ActivityLogRepository
	ratingRepo      repositories.RatingRepository
}

func NewUserService(
	userRepo repositories.UserRepository,
	reviewRepo repositories.ReviewRepository,
	gameLogRepo repositories.GameLogRepository,
	listRepo repositories.ListRepository,
	activityLogRepo repositories.ActivityLogRepository,
	ratingRepo repositories.RatingRepository,
) UserService {
	return &userService{
		userRepo:        userRepo,
		reviewRepo:      reviewRepo,
		gameLogRepo:     gameLogRepo,
		listRepo:        listRepo,
		activityLogRepo: activityLogRepo,
		ratingRepo:      ratingRepo,
	}
}

func (s *userService) GetUserProfile(userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
	}

	recentReviews, err := s.fetchRecentReviews(userID, 5)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lấy đánh giá gần đây")
	}

	popularReviews, err := s.fetchPopularReviews(userID, 5)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lấy đánh giá phổ biến")
	}

	backlogGames, err := s.fetchBacklogGames(userID, 5)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lấy backlog")
	}

	diary, err := s.fetchDiaryEntries(userID, 20)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lấy lịch sử chơi")
	}

	averageRating, err := s.fetchAverageRating(userID)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể tính điểm trung bình")
	}

	recentActivity, err := s.fetchRecentActivity(userID, 10)
	if err != nil {
		return nil, dto.NewServiceError("SERVER_ERROR", "không thể lấy hoạt động gần đây")
	}

	return mapper.ToUserProfileResponse(
		user,
		averageRating,
		recentReviews,
		popularReviews,
		backlogGames,
		diary,
		recentActivity,
	), nil
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
	backlogList, err := s.listRepo.GetBacklogByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []dto.GameSummary{}, nil
		}
		return nil, err
	}

	entries, err := s.listRepo.GetBacklogEntries(backlogList.ID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.GameSummary, 0, len(entries))
	for _, entry := range entries {
		result = append(result, mapper.ToGameSummary(&entry.Game))
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
