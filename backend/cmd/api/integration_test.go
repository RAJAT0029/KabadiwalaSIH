package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/database"
	"kabadiconnect/backend/internal/handovers"
)

func testDatabase(t *testing.T) (*mongo.Database, config.Config) {
	t.Helper()
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("set MONGODB_TEST_URI to run real MongoDB integration tests")
	}
	client, err := database.Connect(uri)
	if err != nil {
		t.Fatal(err)
	}
	name := "kc_test_" + primitive.NewObjectID().Hex()
	db := client.Database(name)
	t.Cleanup(func() {
		if strings.HasPrefix(name, "kc_test_") {
			_ = db.Drop(context.Background())
		}
		_ = client.Disconnect(context.Background())
	})
	return db, config.Config{Environment: "development", MongoURI: uri, MongoDatabase: name, AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("b", 32), AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour, FrontendOrigin: "http://localhost:5173"}
}

func request(t *testing.T, server, method, path, token string, body any, cookie *http.Cookie, want int) (map[string]any, *http.Response) {
	t.Helper()
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	r, err := http.NewRequest(method, server+path, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != want {
		t.Fatalf("%s %s: got %d want %d: %s", method, path, resp.StatusCode, want, raw)
	}
	var payload map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatal(err)
		}
	}
	if bytes.Contains(raw, []byte("passwordHash")) {
		t.Fatal("password hash exposed")
	}
	if inner, ok := payload["data"].(map[string]any); ok {
		return inner, resp
	}
	return payload, resp
}

func register(t *testing.T, server, suffix string) (map[string]any, string, *http.Cookie) {
	d, resp := request(t, server, "POST", "/api/v1/auth/register", "", bson.M{"name": "Test Recycler", "companyName": "Test E-Waste Facility", "email": suffix + "@example.test", "phone": suffix, "password": "Integration-password-123", "termsAccepted": true, "acceptedMaterials": []string{"PCB"}}, nil, 201)
	if len(resp.Cookies()) != 1 || !resp.Cookies()[0].HttpOnly {
		t.Fatal("missing HTTP-only refresh cookie")
	}
	return d["user"].(map[string]any), d["accessToken"].(string), resp.Cookies()[0]
}

