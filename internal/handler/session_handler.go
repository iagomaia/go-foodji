package handler

import (
	"github.com/iagomaia/go-foodji/internal/service/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	svc session.SessionService
}

func NewSessionHandler(svc session.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	sessions := rg.Group("/sessions")
	sessions.POST("", h.create)
}

func (h *SessionHandler) create(c *gin.Context) {
	createdSession, err := h.svc.CreateSession(c.Request.Context())
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}
	c.JSON(http.StatusCreated, *createdSession)
}
