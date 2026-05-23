package models

import "time"

type RefreshToken struct {
    ID        uint      `gorm:"primarykey;autoIncrement"`
    CreatedAt time.Time
    UserID     uint    `gorm:"not null;index"`
    User       User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
    Token      string  `gorm:"type:varchar(500);not null;uniqueIndex"`
    ExpiresAt  time.Time `gorm:"not null"`
    Revoked    bool    `gorm:"default:false"`
}