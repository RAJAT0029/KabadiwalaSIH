package prices

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Location struct {
	City        string    `bson:"city" json:"city"`
	State       string    `bson:"state" json:"state"`
	Coordinates []float64 `bson:"coordinates,omitempty" json:"coordinates,omitempty"`
}

type PriceRecord struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	MaterialCategory    string              `bson:"materialCategory" json:"materialCategory"`
	MaterialSubCategory string              `bson:"materialSubCategory,omitempty" json:"materialSubCategory,omitempty"`
	Location            Location            `bson:"location" json:"location"`
	Timestamp           time.Time           `bson:"timestamp" json:"timestamp"`
	BuyingPrice         float64             `bson:"buyingPrice" json:"buyingPrice"`
	QuotedPrice         float64             `bson:"quotedPrice" json:"quotedPrice"`
	Unit                string              `bson:"unit" json:"unit"`
	RecyclerID          *primitive.ObjectID `bson:"recyclerId,omitempty" json:"recyclerId,omitempty"`
	AggregatorID        *primitive.ObjectID `bson:"aggregatorId,omitempty" json:"aggregatorId,omitempty"`
	SourceType          string              `bson:"sourceType" json:"sourceType"`
	IsDemo              bool                `bson:"isDemo,omitempty" json:"isDemo"`
	CreatedAt           time.Time           `bson:"createdAt" json:"createdAt"`
}

type BoardEntry struct {
	MaterialCategory    string    `json:"materialCategory"`
	MaterialSubCategory string    `json:"materialSubCategory,omitempty"`
	CurrentBuyingRate   float64   `json:"currentBuyingRate"`
	QuotedPrice         float64   `json:"quotedPrice"`
	Unit                string    `json:"unit"`
	City                string    `json:"city"`
	State               string    `json:"state"`
	MarketRangeMin      float64   `json:"marketRangeMin"`
	MarketRangeMax      float64   `json:"marketRangeMax"`
	Trend               string    `json:"trend"`
	UpdatedAt           time.Time `json:"updatedAt"`
	SourceType          string    `json:"sourceType"`
	IsDemo              bool      `json:"isDemo"`
}

type Board struct {
	Currency string       `json:"currency"`
	Notice   string       `json:"notice"`
	Entries  []BoardEntry `json:"entries"`
}
