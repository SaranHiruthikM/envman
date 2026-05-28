package members

import "context"

func (h *Handler) inviteMembers(ctx context.Context, userID, projectID, role string) error {
	query := `INSERT INTO user_projects VALUES (
		gen_random_uuid(), $1, $2, $3
	)`

	_, err := h.db.Exec(ctx, query, userID, projectID, role)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) updateRole(ctx context.Context, userID, projectID, role string) error {
	query := `UPDATE user_projects SET role=$1 WHERE user_id=$2 AND project_id=$3`

	_, err := h.db.Exec(ctx, query, role, userID, projectID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) removeMember(ctx context.Context, userID, projectID string) error {
	_, err := h.db.Exec(ctx, `DELETE FROM user_projects WHERE user_id=$1 AND project_id=$2`, userID, projectID)
	return err
}

func (h *Handler) countOwners(ctx context.Context, projectID string) (int, error) {
	var count int
	err := h.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_projects WHERE project_id=$1 AND role='owner'`, projectID).Scan(&count)
	return count, err
}
