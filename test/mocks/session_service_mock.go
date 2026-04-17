package mocks

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type SessionServiceMock struct {
	CreateSessionFn func(ctx context.Context) (*domain.Session, error)
}

func (m *SessionServiceMock) CreateSession(ctx context.Context) (*domain.Session, error) {
	return m.CreateSessionFn(ctx)
}
