package models

import "time"

type UserRole string

const (
    RoleUser       UserRole = "user"
    RoleInfluencer UserRole = "influencer"
    RoleAdmin      UserRole = "admin"
)

type User struct {
    ID        uint      `gorm:"primarykey;autoIncrement" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Email        string   `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
    Password string   `gorm:"type:varchar(255);not null"             json:"-"`
    Username     string   `gorm:"type:varchar(50);uniqueIndex;not null"  json:"username"`
    AvatarURL    *string  `gorm:"type:text"                              json:"avatar_url"`
    Bio          *string  `gorm:"type:text"                              json:"bio"`
    Role         UserRole `gorm:"type:enum('user','influencer','admin');default:'user'" json:"role"`
    SteamID        string    `gorm:"default:null"                     json:"steam_id"`
    FollowerCount  int     `gorm:"default:0" json:"follower_count"`
    FollowingCount int     `gorm:"default:0" json:"following_count"`
    ReviewCount    int     `gorm:"default:0" json:"review_count"`
    ListCount      int     `gorm:"default:0" json:"list_count"`
    GameLogsCount  int     `gorm:"default:0" json:"game_logs_count"`
    AverageRating  float64 `gorm:"type:decimal(3,2);default:0" json:"average_rating_count"`
    Status         string  `gorm:"type:varchar(20);default:'active'" json:"status"`
}
