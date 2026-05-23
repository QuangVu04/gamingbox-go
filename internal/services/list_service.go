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
	GetTrendingLists(ctx context.Context, userID uint, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error)
	CreateList(ctx context.Context, userID uint, req dto.CreateListRequest) (*dto.ListDetailResponse, error)
	UpdateList(ctx context.Context, userID, listID uint, req dto.UpdateListRequest) (*dto.ListDetailResponse, error)
	DeleteList(ctx context.Context, userID, listID uint) error
	GetListDetail(ctx context.Context, userID uint, listID uint) (*dto.ListDetailResponse, error)
	GetGameLists(ctx context.Context, userID uint, gameID uint, page, limit int, sort string) (*dto.GameListsResponse, error)
	GetListComments(ctx context.Context, listID uint) ([]dto.CommentResponse, error)
	AddListComment(ctx context.Context, userID, listID uint, req dto.CommentRequest) (*dto.CommentResponse, error)
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

func (s *listService) GetTrendingLists(ctx context.Context, userID uint, page, limit int) ([]dto.ListTrendingResponse, *dto.PaginationDTO, error) {
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
		if userID > 0 {
			s.populateLikedStatus(userID, cached.Data)
		}
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
	_ = redisUtil.SetCached(ctx, s.rdb, cacheKey, cacheData, CacheTTL)

	if userID > 0 {
		s.populateLikedStatus(userID, responses)
	}

	return responses, pagination, nil
}

func (s *listService) CreateList(ctx context.Context, userID uint, req dto.CreateListRequest) (*dto.ListDetailResponse, error) {
	var entries []models.ListEntry
	var gameCount int

	if len(req.Entries) > 0 {
		for i, entry := range req.Entries {
			entries = append(entries, models.ListEntry{
				GameID:   entry.GameID,
				Position: i + 1,
				GhiChu:   entry.Note,
			})
		}
		gameCount = len(req.Entries)
	} else if len(req.GameIDs) > 0 {
		for i, gameID := range req.GameIDs {
			entries = append(entries, models.ListEntry{
				GameID:   gameID,
				Position: i + 1,
			})
		}
		gameCount = len(req.GameIDs)
	}

	list := &models.List{
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailImg: req.ThumbnailImg,
		IsPublic:     req.IsPublic,
		GameCount:    gameCount,
		Entries:      entries,
	}

	if err := s.listRepo.Create(list); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể tạo danh sách")
	}
	s.clearTrendingListsCache(ctx)

	return s.GetListDetail(ctx, userID, list.ID)
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

	if len(req.Entries) > 0 {
		entries := make([]models.ListEntry, 0, len(req.Entries))
		for i, entry := range req.Entries {
			entries = append(entries, models.ListEntry{
				ListID:   listID,
				GameID:   entry.GameID,
				Position: i + 1,
				GhiChu:   entry.Note,
			})
		}
		list.Entries = entries
		list.GameCount = len(req.Entries)
	} else if len(req.GameIDs) > 0 {
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
	s.clearTrendingListsCache(ctx)

	return s.GetListDetail(ctx, userID, listID)
}

func (s *listService) DeleteList(ctx context.Context, userID, listID uint) error {
	list, err := s.listRepo.FindByID(listID)
	if err != nil {
		return dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	if list.UserID != userID {
		return dto.NewServiceError("FORBIDDEN", "không có quyền xóa")
	}

	if err := s.listRepo.Delete(listID); err != nil {
		return err
	}
	s.clearTrendingListsCache(ctx)
	return nil
}

func (s *listService) GetListDetail(ctx context.Context, userID uint, listID uint) (*dto.ListDetailResponse, error) {
	list, err := s.listRepo.FindDetailByID(listID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	comments, _ := s.listRepo.GetComments(listID)
	res := mapper.ToListDetailResponse(list, len(comments))
	if userID > 0 {
		res.IsLiked = s.listRepo.IsListLiked(userID, list.ID)
	}
	return res, nil
}

func (s *listService) GetGameLists(ctx context.Context, userID uint, gameID uint, page, limit int, sort string) (*dto.GameListsResponse, error) {
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

	listIDs := make([]uint, len(lists))
	for i, l := range lists {
		listIDs[i] = l.ID
	}
	commentCounts, _ := s.listRepo.GetCommentCounts(listIDs)

	responses := mapper.ToTrendingListResponsesFromModels(lists, commentCounts)
	if userID > 0 {
		s.populateLikedStatus(userID, responses)
	}

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

func (s *listService) GetListComments(ctx context.Context, listID uint) ([]dto.CommentResponse, error) {
	comments, err := s.listRepo.GetComments(listID)
	if err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể lấy bình luận")
	}

	usersMap := make(map[uint]models.User)
	for _, c := range comments {
		if c.User.ID != 0 {
			usersMap[c.UserID] = c.User
		}
	}

	return mapper.ToCommentResponses(comments, usersMap, nil), nil
}

func (s *listService) AddListComment(ctx context.Context, userID, listID uint, req dto.CommentRequest) (*dto.CommentResponse, error) {
	_, err := s.listRepo.FindByID(listID)
	if err != nil {
		return nil, dto.NewServiceError("NOT_FOUND", "không tìm thấy danh sách")
	}

	comment := &models.Comment{
		ListID:   &listID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := s.listRepo.AddComment(comment); err != nil {
		return nil, dto.NewServiceError("DATABASE_ERROR", "không thể thêm bình luận")
	}

	// Fetch again to get User info preloaded (simpler way)
	comments, _ := s.listRepo.GetComments(listID)
	for _, c := range comments {
		if c.ID == comment.ID {
			s.clearTrendingListsCache(ctx)
			res := mapper.ToCommentResponse(&c, &c.User, false)
			return res, nil
		}
	}

	return nil, dto.NewServiceError("SERVER_ERROR", "lỗi không xác định")
}

func (s *listService) populateLikedStatus(userID uint, lists []dto.ListTrendingResponse) {
	if len(lists) == 0 {
		return
	}
	var listIDs []uint
	for _, l := range lists {
		listIDs = append(listIDs, l.ListID)
	}
	likedIDs := s.listRepo.GetLikedListIDs(userID, listIDs)
	likedMap := make(map[uint]bool)
	for _, id := range likedIDs {
		likedMap[id] = true
	}
	for i := range lists {
		lists[i].UserHasLiked = likedMap[lists[i].ListID]
	}
}

func (s *listService) clearTrendingListsCache(ctx context.Context) {
	if s.rdb == nil {
		return
	}

	pattern := "trending:lists:*"
	iter := s.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		_ = s.rdb.Del(ctx, iter.Val())
	}
}
