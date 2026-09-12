package offers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"kabadiconnect/backend/internal/listings"
	"kabadiconnect/backend/internal/users"
)

type Service struct {
	repo      *Repository
	listings  *listings.Repository
	users     *users.Repository
	allowDemo bool
}

func NewService(r *Repository, l *listings.Repository, u *users.Repository, allowDemo ...bool) *Service {
	return &Service{repo: r, listings: l, users: u, allowDemo: len(allowDemo) > 0 && allowDemo[0]}
}

type CreateInput struct {
	canonicalWeightSet  bool
	canonicalPriceSet   bool
	ApproximateWeight   float64 `json:"approximateWeight"`
	OfferedPricePerUnit float64 `json:"offeredPricePerUnit"`
	PickupAvailable     bool    `json:"pickupAvailable"`
	Message             string  `json:"message"`
	// V1 request compatibility.
	Quantity          float64 `json:"quantity"`
	OfferedPricePerKg float64 `json:"offeredPricePerKg"`
}

// Presence, not positivity, decides which contract version is authoritative.
// Explicit canonical zero/null cannot be bypassed by also sending a legacy field.
func (in *CreateInput) UnmarshalJSON(raw []byte) error {
	type plain CreateInput
	var value plain
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	*in = CreateInput(value)
	_, in.canonicalWeightSet = fields["approximateWeight"]
	_, in.canonicalPriceSet = fields["offeredPricePerUnit"]
	return nil
}

func (in CreateInput) weight() float64 {
	if in.canonicalWeightSet || in.ApproximateWeight != 0 {
		return in.ApproximateWeight
	}
	return in.Quantity
}
func (in CreateInput) price() float64 {
	if in.canonicalPriceSet || in.OfferedPricePerUnit != 0 {
		return in.OfferedPricePerUnit
	}
	return in.OfferedPricePerKg
}

func ValidateOffer(lot *listings.EWasteLot, in CreateInput) error {
	if !lot.IsAvailable() {
		return ErrLotUnavailable
	}
	if math.IsNaN(in.weight()) || math.IsInf(in.weight(), 0) || math.IsNaN(in.price()) || math.IsInf(in.price(), 0) || math.IsInf(in.weight()*in.price()*100, 0) || in.weight()*in.price() > 1e12 {
		return fmt.Errorf("offer values must be finite and total at most 1 trillion INR")
	}
	if len(in.Message) > 2000 {
		return fmt.Errorf("message must be at most 2000 bytes")
	}
	if lot.Quantity.Unit == "unit" && in.weight() != math.Trunc(in.weight()) {
		return fmt.Errorf("unit quantity must be a whole number")
	}
	if in.weight() <= 0 {
		return fmt.Errorf("offer weight/quantity must be positive")
	}
	if in.price() <= 0 {
		return fmt.Errorf("offered price must be positive")
	}
	if math.Round(in.weight()*in.price()*100) < 1 {
		return fmt.Errorf("offered total must be at least INR 0.01")
	}
	if in.weight() > lot.AvailableAmount() {
		return ErrQuantityUnavailable
	}
	return nil
}
func ValidateRecyclerForLot(recycler *users.User, lot *listings.EWasteLot, now time.Time, allowDemo ...bool) error {
	if !recycler.CanParticipate(now, allowDemo...) {
		return ErrRecyclerUnauthorized
	}
	if !recycler.AcceptsMaterial(lot.Material.Category, lot.Material.SubCategory) {
		return ErrMaterialNotAccepted
	}
	return nil
}
func (s *Service) Create(ctx context.Context, recyclerID, lotID primitive.ObjectID, in CreateInput) (*Offer, error) {
	lot, err := s.listings.RawByID(ctx, lotID)
	if err != nil {
		return nil, err
	}
	if err := ValidateOffer(lot, in); err != nil {
		return nil, err
	}
	recycler, err := s.users.ByID(ctx, recyclerID)
	if err != nil {
		return nil, err
	}
	if err := ValidateRecyclerForLot(recycler, lot, time.Now().UTC(), s.allowDemo); err != nil {
		return nil, err
	}
	n, err := s.repo.PendingByRecyclerAndLot(ctx, recyclerID, lotID)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, ErrDuplicatePending
	}
	now := time.Now().UTC()
	weight, price := in.weight(), in.price()
	total := math.Round(weight*price*100) / 100
	o := &Offer{LotID: lot.ID, ListingID: lot.ID, CollectorID: lot.CollectorID, KabadiwalaID: lot.CollectorID, RecyclerID: recyclerID, ApproximateWeight: weight, Quantity: weight, OfferedPricePerUnit: price, OfferedPricePerKg: price, OfferedTotal: total, TotalAmount: total, Unit: lot.Quantity.Unit, PickupAvailable: in.PickupAvailable, Message: strings.TrimSpace(in.Message), Status: "pending", CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
func (s *Service) List(ctx context.Context, recyclerID primitive.ObjectID, status string) ([]View, error) {
	items, err := s.repo.ListByRecycler(ctx, recyclerID, status)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(items))
	for _, o := range items {
		v := View{Offer: o, MaterialName: "E-waste lot (historical reference)", CollectorName: "Collector (historical reference)", Unit: o.Unit}
		lot, err := s.listings.RawByID(ctx, o.EffectiveLotID())
		if err == nil {
			v.MaterialName = lot.Material.SubCategory
			if v.MaterialName == "" {
				v.MaterialName = lot.Material.Category
			}
			v.EstimatedPricePerUnit = lot.UnitPrice()
		} else if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		collector, err := s.users.ByID(ctx, o.CollectorID)
		if err == nil {
			v.CollectorName = collector.BusinessName
			if v.CollectorName == "" {
				v.CollectorName = collector.Name
			}
		} else if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		v.SupplierName = v.CollectorName
		v.SellerPricePerKg = v.EstimatedPricePerUnit
		out = append(out, v)
	}

	return out, nil
}

var ErrLotUnavailable = errors.New("lot cannot receive offers")
var ErrListingUnavailable = ErrLotUnavailable // compatibility for existing tests/callers
var ErrQuantityUnavailable = errors.New("quantity exceeds available lot quantity")
var ErrDuplicatePending = errors.New("a pending offer already exists for this lot")
var ErrRecyclerUnauthorized = errors.New("recycler authorization is not valid for transactions")
var ErrMaterialNotAccepted = errors.New("this material is not included in the recycler accepted-material profile")
