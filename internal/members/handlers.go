package members

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Post("/", h.InviteMembers)
	r.Patch("/{memberUserID}", h.UpdateRole)
	r.Delete("/{memberUserID}", h.RemoveMember)
	return r
}