func TestMongoRecyclerFlow(t *testing.T) {
	db, cfg := testDatabase(t)
	ctx := context.Background()
	// Reproduce historical geo options and custom names before normal API startup.
	for _, col := range []string{"users", "scrapListings"} {
		_, err := db.Collection(col).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "location", Value: "2dsphere"}}, Options: options.Index().SetName("location_2dsphere").SetSparse(true)})
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetName("historical_email_unique").SetUnique(true)})
	if err != nil {
		t.Fatal(err)
	}
	h, err := newHandler(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h)
	defer server.Close()
	t.Run("repeated startup with legacy indexes", func(t *testing.T) {
		if _, err := newHandler(cfg, db); err != nil {
			t.Fatal(err)
		}
	})
	u, token, refresh := register(t, server.URL, "first")
	other, otherToken, _ := register(t, server.URL, "second")
	rid, _ := primitive.ObjectIDFromHex(u["id"].(string))
	otherID, _ := primitive.ObjectIDFromHex(other["id"].(string))
	collectorID := primitive.NewObjectID()
	lotID := primitive.NewObjectID()
	now := time.Now().UTC()
	_, err = db.Collection("users").InsertOne(ctx, bson.M{"_id": collectorID, "email": "collector@example.test", "phone": "collector", "role": "kabadiwala", "businessName": "Fictional E-Waste Collector", "createdAt": now})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Collection("scrapListings").InsertOne(ctx, bson.M{"_id": lotID, "referenceId": "TEST-PCB-001", "domain": "e-waste", "collectorId": collectorID, "material": bson.M{"category": "PCB", "subCategory": "Mixed PCBs"}, "quantity": bson.M{"approximateWeight": 42, "availableWeight": 42, "unit": "kg"}, "valuation": bson.M{"estimatedPricePerUnit": 310}, "status": "available", "createdAt": now, "updatedAt": now})
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/lots/" + lotID.Hex() + "/offers"
	t.Run("login and empty dashboard feeds", func(t *testing.T) {
		request(t, server.URL, "POST", "/api/v1/auth/login", "", bson.M{"identifier": "first@example.test", "password": "Integration-password-123", "rememberMe": false}, nil, 200)
		for _, p := range []string{"offers", "transactions", "handovers", "payments"} {
			d, _ := request(t, server.URL, "GET", "/api/v1/"+p, token, nil, nil, 200)
			a, ok := d["items"].([]any)
			if !ok || len(a) != 0 {
				t.Fatalf("%s must return empty array", p)
			}
		}
		d, _ := request(t, server.URL, "GET", "/api/v1/me", token, nil, nil, 200)
		if d["participationAllowed"] != true {
			t.Fatal("demo should be explicitly enabled in development")
		}
		request(t, server.URL, "GET", "/api/v1/lots", token, nil, nil, 200)
		request(t, server.URL, "GET", "/api/v1/lots/"+lotID.Hex(), token, nil, nil, 200)
	})
	t.Run("invalid quantities rates and forged ownership", func(t *testing.T) {
		for _, body := range []bson.M{{"approximateWeight": 0, "offeredPricePerUnit": 1}, {"approximateWeight": -1, "offeredPricePerUnit": 1}, {"approximateWeight": 1, "offeredPricePerUnit": 0}, {"approximateWeight": 1, "offeredPricePerUnit": -1}, {"approximateWeight": 1, "offeredPricePerUnit": 1, "recyclerId": otherID.Hex()}, {"approximateWeight": 1, "offeredPricePerUnit": 1, "offeredTotal": 999}} {
			request(t, server.URL, "POST", path, token, body, nil, 400)
		}
		request(t, server.URL, "POST", path, token, bson.M{"approximateWeight": 43, "offeredPricePerUnit": 1}, nil, 409)
		request(t, server.URL, "GET", "/api/v1/me", "invalid-token", nil, nil, 401)
	})
	t.Run("authorization cannot be self issued or extended", func(t *testing.T) {
		for _, a := range []bson.M{{"status": "authorized"}, {"isDemo": true}, {"validUntil": now.AddDate(10, 0, 0)}} {
			request(t, server.URL, "PATCH", "/api/v1/me", token, bson.M{"authorization": a}, nil, 400)
		}
		prod := cfg
		prod.Environment = "production"
		ph, err := newHandler(prod, db)
		if err != nil {
			t.Fatal(err)
		}
		ps := httptest.NewServer(ph)
		defer ps.Close()
		request(t, ps.URL, "POST", path, token, bson.M{"approximateWeight": 1, "offeredPricePerUnit": 1}, nil, 403)
		pu, _, _ := register(t, ps.URL, "production")
		if pu["authorization"].(map[string]any)["status"] != "pending_verification" || pu["participationAllowed"] != false {
			t.Fatal("production registration must be pending")
		}
	})
	t.Run("pending expired inactive and incompatible profiles", func(t *testing.T) {
		for _, update := range []bson.M{{"authorization.status": "pending_verification"}, {"authorization.status": "demo_verified", "authorization.validUntil": now.Add(-time.Hour)}, {"authorization.validUntil": nil, "acceptedMaterials": []string{"Cable"}}} {
			_, err := db.Collection("users").UpdateByID(ctx, rid, bson.M{"$set": update})
			if err != nil {
				t.Fatal(err)
			}
			request(t, server.URL, "POST", path, token, bson.M{"approximateWeight": 1, "offeredPricePerUnit": 1}, nil, 403)
		}
		_, err := db.Collection("users").UpdateByID(ctx, rid, bson.M{"$set": bson.M{"acceptedMaterials": []string{"PCB"}, "isActive": false}})
		if err != nil {
			t.Fatal(err)
		}
		request(t, server.URL, "GET", "/api/v1/me", token, nil, nil, 401)
		_, err = db.Collection("users").UpdateByID(ctx, rid, bson.M{"$set": bson.M{"isActive": true}})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("cancelled and depleted lots", func(t *testing.T) {
		for _, update := range []bson.M{{"status": "cancelled"}, {"status": "available", "quantity.availableWeight": 0}} {
			_, err := db.Collection("scrapListings").UpdateByID(ctx, lotID, bson.M{"$set": update})
			if err != nil {
				t.Fatal(err)
			}
			request(t, server.URL, "POST", path, token, bson.M{"approximateWeight": 1, "offeredPricePerUnit": 1}, nil, 409)
		}
		_, err := db.Collection("scrapListings").UpdateByID(ctx, lotID, bson.M{"$set": bson.M{"quantity.availableWeight": 42}})
		if err != nil {
			t.Fatal(err)
		}
	})
	var offerID string
	t.Run("offer persisted pending and owner cancel enforced", func(t *testing.T) {
		d, _ := request(t, server.URL, "POST", path, token, bson.M{"approximateWeight": 20, "offeredPricePerUnit": 295}, nil, 201)
		offerID = d["id"].(string)
		if d["offeredTotal"] != float64(5900) || d["recyclerId"] != rid.Hex() || d["collectorId"] != collectorID.Hex() || d["status"] != "pending" {
			t.Fatalf("bad server derived offer: %v", d)
		}
		oid, _ := primitive.ObjectIDFromHex(offerID)
		n, err := db.Collection("offers").CountDocuments(ctx, bson.M{"_id": oid, "status": "pending"})
		if err != nil || n != 1 {
			t.Fatal("offer not persisted")
		}
		list, _ := request(t, server.URL, "GET", "/api/v1/offers?status=pending", token, nil, nil, 200)
		if len(list["items"].([]any)) != 1 {
			t.Fatal("pending offer missing")
		}
		request(t, server.URL, "PATCH", "/api/v1/offers/"+offerID+"/cancel", otherToken, nil, nil, 409)
		request(t, server.URL, "PATCH", "/api/v1/offers/"+offerID+"/cancel", token, nil, nil, 204)
	})
	t.Run("concurrent offers have one winner", func(t *testing.T) {
		var wg sync.WaitGroup
		codes := make(chan int, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("POST", server.URL+path, strings.NewReader(`{"approximateWeight":2,"offeredPricePerUnit":300}`))
				req.Header.Set("Authorization", "Bearer "+token)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					codes <- 0
					return
				}
				_ = resp.Body.Close()
				codes <- resp.StatusCode
			}()
		}
		wg.Wait()
		close(codes)
		created := 0
		for code := range codes {
			if code == 201 {
				created++
			} else if code != 409 {
				t.Fatalf("unexpected concurrent status %d", code)
			}
		}
		if created != 1 {
			t.Fatalf("created %d offers", created)
		}
	})
	t.Run("refresh rotates and replay fails", func(t *testing.T) {
		_, resp := request(t, server.URL, "POST", "/api/v1/auth/refresh", "", nil, refresh, 200)
		rotated := resp.Cookies()[0]
		if rotated.Value == refresh.Value {
			t.Fatal("refresh did not rotate")
		}
		request(t, server.URL, "POST", "/api/v1/auth/refresh", "", nil, refresh, 401)
		request(t, server.URL, "POST", "/api/v1/auth/refresh", "", nil, rotated, 200)
	})
	t.Run("handover atomic completion ownership and retry", func(t *testing.T) {
		id := primitive.NewObjectID()
		h := handovers.HandoverRecord{ID: id, HandoverReference: "TEST-HANDOVER", LotID: lotID, TransactionID: primitive.NewObjectID(), CollectorID: collectorID, RecyclerID: rid, Status: "collector_confirmed", CollectorConfirmation: true, StatusHistory: []handovers.StatusEvent{{Status: "collector_confirmed", Timestamp: now}}}
		if _, err := db.Collection("handoverRecords").InsertOne(ctx, h); err != nil {
			t.Fatal(err)
		}
		hp := "/api/v1/handovers/" + id.Hex() + "/confirm"
		request(t, server.URL, "POST", hp, otherToken, nil, nil, 409)
		for i := 0; i < 2; i++ {
			d, _ := request(t, server.URL, "POST", hp, token, nil, nil, 200)
			if d["status"] != "completed" || len(d["statusHistory"].([]any)) != 3 {
				t.Fatal("confirmation must append exactly once and complete atomically")
			}
		}
	})
	t.Run("transaction and payment ownership", func(t *testing.T) {
		id := primitive.NewObjectID()
		_, err := db.Collection("transactions").InsertOne(ctx, bson.M{"_id": id, "reference": "TEST-TX", "recyclerId": rid})
		if err != nil {
			t.Fatal(err)
		}
		request(t, server.URL, "GET", "/api/v1/transactions/"+id.Hex(), otherToken, nil, nil, 404)
		_, err = db.Collection("payments").InsertOne(ctx, bson.M{"paymentId": "TEST-PAY", "recyclerId": rid, "paymentMethod": "cash"})
		if err != nil {
			t.Fatal(err)
		}
		d, _ := request(t, server.URL, "GET", "/api/v1/payments", otherToken, nil, nil, 200)
		if len(d["items"].([]any)) != 0 {
			t.Fatal("payment leaked to another recycler")
		}
	})
	t.Run("wrong browser origin denied", func(t *testing.T) {
		r, _ := http.NewRequest("POST", server.URL+"/api/v1/auth/logout", nil)
		r.Header.Set("Origin", "https://untrusted.example")
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 403 {
			t.Fatal("untrusted origin permitted")
		}
	})
}

