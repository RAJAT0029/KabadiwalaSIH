package matching

import (
	"kabadiconnect/backend/internal/listings"
	"kabadiconnect/backend/internal/users"
	"testing"
	"time"
)

func TestUnauthorizedRecyclerNeverEligible(t *testing.T) {
	lot := &listings.EWasteLot{Material: listings.Material{Category: "PCB"}}
	u := &users.User{Role: "recycler", IsActive: true, AcceptedMaterials: []string{"PCB"}, Authorization: users.Authorization{Status: "pending_verification"}}
	if Score(lot, u, Input{}, time.Now()).Eligible {
		t.Fatal("unauthorized recycler must not be eligible")
	}
}
func TestDeterministicScoreRewardsPickupAndDistance(t *testing.T) {
	lot := &listings.EWasteLot{Material: listings.Material{Category: "PCB"}, Valuation: listings.Valuation{EstimatedPricePerUnit: 300}}
	u := &users.User{Role: "recycler", IsActive: true, AcceptedMaterials: []string{"PCB"}, Authorization: users.Authorization{Status: "demo_verified", IsDemo: true}, Pickup: users.PickupProfile{Available: true}}
	d := 8.0
	r := 295.0
	got := Score(lot, u, Input{AllowDemo: true, DistanceKm: &d, RecyclerOfferedRate: &r}, time.Now())
	if !got.Eligible || got.Score < 90 {
		t.Fatalf("unexpected score: %+v", got)
	}
}
