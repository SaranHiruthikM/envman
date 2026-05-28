package projects

import (
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/members"
	"github.com/SaranHiruthikM/envman/internal/secrets"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db            *pgxpool.Pool
	memberHandler *members.Handler
	secretHandler *secrets.Handler
}

func NewHandler(db *pgxpool.Pool, memberHandler *members.Handler, secretHandler *secrets.Handler) *Handler {
	return &Handler{
		db:            db,
		memberHandler: memberHandler,
		secretHandler: secretHandler,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetProjectByID)
	r.Post("/", h.CreateProject)
	r.Mount("/{id}/members", h.memberHandler.Routes())
	r.Mount("/{id}/secrets", h.secretHandler.Routes())
	return r
}
