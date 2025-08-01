package middleware

import (
	"context"
	"net/http"

	"github.com/andrasnagy-data/timelog/internal/shared/cookie"
	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userIDKey contextKey = "userID"

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) uuid.UUID {
	userID, _ := ctx.Value(userIDKey).(uuid.UUID)
	return userID
}

// NewAuthMiddleware creates authentication middleware that validates session cookies
// and protects routes from unauthorized access. It extracts the user ID from encrypted
// cookies and adds it to the request context for downstream handlers.
func NewAuthMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := cookie.GetCookie(r, secretKey)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, *userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
