package handlers

import (
	"net/http"

	"vault/be/internal/middleware"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "Unauthorized")
		return
	}

	result, err := h.userService.GetUserProfile(userID)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, result)
}