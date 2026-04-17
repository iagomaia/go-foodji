package session

import (
	"context"
	"fmt"
	"time"

	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/internal/repository"
)

type sessionService struct {
	repo repository.SessionRepository
}

func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionService{repo: repo}
}

func (s *sessionService) CreateSession(ctx context.Context) (*domain.Session, error) {
	now := time.Now().UTC()
	session := &domain.Session{
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}
