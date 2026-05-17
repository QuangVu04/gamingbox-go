package database

import (
    "log"
    "vault/be/internal/models"
    "gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
    log.Println("Running database migrations...")

    err := db.AutoMigrate(
        &models.User{},
        &models.RefreshToken{},
		&models.Game{}, &models.Studio{}, &models.Genre{}, &models.Platform{}, &models.GameImg{},
        &models.Review{}, &models.Rating{}, &models.Comment{}, &models.Like{}, &models.Follow{},
        &models.List{}, &models.ListEntry{}, &models.GameLog{}, &models.Notification{}, &models.ActivityLog{},
    )

    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    // Đảm bảo mở rộng enum img_type trong MySQL
    db.Exec("ALTER TABLE game_imgs MODIFY COLUMN img_type ENUM('header', 'screenshot', 'background', 'cover')")

    log.Println("Migrations completed successfully!")
}