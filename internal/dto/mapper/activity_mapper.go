package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToActivitySummary(activity *models.ActivityLog) dto.ActivitySummary {
	return dto.ActivitySummary{
		ID:         uint64(activity.ID),
		ActionType: activity.ActionType,
		TargetType: activity.TargetType,
		TargetID:   activity.TargetID,
		Preview:    activity.Preview,
		CreatedAt:  activity.CreatedAt,
	}
}