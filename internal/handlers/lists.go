package handlers

import (
	"context"
	"net/http"

	"vault/be/internal/dto"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ListHandler struct {
	listService services.ListService
}

func NewListHandler(listService services.ListService) *ListHandler {
	return &ListHandler{
		listService: listService,
	}
}

func (h *ListHandler) TrendingLists(c *gin.Context) {
	// Parse pagination parameters using request utility
	page := utils.GetQueryIntWithRange(c, "page", 1, 1, 1000)
	limit := utils.GetQueryIntWithRange(c, "limit", 10, 1, 100)

	// Call service with caching
	ctx := context.Background()
	lists, pagination, err := h.listService.GetTrendingLists(ctx, page, limit)
	if err != nil {
		if serviceErr, ok := err.(*dto.ServiceError); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"code":   serviceErr.Code,
				"error":  serviceErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"code":   "SERVER_ERROR",
			"error":  "đã xảy ra lỗi",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"pagination": pagination,
		"data":       lists,
	})
}
