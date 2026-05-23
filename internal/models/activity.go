package models

import (
	"time"

	"gorm.io/gorm"
)

type List struct {
	gorm.Model
	UserID       uint   `gorm:"index"`
	Title        string `gorm:"not null"`
	Description  string `gorm:"type:text"`
	ThumbnailImg string
	IsPublic     bool `gorm:"default:true"`
	LikeCount    int  `gorm:"default:0"`
	GameCount    int  `gorm:"default:0"`

	User    User        `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	Entries []ListEntry `gorm:"foreignKey:ListID;constraint:OnDelete:CASCADE;"`
}

type ListEntry struct {
	ListID   uint   `gorm:"primaryKey"`
	GameID   uint   `gorm:"primaryKey"`
	GhiChu   string `gorm:"column:ghichu"`
	Position int

	Game Game `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE;"`
}

type GameLog struct {
	UserID   uint      `gorm:"primaryKey"`
	GameID   uint      `gorm:"primaryKey"`
	LoggedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Status   string    `gorm:"type:enum('playing', 'completed', 'dropped', 'backlog')"`

	Game Game `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE;"`
}

type Notification struct {
	ID         uint64 `gorm:"primaryKey"`
	ReceiverID uint   `gorm:"index"`
	SenderID   uint   `gorm:"index"`
	ActionType string
	TargetID   uint
	TargetType string
	IsRead     bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type ActivityLog struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"index"`
	ActionType string `gorm:"type:enum('create', 'update', 'delete', 'like', 'follow')"`
	TargetID   uint
	TargetType string
	Preview    string
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	User       User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}
