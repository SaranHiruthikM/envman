package permissions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Sole purpose of this pkg is to check the user role and return it
// Will be used by pkgs which needs to know the user role before doing works
func GetUserRole(ctx context.Context, db *pgxpool.Pool, userID, projectID string) (string, error) {

	query := "SELECT role FROM user_projects WHERE user_id=$1 AND project_id=$2"

	var role string

	err := db.QueryRow(ctx, query, userID, projectID).Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}
