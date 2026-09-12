package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/database"
	"kabadiconnect/backend/internal/listings"
	"kabadiconnect/backend/internal/prices"
	"kabadiconnect/backend/internal/users"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Environment != "development" {
		log.Fatal("seed is restricted to ENVIRONMENT=development")
	}
	client, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.MongoDatabase)
	ctx := context.Background()
	now := time.Now().UTC()
	usersCol := db.Collection("users")
	lotsCol := db.Collection("scrapListings") // physical V1 name retained non-destructively
	pricesCol := db.Collection("priceRecords")

	ramesh := upsertCollector(ctx, usersCol, collectorSeed{Email: "ramesh.ewaste.demo@kabadiconnect.local", Phone: "+919999900011", Name: "Ramesh Kumar", Business: "Ramesh E-Waste Collector", City: "Kurukshetra", State: "Haryana", Pincode: "136118", Language: "hi", Coordinates: []float64{76.8783, 29.9695}}, now)
	sharma := upsertCollector(ctx, usersCol, collectorSeed{Email: "sharma.ewaste.demo@kabadiconnect.local", Phone: "+919999900012", Name: "Amit Sharma", Business: "Sharma Electronics Scrap Aggregator", City: "Karnal", State: "Haryana", Pincode: "132001", Language: "hi", Coordinates: []float64{76.9905, 29.6857}}, now)
	upsertDemoRecycler(ctx, usersCol, now)

	lots := []listings.EWasteLot{
		lot("KC-EW-DEMO-001", ramesh.ID, "PCB", "Mixed PCBs", "Mixed printed circuit boards from dismantled consumer electronics", "Dismantled", "consumer electronics", 42, "kg", 310, 260, 360, "Kurukshetra, Haryana", []float64{76.8783, 29.9695}, map[string]any{"boardType": "mixed consumer electronics", "visibleComponents": "mixed", "condition": "dismantled"}, now.Add(-2*time.Hour)),
		lot("KC-EW-DEMO-002", sharma.ID, "Cable", "Copper Cable", "Copper-bearing cables from end-of-life electrical/electronic equipment", "Loose / sorted", "electrical and electronic equipment", 65, "kg", 128, 105, 150, "Karnal, Haryana", []float64{76.9905, 29.6857}, map[string]any{"cableType": "mixed power/data cable", "insulationPresent": true, "condition": "sorted"}, now.Add(-4*time.Hour)),
		lot("KC-EW-DEMO-003", ramesh.ID, "LCD Panel", "Damaged LCD Panels", "Damaged LCD display panels removed from end-of-life devices", "Damaged", "display equipment", 24, "unit", 180, 150, 220, "Panipat, Haryana", []float64{76.9635, 29.3909}, map[string]any{"panelType": "mixed LCD", "cracked": true, "sizeClass": "mixed"}, now.Add(-6*time.Hour)),
		lot("KC-EW-DEMO-004", sharma.ID, "Battery", "Lithium-Ion Battery Lot", "Mixed lithium-ion batteries from end-of-life portable electronics", "Mixed condition", "portable electronics", 28, "kg", 145, 120, 175, "Ambala, Haryana", []float64{76.7794, 30.3782}, map[string]any{"batteryType": "portable device battery", "chemistry": "lithium-ion", "damaged": false, "swollen": false, "hazardousHandling": true}, now.Add(-8*time.Hour)),
		lot("KC-EW-DEMO-005", ramesh.ID, "Motor", "Electronic Motors", "Small motors recovered from discarded electrical/electronic equipment", "For recovery", "electrical equipment", 75, "kg", 96, 80, 115, "Rohtak, Haryana", []float64{76.6066, 28.8955}, map[string]any{"motorType": "mixed small motors", "condition": "for material recovery"}, now.Add(-11*time.Hour)),
		lot("KC-EW-DEMO-006", sharma.ID, "Magnet-bearing Assembly", "Magnet-Bearing Assemblies", "Mixed assemblies containing recoverable magnets from end-of-life equipment", "Dismantled", "electrical and electronic equipment", 18, "kg", 170, 145, 205, "Gurugram, Haryana", []float64{77.0266, 28.4595}, map[string]any{"assemblyType": "mixed speaker/drive assemblies", "condition": "dismantled"}, now.Add(-14*time.Hour)),
		lot("KC-EW-DEMO-007", ramesh.ID, "Mixed Plastics from EEE", "Mixed E-Waste Plastics", "Plastic fractions generated from dismantled end-of-life electrical/electronic equipment", "Sorted fractions", "dismantled EEE", 95, "kg", 32, 24, 40, "Kurukshetra, Haryana", []float64{76.8783, 29.9695}, map[string]any{"source": "EEE housings and components", "sorted": true}, now.Add(-18*time.Hour)),
		lot("KC-EW-DEMO-008", sharma.ID, "CRT", "CRT Units", "Intact and damaged CRT units awaiting formal recycler handover", "Mixed condition", "legacy display equipment", 12, "unit", 260, 220, 310, "Karnal, Haryana", []float64{76.9905, 29.6857}, map[string]any{"intact": 8, "damaged": 4, "hazardousHandling": true}, now.Add(-20*time.Hour)),
	}
	for _, l := range lots {
		if _, err := lotsCol.UpdateOne(ctx, bson.M{"referenceId": l.ReferenceID}, bson.M{"$setOnInsert": withSeedLot(l)}, options.Update().SetUpsert(true)); err != nil {
			log.Fatal(err)
		}
	}
	seedPrices(ctx, pricesCol, now)
	log.Printf("ensured %d e-waste demo lots and demo price history without overwriting existing records", len(lots))
}

