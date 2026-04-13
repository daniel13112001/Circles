package friends

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	circlesauth "github.com/danielyakubu/circles/internal/auth"
)

func Routes(svc *Service) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		userID, _ := circlesauth.UserIDFromCtx(r.Context())

		friends, err := svc.List(r.Context(), userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if friends == nil {
			friends = []Friend{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(friends)
	})
	return r
}
