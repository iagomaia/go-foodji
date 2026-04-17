package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/iagomaia/go-foodji/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const voteCollectionName = "votes"

type voteDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	SessionID bson.ObjectID `bson:"session_id"`
	ProductID string        `bson:"product_id"`
	VoteType  string        `bson:"vote_type"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type voteReportDocument struct {
	ProductID    string `bson:"_id"`
	LikeCount    int    `bson:"like_count"`
	DislikeCount int    `bson:"dislike_count"`
}

type VoteRepository struct {
	col *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	return &VoteRepository{col: db.Collection(voteCollectionName)}
}

func (r *VoteRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{"session_id", 1}, {"product_id", 1}},
			Options: options.Index().SetUnique(true).SetName("ux_votes_session_product"),
		},
	}
	_, err := r.col.Indexes().CreateMany(ctx, models)
	return err
}

func (r *VoteRepository) Upsert(ctx context.Context, vote *domain.Vote) (bool, error) {
	sessionOID, err := bson.ObjectIDFromHex(vote.SessionID)
	if err != nil {
		return false, fmt.Errorf("invalid session id: %w", domain.ErrBadRequest)
	}

	candidateID := bson.NewObjectID()
	now := time.Now().UTC()
	filter := bson.D{{"session_id", sessionOID}, {"product_id", vote.ProductID}}
	update := bson.D{
		{"$set", bson.D{
			{"vote_type", string(vote.VoteType)},
			{"updated_at", now},
		}},
		{"$setOnInsert", bson.D{
			{"_id", candidateID},
			{"session_id", sessionOID},
			{"product_id", vote.ProductID},
			{"created_at", now},
		}},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc voteDocument
	if err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		return false, fmt.Errorf("upsert vote: %w", err)
	}

	created := doc.ID.Hex() == candidateID.Hex()
	vote.ID = doc.ID.Hex()
	vote.CreatedAt = doc.CreatedAt
	vote.UpdatedAt = doc.UpdatedAt

	return created, nil
}

func (r *VoteRepository) Get(ctx context.Context, filter *domain.GetVoteFilter) ([]*domain.Vote, error) {
	query, err := toVoteQuery(filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	cursor, err := r.col.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find votes: %w", err)
	}
	defer cursor.Close(ctx)

	votes := []*domain.Vote{}
	for cursor.Next(ctx) {
		doc := voteDocument{}
		err = cursor.Decode(&doc)
		if err != nil {
			return nil, fmt.Errorf("decode vote: %w", err)
		}
		vote := toVote(doc)
		votes = append(votes, &vote)
	}
	return votes, nil
}

func (r *VoteRepository) GetReport(ctx context.Context) (domain.VoteReportResponse, error) {
	pipeline := bson.A{
		bson.D{
			{"$group",
				bson.D{
					{"_id", "$product_id"},
					{"like_count",
						bson.D{
							{"$sum",
								bson.D{
									{"$cond",
										bson.A{
											bson.D{
												{"$eq",
													bson.A{
														"$vote_type",
														domain.Like,
													},
												},
											},
											1,
											0,
										},
									},
								},
							},
						},
					},
					{"dislike_count",
						bson.D{
							{"$sum",
								bson.D{
									{"$cond",
										bson.A{
											bson.D{
												{"$eq",
													bson.A{
														"$vote_type",
														domain.Dislike,
													},
												},
											},
											1,
											0,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return domain.VoteReportResponse{}, fmt.Errorf("aggregate votes: %w", err)
	}

	defer cursor.Close(ctx)
	voteReport := domain.VoteReportResponse{}
	for cursor.Next(ctx) {
		var doc voteReportDocument
		err = cursor.Decode(&doc)
		if err != nil {
			return domain.VoteReportResponse{}, fmt.Errorf("decode vote report: %w", err)
		}
		voteReport = append(voteReport, toVoteReportItem(doc))
	}

	return voteReport, nil
}

func toVoteQuery(filter *domain.GetVoteFilter) (bson.M, error) {
	query := bson.M{}
	if filter.SessionID != "" {
		sessionID, err := bson.ObjectIDFromHex(filter.SessionID)
		if err != nil {
			return nil, err
		}
		query["session_id"] = sessionID
	}
	return query, nil
}

func toVote(doc voteDocument) domain.Vote {
	return domain.Vote{
		ID:        doc.ID.Hex(),
		SessionID: doc.SessionID.Hex(),
		ProductID: doc.ProductID,
		VoteType:  domain.VoteType(doc.VoteType),
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func toVoteReportItem(doc voteReportDocument) domain.VoteReportItem {
	return domain.VoteReportItem{
		ProductID:    doc.ProductID,
		LikeCount:    doc.LikeCount,
		DislikeCount: doc.DislikeCount,
	}
}
