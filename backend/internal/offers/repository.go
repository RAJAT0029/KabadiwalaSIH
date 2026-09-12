package offers

import (
	"context"
	"kabadiconnect/backend/internal/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository { return &Repository{c: db.Collection("offers")} }
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "lotId", Value: 1}, {Key: "recyclerId", Value: 1}}, Options: options.Index().SetName("one_pending_offer_per_recycler_lot").SetUnique(true).SetPartialFilterExpression(bson.M{"status": StatusPending, "lotId": bson.M{"$type": "objectId"}})}}); err != nil {
		return err
	}
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "recyclerId", Value: 1}, {Key: "createdAt", Value: -1}}}, {Keys: bson.D{{Key: "collectorId", Value: 1}, {Key: "createdAt", Value: -1}}}, {Keys: bson.D{{Key: "lotId", Value: 1}, {Key: "status", Value: 1}}}, {Keys: bson.D{{Key: "listingId", Value: 1}, {Key: "status", Value: 1}}}})
	return err
}
func (r *Repository) Create(ctx context.Context, o *Offer) error {
	res, err := r.c.InsertOne(ctx, o)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicatePending
	}
	if err == nil {
		o.ID = res.InsertedID.(primitive.ObjectID)
	}
	return err
}
func (r *Repository) ListByRecycler(ctx context.Context, id primitive.ObjectID, status string) ([]Offer, error) {
	filter := bson.M{"recyclerId": id}
	if status != "" && status != "all" {
		filter["status"] = status
	}
	cur, err := r.c.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var items []Offer
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Normalize()
	}
	return items, nil
}
func (r *Repository) PendingByRecyclerAndLot(ctx context.Context, rid, lid primitive.ObjectID) (int64, error) {
	return r.c.CountDocuments(ctx, bson.M{"recyclerId": rid, "status": "pending", "$or": []bson.M{{"lotId": lid}, {"listingId": lid}}})
}
func (r *Repository) CancelOwned(ctx context.Context, id, recyclerID primitive.ObjectID) error {
	res, err := r.c.UpdateOne(ctx, bson.M{"_id": id, "recyclerId": recyclerID, "status": "pending"}, bson.M{"$set": bson.M{"status": "cancelled", "updatedAt": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
