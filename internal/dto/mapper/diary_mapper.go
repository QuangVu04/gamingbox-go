package mapper

import (
	"vault/be/internal/dto"
	"vault/be/internal/models"
)

func ToDiaryEntry(log *models.GameLog) dto.DiaryEntry {
	return dto.DiaryEntry{
		Game:     ToGameSummary(&log.Game),
		Status:   log.Status,
		LoggedAt: log.LoggedAt,
	}
}