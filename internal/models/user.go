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
    Password     *string  `gorm:"type:varchar(255)"                      json:"-"`
    Username     string   `gorm:"type:varchar(50);uniqueIndex;not null"  json:"username"`
    AvatarURL    *string  `gorm:"type:text"                              json:"avatar_url"`
    Bio          *string  `gorm:"type:text"                              json:"bio"`
    Location     *string  `gorm:"type:varchar(255)"                      json:"location"`
    Role         UserRole `gorm:"type:enum('user','influencer','admin');default:'user'" json:"role"`
    SteamID      string   `gorm:"default:null;index"                     json:"steam_id"`
    GoogleID     string   `gorm:"default:null;index"                     json:"google_id"`
    FacebookID   string   `gorm:"default:null;index"                     json:"facebook_id"`
    FollowerCount  int     `gorm:"default:0" json:"follower_count"`
    FollowingCount int     `gorm:"default:0" json:"following_count"`
    ReviewCount    int     `gorm:"default:0" json:"review_count"`
    ListCount      int     `gorm:"default:0" json:"list_count"`
    GameLogsCount  int     `gorm:"default:0" json:"game_logs_count"`
    AverageRating  float64 `gorm:"type:decimal(3,2);default:0" json:"average_rating_count"`
    Status         string  `gorm:"type:varchar(20);default:'active'" json:"status"`
    FavoriteGameIDs []uint `gorm:"type:json;serializer:json;default:null" json:"favorite_game_ids"`
}
