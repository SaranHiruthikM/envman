package projects

import "context"

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
