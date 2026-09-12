package requestctx

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type key string

const userIDKey key = "userID"
const roleKey key = "role"

func WithIdentity(ctx context.Context, userID primitive.ObjectID, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, roleKey, role)
}

func UserID(ctx context.Context) primitive.ObjectID {
	id, _ := ctx.Value(userIDKey).(primitive.ObjectID)
	return id
}

func Role(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}