type collectorSeed struct {
	Email, Phone, Name, Business, City, State, Pincode, Language string
	Coordinates                                                  []float64
}

func upsertCollector(ctx context.Context, col *mongo.Collection, s collectorSeed, now time.Time) users.User {
	id := primitive.NewObjectID()
	_, err := col.UpdateOne(ctx, bson.M{"email": s.Email}, bson.M{"$setOnInsert": bson.M{
		"role": "kabadiwala", "name": s.Name, "phone": s.Phone, "email": s.Email, "businessName": s.Business,
		"address":  users.Address{Street: "Development seed location", City: s.City, State: s.State, Pincode: s.Pincode},
		"location": users.GeoPoint{Type: "Point", Coordinates: s.Coordinates}, "preferredLanguage": s.Language,
		"verificationStatus": "development_profile", "passwordHash": "seed-account-not-for-login", "isActive": true,
		"updatedAt": now,
		"_id":       id, "createdAt": now.AddDate(-1, 0, 0)}}, options.Update().SetUpsert(true))
	if err != nil {
		log.Fatal(err)
	}
	var u users.User
	if err := col.FindOne(ctx, bson.M{"email": s.Email}).Decode(&u); err != nil {
		log.Fatal(err)
	}
	return u
}

func upsertDemoRecycler(ctx context.Context, col *mongo.Collection, now time.Time) {
	id := primitive.NewObjectID()
	validUntil := now.AddDate(1, 0, 0)
	_, err := col.UpdateOne(ctx, bson.M{"email": "greencycle.demo@kabadiconnect.local"}, bson.M{"$setOnInsert": bson.M{
		"role": "recycler", "name": "Development Recycler Contact", "phone": "+919999900021", "email": "greencycle.demo@kabadiconnect.local",
		"companyName": "GreenCycle E-Waste Recycling Facility", "facilityName": "GreenCycle E-Waste Recycling Facility", "businessType": "E-waste Recycler / Processing Unit",
		"address":            users.Address{Street: "Development seed facility", City: "Kurukshetra", State: "Haryana", Pincode: "136118"},
		"acceptedMaterials":  listings.SupportedCategories,
		"authorization":      users.Authorization{RegistrationNumber: "DEMO-AUTH-NOT-GOVERNMENT", Authority: "Development data only", ValidUntil: &validUntil, Status: "demo_verified", IsDemo: true},
		"pickup":             users.PickupProfile{Available: true, ServiceArea: []string{"Kurukshetra", "Karnal", "Ambala"}},
		"verificationStatus": "demo_verified", "passwordHash": "seed-account-not-for-login", "isActive": true, "updatedAt": now,
		"_id": id, "createdAt": now.AddDate(-1, 0, 0)}}, options.Update().SetUpsert(true))
	if err != nil {
		log.Fatal(err)
	}
}

