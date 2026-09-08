package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/gin-gonic/gin"
)

type apiErrorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, apiErrorBody{Error: apiError{
		Code:      code,
		Message:   message,
		RequestID: RequestIDFrom(c.Request.Context()),
	}})
}

func writeDomainError(c *gin.Context, logger *slog.Logger, err error) {
	switch {
	case domain.IsInvalid(err):
		var v domain.ValidationError
		msg := "invalid request"
		if errors.As(err, &v) && v.Message != "" {
			msg = v.Error()
		}
		writeError(c, http.StatusBadRequest, "INVALID_INPUT", msg)
	case domain.IsNotFound(err):
		var nf domain.NotFoundError
		code := "NOT_FOUND"
		msg := "resource not found"
		if errors.As(err, &nf) {
			switch nf.Resource {
			case "service":
				code = "SERVICE_NOT_FOUND"
				msg = "service not found"
			case "release":
				code = "RELEASE_NOT_FOUND"
				msg = "release not found"
			}
		}
		writeError(c, http.StatusNotFound, code, msg)
	case domain.IsAlreadyExists(err):
		writeError(c, http.StatusConflict, "ALREADY_EXISTS", "resource already exists")
	default:
		logger.ErrorContext(c.Request.Context(), "unhandled error", slog.String("err", err.Error()))
		writeError(c, http.StatusInternalServerError, "INTERNAL", "internal error")
	}
}
