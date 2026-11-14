package projects

import (
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/audit"
	"github.com/SaranHiruthikM/envman/internal/members"
	"github.com/SaranHiruthikM/envman/internal/secrets"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db            *pgxpool.Pool
	memberHandler *members.Handler
	secretHandler *secrets.Handler
	logsHandler   *audit.Handler
}

func NewHandler(db *pgxpool.Pool, memberHandler *members.Handler, secretHandler *secrets.Handler, logsHandler *audit.Handler) *Handler {
	return &Handler{
		db:            db,
		memberHandler: memberHandler,
		secretHandler: secretHandler,
		logsHandler:   logsHandler,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/{projectID}", h.GetProjectByID)
	r.Post("/", h.CreateProject)
	r.Post("/{projectID}/keys", h.memberHandler.SetSelfKey)
	r.Get("/{projectID}/keys/me", h.memberHandler.GetSelfKey)
	r.Mount("/{projectID}/members", h.memberHandler.Routes())
	r.Mount("/{projectID}/secrets", h.secretHandler.Routes())
	r.Mount("/{projectID}/audit-logs", h.logsHandler.Routes())
	return r
}