func lot(ref string, collectorID primitive.ObjectID, category, subCategory, description, condition, sourceType string, weight float64, unit string, price, min, max float64, label string, coords []float64, attrs map[string]any, created time.Time) listings.EWasteLot {
	point := listings.GeoPoint{Type: "Point", Coordinates: coords}
	return listings.EWasteLot{ReferenceID: ref, CollectorID: collectorID, KabadiwalaID: collectorID, Domain: listings.DomainEWaste, Material: listings.Material{Category: category, SubCategory: subCategory, SubType: subCategory, Description: description, Grade: description, Condition: condition, SourceType: sourceType, Attributes: attrs}, Quantity: listings.Quantity{ApproximateWeight: weight, AvailableWeight: weight, Unit: unit, Total: weight, Available: weight}, Valuation: listings.Valuation{EstimatedValue: weight * price, EstimatedPricePerUnit: price, MarketRangeMin: min, MarketRangeMax: max, EstimationSource: "demo_historical_reference"}, Pricing: listings.LegacyPricing{ExpectedPricePerKg: price, Negotiable: true, AllowDirectPurchase: false}, Collection: listings.CollectionDetails{CollectedAt: created.Add(-time.Hour), Location: point, Label: label}, Location: point, LocationLabel: label, Status: listings.LotStatusAvailable, CreatedAt: created, UpdatedAt: time.Now().UTC()}
}
func withSeedLot(l listings.EWasteLot) bson.M {
	b, _ := bson.Marshal(l)
	var m bson.M
	_ = bson.Unmarshal(b, &m)
	m["seedTag"] = "ewaste-demo-v2"
	return m
}

func seedPrices(ctx context.Context, col *mongo.Collection, now time.Time) {
	rows := []prices.PriceRecord{
		price("PCB", "Mixed PCBs", 310, 325, "kg", "Kurukshetra", "Haryana", now), price("PCB", "Mixed PCBs", 300, 318, "kg", "Kurukshetra", "Haryana", now.Add(-24*time.Hour)),
		price("Cable", "Copper Cable", 128, 138, "kg", "Karnal", "Haryana", now), price("Cable", "Copper Cable", 124, 136, "kg", "Karnal", "Haryana", now.Add(-24*time.Hour)),
		price("LCD Panel", "Damaged LCD Panels", 180, 195, "unit", "Panipat", "Haryana", now), price("LCD Panel", "Damaged LCD Panels", 185, 198, "unit", "Panipat", "Haryana", now.Add(-24*time.Hour)),
		price("CRT", "CRT Units", 260, 280, "unit", "Karnal", "Haryana", now), price("CRT", "CRT Units", 260, 280, "unit", "Karnal", "Haryana", now.Add(-24*time.Hour)),
		price("Battery", "Lithium-Ion Battery Lot", 145, 158, "kg", "Ambala", "Haryana", now), price("Battery", "Lithium-Ion Battery Lot", 139, 152, "kg", "Ambala", "Haryana", now.Add(-24*time.Hour)),
		price("Motor", "Electronic Motors", 96, 104, "kg", "Rohtak", "Haryana", now), price("Motor", "Electronic Motors", 98, 106, "kg", "Rohtak", "Haryana", now.Add(-24*time.Hour)),
		price("Magnet-bearing Assembly", "Magnet-Bearing Assemblies", 170, 185, "kg", "Gurugram", "Haryana", now), price("Magnet-bearing Assembly", "Magnet-Bearing Assemblies", 165, 181, "kg", "Gurugram", "Haryana", now.Add(-24*time.Hour)),
		price("Mixed Plastics from EEE", "Mixed E-Waste Plastics", 32, 36, "kg", "Kurukshetra", "Haryana", now), price("Mixed Plastics from EEE", "Mixed E-Waste Plastics", 31, 35, "kg", "Kurukshetra", "Haryana", now.Add(-24*time.Hour)),
	}
	for i, r := range rows {
		b, err := bson.Marshal(r)
		if err != nil {
			log.Fatal(err)
		}
		var m bson.M
		if err := bson.Unmarshal(b, &m); err != nil {
			log.Fatal(err)
		}
		m["seedTag"] = "ewaste-demo-v2"
		// Stable slots also match the previous seed's rows, preserving observation time.
		existing := bson.M{"seedTag": "ewaste-demo-v2", "materialCategory": r.MaterialCategory, "materialSubCategory": r.MaterialSubCategory}
		count, err := col.CountDocuments(ctx, existing)
		if err != nil {
			log.Fatal(err)
		}
		if count >= int64(i%2+1) {
			continue
		}
		m["seedKey"] = r.MaterialCategory + "-" + fmt.Sprint(i%2)
		if _, err := col.UpdateOne(ctx, bson.M{"seedKey": m["seedKey"]}, bson.M{"$setOnInsert": m}, options.Update().SetUpsert(true)); err != nil {
			log.Fatal(err)
		}
	}

}
func price(category, sub string, buy, quote float64, unit, city, state string, ts time.Time) prices.PriceRecord {
	return prices.PriceRecord{MaterialCategory: category, MaterialSubCategory: sub, Location: prices.Location{City: city, State: state}, Timestamp: ts, BuyingPrice: buy, QuotedPrice: quote, Unit: unit, SourceType: "demo_reference", IsDemo: true, CreatedAt: ts}
}
