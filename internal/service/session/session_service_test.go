package session_test

import (
	"context"
	"errors"
	"github.com/iagomaia/go-foodji/internal/service/session"
	"testing"

	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSession(t *testing.T) {
	repo := &mocks.SessionRepositoryMock{
		CreateFn: func(ctx context.Context, session *domain.Session) error {
			session.ID = "abc123"
			return nil
		},
	}

	svc := session.NewSessionService(repo)

	session, err := svc.CreateSession(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "abc123", session.ID)
	assert.False(t, session.CreatedAt.IsZero())
	assert.False(t, session.UpdatedAt.IsZero())
	assert.Equal(t, session.CreatedAt, session.UpdatedAt)
}

func TestCreateSessionError(t *testing.T) {
	dbErr := errors.New("database error")

	repo := &mocks.SessionRepositoryMock{
		CreateFn: func(ctx context.Context, session *domain.Session) error {
			return dbErr
		},
	}

	svc := session.NewSessionService(repo)
	createdSession, err := svc.CreateSession(context.Background())

	require.Error(t, err)
	assert.Nil(t, createdSession)
	assert.ErrorIs(t, err, dbErr)
}
