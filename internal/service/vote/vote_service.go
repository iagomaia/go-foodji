package vote

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type VoteService interface {
	UpsertVote(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*domain.Vote, error)
	GetVoteReport(ctx context.Context) (domain.VoteReportResponse, error)
}
