package matching

import (
	"math"
	"time"

	"kabadiconnect/backend/internal/listings"
	"kabadiconnect/backend/internal/users"
)

type Input struct {
	AllowDemo           bool // Explicit runtime policy; false outside development.
	DistanceKm          *float64
	RecyclerOfferedRate *float64
}

type Result struct {
	Eligible bool    `json:"eligible"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason,omitempty"`
}

// Score applies deterministic V1 ranking. Authorization and material
// compatibility are hard gates; location/rate/pickup only rank eligible recyclers.
func Score(lot *listings.EWasteLot, recycler *users.User, in Input, now time.Time) Result {
	if !recycler.CanParticipate(now, in.AllowDemo) {
		return Result{Reason: "authorization_not_valid"}
	}
	if !recycler.AcceptsMaterial(lot.Material.Category, lot.Material.SubCategory) {
		return Result{Reason: "material_not_accepted"}
	}
	score := 50.0
	if recycler.Pickup.Available {
		score += 10
	}
	if in.DistanceKm != nil {
		d := math.Max(0, *in.DistanceKm)
		switch {
		case d <= 10:
			score += 25
		case d <= 25:
			score += 18
		case d <= 50:
			score += 10
		case d <= 100:
			score += 4
		}
	}
	if in.RecyclerOfferedRate != nil && lot.UnitPrice() > 0 {
		ratio := *in.RecyclerOfferedRate / lot.UnitPrice()
		if ratio >= 1 {
			score += 15
		} else if ratio >= 0.9 {
			score += 10
		} else if ratio >= 0.8 {
			score += 5
		}
	}
	if score > 100 {
		score = 100
	}
	return Result{Eligible: true, Score: score}
}
