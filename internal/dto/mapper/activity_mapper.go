package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToActivitySummary(activity *models.ActivityLog) dto.ActivitySummary {
	var userSummary dto.UserSummary
	if activity.User.ID != 0 {
		userSummary = dto.UserSummary{
			UserID:   activity.User.ID,
			Username: activity.User.Username,
			Avatar:   activity.User.AvatarURL,
		}
	}

	return dto.ActivitySummary{
		ID:         uint64(activity.ID),
		ActionType: activity.ActionType,
		TargetType: activity.TargetType,
		TargetID:   activity.TargetID,
		Preview:    activity.Preview,
		CreatedAt:  activity.CreatedAt,
		User:       userSummary,
	}
}