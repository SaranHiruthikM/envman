package auth

import (
	"github.com/SaranHiruthikM/envman/internal/token"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db           *pgxpool.Pool
	tokenManager *token.Manager
}

func NewHandler(db *pgxpool.Pool, tokenManager *token.Manager) *Handler {
	return &Handler{
		db:           db,
		tokenManager: tokenManager,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
}
