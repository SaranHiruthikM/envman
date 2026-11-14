package secrets

import (
	"context"
	"fmt"
)

func (h *Handler) pushSecrets(ctx context.Context, projectID, ciphertext, userID string) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var secretID string
	err = tx.QueryRow(ctx, `
		INSERT INTO secrets (id, project_id, pushed_by, encrypted_data, pushed_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW()) RETURNING id`,
		projectID, userID, ciphertext,
	).Scan(&secretID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_logs (id, project_id, pushed_by, secrets_id, pushed_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		projectID, userID, secretID,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

func (h *Handler) getSecrets(ctx context.Context, projectID string) (string, error) {
	var ciphertext string
	err := h.db.QueryRow(ctx,
		`SELECT encrypted_data FROM secrets WHERE project_id=$1 ORDER BY pushed_at DESC LIMIT 1`,
		projectID).Scan(&ciphertext)
	return ciphertext, err
}
