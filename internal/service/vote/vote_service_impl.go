package vote

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/iagomaia/go-foodji/internal/domain"
	"github.com/iagomaia/go-foodji/internal/repository"
)

type voteService struct {
	voteRepo    repository.VoteRepository
	sessionRepo repository.SessionRepository
}

func NewVoteService(voteRepo repository.VoteRepository, sessionRepo repository.SessionRepository) VoteService {
	return &voteService{voteRepo: voteRepo, sessionRepo: sessionRepo}
}

func (s *voteService) UpsertVote(ctx context.Context, input domain.UpsertVoteInput) (*domain.Vote, error) {
	if input.VoteType == "" {
		return nil, fmt.Errorf("vote type is required: %w", domain.ErrBadRequest)
	}
	if !slices.Contains(domain.VoteTypes, input.VoteType) {
		return nil, fmt.Errorf("invalid vote type: %w", domain.ErrBadRequest)
	}

	_, err := s.sessionRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		return nil, fmt.Errorf("upsert vote: %w", err)
	}

	now := time.Now().UTC()
	vote := &domain.Vote{
		SessionID: input.SessionID,
		ProductID: input.ProductID,
		VoteType:  input.VoteType,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = s.voteRepo.Upsert(ctx, vote)
	if err != nil {
		return nil, fmt.Errorf("upsert vote: %w", err)
	}

	return vote, nil
}

func (s *voteService) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.Vote, error) {
	filter := &domain.GetVoteFilter{SessionID: sessionID}
	votes, err := s.voteRepo.Get(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get votes by session id: %w", err)
	}
	return votes, nil
}

func (s *voteService) GetVoteReport(ctx context.Context) (domain.VoteReportResponse, error) {
	report, err := s.voteRepo.GetReport(ctx)
	if err != nil {
		return domain.VoteReportResponse{}, fmt.Errorf("get vote report: %w", err)
	}

	for i, v := range report {
		sum := v.DislikeCount + v.LikeCount
		if sum == 0 {
			report[i].LikeRatio = 0
			report[i].DislikeRatio = 0
		} else {
			report[i].LikeRatio = float64(v.LikeCount) / float64(sum)
			report[i].DislikeRatio = float64(v.DislikeCount) / float64(sum)
		}
	}

	return report, nil
}
