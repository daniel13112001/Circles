package posts

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	circlesauth "github.com/danielyakubu/circles/internal/auth"
)

func GroupRoutes(svc *Service) http.Handler {
	r := chi.NewRouter()
	h := &handler{svc: svc}
	r.Post("/", h.create)
	r.Get("/", h.groupFeed)
	return r
}

func GlobalFeedHandler(svc *Service) http.HandlerFunc {
	h := &handler{svc: svc}
	return h.globalFeed
}

type handler struct{ svc *Service }

// POST /groups/:id/posts
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())
	groupID := chi.URLParam(r, "id")

	var body struct {
		ImageURL string  `json:"image_url"`
		Caption  *string `json:"caption"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	post, err := h.svc.Create(r.Context(), groupID, userID, body.ImageURL, body.Caption)
	if errors.Is(err, ErrForbidden) {
		http.Error(w, "not a member of this group", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respond(w, http.StatusCreated, post)
}

// GET /groups/:id/posts
func (h *handler) groupFeed(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())
	groupID := chi.URLParam(r, "id")

	posts, err := h.svc.GroupFeed(r.Context(), groupID, userID)
	if errors.Is(err, ErrForbidden) {
		http.Error(w, "not a member of this group", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []Post{}
	}

	respond(w, http.StatusOK, posts)
}

// GET /feed
func (h *handler) globalFeed(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())

	posts, err := h.svc.GlobalFeed(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []Post{}
	}

	respond(w, http.StatusOK, posts)
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
