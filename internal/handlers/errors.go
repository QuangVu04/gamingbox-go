package handlers

import (
	"errors"
	"net/http"
	"vault/be/internal/dto"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
)

func handleAuthServiceError(c *gin.Context, err error) {
	var svcErr *dto.ServiceError
	if !errors.As(err, &svcErr) {
		utils.InternalError(c)
		return
	}

	switch svcErr.Code {
	case "INVALID_CREDENTIALS",
		"INVALID_REFRESH_TOKEN",
		"TOKEN_REVOKED":
		utils.Unauthorized(c, svcErr.Message)

	case "EMAIL_EXISTS",
		"USERNAME_EXISTS":
		utils.Conflict(c, svcErr.Code, svcErr.Message, svcErr.Field)

	case "USERNAME_INVALID":
		utils.ErrorWithField(c, http.StatusBadRequest, svcErr.Code, svcErr.Message, svcErr.Field)

	case "USER_NOT_FOUND":
		utils.NotFound(c, svcErr.Message)

	case "SERVER_ERROR":
		utils.InternalError(c)

	default:
		utils.InternalError(c)
	}
}

func handleUserServiceError(c *gin.Context, err error) {
	var svcErr *dto.ServiceError
	if !errors.As(err, &svcErr) {
		utils.InternalError(c)
		return
	}

	switch svcErr.Code {
	case "USER_NOT_FOUND":
		utils.NotFound(c, svcErr.Message)
	case "SERVER_ERROR":
		utils.InternalError(c)
	default:
		utils.InternalError(c)
	}
}

func handleGameInteractionError(c *gin.Context, err error) {
	if serviceErr, ok := err.(*dto.ServiceError); ok {
		switch serviceErr.Code {
		case "VALIDATION_ERROR":
			utils.ValidationError(c, serviceErr.Message)
		case "SERVER_ERROR":
			utils.InternalError(c)
		case "NOT_FOUND":
			utils.NotFound(c, serviceErr.Message)
		default:
			utils.InternalError(c)
		}
	} else {
		utils.InternalError(c)
	}
}

func handleLikeError(c *gin.Context, err error) {
	// Handle service errors
	if svcErr, ok := err.(*dto.ServiceError); ok {
		switch svcErr.Code {
		case "VALIDATION_ERROR":
			utils.ValidationError(c, svcErr.Message)
		case "SERVER_ERROR":
			utils.InternalError(c)
		default:
			utils.InternalError(c)
		}
		return
	}

	utils.InternalError(c)
}
