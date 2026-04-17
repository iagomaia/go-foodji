package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/internal/handler"
	"github.com/iagomaia/go-foodji/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVoteRouter(svc *mocks.VoteServiceMock) *gin.Engine {
	r := gin.New()
	h := handler.NewVoteHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

// --- PUT /votes (upsert) ---

func TestUpsertVote_Created(t *testing.T) {
	now := time.Now().UTC()
	svc := &mocks.VoteServiceMock{
		UpsertVoteFn: func(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
			return &domain.Vote{
				ID:        "v1",
				SessionID: input.SessionID,
				ProductID: input.ProductID,
				VoteType:  input.VoteType,
				CreatedAt: now,
				UpdatedAt: now,
				Created:   true,
			}, nil
		},
	}

	body, _ := json.Marshal(domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Like})
	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/votes", bytes.NewReader(body)))

	assert.Equal(t, http.StatusCreated, w.Code)

	var v domain.Vote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
	assert.Equal(t, "v1", v.ID)
}

func TestUpsertVote_Updated(t *testing.T) {
	past := time.Now().UTC().Add(-time.Hour)
	svc := &mocks.VoteServiceMock{
		UpsertVoteFn: func(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
			return &domain.Vote{
				ID:        "v1",
				VoteType:  input.VoteType,
				CreatedAt: past,
				UpdatedAt: time.Now().UTC(),
				Created:   false,
			}, nil
		},
	}

	body, _ := json.Marshal(domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Dislike})
	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/votes", bytes.NewReader(body)))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpsertVote_BadRequest_MissingFields(t *testing.T) {
	svc := &mocks.VoteServiceMock{}
	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/votes", bytes.NewReader([]byte(`{}`))))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpsertVote_SessionNotFound(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		UpsertVoteFn: func(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
			return nil, domain.ErrNotFound
		},
	}

	body, _ := json.Marshal(domain.UpsertVoteInput{SessionID: "missing", ProductID: "p1", VoteType: domain.Like})
	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/votes", bytes.NewReader(body)))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpsertVote_ServerError(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		UpsertVoteFn: func(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
			return nil, errors.New("db failure")
		},
	}

	body, _ := json.Marshal(domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Like})
	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/votes", bytes.NewReader(body)))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GET /votes/sessions/:session_id ---

func TestGetBySessionID_OK(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		GetBySessionIDFn: func(ctx context.Context, sessionID string) ([]*domain.Vote, error) {
			assert.Equal(t, "s1", sessionID)
			return []*domain.Vote{
				{ID: "v1", SessionID: "s1", ProductID: "p1", VoteType: domain.Like},
			}, nil
		},
	}

	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/votes/sessions/s1", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var votes []*domain.Vote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &votes))
	assert.Len(t, votes, 1)
}

func TestGetBySessionID_ServerError(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		GetBySessionIDFn: func(ctx context.Context, sessionID string) ([]*domain.Vote, error) {
			return nil, errors.New("db failure")
		},
	}

	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/votes/sessions/s1", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GET /votes/report ---

func TestGetReport_OK(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		GetVoteReportFn: func(ctx context.Context) (domain.VoteReportResponse, error) {
			return domain.VoteReportResponse{
				{ProductID: "p1", LikeCount: 3, DislikeCount: 1, LikeRatio: 0.75, DislikeRatio: 0.25},
			}, nil
		},
	}

	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/votes/report", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var report domain.VoteReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &report))
	assert.Len(t, report, 1)
	assert.Equal(t, "p1", report[0].ProductID)
}

func TestGetReport_ServerError(t *testing.T) {
	svc := &mocks.VoteServiceMock{
		GetVoteReportFn: func(ctx context.Context) (domain.VoteReportResponse, error) {
			return nil, errors.New("db failure")
		},
	}

	r := setupVoteRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/votes/report", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