func TestDevelopmentSeedPreservesExistingRecords(t *testing.T) {
	db, cfg := testDatabase(t)
	ctx := context.Background()
	run := func() {
		t.Helper()
		cmd := exec.Command("go", "run", "../seed")
		cmd.Env = append(os.Environ(), "ENVIRONMENT=development", "MONGODB_URI="+cfg.MongoURI, "MONGODB_DATABASE="+db.Name(), "JWT_ACCESS_SECRET="+cfg.AccessSecret, "JWT_REFRESH_SECRET="+cfg.RefreshSecret)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("seed: %v %s", err, out)
		}
	}
	run()
	var before bson.M
	if err := db.Collection("scrapListings").FindOne(ctx, bson.M{"referenceId": "KC-EW-DEMO-001"}).Decode(&before); err != nil {
		t.Fatal(err)
	}
	_, err := db.Collection("scrapListings").UpdateByID(ctx, before["_id"], bson.M{"$set": bson.M{"status": "matched"}})
	if err != nil {
		t.Fatal(err)
	}
	run()
	var after bson.M
	if err := db.Collection("scrapListings").FindOne(ctx, bson.M{"_id": before["_id"]}).Decode(&after); err != nil {
		t.Fatal(err)
	}
	if after["status"] != "matched" {
		t.Fatal("reseed reset existing lot status")
	}
	for col, want := range map[string]int64{"users": 3, "scrapListings": 8, "priceRecords": 16} {
		n, err := db.Collection(col).CountDocuments(ctx, bson.M{})
		if err != nil || n != want {
			t.Fatal(fmt.Sprintf("%s count=%d want=%d err=%v", col, n, want, err))
		}
	}
}
