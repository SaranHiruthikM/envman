package secrets

import (
	"context"
)

func (h *Handler) pushSecrets(ctx context.Context, projectID, encryptedData, userId string) error {
	query1 := `INSERT INTO secrets VALUES (
		gen_random_uuid(), $1, $2, $3, NOW()
	) RETURNING id`

	query2 := `INSERT INTO audit_logs VALUES (
		gen_random_uuid(), $1, $2, $3, NOW()
	)`

	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}

	var secret string

	err = tx.QueryRow(ctx, query1, projectID, userId, encryptedData).Scan(&secret)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query2, projectID, userId, secret)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (h *Handler) getSecrets(ctx context.Context, projectID string) (string, error) {
	query := `SELECT encrypted_data FROM secrets WHERE project_id=$1 ORDER BY pushed_at DESC LIMIT 1`

	var secret string

	err := h.db.QueryRow(ctx, query, projectID).Scan(&secret)
	if err != nil {
		return "", err
	}

	return secret, nil
}
