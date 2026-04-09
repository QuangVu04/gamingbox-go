package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToUserResponse(user *models.User) *dto.UserResponse {
    return &dto.UserResponse{
        ID:                      user.ID,
        Email:                   user.Email,
        Username:                user.Username,
        AvatarURL:               user.AvatarURL,
        Bio:                     user.Bio,
        Role:                    user.Role,
        SteamID:                 user.SteamID,
        FollowerCount:           user.FollowerCount,
        FollowingCount:          user.FollowingCount,
        ReviewCount:             user.ReviewCount,
        ListCount:               user.ListCount,
        GameLogsCount:           user.GameLogsCount,
        AverageRating:           user.AverageRating,
        CreatedAt:               user.CreatedAt,
    }
}

func ToUserProfileResponse(user *models.User, averageRating float64, recentReviews []dto.ReviewSummary, popularReviews []dto.ReviewSummary, backlogGames []dto.GameSummary, diary []dto.DiaryEntry, recentActivity []dto.ActivitySummary) *dto.UserProfileResponse {
	return &dto.UserProfileResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		Bio:            user.Bio,
		Role:           user.Role,
		SteamID:        user.SteamID,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		ReviewCount:    user.ReviewCount,
		ListCount:      user.ListCount,
		GameLogsCount:  user.GameLogsCount,
		AverageRating:  averageRating,
		CreatedAt:      user.CreatedAt,
		RecentReviews:  recentReviews,
		PopularReviews: popularReviews,
		BacklogGames:   backlogGames,
		Diary:          diary,
		RecentActivity: recentActivity,
	}
}