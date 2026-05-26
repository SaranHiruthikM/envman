package auth

import (
	"context"

	"github.com/SaranHiruthikM/envman/internal/models"
)

func (h *Handler) getUserByEmail(ctx context.Context, emailID string) (*models.User, error) {
	query := `SELECT id, email, name, password_hash, created_at from users where email = $1`

	var user models.User

	err := h.db.QueryRow(ctx, query, emailID).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *Handler) createUser(ctx context.Context, name, email, passwordHash string) (string, error) {
	query := `INSERT INTO users VALUES (
		gen_random_uuid(), $1, $2, $3, NOW()
	) RETURNING id`

	var userId string
	row := h.db.QueryRow(ctx, query, email, name, passwordHash)
	err := row.Scan(&userId)
	if err != nil {
		return "", err
	}

	return userId, nil
}
