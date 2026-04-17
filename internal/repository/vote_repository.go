package repository

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type VoteRepository interface {
	Upsert(ctx context.Context, vote *domain.Vote) error
	EnsureIndexes(ctx context.Context) error
	Get(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error)
	GetReport(ctx context.Context) (domain.VoteReportResponse, error)
}
