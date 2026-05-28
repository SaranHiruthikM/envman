package auth

import (
	"context"
)

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
