package offers

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"kabadiconnect/backend/internal/listings"
	"kabadiconnect/backend/internal/users"
)

func TestValidateOffer(t *testing.T) {
	base := listings.EWasteLot{Status: listings.LotStatusAvailable, Quantity: listings.Quantity{AvailableWeight: 50, Unit: "kg"}}
	cases := []struct {
		name string
		in   CreateInput
		want error
	}{
		{"valid", CreateInput{ApproximateWeight: 20, OfferedPricePerUnit: 25}, nil},
		{"negative quantity", CreateInput{ApproximateWeight: -1, OfferedPricePerUnit: 25}, errors.New("x")},
		{"over available", CreateInput{ApproximateWeight: 51, OfferedPricePerUnit: 25}, ErrQuantityUnavailable},
		{"zero price", CreateInput{ApproximateWeight: 10, OfferedPricePerUnit: 0}, errors.New("x")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOffer(&base, tc.in)
			if tc.want == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want != nil && err == nil {
				t.Fatal("expected error")
			}
			if errors.Is(tc.want, ErrQuantityUnavailable) && !errors.Is(err, ErrQuantityUnavailable) {
				t.Fatalf("expected quantity unavailable, got %v", err)
			}
		})
	}
}

func TestCanonicalZeroCannotUseLegacyFallback(t *testing.T) {
	for _, raw := range []string{`{"approximateWeight":0,"quantity":10,"offeredPricePerUnit":2}`, `{"approximateWeight":10,"offeredPricePerUnit":null,"offeredPricePerKg":2}`} {
		var in CreateInput
		if err := json.Unmarshal([]byte(raw), &in); err != nil {
			t.Fatal(err)
		}
		lot := &listings.EWasteLot{Status: listings.LotStatusAvailable, Quantity: listings.Quantity{AvailableWeight: 20}}
		if ValidateOffer(lot, in) == nil {
			t.Fatal("canonical zero bypassed")
		}
	}
}

func TestRejectNonFiniteAndInvalidQuantities(t *testing.T) {
	l := &listings.EWasteLot{Status: listings.LotStatusAvailable, Quantity: listings.Quantity{ApproximateWeight: 50, AvailableWeight: 50, Unit: "kg"}}
	for _, in := range []CreateInput{{ApproximateWeight: math.NaN(), OfferedPricePerUnit: 1}, {ApproximateWeight: 1, OfferedPricePerUnit: math.Inf(1)}, {ApproximateWeight: 1, OfferedPricePerUnit: 1e308}, {ApproximateWeight: -1, Quantity: 10, OfferedPricePerUnit: 1}, {ApproximateWeight: 0, OfferedPricePerUnit: 1}, {ApproximateWeight: 1, OfferedPricePerUnit: -1}} {
		if ValidateOffer(l, in) == nil {
			t.Fatalf("invalid offer accepted: %+v", in)
		}
	}
	l.Quantity.Unit = "unit"
	if ValidateOffer(l, CreateInput{ApproximateWeight: 1.5, OfferedPricePerUnit: 1}) == nil {
		t.Fatal("fractional unit accepted")
	}
}

func TestValidateOfferRejectsUnavailableLot(t *testing.T) {
	lot := listings.EWasteLot{Status: listings.LotStatusCancelled, Quantity: listings.Quantity{AvailableWeight: 10}}
	if !errors.Is(ValidateOffer(&lot, CreateInput{ApproximateWeight: 1, OfferedPricePerUnit: 1}), ErrLotUnavailable) {
		t.Fatal("expected unavailable lot to be rejected")
	}
}

func TestValidateRecyclerForLot(t *testing.T) {
	now := time.Now().UTC()
	lot := &listings.EWasteLot{Material: listings.Material{Category: "PCB", SubCategory: "Mixed PCBs"}}
	recycler := &users.User{Role: "recycler", IsActive: true, AcceptedMaterials: []string{"PCB"}, Authorization: users.Authorization{Status: "demo_verified", IsDemo: true}}
	if err := ValidateRecyclerForLot(recycler, lot, now, true); err != nil {
		t.Fatalf("expected eligible development recycler: %v", err)
	}
	recycler.Authorization.Status = "pending_verification"
	if !errors.Is(ValidateRecyclerForLot(recycler, lot, now, true), ErrRecyclerUnauthorized) {
		t.Fatal("pending recycler must not be eligible")
	}
	recycler.Authorization.Status = "demo_verified"
	recycler.AcceptedMaterials = []string{"Cable"}
	if !errors.Is(ValidateRecyclerForLot(recycler, lot, now, true), ErrMaterialNotAccepted) {
		t.Fatal("material mismatch must be rejected")
	}
}
