package groups

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	circlesauth "github.com/danielyakubu/circles/internal/auth"
)

func Routes(svc *Service) http.Handler {
	r := chi.NewRouter()
	h := &handler{svc: svc}

	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}/members", h.listMembers)
	r.Post("/{id}/members", h.addMember)
	r.Delete("/{id}/members/me", h.leave)

	return r
}

type handler struct{ svc *Service }

// POST /groups
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	g, err := h.svc.Create(r.Context(), body.Name, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respond(w, http.StatusCreated, g)
}

// GET /groups
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())

	groups, err := h.svc.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = []Group{}
	}

	respond(w, http.StatusOK, groups)
}

// GET /groups/:id/members
func (h *handler) listMembers(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())
	groupID := chi.URLParam(r, "id")

	members, err := h.svc.ListMembers(r.Context(), groupID, userID)
	if errors.Is(err, ErrForbidden) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if members == nil {
		members = []Member{}
	}

	respond(w, http.StatusOK, members)
}

// POST /groups/:id/members
func (h *handler) addMember(w http.ResponseWriter, r *http.Request) {
	requesterID, _ := circlesauth.UserIDFromCtx(r.Context())
	groupID := chi.URLParam(r, "id")

	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	err := h.svc.AddMember(r.Context(), groupID, requesterID, body.UserID)
	if errors.Is(err, ErrForbidden) {
		http.Error(w, "you are not a member of this group", http.StatusForbidden)
		return
	}
	if errors.Is(err, ErrNotFriend) {
		http.Error(w, "that user is not your friend", http.StatusForbidden)
		return
	}
	if errors.Is(err, ErrCircleCheck) {
		http.Error(w, "user is not friends with all current members", http.StatusForbidden)
		return
	}
	if errors.Is(err, ErrAlreadyMember) {
		http.Error(w, "user is already a member", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /groups/:id/members/me
func (h *handler) leave(w http.ResponseWriter, r *http.Request) {
	userID, _ := circlesauth.UserIDFromCtx(r.Context())
	groupID := chi.URLParam(r, "id")

	if err := h.svc.Leave(r.Context(), groupID, userID); errors.Is(err, ErrNotFound) {
		http.Error(w, "not a member of this group", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
