package projects

import (
	"context"
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/models"
)

func (h *Handler) createProject(ctx context.Context, userID, name string) (string, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var projectID string
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (id, name, created_by, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW()) RETURNING id`,
		name, userID,
	).Scan(&projectID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_projects (id, user_id, project_id, role)
		VALUES (gen_random_uuid(), $1, $2, 'owner')`,
		userID, projectID,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit failed: %w", err)
	}
	return projectID, nil
}

func (h *Handler) getProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	var p models.Project
	err := h.db.QueryRow(ctx,
		`SELECT id, name, created_by, created_at FROM projects WHERE id=$1`,
		projectID,
	).Scan(&p.ID, &p.Name, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
