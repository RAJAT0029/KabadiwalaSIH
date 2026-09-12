package users

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	Street  string `bson:"street" json:"street"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Pincode string `bson:"pincode" json:"pincode"`
}

type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type Authorization struct {
	RegistrationNumber string     `bson:"registrationNumber,omitempty" json:"registrationNumber,omitempty"`
	Authority          string     `bson:"authority,omitempty" json:"authority,omitempty"`
	ValidFrom          *time.Time `bson:"validFrom,omitempty" json:"validFrom,omitempty"`
	ValidUntil         *time.Time `bson:"validUntil,omitempty" json:"validUntil,omitempty"`
	Status             string     `bson:"status" json:"status"`
	DocumentReference  string     `bson:"documentReference,omitempty" json:"documentReference,omitempty"`
	IsDemo             bool       `bson:"isDemo,omitempty" json:"isDemo,omitempty"`
}

type PickupProfile struct {
	Available   bool     `bson:"available" json:"available"`
	ServiceArea []string `bson:"serviceArea,omitempty" json:"serviceArea,omitempty"`
}

type OfferedRate struct {
	MaterialCategory    string  `bson:"materialCategory" json:"materialCategory"`
	MaterialSubCategory string  `bson:"materialSubCategory,omitempty" json:"materialSubCategory,omitempty"`
	PricePerUnit        float64 `bson:"pricePerUnit" json:"pricePerUnit"`
	Unit                string  `bson:"unit" json:"unit"`
}

type User struct {
	ID                           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Role                         string             `bson:"role" json:"role"`
	Name                         string             `bson:"name" json:"name"`
	Phone                        string             `bson:"phone" json:"phone"`
	Email                        string             `bson:"email" json:"email"`
	CompanyName                  string             `bson:"companyName,omitempty" json:"companyName,omitempty"`
	FacilityName                 string             `bson:"facilityName,omitempty" json:"facilityName,omitempty"`
	BusinessName                 string             `bson:"businessName,omitempty" json:"businessName,omitempty"`
	BusinessType                 string             `bson:"businessType,omitempty" json:"businessType,omitempty"`
	GSTNumber                    string             `bson:"gstNumber,omitempty" json:"gstNumber,omitempty"`
	RegistrationNumber           string             `bson:"registrationNumber,omitempty" json:"registrationNumber,omitempty"` // legacy compatibility
	Address                      Address            `bson:"address" json:"address"`
	Location                     *GeoPoint          `bson:"location,omitempty" json:"location,omitempty"`
	AcceptedMaterials            []string           `bson:"acceptedMaterials,omitempty" json:"acceptedMaterials,omitempty"`
	ProcessingCapacityKg         *float64           `bson:"processingCapacityKg,omitempty" json:"processingCapacityKg,omitempty"`
	Authorization                Authorization      `bson:"authorization" json:"authorization"`
	Pickup                       PickupProfile      `bson:"pickup" json:"pickup"`
	OfferedRates                 []OfferedRate      `bson:"offeredRates,omitempty" json:"offeredRates,omitempty"`
	PreferredLanguage            string             `bson:"preferredLanguage,omitempty" json:"preferredLanguage,omitempty"`
	VerificationStatus           string             `bson:"verificationStatus" json:"verificationStatus"` // legacy compatibility
	ProfileImage                 string             `bson:"profileImage,omitempty" json:"profileImage,omitempty"`
	PasswordHash                 string             `bson:"passwordHash" json:"-"`
	IsActive                     bool               `bson:"isActive" json:"isActive"`
	ParticipationAllowed         bool               `bson:"-" json:"participationAllowed"`
	EffectiveAuthorizationStatus string             `bson:"-" json:"effectiveAuthorizationStatus"`
	CreatedAt                    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt                    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func (u *User) AuthorizationStatus(now time.Time) string {
	status := strings.TrimSpace(u.Authorization.Status)
	if status == "" {
		status = "pending_verification"
	}
	if status == "authorized" || status == "demo_verified" {
		if u.Authorization.ValidUntil != nil && !u.Authorization.ValidUntil.After(now) {
			return "expired"
		}
		if u.Authorization.ValidFrom != nil && u.Authorization.ValidFrom.After(now) {
			return "pending_verification"
		}
	}
	return status
}

func (u *User) CanParticipate(now time.Time, allowDemo ...bool) bool {
	status := u.AuthorizationStatus(now)
	demo := len(allowDemo) > 0 && allowDemo[0] && u.Authorization.IsDemo && status == "demo_verified"
	return u.IsActive && u.Role == "recycler" && ((status == "authorized" && !u.Authorization.IsDemo) || demo)
}

func (u *User) SetParticipation(allowDemo bool) {
	u.EffectiveAuthorizationStatus = u.AuthorizationStatus(time.Now().UTC())
	u.ParticipationAllowed = u.CanParticipate(time.Now().UTC(), allowDemo)
}

func (u *User) AcceptsMaterial(category, subCategory string) bool {
	for _, item := range u.AcceptedMaterials {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(category)) || (subCategory != "" && strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(subCategory))) {
			return true
		}
	}
	return false
}
