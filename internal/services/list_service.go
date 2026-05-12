package services

import (
	"context"
	"log"
	"math"
	"vault/be/internal/models"

	"vault/be/internal/dto"
	"vault/be/internal/dto/mapper"
	"vault/be/internal/repositories"
	redisUtil "vault/be/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type ListService interface {
	GetTrendingLists(ctx context.Context, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error)
	CreateList(ctx context.Context, userID uint, req dto.CreateListRequest) (*dto.ListDetailResponse, error)
	UpdateList(ctx context.Context, userID, listID uint, req dto.UpdateListRequest) (*dto.ListDetailResponse, error)
	DeleteList(ctx context.Context, userID, listID uint) error
	GetListDetail(ctx context.Context, listID uint) (*dto.ListDetailResponse, error)
	GetGameLists(ctx context.Context, gameID uint, page, limit int, sort string) (*dto.GameListsResponse, error)
}

type listService struct {
	listRepo repositories.ListRepository
	rdb      *redis.Client
}

func NewListService(listRepo repositories.ListRepository, rdb *redis.Client) ListService {
	return &listService{
		listRepo: listRepo,
		rdb:      rdb,
	}
}

// CachedListResponse is used for marshaling/unmarshaling list data with pagination
type CachedListResponse struct {
	Data       []dto.ListTrendingResponse `json:"data"`
	Pagination *dto.PaginationDTO         `json:"pagination"`
}

func (s *listService) GetTrendingLists(ctx context.Context, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Try to get from cache first
	cacheKey := redisUtil.GetTrendingCacheKey("lists", page, limit)
	cached, err := redisUtil.GetCached[CachedListResponse](ctx, s.rdb, cacheKey, CacheTTL)
	if err == nil && cached != nil {
		log.Printf("✓ Cache hit for trending lists (page=%d, limit=%d)", page, limit)
		return cached.Data, cached.Pagination, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for trending lists (page=%d, limit=%d), fetching from database", page, limit)

	// Fetch trending lists from repository (already includes weekly likes count)
	listsData, total, err := s.listRepo.GetTrendingLists(page, limit)
	if err != nil {
		return nil, nil, err
	}

	// Map to response DTOs
	responses := mapper.ToTrendingListResponses(listsData)

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &dto.PaginationDTO{
		TotalRecords: int(total),
		CurrentPage:  page,
		TotalPages:   totalPages,
		Limit:        limit,
	}

	// Cache the response
	cacheData := &CachedListResponse{
		Data:       responses,
		Pagination: pagination,
	}
	err = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)
	if err != nil {
		log.Printf("⚠ Failed to cache trending lists: %v", err)
	}

	return responses, pagination, nil
}

func (s *listService) CreateList(ctx context.Context, userID uint, req dto.CreateListRequest) (*dto.ListDetailResponse, error) {
	entries := make([]models.ListEntry, 0, len(req.GameIDs))
	for i, gameID := range req.GameIDs {
		entries = append(entries, models.ListEntry{
			GameID:   gameID,
			Position: i + 1,
		})
	}

	list := &models.List{
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailImg: req.ThumbnailImg,
		IsPublic:     req.IsPublic,
		GameCount:    len(req.GameIDs),
		Entries:      entries,
	}

	if err := s.listRepo.Create(list); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể tạo danh sách")
	}

	return s.GetListDetail(ctx, list.ID)
}

func (s *listService) UpdateList(ctx context.Context, userID, listID uint, req dto.UpdateListRequest) (*dto.ListDetailResponse, error) {
	list, err := s.listRepo.FindByID(listID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	if list.UserID != userID {
		return nil, dto.NewServiceError("FORBIDDEN", "không có quyền chỉnh sửa")
	}

	if req.Title != "" {
		list.Title = req.Title
	}
	if req.Description != "" {
		list.Description = req.Description
	}
	if req.ThumbnailImg != "" {
		list.ThumbnailImg = req.ThumbnailImg
	}
	if req.IsPublic != nil {
		list.IsPublic = *req.IsPublic
	}

	if len(req.GameIDs) > 0 {
		entries := make([]models.ListEntry, 0, len(req.GameIDs))
		for i, gameID := range req.GameIDs {
			entries = append(entries, models.ListEntry{
				ListID:   listID,
				GameID:   gameID,
				Position: i + 1,
			})
		}
		list.Entries = entries
		list.GameCount = len(req.GameIDs)
	}

	if err := s.listRepo.Update(list); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể cập nhật danh sách")
	}

	return s.GetListDetail(ctx, listID)
}

func (s *listService) DeleteList(ctx context.Context, userID, listID uint) error {
	list, err := s.listRepo.FindByID(listID)
	if err != nil {
		return dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	if list.UserID != userID {
		return dto.NewServiceError("FORBIDDEN", "không có quyền xóa")
	}

	return s.listRepo.Delete(listID)
}

func (s *listService) GetListDetail(ctx context.Context, listID uint) (*dto.ListDetailResponse, error) {
	list, err := s.listRepo.FindDetailByID(listID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	return mapper.ToListDetailResponse(list), nil
}
func (s *listService) GetGameLists(ctx context.Context, gameID uint, page, limit int, sort string) (*dto.GameListsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	lists, total, err := s.listRepo.GetGameLists(gameID, page, limit, sort)
	if err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy danh sách list")
	}

	responses := mapper.ToTrendingListResponsesFromModels(lists)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &dto.GameListsResponse{
		Lists: responses,
		Pagination: dto.PaginationDTO{
			TotalRecords: int(total),
			CurrentPage:  page,
			TotalPages:   totalPages,
			Limit:        limit,
		},
	}, nil
}
