package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
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

func ToListDetailResponse(list *models.List) *dto.ListDetailResponse {
	if list == nil {
		return nil
	}

	games := make([]dto.ListEntryDTO, 0, len(list.Entries))
	for _, entry := range list.Entries {
		poster := ""
		for _, img := range entry.Game.Images {
			if img.ImgType == "header" {
				poster = img.OgURL
				break
			}
		}
		games = append(games, dto.ListEntryDTO{
			GameID: entry.GameID,
			Title:  entry.Game.Title,
			Poster: poster,
			Note:   entry.GhiChu,
		})
	}

	return &dto.ListDetailResponse{
		ID:          list.ID,
		Title:       list.Title,
		Description: list.Description,
		Author: dto.ListAuthorInfo{
			Username: list.User.Username,
			Avatar:   list.User.AvatarURL,
		},
		ThumbnailImg: list.ThumbnailImg,
		GameCount:    list.GameCount,
		LikeCount:    list.LikeCount,
		CreatedAt:    list.CreatedAt.Format("2006-01-02 15:04:05"),
		Games:        games,
	}
}
func ToTrendingListResponseFromModel(list *models.List) *dto.ListTrendingResponse {
	if list == nil {
		return nil
	}

	// Extract thumbnails from up to 5 games
	thumbnails := make([]string, 0)
	for i, entry := range list.Entries {
		if i >= 5 {
			break
		}
		if entry.Game.ID > 0 {
			for _, img := range entry.Game.Images {
				if img.ImgType == "header" {
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
		GameCount:  list.GameCount,
		Thumbnails: thumbnails,
		TotalLikes: list.LikeCount,
		CreatedAt:  list.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToTrendingListResponsesFromModels(lists []models.List) []dto.ListTrendingResponse {
	responses := make([]dto.ListTrendingResponse, 0, len(lists))
	for i := range lists {
		resp := ToTrendingListResponseFromModel(&lists[i])
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
