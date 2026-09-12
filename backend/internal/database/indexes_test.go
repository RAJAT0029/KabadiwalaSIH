package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestEquivalentIndexPolicy(t *testing.T) {
	geo := mongo.IndexModel{Keys: bson.D{{Key: "location", Value: "2dsphere"}}}
	if !equivalentIndex(geo, bson.M{"key": geo.Keys, "name": "location_2dsphere", "sparse": true}) {
		t.Fatal("redundant geo sparse difference must be accepted")
	}
	for _, field := range []string{"unique", "hidden"} {
		if equivalentIndex(geo, bson.M{"key": geo.Keys, field: true}) {
			t.Fatalf("must not ignore %s", field)
		}
	}
	unique := mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)}
	if equivalentIndex(unique, bson.M{"key": unique.Keys}) {
		t.Fatal("non-unique does not satisfy unique")
	}
	if !equivalentIndex(unique, bson.M{"key": unique.Keys, "unique": true}) {
		t.Fatal("equivalent unique rejected")
	}
}
