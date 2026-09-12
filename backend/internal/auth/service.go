package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/users"
)

type Claims struct {
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}
type Service struct {
	cfg      config.Config
	users    *users.Repository
	sessions *SessionRepository
}

func NewService(cfg config.Config, u *users.Repository, s *SessionRepository) *Service {
	return &Service{cfg: cfg, users: u, sessions: s}
}

type RegisterInput struct {
	Name                   string        `json:"name"`
	Phone                  string        `json:"phone"`
	Email                  string        `json:"email"`
	Password               string        `json:"password"`
	CompanyName            string        `json:"companyName"`
	FacilityName           string        `json:"facilityName"`
	BusinessType           string        `json:"businessType"`
	GSTNumber              string        `json:"gstNumber"`
	RegistrationNumber     string        `json:"registrationNumber"`
	AuthorizationAuthority string        `json:"authorizationAuthority"`
	Address                users.Address `json:"address"`
	AcceptedMaterials      []string      `json:"acceptedMaterials"`
	ProcessingCapacityKg   *float64      `json:"processingCapacityKg"`
	PickupAvailable        bool          `json:"pickupAvailable"`
	ServiceArea            []string      `json:"serviceArea"`
	TermsAccepted          bool          `json:"termsAccepted"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*users.User, string, string, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.CompanyName) == "" || in.Phone == "" {
		return nil, "", "", fmt.Errorf("name, company and phone are required")
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return nil, "", "", fmt.Errorf("valid email is required")
	}
	if len(in.Password) < 10 || len(in.Password) > 72 {
		return nil, "", "", fmt.Errorf("password must be between 10 and 72 bytes")
	}
	if !in.TermsAccepted {
		return nil, "", "", fmt.Errorf("terms must be accepted")
	}
	if err := users.ValidateProfilePatch(users.ProfilePatch{AcceptedMaterials: &in.AcceptedMaterials, ProcessingCapacityKg: in.ProcessingCapacityKg}); err != nil {
		return nil, "", "", err
	}
	exists, err := s.users.ExistsEmailOrPhone(ctx, in.Email, in.Phone)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return nil, "", "", err
	}
	now := time.Now().UTC()
	authorizationStatus := "pending_verification"
	verificationStatus := "unverified"
	isDemoAuthorization := false
	if s.cfg.Environment == "development" {
		authorizationStatus = "demo_verified"
		verificationStatus = "demo_verified"
		isDemoAuthorization = true
	}
	facilityName := strings.TrimSpace(in.FacilityName)
	if facilityName == "" {
		facilityName = strings.TrimSpace(in.CompanyName)
	}
	u := &users.User{
		Role: "recycler", Name: strings.TrimSpace(in.Name), Phone: in.Phone, Email: in.Email,
		CompanyName: strings.TrimSpace(in.CompanyName), FacilityName: facilityName,
		BusinessType: strings.TrimSpace(in.BusinessType), GSTNumber: strings.TrimSpace(in.GSTNumber),
		RegistrationNumber: strings.TrimSpace(in.RegistrationNumber), Address: in.Address,
		AcceptedMaterials: in.AcceptedMaterials, ProcessingCapacityKg: in.ProcessingCapacityKg,
		Authorization:      users.Authorization{RegistrationNumber: strings.TrimSpace(in.RegistrationNumber), Authority: strings.TrimSpace(in.AuthorizationAuthority), Status: authorizationStatus, IsDemo: isDemoAuthorization},
		Pickup:             users.PickupProfile{Available: in.PickupAvailable, ServiceArea: in.ServiceArea},
		VerificationStatus: verificationStatus, PasswordHash: string(hash), IsActive: true,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, "", "", ErrConflict
		}
		return nil, "", "", err
	}
	u.SetParticipation(s.cfg.Environment == "development")
	access, refresh, err := s.issue(ctx, u)
	return u, access, refresh, err
}
func (s *Service) Login(ctx context.Context, identifier, password string, remember ...bool) (*users.User, string, string, error) {
	u, err := s.users.ByEmailOrPhone(ctx, identifier)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, "", "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", "", err
	}
	if !u.IsActive || u.Role != "recycler" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", "", ErrInvalidCredentials
	}
	if s.cfg.Environment == "development" && u.Role == "recycler" && u.Authorization.Status == "" {
		if err := s.users.EnsureDevelopmentAuthorization(ctx, u.ID); err != nil {
			return nil, "", "", err
		}
		u.Authorization.Status = "demo_verified"
		u.Authorization.IsDemo = true
		u.VerificationStatus = "demo_verified"
	}
	ttl := s.cfg.RefreshTTL
	if len(remember) > 0 && !remember[0] {
		ttl = 24 * time.Hour
	}
	access, refresh, err := s.issue(ctx, u, ttl)
	u.SetParticipation(s.cfg.Environment == "development")
	return u, access, refresh, err
}
func (s *Service) Refresh(ctx context.Context, raw string) (string, string, error) {
	claims, err := s.parse(raw, s.cfg.RefreshSecret, "refresh")
	if err != nil {
		return "", "", ErrInvalidToken
	}
	h := hashToken(raw)
	sess, err := s.sessions.ActiveByHash(ctx, h)
	if err != nil {
		return "", "", ErrInvalidToken
	}
	uid, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil || uid != sess.UserID {
		return "", "", ErrInvalidToken
	}
	u, err := s.users.ByID(ctx, uid)
	if err != nil || !u.IsActive {
		return "", "", ErrInvalidToken
	}
	if s.cfg.Environment == "development" && u.Role == "recycler" && u.Authorization.Status == "" {
		if err := s.users.EnsureDevelopmentAuthorization(ctx, u.ID); err != nil {
			return "", "", err
		}
		u.Authorization.Status = "demo_verified"
		u.Authorization.IsDemo = true
		u.VerificationStatus = "demo_verified"
	}
	if err := s.sessions.Consume(ctx, h); err != nil {
		return "", "", ErrInvalidToken
	}
	return s.issue(ctx, u, time.Until(sess.ExpiresAt))
}
func (s *Service) Logout(ctx context.Context, raw string) {
	if raw != "" {
		_ = s.sessions.Revoke(ctx, hashToken(raw))
	}
}
func (s *Service) ParseAccess(raw string) (*Claims, error) {
	return s.parse(raw, s.cfg.AccessSecret, "access")
}
func (s *Service) issue(ctx context.Context, u *users.User, lifetime ...time.Duration) (string, string, error) {
	now := time.Now().UTC()
	ttl := s.cfg.RefreshTTL
	if len(lifetime) > 0 {
		ttl = lifetime[0]
	}
	access, err := s.sign(u, "access", s.cfg.AccessSecret, s.cfg.AccessTTL, now)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.sign(u, "refresh", s.cfg.RefreshSecret, ttl, now)
	if err != nil {
		return "", "", err
	}
	if err := s.sessions.Create(ctx, RefreshSession{UserID: u.ID, TokenHash: hashToken(refresh), ExpiresAt: now.Add(ttl), CreatedAt: now}); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *Service) ActiveIdentity(ctx context.Context, claims *Claims) bool {
	id, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return false
	}
	u, err := s.users.ByID(ctx, id)
	return err == nil && u.IsActive && u.Role == claims.Role
}
func (s *Service) sign(u *users.User, typ, secret string, ttl time.Duration, now time.Time) (string, error) {
	claims := Claims{Role: u.Role, TokenType: typ, RegisteredClaims: jwt.RegisteredClaims{ID: uuid.NewString(), Subject: u.ID.Hex(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
func (s *Service) parse(raw, secret, typ string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid || claims.TokenType != typ {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")
var ErrConflict = errors.New("email or phone already exists")
