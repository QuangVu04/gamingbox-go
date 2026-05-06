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
    ToggleFollow(followerID, followingID uint) (bool, error)
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

func (r *userRepository) ToggleFollow(followerID, followingID uint) (bool, error) {
    var follow models.Follow
    err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error

    if err == nil {
        // Đã follow -> Hủy follow
        err = r.db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Delete(&follow).Error; err != nil {
                return err
            }
            if err := tx.Model(&models.User{}).Where("id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
                return err
            }
            if err := tx.Model(&models.User{}).Where("id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
                return err
            }
            return nil
        })
        return false, err
    } else if err == gorm.ErrRecordNotFound {
        // Chưa follow -> Thêm follow
        newFollow := models.Follow{FollowerID: followerID, FollowingID: followingID}
        err = r.db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Create(&newFollow).Error; err != nil {
                return err
            }
            if err := tx.Model(&models.User{}).Where("id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
                return err
            }
            if err := tx.Model(&models.User{}).Where("id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
                return err
            }
            return nil
        })
        return true, err
    }
    
    return false, err
}
