package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToReviewSummary(review *models.Review, commentCount int) dto.ReviewSummary {
	return dto.ReviewSummary{
		ID:           review.ID,
		Content:      review.Content,
		LikeCount:    review.LikeCount,
		CommentCount: commentCount,
		Recommend:    review.Recommend,
		IsSpoiler:    review.IsSpoiler,
		CreatedAt:    review.CreatedAt,
		Game:         ToGameSummary(&review.Game),
	}
}

func ToReviewTrendingResponse(review *models.Review, commentCount int, userHasLiked bool) *dto.ReviewTrendingResponse {
	if review == nil {
		return nil
	}

	// Extract thumbnail from game images (prioritize cover)
	thumbnail := ""
	for _, img := range review.Game.Images {
		if img.ImgType == "cover" {
			thumbnail = img.OgURL
			if thumbnail == "" {
				thumbnail = img.Thumb
			}
			break
		}
	}
	if thumbnail == "" {
		for _, img := range review.Game.Images {
			if img.ImgType == "header" || img.Thumb != "" {
				thumbnail = img.Thumb
				if img.Thumb == "" {
					thumbnail = img.OgURL
				}
				break
			}
		}
	}

	var userAvatar *string
	if review.User.AvatarURL != nil {
		userAvatar = review.User.AvatarURL
	}

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
		Rating:       review.Rating,
		Content:      review.Content,
		LikeCount:    review.LikeCount,
		CommentCount: commentCount,
		IsSpoiler:    review.IsSpoiler,
		UserHasLiked: userHasLiked,
		CreatedAt:    review.CreatedAt.Format("2006-01-02"),
	}
}

// ToReviewTrendingResponses converts multiple Review models to ReviewTrendingResponse DTOs
func ToReviewTrendingResponses(reviews []models.Review, commentCounts map[uint]int, likedReviews map[uint]bool) []dto.ReviewTrendingResponse {
	responses := make([]dto.ReviewTrendingResponse, 0, len(reviews))
	for i := range reviews {
		count := commentCounts[reviews[i].ID]
		hasLiked := likedReviews[reviews[i].ID]
		resp := ToReviewTrendingResponse(&reviews[i], count, hasLiked)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}

// ToReviewCompactResponse converts a Review model to a ReviewCompactResponse DTO (no game info)
func ToReviewCompactResponse(review *models.Review, commentCount int, userHasLiked bool) *dto.ReviewCompactResponse {
	if review == nil {
		return nil
	}

	var userAvatar *string
	if review.User.AvatarURL != nil {
		userAvatar = review.User.AvatarURL
	}

	return &dto.ReviewCompactResponse{
		ReviewID: review.ID,
		User: dto.ReviewUserInfo{
			ID:       review.User.ID,
			Username: review.User.Username,
			Avatar:   userAvatar,
		},
		Rating:       review.Rating,
		Content:      review.Content,
		LikeCount:    review.LikeCount,
		CommentCount: commentCount,
		IsSpoiler:    review.IsSpoiler,
		UserHasLiked: userHasLiked,
		CreatedAt:    review.CreatedAt.Format("2006-01-02"),
	}
}

// ToReviewCompactResponses converts multiple Review models to ReviewCompactResponse DTOs
func ToReviewCompactResponses(reviews []models.Review, commentCounts map[uint]int, likedReviews map[uint]bool) []dto.ReviewCompactResponse {
	responses := make([]dto.ReviewCompactResponse, 0, len(reviews))
	for i := range reviews {
		count := commentCounts[reviews[i].ID]
		hasLiked := likedReviews[reviews[i].ID]
		resp := ToReviewCompactResponse(&reviews[i], count, hasLiked)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}

func ToCommentResponse(comment *models.Comment, user *models.User, userHasLiked bool) *dto.CommentResponse {
	if comment == nil {
		return nil
	}

	return &dto.CommentResponse{
		ID: comment.ID,
		User: dto.ReviewUserInfo{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   user.AvatarURL,
		},
		Content:      comment.Content,
		ParentID:     comment.ParentID,
		LikeCount:    comment.LikeCount,
		UserHasLiked: userHasLiked,
		CreatedAt:    comment.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToCommentResponses(comments []models.Comment, users map[uint]models.User, likedComments map[uint]bool) []dto.CommentResponse {
	responses := make([]dto.CommentResponse, 0, len(comments))
	for _, c := range comments {
		user := users[c.UserID]
		hasLiked := likedComments[c.ID]
		resp := ToCommentResponse(&c, &user, hasLiked)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}
	return responses
}
