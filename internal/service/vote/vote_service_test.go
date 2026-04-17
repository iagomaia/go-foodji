package vote_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/internal/service/vote"
	"github.com/iagomaia/go-foodji/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UpsertVote ---

func TestUpsertVote_CreatesNewVote(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Session, error) {
			return &domain.Session{ID: id}, nil
		},
	}
	voteRepo := &mocks.VoteRepositoryMock{
		UpsertFn: func(ctx context.Context, v *domain.Vote) (bool, error) {
			now := time.Now().UTC()
			v.ID = "vote-1"
			v.CreatedAt = now
			v.UpdatedAt = now
			return true, nil
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Like}

	result, err := svc.UpsertVote(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "vote-1", result.ID)
	assert.Equal(t, domain.Like, result.VoteType)
	assert.Equal(t, "s1", result.SessionID)
	assert.Equal(t, "p1", result.ProductID)
	assert.True(t, result.Created, "new vote should have Created=true")
}

func TestUpsertVote_UpdatesExistingVote(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Session, error) {
			return &domain.Session{ID: id}, nil
		},
	}
	voteRepo := &mocks.VoteRepositoryMock{
		UpsertFn: func(ctx context.Context, v *domain.Vote) (bool, error) {
			now := time.Now().UTC()
			v.ID = "vote-1"
			v.CreatedAt = now
			v.UpdatedAt = now.Add(time.Minute)
			return false, nil
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Dislike}

	result, err := svc.UpsertVote(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, domain.Dislike, result.VoteType)
	assert.False(t, result.Created, "updated vote should have Created=false")
}

func TestUpsertVote_SessionNotFound(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Session, error) {
			return nil, domain.ErrNotFound
		},
	}
	voteRepo := &mocks.VoteRepositoryMock{}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "missing", ProductID: "p1", VoteType: domain.Like}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpsertVote_EmptySessionID(t *testing.T) {
	svc := vote.NewVoteService(&mocks.VoteRepositoryMock{}, &mocks.SessionRepositoryMock{})
	input := domain.UpsertVoteInput{SessionID: "", ProductID: "p1", VoteType: domain.Like}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBadRequest)
}

func TestUpsertVote_EmptyProductID(t *testing.T) {
	svc := vote.NewVoteService(&mocks.VoteRepositoryMock{}, &mocks.SessionRepositoryMock{})
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "", VoteType: domain.Like}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBadRequest)
}

func TestUpsertVote_InvalidVoteType(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Session, error) {
			return &domain.Session{ID: id}, nil
		},
	}
	voteRepo := &mocks.VoteRepositoryMock{}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: "meh"}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBadRequest)
}

func TestUpsertVote_EmptyVoteType(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: ""}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBadRequest)
}

func TestUpsertVote_RepoError(t *testing.T) {
	dbErr := errors.New("db unavailable")

	sessionRepo := &mocks.SessionRepositoryMock{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Session, error) {
			return &domain.Session{ID: id}, nil
		},
	}
	voteRepo := &mocks.VoteRepositoryMock{
		UpsertFn: func(ctx context.Context, v *domain.Vote) (bool, error) {
			return false, dbErr
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	input := domain.UpsertVoteInput{SessionID: "s1", ProductID: "p1", VoteType: domain.Like}

	_, err := svc.UpsertVote(context.Background(), input)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}

// --- GetBySessionID ---

func TestGetBySessionID_ReturnsList(t *testing.T) {
	expected := []*domain.Vote{
		{ID: "v1", SessionID: "s1", ProductID: "p1", VoteType: domain.Like},
		{ID: "v2", SessionID: "s1", ProductID: "p2", VoteType: domain.Dislike},
	}

	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{
		GetFn: func(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error) {
			assert.Equal(t, "s1", filter.SessionID)
			return expected, nil
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	result, err := svc.GetBySessionID(context.Background(), "s1")

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetBySessionID_RepoError(t *testing.T) {
	dbErr := errors.New("db unavailable")

	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{
		GetFn: func(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error) {
			return nil, dbErr
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	_, err := svc.GetBySessionID(context.Background(), "s1")

	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}

// --- GetVoteReport ---

func TestGetVoteReport_CalculatesRatios(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{
		GetReportFn: func(ctx context.Context) (domain.VoteReportResponse, error) {
			return domain.VoteReportResponse{
				{ProductID: "p1", LikeCount: 3, DislikeCount: 1},
				{ProductID: "p2", LikeCount: 0, DislikeCount: 4},
			}, nil
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	report, err := svc.GetVoteReport(context.Background())

	require.NoError(t, err)
	require.Len(t, report, 2)

	assert.InDelta(t, 0.75, report[0].LikeRatio, 0.001)
	assert.InDelta(t, 0.25, report[0].DislikeRatio, 0.001)

	assert.InDelta(t, 0.0, report[1].LikeRatio, 0.001)
	assert.InDelta(t, 1.0, report[1].DislikeRatio, 0.001)
}

func TestGetVoteReport_ZeroVotesNoNaNRatio(t *testing.T) {
	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{
		GetReportFn: func(ctx context.Context) (domain.VoteReportResponse, error) {
			return domain.VoteReportResponse{
				{ProductID: "p1", LikeCount: 0, DislikeCount: 0},
			}, nil
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	report, err := svc.GetVoteReport(context.Background())

	require.NoError(t, err)
	require.Len(t, report, 1)
	assert.Equal(t, 0.0, report[0].LikeRatio, "zero votes should produce 0.0 like ratio, not NaN")
	assert.Equal(t, 0.0, report[0].DislikeRatio, "zero votes should produce 0.0 dislike ratio, not NaN")
}

func TestGetVoteReport_RepoError(t *testing.T) {
	dbErr := errors.New("db unavailable")

	sessionRepo := &mocks.SessionRepositoryMock{}
	voteRepo := &mocks.VoteRepositoryMock{
		GetReportFn: func(ctx context.Context) (domain.VoteReportResponse, error) {
			return nil, dbErr
		},
	}

	svc := vote.NewVoteService(voteRepo, sessionRepo)
	_, err := svc.GetVoteReport(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}
