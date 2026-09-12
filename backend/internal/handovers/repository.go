package handovers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kabadiconnect/backend/internal/database"
	"time"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{c: db.Collection("handoverRecords")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "handoverReference", Value: 1}}, Options: options.Index().SetUnique(true)}, {Keys: bson.D{{Key: "transactionId", Value: 1}}}, {Keys: bson.D{{Key: "lotId", Value: 1}}}, {Keys: bson.D{{Key: "recyclerId", Value: 1}, {Key: "createdAt", Value: -1}}}})
	return err
}
func (r *Repository) ListByRecycler(ctx context.Context, id primitive.ObjectID) ([]HandoverRecord, error) {
	cur, err := r.c.Find(ctx, bson.M{"recyclerId": id}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	rows := make([]HandoverRecord, 0)
	err = cur.All(ctx, &rows)
	return rows, err
}
func (r *Repository) ByIDForRecycler(ctx context.Context, id, recyclerID primitive.ObjectID) (*HandoverRecord, error) {
	var row HandoverRecord
	err := r.c.FindOne(ctx, bson.M{"_id": id, "recyclerId": recyclerID}).Decode(&row)
	return &row, err
}
func (r *Repository) ConfirmRecycler(ctx context.Context, id, recyclerID primitive.ObjectID) (*HandoverRecord, error) {
	now := time.Now().UTC()
	event := StatusEvent{Status: "recycler_confirmed", Timestamp: now, ActorType: "recycler", ActorID: recyclerID}
	complete := StatusEvent{Status: "completed", Timestamp: now, ActorType: "system"}
	filter := bson.M{"_id": id, "recyclerId": recyclerID, "recyclerConfirmation": bson.M{"$ne": true}, "status": bson.M{"$in": []string{"pending", "collector_confirmed"}}}
	// One atomic document update appends both events when both parties confirm.
	update := mongo.Pipeline{bson.D{{Key: "$set", Value: bson.M{
		"recyclerConfirmation": true,
		"status":               bson.M{"$cond": []any{"$collectorConfirmation", "completed", "recycler_confirmed"}},
		"updatedAt":            now,
		"statusHistory":        bson.M{"$concatArrays": []any{bson.M{"$ifNull": []any{"$statusHistory", bson.A{}}}, bson.A{event}, bson.M{"$cond": []any{"$collectorConfirmation", bson.A{complete}, bson.A{}}}}},
	}}}}
	var row HandoverRecord
	err := r.c.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&row)
	if err == mongo.ErrNoDocuments {
		// Safe retry after a lost success response returns the owned record unchanged.
		existing, readErr := r.ByIDForRecycler(ctx, id, recyclerID)
		if readErr == nil && existing.RecyclerConfirmation {
			return existing, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
