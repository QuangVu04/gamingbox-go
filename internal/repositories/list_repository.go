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
	GetBacklogByUserID(userID uint) (*models.List, error)
	GetBacklogEntries(listID uint, limit int) ([]models.ListEntry, error)
	GetTrendingLists(page, limit int) ([]ListTrendingData, int64, error)
	FindByID(id uint) (*models.List, error)
}

type listRepository struct {
	db *gorm.DB
}

func NewListRepository(db *gorm.DB) ListRepository {
	return &listRepository{db: db}
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
	err := r.db.Preload("Game.Images", "img_type = ?", "header").
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
        Preload("Entries.Game.Images", "img_type = ?", "header").
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