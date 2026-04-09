package repositories

import (
	"vault/be/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LikeRepository interface {
	// Create adds a new like
	Create(like *models.Like) error

	// Delete removes a like
	Delete(userID, gameID uint) error

	// Exists checks if a like exists
	Exists(userID, gameID uint) (bool, error)

	// FindByUserAndGame retrieves a specific like
	FindByUserAndGame(userID, gameID uint) (*models.Like, error)

	// ToggleLike performs Like/Unlike with atomic counter update (polymorphic)
	ToggleLike(userID uint, targetID uint, targetType string) (isLiked bool, err error)

	// CheckLike checks if a user has liked a target (polymorphic)
	CheckLike(userID, targetID uint, targetType string) (bool, error)

	// GetLikeCount gets the like count for a target (polymorphic)
	GetLikeCount(targetID uint, targetType string) (int, error)
}

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(like *models.Like) error {
	return r.db.Create(like).Error
}

func (r *likeRepository) Delete(userID, gameID uint) error {
	return r.db.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, gameID, "game").
		Delete(&models.Like{}).Error
}

func (r *likeRepository) Exists(userID, gameID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("user_id = ? AND target_id = ? AND target_type = ?", userID, gameID, "game").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *likeRepository) FindByUserAndGame(userID, gameID uint) (*models.Like, error) {
	var like models.Like
	err := r.db.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, gameID, "game").
		First(&like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &like, nil
}

// ToggleLike handles Like/Unlike with atomic counter update
// Returns isLiked: true if just liked, false if just unliked
func (r *likeRepository) ToggleLike(userID uint, targetID uint, targetType string) (bool, error) {
	var isLiked bool

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Check if like exists
		var like models.Like
		checkErr := tx.Where("user_id = ? AND target_id = ? AND target_type = ?",
			userID, targetID, targetType).
			First(&like).Error

		if checkErr == gorm.ErrRecordNotFound {
			// Not liked, so create like
			like = models.Like{
				UserID:     userID,
				TargetID:   targetID,
				TargetType: targetType,
			}
			if err := tx.Create(&like).Error; err != nil {
				return err
			}
			isLiked = true

			// Increment counter in target table using atomic expression
			return r.incrementLikeCount(tx, targetID, targetType)
		} else if checkErr != nil {
			return checkErr
		}

		// Like exists, so delete it
		if err := tx.Delete(&like).Error; err != nil {
			return err
		}
		isLiked = false

		// Decrement counter in target table using atomic expression
		return r.decrementLikeCount(tx, targetID, targetType)
	})

	return isLiked, err
}

// CheckLike checks if user has already liked the target
func (r *likeRepository) CheckLike(userID, targetID uint, targetType string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).
		Count(&count).Error

	return count > 0, err
}

// GetLikeCount returns the total number of likes for a target
func (r *likeRepository) GetLikeCount(targetID uint, targetType string) (int, error) {
	var count int
	err := r.db.Model(&models.Like{}).
		Where("target_id = ? AND target_type = ?", targetID, targetType).
		Select("COALESCE(COUNT(*), 0)").
		Scan(&count).Error

	return count, err
}

// Helper: Increment like_count using atomic expression
func (r *likeRepository) incrementLikeCount(tx *gorm.DB, targetID uint, targetType string) error {
	switch targetType {
	case "game":
		return tx.Model(&models.Game{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count + 1"}).Error
	case "review":
		return tx.Model(&models.Review{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count + 1"}).Error
	case "list":
		return tx.Model(&models.List{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count + 1"}).Error
	}
	return nil
}

// Helper: Decrement like_count using atomic expression
func (r *likeRepository) decrementLikeCount(tx *gorm.DB, targetID uint, targetType string) error {
	switch targetType {
	case "game":
		return tx.Model(&models.Game{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count - 1"}).Error
	case "review":
		return tx.Model(&models.Review{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count - 1"}).Error
	case "list":
		return tx.Model(&models.List{}).
			Where("id = ?", targetID).
			Update("like_count", clause.Expr{SQL: "like_count - 1"}).Error
	}
	return nil
}
