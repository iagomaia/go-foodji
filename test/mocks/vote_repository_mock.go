package mocks

import (
	"context"

	"github.com/iagomaia/go-foodji/internal/domain"
)

type VoteRepositoryMock struct {
	UpsertFn        func(ctx context.Context, vote *domain.Vote) (bool, error)
	EnsureIndexesFn func(ctx context.Context) error
	GetFn           func(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error)
	GetReportFn     func(ctx context.Context) (domain.VoteReportResponse, error)
}

func (m *VoteRepositoryMock) Upsert(ctx context.Context, vote *domain.Vote) (bool, error) {
	return m.UpsertFn(ctx, vote)
}

func (m *VoteRepositoryMock) EnsureIndexes(ctx context.Context) error {
	if m.EnsureIndexesFn != nil {
		return m.EnsureIndexesFn(ctx)
	}
	return nil
}

func (m *VoteRepositoryMock) Get(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error) {
	return m.GetFn(ctx, filter)
}

func (m *VoteRepositoryMock) GetReport(ctx context.Context) (domain.VoteReportResponse, error) {
	return m.GetReportFn(ctx)
}
