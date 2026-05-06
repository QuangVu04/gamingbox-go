package repositories

import (
    "vault/be/internal/models"

    "gorm.io/gorm"
)

type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
    FindByUsername(username string) (*models.User, error)
    FindBySteamID(steamID string) (*models.User, error)
    Update(user *models.User) error
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    if err := r.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindBySteamID(steamID string) (*models.User, error) {
    var user models.User
    if err := r.db.Where("steam_id = ?", steamID).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
    var user models.User
    if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}
