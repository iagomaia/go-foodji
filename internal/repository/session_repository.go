package repository

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type SessionRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Session, error)
	Create(ctx context.Context, item *domain.Session) error
}
