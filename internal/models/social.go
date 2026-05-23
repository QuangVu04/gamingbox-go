package models

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID     uint   `gorm:"index"`
	TargetID   uint   `gorm:"index"`
	TargetType string `gorm:"type:enum('game', 'list', 'news')"`
	Content    string `gorm:"type:text"`
	LikeCount  int    `gorm:"default:0"`
	Recommend  string `gorm:"type:enum('recommend', 'mixed', 'not_recommend')"`
	IsSpoiler  bool   `gorm:"default:false"`
	Rating     float64 `gorm:"-" json:"rating"`

	Game Game `gorm:"foreignKey:TargetID;references:ID"`
	User User `gorm:"foreignKey:UserID;references:ID"`
}

type Rating struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_game"`
	GameID    uint      `gorm:"uniqueIndex:idx_user_game"`
	Rating    float64   `gorm:"type:float"`
	CreatedAt time.Time
}

type Comment struct {
	gorm.Model
	ReviewID  *uint  `gorm:"index"` // Nullable to support List comments
	ListID    *uint  `gorm:"index"` // Nullable to support List comments
	UserID    uint   `gorm:"index"`
	Content   string `gorm:"type:text"`
	ParentID  *uint  `gorm:"index"` // Nullable cho comment cấp 1
	LikeCount int    `gorm:"default:0"`

	User User `gorm:"foreignKey:UserID;references:ID"`
}

type Like struct {
	UserID     uint      `gorm:"primaryKey"`
	TargetID   uint      `gorm:"primaryKey"`
	TargetType string    `gorm:"primaryKey"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	// Composite unique index để đảm bảo một user chỉ like một item một lần
	// Indexed as: UNIQUE KEY (user_id, target_id, target_type)
}

type Follow struct {
	FollowerID  uint      `gorm:"primaryKey"`
	FollowingID uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
