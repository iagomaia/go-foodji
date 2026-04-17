package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/internal/handler"
	"github.com/iagomaia/go-foodji/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(svc *mocks.SessionServiceMock) *gin.Engine {
	r := gin.New()
	h := handler.NewSessionHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}
func TestCreateSession_OK(t *testing.T) {
	svc := &mocks.SessionServiceMock{
		CreateSessionFn: func(ctx context.Context) (*domain.Session, error) {
			return &domain.Session{ID: "new-id"}, nil
		},
	}

	r := setupRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/sessions", nil))

	assert.Equal(t, http.StatusCreated, w.Code)

	var session domain.Session
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &session))
	assert.Equal(t, "new-id", session.ID)
}

func TestCreateSession_ServerError(t *testing.T) {
	svc := &mocks.SessionServiceMock{
		CreateSessionFn: func(ctx context.Context) (*domain.Session, error) {
			return nil, errors.New("some error")
		},
	}

	r := setupRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/sessions", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
