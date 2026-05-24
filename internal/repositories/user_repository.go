package repositories

import (
    "time"
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
    GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error)
    GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error)
    Count() (int64, error)
    CountRecent(since time.Time) (int64, error)
    FindByField(field string, value interface{}) (*models.User, error)
    GetAdminUsers(page, limit int, search, role, sort string) ([]models.User, int64, error)
    SearchUsers(search string, page, limit int, sortBy string) ([]models.User, int64, error)
    Delete(id uint) error
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

func (r *userRepository) GetFollowing(userID uint, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total following
	if err := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch users being followed
	err := r.db.
		Joins("JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) GetFollowers(userID uint, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total followers
	if err := r.db.Model(&models.Follow{}).Where("following_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch users who are following this user
	err := r.db.
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.following_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) CountRecent(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *userRepository) FindByField(field string, value interface{}) (*models.User, error) {
    var user models.User
    if err := r.db.Where(field+" = ?", value).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) GetAdminUsers(page, limit int, search, role, sort string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.User{})

	if search != "" {
		db = db.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if role != "" && role != "All" {
		db = db.Where("role = ?", role)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC"
	if sort == "recent" {
		orderClause = "(game_logs_count * 5 + review_count * 10) DESC"
	} else if sort == "alltime" {
		orderClause = "(follower_count * 10 + review_count * 15) DESC"
	}

	if err := db.Order(orderClause).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find all lists of this user
		var listIDs []uint
		if err := tx.Model(&models.List{}).Where("user_id = ?", id).Pluck("id", &listIDs).Error; err != nil {
			return err
		}

		// 2. Delete all list entries belonging to those lists
		if len(listIDs) > 0 {
			if err := tx.Exec("DELETE FROM list_entries WHERE list_id IN ?", listIDs).Error; err != nil {
				return err
			}
		}

		// 3. Delete all lists
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.List{}).Error; err != nil {
			return err
		}

		// 4. Find all reviews of this user
		var reviewIDs []uint
		if err := tx.Model(&models.Review{}).Where("user_id = ?", id).Pluck("id", &reviewIDs).Error; err != nil {
			return err
		}

		// 5. Delete comments on user's reviews, or comments written by user
		if len(reviewIDs) > 0 {
			if err := tx.Unscoped().Where("review_id IN ?", reviewIDs).Delete(&models.Comment{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
			return err
		}

		// 6. Delete likes on user's reviews, or likes written by user
		if len(reviewIDs) > 0 {
			if err := tx.Exec("DELETE FROM likes WHERE target_type = 'review' AND target_id IN ?", reviewIDs).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM likes WHERE user_id = ?", id).Error; err != nil {
			return err
		}

		// 7. Delete all reviews of this user
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.Review{}).Error; err != nil {
			return err
		}

		// 8. Delete all ratings by this user
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.Rating{}).Error; err != nil {
			return err
		}

		// 9. Delete all game logs by this user
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.GameLog{}).Error; err != nil {
			return err
		}

		// 10. Delete follows where user is follower or following
		if err := tx.Unscoped().Where("follower_id = ? OR following_id = ?", id, id).Delete(&models.Follow{}).Error; err != nil {
			return err
		}

		// 11. Delete notifications received or sent by this user
		if err := tx.Unscoped().Where("receiver_id = ? OR sender_id = ?", id, id).Delete(&models.Notification{}).Error; err != nil {
			return err
		}

		// 12. Delete activity logs of this user
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.ActivityLog{}).Error; err != nil {
			return err
		}

		// 13. Delete refresh tokens of this user
		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}

		// 14. Delete the user record
		if err := tx.Unscoped().Delete(&models.User{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *userRepository) SearchUsers(search string, page, limit int, sortBy string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&models.User{}).Where("status = ?", "active")

	if search != "" {
		db = db.Where("username LIKE ? OR bio LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC"
	if sortBy == "active" {
		orderClause = "(game_logs_count + review_count) DESC"
	} else if sortBy == "followers" {
		orderClause = "follower_count DESC"
	} else if sortBy == "newest" {
		orderClause = "created_at DESC"
	}

	if err := db.Order(orderClause).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
