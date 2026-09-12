package auth

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RefreshSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId"`
	TokenHash string             `bson:"tokenHash"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	RevokedAt *time.Time         `bson:"revokedAt,omitempty"`
	CreatedAt time.Time          `bson:"createdAt"`
}
