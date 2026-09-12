package transactions

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kabadiconnect/backend/internal/database"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{c: db.Collection("transactions")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "reference", Value: 1}}, Options: options.Index().SetUnique(true)}, {Keys: bson.D{{Key: "recyclerId", Value: 1}, {Key: "createdAt", Value: -1}}}, {Keys: bson.D{{Key: "collectorId", Value: 1}, {Key: "createdAt", Value: -1}}}, {Keys: bson.D{{Key: "lotId", Value: 1}, {Key: "status", Value: 1}}}})
	return err
}
func (r *Repository) ListByRecycler(ctx context.Context, id primitive.ObjectID, status string) ([]Transaction, error) {
	f := bson.M{"recyclerId": id}
	if status != "" && status != "all" {
		f["status"] = status
	}
	cur, err := r.c.Find(ctx, f, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	rows := make([]Transaction, 0)
	err = cur.All(ctx, &rows)
	return rows, err
}
func (r *Repository) ByIDForRecycler(ctx context.Context, id, recyclerID primitive.ObjectID) (*Transaction, error) {
	var row Transaction
	err := r.c.FindOne(ctx, bson.M{"_id": id, "recyclerId": recyclerID}).Decode(&row)
	return &row, err
}
