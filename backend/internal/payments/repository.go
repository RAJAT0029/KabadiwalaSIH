package payments

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kabadiconnect/backend/internal/database"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository { return &Repository{c: db.Collection("payments")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "paymentId", Value: 1}}, Options: options.Index().SetUnique(true)}, {Keys: bson.D{{Key: "recyclerId", Value: 1}, {Key: "createdAt", Value: -1}}}, {Keys: bson.D{{Key: "transactionId", Value: 1}}}})
	return err
}
func (r *Repository) ListByRecycler(ctx context.Context, id primitive.ObjectID) ([]Payment, error) {
	cur, err := r.c.Find(ctx, bson.M{"recyclerId": id}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	rows := make([]Payment, 0)
	err = cur.All(ctx, &rows)
	return rows, err
}
