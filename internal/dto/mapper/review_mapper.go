package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToReviewSummary(review *models.Review, commentCount int) dto.ReviewSummary {
	return dto.ReviewSummary{
		ID:           review.ID,
		Title:        review.Title,
		Content:      review.Content,
		Img:          review.Img,
		LikeCount:    review.LikeCount,
		CommentCount: commentCount,
		Recommend:    review.Recommend,
		IsSpoiler:    review.IsSpoiler,
		CreatedAt:    review.CreatedAt,
		Game:         ToGameSummary(&review.Game),
	}
}

// ToReviewTrendingResponse converts a Review model to a ReviewTrendingResponse DTO
func ToReviewTrendingResponse(review *models.Review, commentCount int) *dto.ReviewTrendingResponse {
	if review == nil {
		return nil
	}

	// Extract thumbnail from game images
	thumbnail := ""
	for _, img := range review.Game.Images {
		if img.ImgType == "header" || img.Thumb != "" {
			thumbnail = img.Thumb
			if img.Thumb == "" {
				thumbnail = img.OgURL
			}
			break
		}
	}

	// Get user avatar
	userAvatar := review.User.AvatarURL

	return &dto.ReviewTrendingResponse{
		ReviewID: review.ID,
		Game: dto.ReviewGameInfo{
			ID:          review.Game.ID,
			Title:       review.Game.Title,
			Thumbnail:   thumbnail,
			ReleaseDate: review.Game.ReleaseDate.Format("2006-01-02"),
		},
		User: dto.ReviewUserInfo{
			ID:       review.User.ID,
			Username: review.User.Username,
			Avatar:   userAvatar,
		},
		Content:      review.Content,
		LikeCount:    review.LikeCount,
		CommentCount: commentCount,
		IsSpoiler:    review.IsSpoiler,
		CreatedAt:    review.CreatedAt.Format("2006-01-02"),
	}
}

// ToReviewTrendingResponses converts multiple Review models to ReviewTrendingResponse DTOs
func ToReviewTrendingResponses(reviews []models.Review, commentCounts map[uint]int) []dto.ReviewTrendingResponse {
	responses := make([]dto.ReviewTrendingResponse, 0, len(reviews))
	for i := range reviews {
		count := commentCounts[reviews[i].ID]
		resp := ToReviewTrendingResponse(&reviews[i], count)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
