package listings

import (
	"fmt"
	"kabadiconnect/backend/internal/materials"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DomainEWaste               = "e-waste"
	LegacyEWasteCategory       = "E-waste"
	LotStatusDraft             = "draft"
	LotStatusAvailable         = "available"
	LotStatusOfferReceived     = "offer_received"
	LotStatusMatched           = "matched"
	LotStatusHandoverScheduled = "handover_scheduled"
	LotStatusHandedOver        = "handed_over"
	LotStatusCompleted         = "completed"
	LotStatusCancelled         = "cancelled"
)

var SupportedCategories = materials.Categories

type Material struct {
	Category    string         `bson:"category" json:"category"`
	SubCategory string         `bson:"subCategory,omitempty" json:"subCategory,omitempty"`
	Description string         `bson:"description,omitempty" json:"description,omitempty"`
	Condition   string         `bson:"condition,omitempty" json:"condition,omitempty"`
	SourceType  string         `bson:"sourceType,omitempty" json:"sourceType,omitempty"`
	Attributes  map[string]any `bson:"attributes,omitempty" json:"attributes,omitempty"`

	// V1 compatibility fields. New clients should use subCategory/description.
	SubType string `bson:"subType,omitempty" json:"subType,omitempty"`
	Grade   string `bson:"grade,omitempty" json:"grade,omitempty"`
}

type Quantity struct {
	ApproximateWeight float64 `bson:"approximateWeight,omitempty" json:"approximateWeight"`
	AvailableWeight   float64 `bson:"availableWeight" json:"availableWeight"`
	Unit              string  `bson:"unit" json:"unit"`

	// V1 compatibility fields.
	Total     float64 `bson:"total,omitempty" json:"total,omitempty"`
	Available float64 `bson:"available,omitempty" json:"available,omitempty"`
}

type Valuation struct {
	EstimatedValue        float64 `bson:"estimatedValue,omitempty" json:"estimatedValue"`
	EstimatedPricePerUnit float64 `bson:"estimatedPricePerUnit,omitempty" json:"estimatedPricePerUnit"`
	MarketRangeMin        float64 `bson:"marketRangeMin,omitempty" json:"marketRangeMin"`
	MarketRangeMax        float64 `bson:"marketRangeMax,omitempty" json:"marketRangeMax"`
	EstimationSource      string  `bson:"estimationSource,omitempty" json:"estimationSource"`
}

type LegacyPricing struct {
	ExpectedPricePerKg  float64 `bson:"expectedPricePerKg,omitempty" json:"expectedPricePerKg,omitempty"`
	Negotiable          bool    `bson:"negotiable" json:"negotiable"`
	AllowDirectPurchase bool    `bson:"allowDirectPurchase" json:"allowDirectPurchase"`
}

type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type ImageReference struct {
	Reference  string    `bson:"reference" json:"reference"`
	CapturedAt time.Time `bson:"capturedAt,omitempty" json:"capturedAt,omitempty"`
}

type CollectionDetails struct {
	CollectedAt time.Time `bson:"collectedAt,omitempty" json:"collectedAt,omitempty"`
	Location    GeoPoint  `bson:"location" json:"location"`
	Label       string    `bson:"label,omitempty" json:"label,omitempty"`
}

// EWasteLot is the canonical shared entity. The physical V1 collection remains
// scrapListings so existing development data is not destroyed during migration.
type EWasteLot struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReferenceID string             `bson:"referenceId,omitempty" json:"referenceId"`
	CollectorID primitive.ObjectID `bson:"collectorId,omitempty" json:"collectorId"`
	Domain      string             `bson:"domain,omitempty" json:"domain"`
	Material    Material           `bson:"material" json:"material"`
	Images      []string           `bson:"images,omitempty" json:"images,omitempty"` // legacy/simple refs
	ImageRefs   []ImageReference   `bson:"imageReferences,omitempty" json:"imageReferences,omitempty"`
	Quantity    Quantity           `bson:"quantity" json:"quantity"`
	Valuation   Valuation          `bson:"valuation" json:"valuation"`
	Collection  CollectionDetails  `bson:"collection" json:"collection"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`

	// V1 compatibility fields retained while old clients/data are migrated.
	KabadiwalaID  primitive.ObjectID `bson:"kabadiwalaId,omitempty" json:"kabadiwalaId,omitempty"`
	Pricing       LegacyPricing      `bson:"pricing,omitempty" json:"pricing,omitempty"`
	Location      GeoPoint           `bson:"location,omitempty" json:"location,omitempty"`
	LocationLabel string             `bson:"locationLabel,omitempty" json:"locationLabel,omitempty"`
}

