package auth

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type contextKey string

const userIDKey contextKey = "userID"
const firebaseUIDKey contextKey = "firebaseUID"

// Middleware verifies the Firebase JWT on every request.
// On success it injects the internal userID and firebaseUID into the context.
// The userID is looked up by the handler layer after registration; for the
// registration endpoint itself only firebaseUID is available.
func Middleware(client *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
				return
			}

			idToken := strings.TrimPrefix(header, "Bearer ")
			token, err := client.VerifyIDToken(r.Context(), idToken)
			if err != nil {
				http.Error(w, "invalid Firebase token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), firebaseUIDKey, token.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FirebaseUIDFromCtx extracts the Firebase UID injected by Middleware.
func FirebaseUIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(firebaseUIDKey).(string)
	return v, ok
}

// WithUserID injects the resolved internal user UUID into the context.
// Called by handlers after they resolve Firebase UID → internal user.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromCtx extracts the internal user UUID.
func UserIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}
