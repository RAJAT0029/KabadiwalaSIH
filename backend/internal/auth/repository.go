package auth

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kabadiconnect/backend/internal/database"
	"time"
)

type SessionRepository struct{ c *mongo.Collection }

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{c: db.Collection("refreshSessions")}
}
func (r *SessionRepository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetUnique(true)}, {Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)}})
	return err
}
func (r *SessionRepository) Create(ctx context.Context, s RefreshSession) error {
	_, err := r.c.InsertOne(ctx, s)
	return err
}
func (r *SessionRepository) ActiveByHash(ctx context.Context, h string) (*RefreshSession, error) {
	var s RefreshSession
	err := r.c.FindOne(ctx, bson.M{"tokenHash": h, "revokedAt": bson.M{"$exists": false}, "expiresAt": bson.M{"$gt": time.Now().UTC()}}).Decode(&s)
	return &s, err
}
func (r *SessionRepository) Revoke(ctx context.Context, h string) error {
	now := time.Now().UTC()
	_, err := r.c.UpdateOne(ctx, bson.M{"tokenHash": h, "revokedAt": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"revokedAt": now}})
	return err
}

// Consume atomically claims a refresh session; simultaneous replay has one winner.
func (r *SessionRepository) Consume(ctx context.Context, h string) error {
	res, err := r.c.UpdateOne(ctx, bson.M{"tokenHash": h, "revokedAt": bson.M{"$exists": false}, "expiresAt": bson.M{"$gt": time.Now().UTC()}}, bson.M{"$set": bson.M{"revokedAt": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return mongo.ErrNoDocuments
	}
	return nil
}
