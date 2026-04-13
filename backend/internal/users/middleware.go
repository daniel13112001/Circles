package users

import (
	"errors"
	"net/http"

	circlesauth "github.com/danielyakubu/circles/internal/auth"
)

// RequireUser resolves the Firebase UID from the JWT (already verified by
// auth.Middleware) into an internal user record, then injects the user ID into
// the request context. Returns 401 if the user has not registered yet.
//
// Apply this middleware to every route that needs an authenticated, registered user.
func RequireUser(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			firebaseUID, ok := circlesauth.FirebaseUIDFromCtx(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := svc.GetByFirebaseUID(r.Context(), firebaseUID)
			if errors.Is(err, ErrNotFound) {
				http.Error(w, "user not registered", http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			ctx := circlesauth.WithUserID(r.Context(), user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
