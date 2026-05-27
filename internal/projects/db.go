package projects

import (
	"context"

	"github.com/SaranHiruthikM/envman/internal/models"
)

func (h *Handler) createProject(ctx context.Context, userId, name string) (string, error) {
	query1 := `INSERT INTO projects VALUES (
		gen_random_uuid(), $1, $2, NOW()
	) RETURNING id`

	query2 := `INSERT INTO user_projects VALUES (
		gen_random_uuid(), $1, $2, 'owner'
	)`

	tx, err := h.db.Begin(ctx)
	if err != nil {
		return "", err
	}

	var projectId string

	err = tx.QueryRow(ctx, query1, name, userId).Scan(&projectId)
	if err != nil {
		tx.Rollback(ctx)
		return "", err
	}

	_, err = tx.Exec(ctx, query2, userId, projectId)
	if err != nil {
		tx.Rollback(ctx)
		return "", err
	}

	tx.Commit(ctx)

	return projectId, nil
}

func (h *Handler) getProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	query := `SELECT id, name, created_by, created_at from projects WHERE id=$1`

	var project models.Project
	err := h.db.QueryRow(ctx, query, projectID).Scan(&project.ID, &project.Name, &project.CreatedBy, &project.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &project, nil

}
