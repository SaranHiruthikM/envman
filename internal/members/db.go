package members

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrAlreadyMember = errors.New("user is already a member")

type wrappedKey struct {
	WrappedKey      []byte
	WrappedKeyNonce []byte
	EphemeralPubkey []byte
}

// inviteMember inserts the user_projects row AND the per-member wrapped
// project key in one transaction. Partial state is unacceptable: a membership
// row without a key would leave the invitee unable to decrypt.
func (h *Handler) inviteMember(ctx context.Context, userID, projectID, role string, k wrappedKey) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO user_projects (id, user_id, project_id, role)
		 VALUES (gen_random_uuid(), $1, $2, $3)`,
		userID, projectID, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrAlreadyMember
		}
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO project_keys (project_id, user_id, wrapped_key, wrapped_key_nonce, ephemeral_pubkey)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, user_id) DO UPDATE
			SET wrapped_key = EXCLUDED.wrapped_key,
			    wrapped_key_nonce = EXCLUDED.wrapped_key_nonce,
			    ephemeral_pubkey = EXCLUDED.ephemeral_pubkey`,
		projectID, userID, k.WrappedKey, k.WrappedKeyNonce, k.EphemeralPubkey)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

func (h *Handler) updateRole(ctx context.Context, userID, projectID, role string) error {
	_, err := h.db.Exec(ctx,
		`UPDATE user_projects SET role=$1 WHERE user_id=$2 AND project_id=$3`,
		role, userID, projectID)
	return err
}

func (h *Handler) removeMember(ctx context.Context, userID, projectID string) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`DELETE FROM user_projects WHERE user_id=$1 AND project_id=$2`,
		userID, projectID); err != nil {
		return err
	}
	// project_keys rows cascade on ON DELETE CASCADE for the project, but a
	// member being removed from a project doesn't trigger that — delete explicitly.
	if _, err := tx.Exec(ctx,
		`DELETE FROM project_keys WHERE user_id=$1 AND project_id=$2`,
		userID, projectID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (h *Handler) countOwners(ctx context.Context, projectID string) (int, error) {
	var count int
	err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_projects WHERE project_id=$1 AND role='owner'`,
		projectID).Scan(&count)
	return count, err
}

func (h *Handler) setSelfKey(ctx context.Context, userID, projectID string, k wrappedKey) error {
	_, err := h.db.Exec(ctx, `
		INSERT INTO project_keys (project_id, user_id, wrapped_key, wrapped_key_nonce, ephemeral_pubkey)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, user_id) DO UPDATE
			SET wrapped_key = EXCLUDED.wrapped_key,
			    wrapped_key_nonce = EXCLUDED.wrapped_key_nonce,
			    ephemeral_pubkey = EXCLUDED.ephemeral_pubkey`,
		projectID, userID, k.WrappedKey, k.WrappedKeyNonce, k.EphemeralPubkey)
	return err
}

func (h *Handler) getSelfKey(ctx context.Context, userID, projectID string) (*wrappedKey, error) {
	var k wrappedKey
	err := h.db.QueryRow(ctx,
		`SELECT wrapped_key, wrapped_key_nonce, ephemeral_pubkey
		 FROM project_keys WHERE user_id=$1 AND project_id=$2`,
		userID, projectID).Scan(&k.WrappedKey, &k.WrappedKeyNonce, &k.EphemeralPubkey)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}
