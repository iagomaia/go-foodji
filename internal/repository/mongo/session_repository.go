package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iagomaia/go-foodji/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const sessionCollectionName = "sessions"

type sessionDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type SessionRepository struct {
	col *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{col: db.Collection(sessionCollectionName)}
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var doc sessionDocument
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find session by id: %w", err)
	}

	session := toSession(doc)
	return &session, nil
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	doc := fromSession(*session)
	doc.ID = bson.NewObjectID()

	result, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	session.ID = result.InsertedID.(bson.ObjectID).Hex()
	return nil
}

func toSession(doc sessionDocument) domain.Session {
	return domain.Session{
		ID:        doc.ID.Hex(),
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func fromSession(session domain.Session) sessionDocument {
	return sessionDocument{
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}
