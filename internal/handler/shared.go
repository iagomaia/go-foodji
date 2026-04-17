package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/iagomaia/go-foodji/internal/domain"
	"net/http"
)

func statusFromError(err error) int {
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, domain.ErrConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, domain.ErrBadRequest) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
