package listings

import "testing"

func TestNormalizeNeverRestoresDepletedStock(t *testing.T) {
	l := EWasteLot{Status: LotStatusAvailable, Quantity: Quantity{ApproximateWeight: 42, AvailableWeight: 0, Available: 42, Total: 42}}
	l.Normalize()
	l.Normalize()
	if l.AvailableAmount() != 0 {
		t.Fatal("depleted canonical stock restored")
	}
	legacy := EWasteLot{Quantity: Quantity{Total: 42, Available: 12}}
	legacy.Normalize()
	legacy.Normalize()
	if legacy.AvailableAmount() != 12 {
		t.Fatal("legacy stock lost")
	}
}

func TestOfferReceivedRemainsOpenForOffers(t *testing.T) {
	if !(&EWasteLot{Status: LotStatusOfferReceived}).IsAvailable() {
		t.Fatal("open negotiation not available")
	}
}
