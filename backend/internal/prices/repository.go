package prices

import (
	"context"
	"kabadiconnect/backend/internal/database"
	"math"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct{ c *mongo.Collection }

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{c: db.Collection("priceRecords")}
}
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{{Keys: bson.D{{Key: "materialCategory", Value: 1}, {Key: "materialSubCategory", Value: 1}, {Key: "timestamp", Value: -1}}}, {Keys: bson.D{{Key: "location.city", Value: 1}, {Key: "location.state", Value: 1}, {Key: "timestamp", Value: -1}}}})
	return err
}
func (r *Repository) Board(ctx context.Context, city string) (Board, error) {
	filter := bson.M{}
	if strings.TrimSpace(city) != "" {
		filter["location.city"] = bson.M{"$regex": "^" + regexp.QuoteMeta(strings.TrimSpace(city)) + "$", "$options": "i"}
	}
	cur, err := r.c.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(300))
	if err != nil {
		return Board{}, err
	}
	defer cur.Close(ctx)
	rows := make([]PriceRecord, 0)
	if err := cur.All(ctx, &rows); err != nil {
		return Board{}, err
	}
	groups := map[string][]PriceRecord{}
	order := []string{}
	for _, row := range rows {
		k := strings.Join([]string{row.MaterialCategory, row.MaterialSubCategory, row.Unit, row.Location.City, row.Location.State, row.SourceType}, "\x00")
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], row)
	}
	entries := make([]BoardEntry, 0, len(groups))
	for _, k := range order {
		g := groups[k]
		if len(g) == 0 {
			continue
		}
		latest := g[0]
		min, max := latest.BuyingPrice, latest.BuyingPrice
		for _, x := range g {
			if x.BuyingPrice < min {
				min = x.BuyingPrice
			}
			if x.BuyingPrice > max {
				max = x.BuyingPrice
			}
		}
		trend := "unknown"
		if len(g) > 1 {
			trend = "stable"
			delta := latest.BuyingPrice - g[1].BuyingPrice
			if math.Abs(delta) >= 1 {
				if delta > 0 {
					trend = "up"
				} else {
					trend = "down"
				}
			}
		}
		entries = append(entries, BoardEntry{MaterialCategory: latest.MaterialCategory, MaterialSubCategory: latest.MaterialSubCategory, CurrentBuyingRate: latest.BuyingPrice, QuotedPrice: latest.QuotedPrice, Unit: latest.Unit, City: latest.Location.City, State: latest.Location.State, MarketRangeMin: min, MarketRangeMax: max, Trend: trend, UpdatedAt: latest.Timestamp, SourceType: latest.SourceType, IsDemo: latest.IsDemo})
	}
	return Board{Currency: "INR", Notice: "Development/demo price records only. They are not live Indian market prices and must not be treated as production quotations.", Entries: entries}, nil
}
func (r *Repository) History(ctx context.Context, category, subCategory string, limit int) ([]PriceRecord, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	filter := bson.M{}
	if category != "" {
		filter["materialCategory"] = category
	}
	if subCategory != "" {
		filter["materialSubCategory"] = subCategory
	}
	cur, err := r.c.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	rows := make([]PriceRecord, 0)
	err = cur.All(ctx, &rows)
	return rows, err
}
