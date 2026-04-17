package mocks

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type SessionRepositoryMock struct {
	FindByIDFn func(ctx context.Context, id string) (*domain.Session, error)
	CreateFn   func(ctx context.Context, session *domain.Session) error
}

func (m *SessionRepositoryMock) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *SessionRepositoryMock) Create(ctx context.Context, session *domain.Session) error {
	return m.CreateFn(ctx, session)
}
