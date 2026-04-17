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

// upsert godoc
// @Summary      Upsert a vote
// @Description  Registers or updates a vote for a product within a session. Returns 201 on creation, 200 on update.
// @Tags         votes
// @Accept       json
// @Produce      json
// @Param        body  body      domain.UpsertVoteInput  true  "Vote payload"
// @Success      201   {object}  domain.Vote
// @Success      200   {object}  domain.Vote
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /votes [put]
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

	if vote.Created {
		c.JSON(http.StatusCreated, vote)
		return
	}
	c.JSON(http.StatusOK, vote)
}

// getBySessionID godoc
// @Summary      List votes for a session
// @Description  Returns all votes cast within a given session
// @Tags         votes
// @Produce      json
// @Param        session_id  path      string  true  "Session ID"
// @Success      200         {array}   domain.Vote
// @Failure      500         {object}  ErrorResponse
// @Router       /votes/sessions/{session_id} [get]
func (h *VoteHandler) getBySessionID(c *gin.Context) {
	sessionID := c.Param("session_id")
	votes, err := h.svc.GetBySessionID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, votes)
}

// getReport godoc
// @Summary      Vote report
// @Description  Aggregates like/dislike counts and ratios for every product across all sessions
// @Tags         votes
// @Produce      json
// @Success      200  {array}   domain.VoteReportItem
// @Failure      500  {object}  ErrorResponse
// @Router       /votes/report [get]
func (h *VoteHandler) getReport(c *gin.Context) {
	report, err := h.svc.GetVoteReport(c.Request.Context())
	if err != nil {
		c.JSON(statusFromError(err), errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, report)
}
