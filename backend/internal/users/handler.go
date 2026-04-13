package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	circlesauth "github.com/danielyakubu/circles/internal/auth"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes returns the router for the /users prefix.
// POST /users does not require a resolved user (user may not exist yet).
// GET routes do.
func Routes(svc *Service) http.Handler {
	r := chi.NewRouter()
	h := NewHandler(svc)

	r.Post("/", h.register)

	r.Group(func(r chi.Router) {
		r.Use(RequireUser(svc))
		r.Get("/me", h.getMe)
		r.Get("/{id}", h.getUser)
	})

	return r
}

// POST /users
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	firebaseUID, ok := circlesauth.FirebaseUIDFromCtx(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		DisplayName string `json:"display_name"`
		PhoneHash   string `json:"phone_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), firebaseUID, body.DisplayName, body.PhoneHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respond(w, http.StatusOK, user)
}

// GET /users/me
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())
	user, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, user)
}

// GET /users/:id
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	requesterID, _ := circlesauth.UserIDFromCtx(r.Context())
	targetID := chi.URLParam(r, "id")

	user, err := h.svc.GetFriend(r.Context(), requesterID, targetID)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, ErrForbidden) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respond(w, http.StatusOK, user)
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
