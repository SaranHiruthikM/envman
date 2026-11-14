package auth

import (
	"github.com/SaranHiruthikM/envman/internal/token"
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
