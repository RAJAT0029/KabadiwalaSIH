package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/users"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	cfg := config.Config{AccessSecret: "0123456789abcdef0123456789abcdef", AccessTTL: 15 * time.Minute}
	svc := &Service{cfg: cfg}
	u := &users.User{ID: primitive.NewObjectID(), Role: "recycler"}
	raw, err := svc.sign(u, "access", cfg.AccessSecret, cfg.AccessTTL, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.ParseAccess(raw)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if claims.Subject != u.ID.Hex() || claims.Role != "recycler" || claims.TokenType != "access" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestSameSecondRefreshTokensAreUnique(t *testing.T) {
	s := &Service{}
	u := &users.User{ID: primitive.NewObjectID(), Role: "recycler"}
	now := time.Now().UTC()
	a, err := s.sign(u, "refresh", "test-secret", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.sign(u, "refresh", "test-secret", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("session tokens collide within same second")
	}
}

func TestAccessTokenRequiresExpiry(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	s := &Service{cfg: config.Config{AccessSecret: secret}}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{Role: "recycler", TokenType: "access"}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ParseAccess(raw); err == nil {
		t.Fatal("token without expiry accepted")
	}
}

func TestRefreshTokenCannotBeUsedAsAccessToken(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	cfg := config.Config{AccessSecret: secret, AccessTTL: 15 * time.Minute}
	svc := &Service{cfg: cfg}
	u := &users.User{ID: primitive.NewObjectID(), Role: "recycler"}
	raw, err := svc.sign(u, "refresh", secret, time.Hour, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ParseAccess(raw); err == nil {
		t.Fatal("refresh token must not be accepted as an access token")
	}
}
