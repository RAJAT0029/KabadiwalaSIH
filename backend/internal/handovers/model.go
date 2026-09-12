package handovers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"kabadiconnect/backend/internal/listings"
	"time"
)

type StatusEvent struct {
	Status    string             `bson:"status" json:"status"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	ActorType string             `bson:"actorType,omitempty" json:"actorType,omitempty"`
	ActorID   primitive.ObjectID `bson:"actorId,omitempty" json:"actorId,omitempty"`
}
type HandoverRecord struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HandoverReference     string             `bson:"handoverReference" json:"handoverReference"`
	LotID                 primitive.ObjectID `bson:"lotId" json:"lotId"`
	TransactionID         primitive.ObjectID `bson:"transactionId" json:"transactionId"`
	CollectorID           primitive.ObjectID `bson:"collectorId" json:"collectorId"`
	RecyclerID            primitive.ObjectID `bson:"recyclerId" json:"recyclerId"`
	MaterialLabel         string             `bson:"materialLabel" json:"materialLabel"`
	Photographs           []string           `bson:"photographs,omitempty" json:"photographs,omitempty"`
	ApproximateWeight     float64            `bson:"approximateWeight" json:"approximateWeight"`
	ConfirmedWeight       *float64           `bson:"confirmedWeight,omitempty" json:"confirmedWeight,omitempty"`
	WeightUnit            string             `bson:"weightUnit" json:"weightUnit"`
	Timestamp             time.Time          `bson:"timestamp" json:"timestamp"`
	Location              listings.GeoPoint  `bson:"location" json:"location"`
	LocationLabel         string             `bson:"locationLabel,omitempty" json:"locationLabel,omitempty"`
	CollectorConfirmation bool               `bson:"collectorConfirmation" json:"collectorConfirmation"`
	RecyclerConfirmation  bool               `bson:"recyclerConfirmation" json:"recyclerConfirmation"`
	Status                string             `bson:"status" json:"status"`
	StatusHistory         []StatusEvent      `bson:"statusHistory" json:"statusHistory"`
	CreatedAt             time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt             time.Time          `bson:"updatedAt" json:"updatedAt"`
}
