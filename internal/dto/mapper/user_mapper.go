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

func ToListSummary(list *models.List) dto.ListSummary {
	thumbnail := ""
	if len(list.Entries) > 0 && list.Entries[0].Game.Images != nil && len(list.Entries[0].Game.Images) > 0 {
		for _, img := range list.Entries[0].Game.Images {
			if img.ImgType == "cover" {
				thumbnail = img.OgURL
				break
			}
		}
		if thumbnail == "" {
			thumbnail = list.Entries[0].Game.Images[0].OgURL
		}
	}

	// Extract thumbnails from up to 5 games
	thumbnails := make([]string, 0)
	for i, entry := range list.Entries {
		if i >= 5 {
			break
		}
		if entry.Game.ID > 0 {
			thumbnailImg := ""
			for _, img := range entry.Game.Images {
				if img.ImgType == "cover" {
					thumbnailImg = img.Thumb
					if thumbnailImg == "" {
						thumbnailImg = img.OgURL
					}
					break
				}
			}
			if thumbnailImg == "" {
				for _, img := range entry.Game.Images {
					if img.ImgType == "header" {
						thumbnailImg = img.Thumb
						if thumbnailImg == "" {
							thumbnailImg = img.OgURL
						}
						break
					}
				}
			}
			if thumbnailImg != "" {
				thumbnails = append(thumbnails, thumbnailImg)
			}
		}
	}

	if len(thumbnails) == 0 && list.ThumbnailImg != "" {
		thumbnails = append(thumbnails, list.ThumbnailImg)
	}

	return dto.ListSummary{
		ID:         list.ID,
		Title:      list.Title,
		GameCount:  len(list.Entries),
		LikeCount:  list.LikeCount,
		UpdatedAt:  list.UpdatedAt,
		Thumbnail:  thumbnail,
		Thumbnails: thumbnails,
	}
}

func ToUserProfileResponse(user *models.User, averageRating float64, recentReviews []dto.ReviewSummary, popularReviews []dto.ReviewSummary, backlogGames []dto.GameSummary, diary []dto.DiaryEntry, recentActivity []dto.ActivitySummary, lists []dto.ListSummary, favoriteGames []dto.GameSummary) *dto.UserProfileResponse {
	return &dto.UserProfileResponse{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		Bio:            user.Bio,
		Role:           user.Role,
		Status:         user.Status,
		Location:       "",
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
		Lists:          lists,
		FavoriteGames:  favoriteGames,
	}
}