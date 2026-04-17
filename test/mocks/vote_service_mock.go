package mocks

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type VoteServiceMock struct {
	UpsertVoteFn     func(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error)
	GetBySessionIDFn func(ctx context.Context, sessionID string) ([]*domain.Vote, error)
	GetVoteReportFn  func(ctx context.Context) (domain.VoteReportResponse, error)
}

func (m *VoteServiceMock) UpsertVote(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
	return m.UpsertVoteFn(ctx, input)
}

func (m *VoteServiceMock) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.Vote, error) {
	return m.GetBySessionIDFn(ctx, sessionID)
}

func (m *VoteServiceMock) GetVoteReport(ctx context.Context) (domain.VoteReportResponse, error) {
	return m.GetVoteReportFn(ctx)
}
