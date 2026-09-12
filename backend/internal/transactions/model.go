package transactions

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"kabadiconnect/backend/internal/listings"
)

type StatusEvent struct {
	Status    string             `bson:"status" json:"status"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	ActorType string             `bson:"actorType,omitempty" json:"actorType,omitempty"`
	ActorID   primitive.ObjectID `bson:"actorId,omitempty" json:"actorId,omitempty"`
}

type MaterialSnapshot struct {
	ReferenceID string            `bson:"referenceId" json:"referenceId"`
	Material    listings.Material `bson:"material" json:"material"`
	SourceType  string            `bson:"sourceType,omitempty" json:"sourceType,omitempty"`
}

type Transaction struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Reference           string             `bson:"reference" json:"reference"`
	LotID               primitive.ObjectID `bson:"lotId" json:"lotId"`
	OfferID             primitive.ObjectID `bson:"offerId,omitempty" json:"offerId,omitempty"`
	CollectorID         primitive.ObjectID `bson:"collectorId" json:"collectorId"`
	RecyclerID          primitive.ObjectID `bson:"recyclerId" json:"recyclerId"`
	MaterialSnapshot    MaterialSnapshot   `bson:"materialSnapshot" json:"materialSnapshot"`
	ApproximateQuantity float64            `bson:"approximateQuantity" json:"approximateQuantity"`
	FinalQuantity       *float64           `bson:"finalQuantity,omitempty" json:"finalQuantity,omitempty"`
	Unit                string             `bson:"unit" json:"unit"`
	QuotedPricePerUnit  float64            `bson:"quotedPricePerUnit" json:"quotedPricePerUnit"`
	FinalPricePerUnit   *float64           `bson:"finalPricePerUnit,omitempty" json:"finalPricePerUnit,omitempty"`
	QuotedTotal         float64            `bson:"quotedTotal" json:"quotedTotal"`
	FinalTotal          *float64           `bson:"finalTotal,omitempty" json:"finalTotal,omitempty"`
	CollectionLocation  listings.GeoPoint  `bson:"collectionLocation" json:"collectionLocation"`
	CollectionLabel     string             `bson:"collectionLabel,omitempty" json:"collectionLabel,omitempty"`
	HandoverLocation    *listings.GeoPoint `bson:"handoverLocation,omitempty" json:"handoverLocation,omitempty"`
	HandoverLabel       string             `bson:"handoverLabel,omitempty" json:"handoverLabel,omitempty"`
	Status              string             `bson:"status" json:"status"`
	PaymentStatus       string             `bson:"paymentStatus" json:"paymentStatus"`
	StatusHistory       []StatusEvent      `bson:"statusHistory" json:"statusHistory"`
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
}
