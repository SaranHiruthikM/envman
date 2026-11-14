package users

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/common"
	"github.com/SaranHiruthikM/envman/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/by-email/{email}/public-key", h.GetPublicKey)
	return r
}

// GetPublicKey returns {user_id, public_key} for a given email. Required by
// the invite flow so the inviter can wrap the project key for the invitee.
func (h *Handler) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		response.Error(w, http.StatusBadRequest, "Email is required")
		return
	}

	id, pub, err := common.GetUserPublicKey(r.Context(), h.db, email)
	if errors.Is(err, pgx.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response.Json(w, http.StatusOK, map[string]string{
		"user_id":    id,
		"public_key": base64.StdEncoding.EncodeToString(pub),
	})
}
