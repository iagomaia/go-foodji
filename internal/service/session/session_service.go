package session

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type SessionService interface {
	CreateSession(ctx context.Context) (*domain.Session, error)
}
