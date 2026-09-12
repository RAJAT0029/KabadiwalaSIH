package payments

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Payment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PaymentID     string             `bson:"paymentId" json:"paymentId"`
	TransactionID primitive.ObjectID `bson:"transactionId" json:"transactionId"`
	RecyclerID    primitive.ObjectID `bson:"recyclerId" json:"recyclerId"`
	CollectorID   primitive.ObjectID `bson:"collectorId" json:"collectorId"`
	Amount        float64            `bson:"amount" json:"amount"`
	Currency      string             `bson:"currency" json:"currency"`
	PaymentMethod string             `bson:"paymentMethod" json:"paymentMethod"`
	Status        string             `bson:"status" json:"status"`
	PaidAt        *time.Time         `bson:"paidAt,omitempty" json:"paidAt,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
