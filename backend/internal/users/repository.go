package users

import (
	"context"
	"errors"
	"kabadiconnect/backend/internal/database"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	c         *mongo.Collection
	allowDemo bool
}

func NewRepository(db *mongo.Database, allowDemo ...bool) *Repository {
	return &Repository{c: db.Collection("users"), allowDemo: len(allowDemo) > 0 && allowDemo[0]}
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := database.EnsureIndexes(ctx, r.c, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "phone", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "location", Value: "2dsphere"}}},
		{Keys: bson.D{{Key: "role", Value: 1}, {Key: "authorization.status", Value: 1}}},
		{Keys: bson.D{{Key: "role", Value: 1}, {Key: "acceptedMaterials", Value: 1}}},
	})
	return err
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	res, err := r.c.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *Repository) ByEmailOrPhone(ctx context.Context, value string) (*User, error) {
	var u User
	v := strings.ToLower(strings.TrimSpace(value))
	err := r.c.FindOne(ctx, bson.M{"$or": []bson.M{{"email": v}, {"phone": strings.TrimSpace(value)}}}).Decode(&u)
	u.SetParticipation(r.allowDemo)
	return &u, err
}

func (r *Repository) ByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var u User
	err := r.c.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	u.SetParticipation(r.allowDemo)
	return &u, err
}

func (r *Repository) ExistsEmailOrPhone(ctx context.Context, email, phone string) (bool, error) {
	n, err := r.c.CountDocuments(ctx, bson.M{"$or": []bson.M{{"email": strings.ToLower(email)}, {"phone": phone}}})
	return n > 0, err
}

func (r *Repository) PublicByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	u, err := r.ByID(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return u, err
}

type ProfilePatch struct {
	Name                 *string        `json:"name"`
	CompanyName          *string        `json:"companyName"`
	FacilityName         *string        `json:"facilityName"`
	BusinessType         *string        `json:"businessType"`
	GSTNumber            *string        `json:"gstNumber"`
	Address              *Address       `json:"address"`
	AcceptedMaterials    *[]string      `json:"acceptedMaterials"`
	ProcessingCapacityKg *float64       `json:"processingCapacityKg"`
	Authorization        *Authorization `json:"authorization"`
	Pickup               *PickupProfile `json:"pickup"`
	OfferedRates         *[]OfferedRate `json:"offeredRates"`
}

func (r *Repository) UpdateRecyclerProfile(ctx context.Context, id primitive.ObjectID, in ProfilePatch) (*User, error) {
	if err := ValidateProfilePatch(in); err != nil {
		return nil, err
	}
	set := bson.M{"updatedAt": time.Now().UTC()}
	if in.Name != nil {
		set["name"] = strings.TrimSpace(*in.Name)
	}
	if in.CompanyName != nil {
		set["companyName"] = strings.TrimSpace(*in.CompanyName)
	}
	if in.FacilityName != nil {
		set["facilityName"] = strings.TrimSpace(*in.FacilityName)
	}
	if in.BusinessType != nil {
		set["businessType"] = strings.TrimSpace(*in.BusinessType)
	}
	if in.GSTNumber != nil {
		set["gstNumber"] = strings.TrimSpace(*in.GSTNumber)
	}
	if in.Address != nil {
		set["address"] = *in.Address
	}
	if in.AcceptedMaterials != nil {
		set["acceptedMaterials"] = *in.AcceptedMaterials
	}
	if in.ProcessingCapacityKg != nil {
		set["processingCapacityKg"] = *in.ProcessingCapacityKg
	}
	if in.Pickup != nil {
		set["pickup"] = *in.Pickup
	}
	if in.OfferedRates != nil {
		set["offeredRates"] = *in.OfferedRates
	}
	if in.Authorization != nil {
		// Recycler may maintain reference details, but cannot self-promote authorization status.
		set["authorization.registrationNumber"] = strings.TrimSpace(in.Authorization.RegistrationNumber)
		set["authorization.authority"] = strings.TrimSpace(in.Authorization.Authority)
		set["authorization.documentReference"] = strings.TrimSpace(in.Authorization.DocumentReference)
		current, err := r.ByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if !current.Authorization.IsDemo && (current.Authorization.RegistrationNumber != strings.TrimSpace(in.Authorization.RegistrationNumber) || current.Authorization.Authority != strings.TrimSpace(in.Authorization.Authority) || current.Authorization.DocumentReference != strings.TrimSpace(in.Authorization.DocumentReference)) {
			set["authorization.status"] = "pending_verification"
			set["verificationStatus"] = "unverified"
		}
	}
	var u User
	err := r.c.FindOneAndUpdate(ctx, bson.M{"_id": id, "role": "recycler"}, bson.M{"$set": set}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&u)
	u.SetParticipation(r.allowDemo)
	return &u, err
}

func (r *Repository) EnsureDevelopmentAuthorization(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.c.UpdateOne(ctx, bson.M{"_id": id, "role": "recycler", "$or": []bson.M{{"authorization.status": bson.M{"$exists": false}}, {"authorization.status": ""}}}, bson.M{"$set": bson.M{"authorization.status": "demo_verified", "authorization.isDemo": true, "verificationStatus": "demo_verified", "updatedAt": time.Now().UTC()}})
	return err
}

func NowUTC() time.Time { return time.Now().UTC() }
