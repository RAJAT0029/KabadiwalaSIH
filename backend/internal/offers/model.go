package offers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending   = "pending"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

type Offer struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	LotID       primitive.ObjectID `bson:"lotId,omitempty" json:"lotId"`
	CollectorID primitive.ObjectID `bson:"collectorId,omitempty" json:"collectorId"`
	RecyclerID  primitive.ObjectID `bson:"recyclerId" json:"recyclerId"`

	ApproximateWeight   float64   `bson:"approximateWeight,omitempty" json:"approximateWeight"`
	OfferedPricePerUnit float64   `bson:"offeredPricePerUnit,omitempty" json:"offeredPricePerUnit"`
	OfferedTotal        float64   `bson:"offeredTotal,omitempty" json:"offeredTotal"`
	Unit                string    `bson:"unit,omitempty" json:"unit"`
	PickupAvailable     bool      `bson:"pickupAvailable" json:"pickupAvailable"`
	Message             string    `bson:"message,omitempty" json:"message,omitempty"`
	Status              string    `bson:"status" json:"status"`
	CreatedAt           time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time `bson:"updatedAt" json:"updatedAt"`

	// V1 compatibility fields.
	ListingID         primitive.ObjectID `bson:"listingId,omitempty" json:"listingId,omitempty"`
	KabadiwalaID      primitive.ObjectID `bson:"kabadiwalaId,omitempty" json:"kabadiwalaId,omitempty"`
	Quantity          float64            `bson:"quantity,omitempty" json:"quantity,omitempty"`
	OfferedPricePerKg float64            `bson:"offeredPricePerKg,omitempty" json:"offeredPricePerKg,omitempty"`
	TotalAmount       float64            `bson:"totalAmount,omitempty" json:"totalAmount,omitempty"`
}

func (o *Offer) Normalize() {
	if o.LotID.IsZero() {
		o.LotID = o.ListingID
	}
	if o.ListingID.IsZero() {
		o.ListingID = o.LotID
	}
	if o.CollectorID.IsZero() {
		o.CollectorID = o.KabadiwalaID
	}
	if o.KabadiwalaID.IsZero() {
		o.KabadiwalaID = o.CollectorID
	}
	if o.ApproximateWeight == 0 {
		o.ApproximateWeight = o.Quantity
	}
	if o.Quantity == 0 {
		o.Quantity = o.ApproximateWeight
	}
	if o.OfferedPricePerUnit == 0 {
		o.OfferedPricePerUnit = o.OfferedPricePerKg
	}
	if o.OfferedPricePerKg == 0 {
		o.OfferedPricePerKg = o.OfferedPricePerUnit
	}
	if o.OfferedTotal == 0 {
		o.OfferedTotal = o.TotalAmount
	}
	if o.TotalAmount == 0 {
		o.TotalAmount = o.OfferedTotal
	}
	if o.Unit == "" {
		o.Unit = "kg"
	}
}

func (o Offer) EffectiveLotID() primitive.ObjectID {
	if !o.LotID.IsZero() {
		return o.LotID
	}
	return o.ListingID
}

type View struct {
	Offer                 Offer   `json:"offer"`
	MaterialName          string  `json:"materialName"`
	CollectorName         string  `json:"collectorName"`
	EstimatedPricePerUnit float64 `json:"estimatedPricePerUnit"`
	Unit                  string  `json:"unit"`
	SupplierName          string  `json:"supplierName"`     // compatibility
	SellerPricePerKg      float64 `json:"sellerPricePerKg"` // compatibility
}
