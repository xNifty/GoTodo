package storage

import (
	"context"
	"fmt"
	"time"
)

// Project represents a user-owned project that can contain tasks.
type Project struct {
	ID        int
	UserID    int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateProject inserts a new project for the given user and returns it.
func CreateProject(userID int, name string) (*Project, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var p Project
	err = pool.QueryRow(context.Background(), "INSERT INTO projects (user_id, name) VALUES ($1, $2) RETURNING id, user_id, name, created_at, updated_at", userID, name).Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}
	return &p, nil
}

// UpdateProject updates the name of a project owned by the user.
func UpdateProject(id int, userID int, name string) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), "UPDATE projects SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3", name, id, userID)
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

// GetProjectsForUser returns all projects owned by a user.
func GetProjectsForUser(userID int) ([]Project, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(), "SELECT id, user_id, name, created_at, updated_at FROM projects WHERE user_id = $1 ORDER BY name", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
	err = pool.QueryRow(context.Background(), "SELECT id, user_id, name, created_at, updated_at FROM projects WHERE id = $1 AND user_id = $2", id, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}
	return &p, nil
}
