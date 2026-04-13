package contacts

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

	r.Post("/", h.add)
	r.Get("/", h.list)
	r.Delete("/{id}", h.delete)

	return r
}

type handler struct{ svc *Service }

// POST /contacts
func (h *handler) add(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := circlesauth.UserIDFromCtx(r.Context())

	var body struct {
		PhoneHash string `json:"phone_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PhoneHash == "" {
		http.Error(w, "phone_hash is required", http.StatusBadRequest)
		return
	}

	contact, err := h.svc.Add(r.Context(), ownerID, body.PhoneHash)
	if errors.Is(err, ErrSelfAdd) {
		http.Error(w, "cannot add your own phone hash", http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrDuplicate) {
		http.Error(w, "contact already exists", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respond(w, http.StatusCreated, contact)
}

// GET /contacts
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := circlesauth.UserIDFromCtx(r.Context())

	contacts, err := h.svc.List(r.Context(), ownerID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if contacts == nil {
		contacts = []Contact{}
	}

	respond(w, http.StatusOK, contacts)
}

// DELETE /contacts/:id
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := circlesauth.UserIDFromCtx(r.Context())
	contactID := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), ownerID, contactID); errors.Is(err, ErrNotFound) {
		http.Error(w, "contact not found", http.StatusNotFound)
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
