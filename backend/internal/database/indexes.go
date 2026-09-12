package database

import (
	"bytes"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EnsureIndexes preserves existing equivalent indexes, including historical
// names and the redundant sparse option on single-field 2dsphere indexes.
// It never drops indexes or weakens uniqueness/TTL/partial-filter guarantees.
func EnsureIndexes(ctx context.Context, c *mongo.Collection, models []mongo.IndexModel) ([]string, error) {
	var names []string
	for _, model := range models {
		name, err := c.Indexes().CreateOne(ctx, model)
		if err == nil {
			names = append(names, name)
			continue
		}
		commandErr, ok := err.(mongo.CommandError)
		if !ok || (commandErr.Code != 85 && commandErr.Code != 86) {
			return nil, fmt.Errorf("%s index: %w", c.Name(), err)
		}
		cur, listErr := c.Indexes().List(ctx)
		if listErr != nil {
			return nil, listErr
		}
		var existing []bson.Raw
		listErr = cur.All(ctx, &existing)
		_ = cur.Close(ctx)
		if listErr != nil {
			return nil, listErr
		}
		found := false
		for _, index := range existing {
			var decoded bson.M
			_ = bson.Unmarshal(index, &decoded)
			decoded["key"] = index.Lookup("key").Document()
			if equivalentIndex(model, decoded) {
				names = append(names, fmt.Sprint(decoded["name"]))
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%s: incompatible index requires explicit migration (no data was removed): %w", c.Name(), err)
		}
	}
	return names, nil
}

func equivalentIndex(model mongo.IndexModel, existing bson.M) bool {
	// BSON encoding normalizes int/int32 key directions and preserves key order.
	wantKeys, _ := bson.Marshal(model.Keys)
	gotKeys, _ := bson.Marshal(existing["key"])
	if !bytes.Equal(wantKeys, gotKeys) {
		return false
	}
	wanted := bson.M{}
	if o := model.Options; o != nil {
		if o.Unique != nil {
			wanted["unique"] = *o.Unique
		}
		if o.Sparse != nil {
			wanted["sparse"] = *o.Sparse
		}
		if o.Hidden != nil {
			wanted["hidden"] = *o.Hidden
		}
		if o.ExpireAfterSeconds != nil {
			wanted["expireAfterSeconds"] = *o.ExpireAfterSeconds
		}
		if o.PartialFilterExpression != nil {
			wanted["partialFilterExpression"] = o.PartialFilterExpression
		}
		if o.Collation != nil {
			wanted["collation"] = o.Collation
		}
	}
	geo := false
	var keys bson.D
	_ = bson.Unmarshal(wantKeys, &keys)
	if len(keys) == 1 {
		geo = keys[0].Value == "2dsphere"
	}
	for _, field := range []string{"unique", "sparse", "expireAfterSeconds", "partialFilterExpression", "collation", "hidden"} {
		if geo && field == "sparse" {
			continue
		}
		a, b := wanted[field], existing[field]
		if field == "unique" || field == "sparse" || field == "hidden" {
			if a == nil {
				a = false
			}
			if b == nil {
				b = false
			}
		}
		x, _ := bson.Marshal(bson.M{"v": a})
		y, _ := bson.Marshal(bson.M{"v": b})
		if !bytes.Equal(x, y) {
			return false
		}
	}
	return true
}
