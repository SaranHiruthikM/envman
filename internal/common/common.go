package common

import (
	"context"
	"errors"

	"github.com/SaranHiruthikM/envman/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetUserByEmail returns the full user row, including the bcrypt hash.
// Callers that only need the user ID should use GetUserIDByEmail instead.
func GetUserByEmail(ctx context.Context, db *pgxpool.Pool, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at,
		       public_key, encrypted_private_key, encrypted_private_key_nonce,
		       kdf_salt, kdf_time, kdf_memory, kdf_parallelism
		FROM users WHERE email = $1`

	var u models.User
	err := db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt,
		&u.PublicKey, &u.EncryptedPrivateKey, &u.EncryptedPrivateKeyNonce,
		&u.KDFSalt, &u.KDFTime, &u.KDFMemory, &u.KDFParallelism,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserIDByEmail returns only the user ID — safe to use in handlers that
// don't need the password hash or key material.
func GetUserIDByEmail(ctx context.Context, db *pgxpool.Pool, email string) (string, error) {
	var id string
	err := db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	return id, err
}

// GetUserPublicKey returns {id, public_key} for an email. Used by the invite flow.
func GetUserPublicKey(ctx context.Context, db *pgxpool.Pool, email string) (string, []byte, error) {
	var id string
	var pub []byte
	err := db.QueryRow(ctx,
		`SELECT id, public_key FROM users WHERE email = $1`,
		email).Scan(&id, &pub)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, err
	}
	return id, pub, err
}
