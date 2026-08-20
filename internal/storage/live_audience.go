package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

// TaskOwnerAndProject returns the task owner and project (0 if personal).
func TaskOwnerAndProject(taskID int) (ownerID, projectID int, err error) {
	pool, err := OpenDatabase()
	if err != nil {
		return 0, 0, err
	}
	defer CloseDatabase(pool)

	var proj sql.NullInt64
	err = pool.QueryRow(context.Background(),
		`SELECT user_id, project_id FROM tasks WHERE id = $1`, taskID).Scan(&ownerID, &proj)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return 0, 0, err
		}
		return 0, 0, err
	}
	if proj.Valid {
		projectID = int(proj.Int64)
	}
	return ownerID, projectID, nil
}

// ProjectMemberUserIDs returns user IDs for every project_members row.
func ProjectMemberUserIDs(projectID int) ([]int, error) {
	members, err := ListProjectMembers(projectID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(members))
	for _, m := range members {
		if m.UserID > 0 {
			ids = append(ids, m.UserID)
		}
	}
	return ids, nil
}
