package audit

import (
	"context"

	"github.com/SaranHiruthikM/envman/internal/models"
)

func (h *Handler) getLogs(ctx context.Context, projectID string) ([]models.Logs, error) {
	query := `SELECT id, project_id, pushed_by, secrets_id, pushed_at FROM audit_logs WHERE project_id=$1 ORDER BY pushed_at DESC`

	rows, err := h.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}

	var logs []models.Logs

	for rows.Next() {
		var log models.Logs
		err := rows.Scan(&log.ID, &log.ProjectID, &log.PushedBy, &log.SecretsID, &log.PushedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
