package listings

import (
	"context"
	"kabadiconnect/backend/internal/database"
	"math"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kabadiconnect/backend/internal/users"
)

type Repository struct {
	c     *mongo.Collection
	users *users.Repository
}

func NewRepository(db *mongo.Database, u *users.Repository) *Repository {
	return &Repository{c: db.Collection("scrapListings"), users: u}
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{
		{Keys: bson.D{{Key: "collectorId", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "material.category", Value: 1}, {Key: "material.subCategory", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "collection.location", Value: "2dsphere"}}},
		{Keys: bson.D{{Key: "location", Value: "2dsphere"}}},
	})
	return err
}

type Query struct {
	Search, Category, SubCategory, Condition, City, Sort string
	MinQuantity, MaxPrice                                *float64
	Page, Limit                                          int
	Longitude, Latitude                                  *float64
}

func eWasteBaseClauses() []bson.M {
	return []bson.M{
		{"$or": []bson.M{{"domain": DomainEWaste}, {"material.category": LegacyEWasteCategory}, {"material.category": bson.M{"$in": SupportedCategories}}}},
		{"status": bson.M{"$in": []string{LotStatusAvailable, LotStatusOfferReceived, "active"}}},
		{"$expr": bson.M{"$gt": []any{availableExpression(), 0}}},
	}
}

func (r *Repository) List(ctx context.Context, q Query) ([]View, int64, error) {
	clauses := eWasteBaseClauses()
	if q.Category != "" {
		clauses = append(clauses, bson.M{"material.category": exactRegex(q.Category)})
	}
	if q.SubCategory != "" {
		clauses = append(clauses, bson.M{"$or": []bson.M{{"material.subCategory": exactRegex(q.SubCategory)}, {"material.subType": exactRegex(q.SubCategory)}}})
	}
	if q.Condition != "" {
		clauses = append(clauses, bson.M{"material.condition": exactRegex(q.Condition)})
	}
	if q.City != "" {
		clauses = append(clauses, bson.M{"$or": []bson.M{{"collection.label": containsRegex(q.City)}, {"locationLabel": containsRegex(q.City)}}})
	}
	if q.MinQuantity != nil {
		clauses = append(clauses, bson.M{"$expr": bson.M{"$gte": []any{availableExpression(), *q.MinQuantity}}})
	}
	if q.MaxPrice != nil {
		clauses = append(clauses, bson.M{"$or": []bson.M{{"valuation.estimatedPricePerUnit": bson.M{"$lte": *q.MaxPrice, "$gt": 0}}, {"pricing.expectedPricePerKg": bson.M{"$lte": *q.MaxPrice, "$gt": 0}}}})
	}
	if q.Search != "" {
		s := containsRegex(q.Search)
		clauses = append(clauses, bson.M{"$or": []bson.M{
			{"referenceId": s}, {"material.category": s}, {"material.subCategory": s}, {"material.subType": s},
			{"material.description": s}, {"material.grade": s}, {"material.condition": s}, {"material.sourceType": s},
			{"collection.label": s}, {"locationLabel": s},
		}})
	}
	filter := bson.M{"$and": clauses}
	total, err := r.c.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSkip(int64((q.Page - 1) * q.Limit)).SetLimit(int64(q.Limit))
	switch q.Sort {
	case "lowest_price":
		opts.SetSort(bson.D{{Key: "valuation.estimatedPricePerUnit", Value: 1}, {Key: "createdAt", Value: -1}})
	case "highest_quantity":
		opts.SetSort(bson.D{{Key: "quantity.availableWeight", Value: -1}, {Key: "createdAt", Value: -1}})
	default:
		opts.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}
	cur, err := r.c.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var lots []EWasteLot
	if err := cur.All(ctx, &lots); err != nil {
		return nil, 0, err
	}
	views := make([]View, 0, len(lots))
	for i := range lots {
		lots[i].Normalize()
		v, err := r.toView(ctx, lots[i], q.Longitude, q.Latitude)
		if err == nil {
			views = append(views, v)
		}
	}
	if q.Sort == "nearest" && q.Longitude != nil && q.Latitude != nil {
		sortViewsByDistance(views)
	}
	return views, total, nil
}

func (r *Repository) ByID(ctx context.Context, id primitive.ObjectID, lon, lat *float64) (*View, error) {
	lot, err := r.RawByID(ctx, id)
	if err != nil || !lot.IsAvailable() {
		if err != nil {
			return nil, err
		}
		return nil, mongo.ErrNoDocuments
	}
	v, err := r.toView(ctx, *lot, lon, lat)
	return &v, err
}

func (r *Repository) RawByID(ctx context.Context, id primitive.ObjectID) (*EWasteLot, error) {
	var lot EWasteLot
	clauses := eWasteBaseClauses()[:1]
	clauses = append(clauses, bson.M{"_id": id})
	err := r.c.FindOne(ctx, bson.M{"$and": clauses}).Decode(&lot)
	if err == nil {
		lot.Normalize()
	}
	return &lot, err
}

func (r *Repository) toView(ctx context.Context, lot EWasteLot, lon, lat *float64) (View, error) {
	collector, err := r.users.PublicByID(ctx, lot.CollectorID)
	if err != nil {
		return View{}, err
	}
	business := collector.BusinessName
	if business == "" {
		business = collector.Name
	}
	location := collector.Address.City
	if collector.Address.State != "" {
		if location != "" {
			location += ", "
		}
		location += collector.Address.State
	}
	summary := CollectorSummary{ID: collector.ID, BusinessName: business, OperatingLocation: location, ProfileStatus: collector.VerificationStatus, PreferredLanguage: collector.PreferredLanguage, MemberSince: collector.CreatedAt}
	v := View{Lot: lot, Listing: lot, Collector: summary, Supplier: summary}
	point := lot.Collection.Location
	if len(point.Coordinates) != 2 {
		point = lot.Location
	}
	if lon != nil && lat != nil && len(point.Coordinates) == 2 {
		d := haversine(*lat, *lon, point.Coordinates[1], point.Coordinates[0])
		v.DistanceKm = &d
	}
	return v, nil
}

func exactRegex(s string) bson.M {
	return bson.M{"$regex": "^" + regexpEscape(s) + "$", "$options": "i"}
}
func containsRegex(s string) bson.M { return bson.M{"$regex": regexpEscape(s), "$options": "i"} }
func regexpEscape(s string) string {
	r := strings.NewReplacer("\\", "\\\\", ".", "\\.", "*", "\\*", "+", "\\+", "?", "\\?", "(", "\\(", ")", "\\)", "[", "\\[", "]", "\\]", "{", "\\{", "}", "\\}", "^", "\\^", "$", "\\$", "|", "\\|")
	return r.Replace(strings.TrimSpace(s))
}
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return math.Round((R*2*math.Atan2(math.Sqrt(a), math.Sqrt(1-a)))*10) / 10
}
func sortViewsByDistance(v []View) {
	for i := 0; i < len(v); i++ {
		for j := i + 1; j < len(v); j++ {
			ai, aj := math.MaxFloat64, math.MaxFloat64
			if v[i].DistanceKm != nil {
				ai = *v[i].DistanceKm
			}
			if v[j].DistanceKm != nil {
				aj = *v[j].DistanceKm
			}
			if aj < ai {
				v[i], v[j] = v[j], v[i]
			}
		}
	}
}
func ParseFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func availableExpression() bson.M {
	return bson.M{"$ifNull": []any{"$quantity.availableWeight", bson.M{"$ifNull": []any{"$quantity.available", 0}}}}
}
