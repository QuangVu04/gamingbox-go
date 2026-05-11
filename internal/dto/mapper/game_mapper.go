package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToGameSummary(game *models.Game) dto.GameSummary {
	return dto.GameSummary{
		ID:          game.ID,
		SteamID:     game.SteamID,
		Title:       game.Title,
		Poster:      firstPoster(game),
		ReleaseDate: game.ReleaseDate,
		Price:       game.Price,
		IsFree:      game.IsFree,
		AvgRating:   game.AvgRating,
		ReviewCount: game.ReviewCount,
	}
}

func firstPoster(game *models.Game) string {
	if len(game.Images) == 0 {
		return ""
	}

	for _, img := range game.Images {
		if img.ImgType == "header" {
			return img.OgURL
		}
	}

	return game.Images[0].OgURL
}

// ToGameTrendingResponse converts a GameTrending model to a GameTrendingResponse DTO
func ToGameTrendingResponse(game *models.GameTrending) *dto.GameTrendingResponse {
	if game == nil {
		return nil
	}

	// Extract thumbnail from images
	thumbnail := ""
	for _, img := range game.Game.Images {
		if img.ImgType == "header" || img.Thumb != "" {
			thumbnail = img.Thumb
			if img.Thumb == "" {
				thumbnail = img.OgURL
			}
			break
		}
	}

	// Extract studio names
	studios := make([]string, 0)
	if game.Game.Studio.ID > 0 {
		studios = append(studios, game.Game.Studio.Name)
	}

	return &dto.GameTrendingResponse{
		GameID:        game.Game.ID,
		Title:         game.Game.Title,
		Thumbnail:     thumbnail,
		TrendingScore: game.TrendingScore,
		AvgRating:     game.Game.AvgRating,
		TotalReviews:  game.Game.ReviewCount,
		ReleaseDate:   game.Game.ReleaseDate,
		Studios:       studios,
	}
}

// ToGameTrendingResponses converts multiple GameTrending models to GameTrendingResponse DTOs
func ToGameTrendingResponses(games []models.GameTrending) []dto.GameTrendingResponse {
	responses := make([]dto.GameTrendingResponse, 0, len(games))
	for i := range games {
		resp := ToGameTrendingResponse(&games[i])
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}

func ToGameDetailResponse(game *models.Game) *dto.GameDetailResponse {
	if game == nil {
		return nil
	}

	genres := make([]dto.GenreDTO, 0, len(game.Genres))
	for _, g := range game.Genres {
		genres = append(genres, dto.GenreDTO{ID: g.ID, Name: g.Name})
	}

	platforms := make([]dto.PlatformDTO, 0, len(game.Platforms))
	for _, p := range game.Platforms {
		platforms = append(platforms, dto.PlatformDTO{ID: p.ID, Name: p.Name})
	}

	images := make([]string, 0, len(game.Images))
	for _, img := range game.Images {
		images = append(images, img.OgURL)
	}

	var studio *dto.StudioDTO
	if game.Studio.ID > 0 {
		studio = &dto.StudioDTO{
			ID:   game.Studio.ID,
			Name: game.Studio.Name,
		}
	}

	return &dto.GameDetailResponse{
		ID:              game.ID,
		SteamID:         game.SteamID,
		Title:           game.Title,
		Description:     game.Description,
		ReleaseDate:     game.ReleaseDate,
		Price:           game.Price,
		IsFree:          game.IsFree,
		AvgRating:       game.AvgRating,
		ReviewCount:     game.ReviewCount,
		LikeCount:       game.LikeCount,
		AveragePlaytime: game.AveragePlaytime,
		Studio:          studio,
		Genres:          genres,
		Platforms:       platforms,
		Images:          images,
	}
}

func ToSimpleGameResponse(game *models.Game) *dto.GameTrendingResponse {
	if game == nil {
		return nil
	}

	thumbnail := ""
	for _, img := range game.Images {
		if img.ImgType == "header" {
			thumbnail = img.OgURL
			break
		}
	}

	studios := make([]string, 0)
	if game.Studio.ID > 0 {
		studios = append(studios, game.Studio.Name)
	}

	return &dto.GameTrendingResponse{
		GameID:       game.ID,
		Title:        game.Title,
		Thumbnail:    thumbnail,
		AvgRating:    game.AvgRating,
		TotalReviews: game.ReviewCount,
		ReleaseDate:  game.ReleaseDate,
		Studios:      studios,
	}
}

func ToSimpleGameResponses(games []models.Game) []dto.GameTrendingResponse {
	responses := make([]dto.GameTrendingResponse, 0, len(games))
	for i := range games {
		resp := ToSimpleGameResponse(&games[i])
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
