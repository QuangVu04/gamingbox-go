package models

import (
	"time"

	"gorm.io/gorm"
)

type Game struct {
	gorm.Model
	SteamID         int    `gorm:"index"`
	Title           string `gorm:"not null"`
	Description     string `gorm:"type:text"`
	ReleaseDate     time.Time
	Price           float64
	IsFree          bool    `gorm:"default:false"`
	AvgRating       float64 `gorm:"default:0"`
	ReviewCount     int     `gorm:"default:0"`
	LikeCount       int     `gorm:"default:0"`
	RatingBreakdown []byte  `gorm:"type:json"`
	AveragePlaytime float64
	StudioID        uint

	Studio    Studio     `gorm:"foreignKey:StudioID"`
	Images    []GameImg  `gorm:"foreignKey:GameID"`
	Genres    []Genre    `gorm:"many2many:game_genres;"`
	Platforms []Platform `gorm:"many2many:game_platforms;"`
}

type Studio struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Games       []Game `gorm:"foreignKey:StudioID"`
}

type Genre struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"`
}

type Platform struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);uniqueIndex;not null"`
}

type GameImg struct {
	ID      uint   `gorm:"primaryKey"`
	OgURL   string `gorm:"column:ogUrl"`
	Thumb   string
	ImgType string `gorm:"type:enum('header', 'screenshot', 'background', 'cover')"`
	GameID  uint
}

// GameTrending is a helper struct for trending calculations
type GameTrending struct {
	Game          Game
	TrendingScore int
	ReviewCount7d int
	RatingCount7d int
}