// Listing keeps source compatibility for existing internal tests/callers.
type Listing = EWasteLot

func (l *EWasteLot) Normalize() {
	if l.CollectorID.IsZero() {
		l.CollectorID = l.KabadiwalaID
	}
	if l.KabadiwalaID.IsZero() {
		l.KabadiwalaID = l.CollectorID
	}
	if l.Domain == "" {
		l.Domain = DomainEWaste
	}
	if l.ReferenceID == "" && !l.ID.IsZero() {
		l.ReferenceID = fmt.Sprintf("KC-EW-%s", strings.ToUpper(l.ID.Hex()[16:]))
	}
	if l.Material.Category == LegacyEWasteCategory {
		switch strings.ToLower(strings.TrimSpace(l.Material.SubType)) {
		case "printed circuit boards":
			l.Material.Category = "PCB"
		case "cables & chargers":
			l.Material.Category = "Cable"
		case "batteries":
			l.Material.Category = "Battery"
		default:
			l.Material.Category = "Other E-Waste"
		}
	}
	if l.Material.SubCategory == "" {
		l.Material.SubCategory = l.Material.SubType
	}
	if l.Material.SubType == "" {
		l.Material.SubType = l.Material.SubCategory
	}
	if l.Material.Description == "" {
		l.Material.Description = l.Material.Grade
	}
	if l.Material.Grade == "" {
		l.Material.Grade = l.Material.Description
	}
	if l.Quantity.ApproximateWeight == 0 {
		l.Quantity.AvailableWeight = l.Quantity.Available
		l.Quantity.ApproximateWeight = l.Quantity.Total
	}
	if l.Quantity.Total == 0 {
		l.Quantity.Total = l.Quantity.ApproximateWeight
	}
	if l.Quantity.Available == 0 {
		l.Quantity.Available = l.Quantity.AvailableWeight
	}
	if l.Valuation.EstimatedPricePerUnit == 0 {
		l.Valuation.EstimatedPricePerUnit = l.Pricing.ExpectedPricePerKg
	}
	if l.Pricing.ExpectedPricePerKg == 0 {
		l.Pricing.ExpectedPricePerKg = l.Valuation.EstimatedPricePerUnit
	}
	if l.Valuation.EstimatedValue == 0 {
		l.Valuation.EstimatedValue = l.Quantity.ApproximateWeight * l.Valuation.EstimatedPricePerUnit
	}
	if len(l.Collection.Location.Coordinates) != 2 && len(l.Location.Coordinates) == 2 {
		l.Collection.Location = l.Location
	}
	if len(l.Location.Coordinates) != 2 && len(l.Collection.Location.Coordinates) == 2 {
		l.Location = l.Collection.Location
	}
	if l.Collection.Label == "" {
		l.Collection.Label = l.LocationLabel
	}
	if l.LocationLabel == "" {
		l.LocationLabel = l.Collection.Label
	}
	if l.Status == "active" {
		l.Status = LotStatusAvailable
	}
}

func (l *EWasteLot) IsAvailable() bool {
	return l.Status == LotStatusAvailable || l.Status == LotStatusOfferReceived || l.Status == "active"
}

func (l *EWasteLot) AvailableAmount() float64 {
	return l.Quantity.AvailableWeight
}

func (l *EWasteLot) UnitPrice() float64 {
	if l.Valuation.EstimatedPricePerUnit > 0 {
		return l.Valuation.EstimatedPricePerUnit
	}
	return l.Pricing.ExpectedPricePerKg
}

type CollectorSummary struct {
	ID                primitive.ObjectID `json:"id"`
	BusinessName      string             `json:"businessName"`
	OperatingLocation string             `json:"operatingLocation,omitempty"`
	ProfileStatus     string             `json:"profileStatus"`
	PreferredLanguage string             `json:"preferredLanguage,omitempty"`
	MemberSince       time.Time          `json:"memberSince"`
}

type View struct {
	Lot        EWasteLot        `json:"lot"`
	Listing    EWasteLot        `json:"listing"` // compatibility alias for V1 recycler builds
	Collector  CollectorSummary `json:"collector"`
	Supplier   CollectorSummary `json:"supplier"` // compatibility alias
	DistanceKm *float64         `json:"distanceKm,omitempty"`
}
