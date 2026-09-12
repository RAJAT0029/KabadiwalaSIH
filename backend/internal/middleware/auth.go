package middleware

import (
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"kabadiconnect/backend/internal/auth"
	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
)

func Auth(service *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
			return
		}

		claims, err := service.ParseAccess(parts[1])
		if err != nil || !service.ActiveIdentity(r.Context(), claims) {
			httpx.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Access token is invalid or expired")
			return
		}

		id, err := primitive.ObjectIDFromHex(claims.Subject)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Access token is invalid")
			return
		}

		ctx := requestctx.WithIdentity(r.Context(), id, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RecyclerOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestctx.Role(r.Context()) != "recycler" {
			httpx.Error(w, http.StatusForbidden, "FORBIDDEN", "Recycler account required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
