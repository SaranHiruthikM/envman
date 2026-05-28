package common

import (
	"context"

	"github.com/SaranHiruthikM/envman/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserByEmail(ctx context.Context, db *pgxpool.Pool, emailID string) (*models.User, error) {
	query := `SELECT id, email, name, password_hash, created_at from users where email = $1`

	var user models.User

	err := db.QueryRow(ctx, query, emailID).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
