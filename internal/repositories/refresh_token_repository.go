package repositories

import (
    "time"

    "vault/be/internal/models"

    "gorm.io/gorm"
)

type RefreshTokenRepository interface {
    Save(token *models.RefreshToken) error
    Find(tokenString string) (*models.RefreshToken, error)
    Revoke(tokenString string) error
}

type refreshTokenRepository struct {
    db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
    return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(token *models.RefreshToken) error {
    return r.db.Create(token).Error
}

func (r *refreshTokenRepository) Find(tokenString string) (*models.RefreshToken, error) {
    var token models.RefreshToken
    err := r.db.
        Where("token = ? AND revoked = false AND expires_at > ?", tokenString, time.Now()).
        First(&token).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

func (r *refreshTokenRepository) Revoke(tokenString string) error {
    return r.db.Model(&models.RefreshToken{}).
        Where("token = ?", tokenString).
        Update("revoked", true).
        Error
}