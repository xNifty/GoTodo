package storage

import (
	"context"
	"fmt"
	"time"
)

// Project represents a user-owned project that can contain tasks.
type Project struct {
	ID           int
	UserID       int
	Name         string
	Description  string
	WorkflowMode string
	Position     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateProject inserts a new project for the given user and returns it.
// New projects are appended at the end of the owner's ordered list.
func CreateProject(userID int, name, description string) (*Project, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var p Project
	err = pool.QueryRow(context.Background(),
		`INSERT INTO projects (user_id, name, description, position)
		 VALUES (
		   $1, $2, $3,
		   COALESCE((SELECT MAX(position) FROM projects WHERE user_id = $1), -1) + 1
		 )
		 RETURNING id, user_id, name, COALESCE(description, ''), COALESCE(workflow_mode, 'classic'),
		           COALESCE(position, 0), created_at, updated_at`,
		userID, name, description).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.WorkflowMode, &p.Position, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}
	if p.WorkflowMode == "" {
		p.WorkflowMode = WorkflowClassic
	}
	if err := EnsureProjectOwnerMember(p.ID, userID); err != nil {
		return nil, fmt.Errorf("failed to create project owner membership: %v", err)
	}
	return &p, nil
}

// UpdateProject updates name and/or description of a project owned by the user.
// Nil pointers leave that field unchanged.
func UpdateProject(id int, userID int, name *string, description *string) error {
	if name == nil && description == nil {
		return nil
	}
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	switch {
	case name != nil && description != nil:
		_, err = pool.Exec(context.Background(),
			`UPDATE projects SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
			 WHERE id = $3 AND user_id = $4`, *name, *description, id, userID)
	case name != nil:
		_, err = pool.Exec(context.Background(),
			`UPDATE projects SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3`,
			*name, id, userID)
	default:
		_, err = pool.Exec(context.Background(),
			`UPDATE projects SET description = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3`,
			*description, id, userID)
	}
	if err != nil {
		return fmt.Errorf("failed to update project: %v", err)
	}
	return nil
}

// DeleteProject removes a project owned by the user.
func DeleteProject(id int, userID int) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}
	return nil
}

// GetProjectsForUser returns all projects owned by a user, ordered by position.
func GetProjectsForUser(userID int) ([]Project, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(),
		`SELECT id, user_id, name, COALESCE(description, ''), COALESCE(workflow_mode, 'classic'),
		        COALESCE(position, 0), created_at, updated_at
		 FROM projects WHERE user_id = $1
		 ORDER BY position ASC, LOWER(name) ASC, id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.WorkflowMode, &p.Position, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %v", err)
		}
		out = append(out, p)
	}
	return out, nil
}

// GetProjectByID returns a project by id for the given user.
func GetProjectByID(id int, userID int) (*Project, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var p Project
	err = pool.QueryRow(context.Background(),
		`SELECT id, user_id, name, COALESCE(description, ''), COALESCE(workflow_mode, 'classic'),
		        COALESCE(position, 0), created_at, updated_at
		 FROM projects WHERE id = $1 AND user_id = $2`, id, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.WorkflowMode, &p.Position, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	return &p, nil
}

// ReorderProjects sets positions from ordered owned project IDs for a user.
// orderedIDs must be the full set of projects owned by userID.
func ReorderProjects(userID int, orderedIDs []int) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	existing, err := GetProjectsForUser(userID)
	if err != nil {
		return err
	}
	if len(orderedIDs) != len(existing) {
		return fmt.Errorf("project list mismatch")
	}
	have := make(map[int]struct{}, len(existing))
	for _, p := range existing {
		have[p.ID] = struct{}{}
	}
	for _, id := range orderedIDs {
		if _, ok := have[id]; !ok {
			return fmt.Errorf("project %d not owned by user", id)
		}
	}

	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	for i, id := range orderedIDs {
		if _, err := tx.Exec(context.Background(),
			`UPDATE projects SET position = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3`,
			i, id, userID); err != nil {
			return err
		}
	}
	return tx.Commit(context.Background())
}
