package seeders

import (
	"log"

	"vault/be/internal/models"
	"vault/be/pkg/utils"

	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) {
	email := "admin@gmail.com"

	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		log.Println("Admin already exists, skipping seed")
		return
	}

	hashedPassword, err := utils.HashPassword("admin12345678")
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	admin := &models.User{
		Email:        email,
		Username:     "admin",
		Password: 	  &hashedPassword,
		Role:         models.RoleAdmin,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	log.Println("Admin seeded successfully")
}