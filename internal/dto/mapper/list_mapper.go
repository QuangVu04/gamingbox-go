package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/repositories"
)

// ToTrendingListResponse converts a ListTrendingData to a ListTrendingResponse DTO
func ToTrendingListResponse(listData *repositories.ListTrendingData) *dto.ListTrendingResponse {
	if listData == nil {
		return nil
	}

	list := &listData.List

	// Extract thumbnails from up to 5 games
	thumbnails := make([]string, 0)
	for i, entry := range list.Entries {
		if i >= 5 {
			break
		}
		if entry.Game.ID > 0 {
			for _, img := range entry.Game.Images {
				if img.ImgType == "header"{
					thumbnail := img.Thumb
					if img.Thumb == "" {
						thumbnail = img.OgURL
					}
					if thumbnail != "" {
						thumbnails = append(thumbnails, thumbnail)
					}
					break
				}
			}
		}
	}

	// If no thumbnails from games, use the list's thumbnail image
	if len(thumbnails) == 0 && list.ThumbnailImg != "" {
		thumbnails = append(thumbnails, list.ThumbnailImg)
	}

	return &dto.ListTrendingResponse{
		ListID: list.ID,
		Title:  list.Title,
		Author: dto.ListAuthorInfo{
			Username: list.User.Username,
			Avatar:   list.User.AvatarURL,
		},
		GameCount:        list.GameCount,
		Thumbnails:       thumbnails,
		WeeklyLikesCount: listData.WeeklyLikesCount,
		TotalLikes:       list.LikeCount,
		CreatedAt:        list.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToTrendingListResponses converts multiple ListTrendingData to ListTrendingResponse DTOs
func ToTrendingListResponses(listsData []repositories.ListTrendingData) []dto.ListTrendingResponse {
	responses := make([]dto.ListTrendingResponse, 0, len(listsData))
	for i := range listsData {
		resp := ToTrendingListResponse(&listsData[i])
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
