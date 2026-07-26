package domain

import (
	"context"
	"fmt"
	"strings"

	"GoTodo/internal/storage"
)

// MaxProjectNameLength is the maximum length of a project name.
const MaxProjectNameLength = 50

// CreateProject validates and creates a project for the user.
func CreateProject(ctx context.Context, userID int, name string) (*storage.Project, error) {
	_ = ctx
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: project name is required", ErrValidation)
	}
	if len(name) > MaxProjectNameLength {
		return nil, fmt.Errorf("%w: project name must be %d characters or less", ErrValidation, MaxProjectNameLength)
	}
	return storage.CreateProject(userID, name)
}

// RenameProject updates a project name (owner only) and returns the updated project.
func RenameProject(ctx context.Context, userID, projectID int, name string) (*storage.Project, error) {
	_ = ctx
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: project name is required", ErrValidation)
	}
	if len(name) > MaxProjectNameLength {
		return nil, fmt.Errorf("%w: project name must be %d characters or less", ErrValidation, MaxProjectNameLength)
	}
	proj, err := storage.GetAccessibleProjectByID(projectID, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !storage.RoleCanManage(proj.Role) {
		return nil, ErrForbidden
	}
	if err := storage.UpdateProject(projectID, proj.OwnerUserID, name); err != nil {
		return nil, err
	}
	return storage.GetProjectByID(projectID, proj.OwnerUserID)
}

// DeleteProject removes a project (owner only).
func DeleteProject(ctx context.Context, userID, projectID int) error {
	_ = ctx
	proj, err := storage.GetAccessibleProjectByID(projectID, userID)
	if err != nil {
		return ErrNotFound
	}
	if !storage.RoleCanManage(proj.Role) {
		return ErrForbidden
	}
	return storage.DeleteProject(projectID, proj.OwnerUserID)
}

// RequireProjectWriteAccess ensures the user can create/edit tasks in the project.
func RequireProjectWriteAccess(projectID, userID int) error {
	if projectID <= 0 {
		return fmt.Errorf("%w: invalid project_id", ErrValidation)
	}
	proj, err := storage.GetAccessibleProjectByID(projectID, userID)
	if err != nil {
		return fmt.Errorf("%w: invalid project_id", ErrValidation)
	}
	if !storage.RoleCanWrite(proj.Role) {
		return ErrForbidden
	}
	return nil
}
