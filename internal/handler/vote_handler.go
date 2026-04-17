package handler

import (
	service "github.com/iagomaia/go-foodji/internal/service/vote"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iagomaia/go-foodji/internal/domain"
)

type VoteHandler struct {
	svc service.VoteService
}

func NewVoteHandler(svc service.VoteService) *VoteHandler {
	return &VoteHandler{svc: svc}
}

func (h *VoteHandler) RegisterRoutes(rg *gin.RouterGroup) {
	votes := rg.Group("/votes")
	votes.PUT("", h.upsert)
	votes.GET("sessions/:session_id", h.getBySessionID)
	votes.GET("report", h.getReport)
}

func (h *VoteHandler) upsert(c *gin.Context) {
	var input domain.UpsertVoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	vote, err := h.svc.UpsertVote(c.Request.Context(), input)
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}

	if vote.CreatedAt.Equal(vote.UpdatedAt) {
		c.JSON(http.StatusCreated, vote)
		return
	}
	c.JSON(http.StatusOK, vote)
}

func (h *VoteHandler) getBySessionID(c *gin.Context) {
	sessionID := c.Param("session_id")
	votes, err := h.svc.GetBySessionID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, votes)
}

func (h *VoteHandler) getReport(c *gin.Context) {
	report, err := h.svc.GetVoteReport(c.Request.Context())
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, report)
}
