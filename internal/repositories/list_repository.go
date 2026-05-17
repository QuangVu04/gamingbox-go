package repositories

import (
	// "fmt"
	"time"
	"vault/be/internal/models"

	"gorm.io/gorm"
)

// ListTrendingData holds list data with weekly likes count
type ListTrendingData struct {
	List             models.List
	WeeklyLikesCount int64
}

type WeeklyCount struct {
        TargetID uint
        Count    int64
    }

type ListRepository interface {
	GetByUserID(userID uint) ([]models.List, error)
	GetByUserIDPaginated(userID uint, page, limit int) ([]models.List, int64, error)
	GetBacklogByUserID(userID uint) (*models.List, error)
	GetBacklogEntries(listID uint, limit int) ([]models.ListEntry, error)
	GetTrendingLists(page, limit int) ([]ListTrendingData, int64, error)
	FindByID(id uint) (*models.List, error)
	FindDetailByID(id uint) (*models.List, error)
	Create(list *models.List) error
	Update(list *models.List) error
	Delete(id uint) error
	GetGameLists(gameID uint, page, limit int, sort string) ([]models.List, int64, error)
}

type listRepository struct {
	db *gorm.DB
}

func NewListRepository(db *gorm.DB) ListRepository {
	return &listRepository{db: db}
}

func (r *listRepository) GetByUserID(userID uint) ([]models.List, error) {
	var lists []models.List
	err := r.db.Preload("Entries").Preload("Entries.Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND is_public = ? AND title NOT IN ?", userID, true, []string{"Backlog", "Muốn chơi"}).
		Order("created_at desc").Find(&lists).Error
	return lists, err
}

func (r *listRepository) GetByUserIDPaginated(userID uint, page, limit int) ([]models.List, int64, error) {
	var lists []models.List
	var total int64
	offset := (page - 1) * limit
	db := r.db.Model(&models.List{}).Where("user_id = ? AND is_public = ? AND title NOT IN ?", userID, true, []string{"Backlog", "Muốn chơi"})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Entries").Preload("Entries.Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("user_id = ? AND is_public = ? AND title NOT IN ?", userID, true, []string{"Backlog", "Muốn chơi"}).
		Order("created_at desc").Offset(offset).Limit(limit).Find(&lists).Error
	return lists, total, err
}

func (r *listRepository) GetBacklogByUserID(userID uint) (*models.List, error) {
	var list models.List
	err := r.db.Where("user_id = ? AND title IN ?", userID, []string{"Backlog", "Muốn chơi"}).
		Order("created_at desc").First(&list).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *listRepository) GetBacklogEntries(listID uint, limit int) ([]models.ListEntry, error) {
	var entries []models.ListEntry
	err := r.db.Preload("Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("list_id = ?", listID).
		Order("position asc").Limit(limit).Find(&entries).Error
	return entries, err
}

// GetTrendingLists retrieves trending lists ordered by weekly likes count
func (r *listRepository) GetTrendingLists(page, limit int) ([]ListTrendingData, int64, error) {
    var lists []models.List
    var total int64
    offset := (page - 1) * limit
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)

    // 1. Đếm tổng số list công khai
    if err := r.db.Model(&models.List{}).Where("is_public = ?", true).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    err := r.db.
        Preload("User").
        // Lưu ý: Bỏ Limit(5) ở đây để tránh lỗi mất data, 
        // chúng ta sẽ xử lý cắt slice (truncate) ở phần sau hoặc dùng SQL đặc biệt.
        Preload("Entries", func(db *gorm.DB) *gorm.DB {
            return db.Order("position ASC")
        }).
        Preload("Entries.Game.Images", "img_type IN ?", []string{"header", "cover"}).
        Where("is_public = ?", true).
        Order(r.db.Model(&models.Like{}).Select("COUNT(*)").Where("target_id = lists.id AND target_type = 'list' AND created_at > ?", sevenDaysAgo).Order("COUNT(*) DESC")).
        Order("like_count DESC").
        Offset(offset).
        Limit(limit).
        Find(&lists).Error

    if err != nil {
        return nil, 0, err
    }

    // 3. Giải quyết N+1 bằng cách lấy toàn bộ WeeklyLikes trong 1 lần query
    listIDs := make([]uint, len(lists))
    for i, l := range lists {
        listIDs[i] = l.ID
    }

    var counts []WeeklyCount
    r.db.Table("likes").
        Select("target_id, COUNT(*) as count").
        Where("target_id IN ? AND target_type = 'list' AND created_at > ?", listIDs, sevenDaysAgo).
        Group("target_id").
        Scan(&counts)

    // Map lại kết quả để truy xuất nhanh
    countMap := make(map[uint]int64)
    for _, c := range counts {
        countMap[c.TargetID] = c.Count
    }

    // 4. Build dữ liệu trả về và giới hạn 5 entries mỗi list bằng code Go
    listsData := make([]ListTrendingData, 0, len(lists))
    for _, list := range lists {
        // Giới hạn 5 entries tại đây (an toàn và chính xác hơn Limit của GORM)
        if len(list.Entries) > 5 {
            list.Entries = list.Entries[:5]
        }

        listsData = append(listsData, ListTrendingData{
            List:             list,
            WeeklyLikesCount: countMap[list.ID],
        })
    }

    return listsData, total, nil
}

func (r *listRepository) FindByID(id uint) (*models.List, error) {
	var list models.List
	err := r.db.First(&list, id).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *listRepository) FindDetailByID(id uint) (*models.List, error) {
	var list models.List
	err := r.db.Preload("User").
		Preload("Entries.Game.Images", "img_type IN ?", []string{"header", "cover"}).
		First(&list, id).Error
	return &list, err
}

func (r *listRepository) Create(list *models.List) error {
	return r.db.Create(list).Error
}

func (r *listRepository) Update(list *models.List) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update main list info
		if err := tx.Save(list).Error; err != nil {
			return err
		}

		// Handle entries update
		if len(list.Entries) > 0 {
			// Delete old entries
			if err := tx.Where("list_id = ?", list.ID).Delete(&models.ListEntry{}).Error; err != nil {
				return err
			}
			// Create new entries
			if err := tx.Create(&list.Entries).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *listRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("list_id = ?", id).Delete(&models.ListEntry{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.List{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}
func (r *listRepository) GetGameLists(gameID uint, page, limit int, sort string) ([]models.List, int64, error) {
	var lists []models.List
	var total int64
	offset := (page - 1) * limit

	// Subquery to find list IDs containing the game
	subQuery := r.db.Model(&models.ListEntry{}).Select("list_id").Where("game_id = ?", gameID)

	db := r.db.Model(&models.List{}).
		Preload("User").
		Preload("Entries", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Entries.Game.Images", "img_type IN ?", []string{"header", "cover"}).
		Where("id IN (?) AND is_public = ?", subQuery, true)

	// Get total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort logic
	switch sort {
	case "list_name":
		db = db.Order("title ASC")
	case "popularity":
		db = db.Order("like_count DESC")
	case "recently_updated":
		db = db.Order("updated_at DESC")
	case "oldest":
		db = db.Order("created_at ASC")
	default: // newest
		db = db.Order("created_at DESC")
	}

	err := db.Offset(offset).Limit(limit).Find(&lists).Error
	
	// Truncate entries to first 5 for thumbnails in response
	for i := range lists {
		if len(lists[i].Entries) > 5 {
			lists[i].Entries = lists[i].Entries[:5]
		}
	}

	return lists, total, err
}
