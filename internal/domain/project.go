package domain

import (
	"context"
	"fmt"
	"strings"

	"GoTodo/internal/storage"
)

// MaxProjectNameLength is the maximum length of a project name.
const MaxProjectNameLength = 50

// MaxProjectDescriptionLength is the maximum length of a project description.
const MaxProjectDescriptionLength = 1000

// CreateProject validates and creates a project for the user.
func CreateProject(ctx context.Context, userID int, name, description string) (*storage.Project, error) {
	_ = ctx
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: project name is required", ErrValidation)
	}
	if len(name) > MaxProjectNameLength {
		return nil, fmt.Errorf("%w: project name must be %d characters or less", ErrValidation, MaxProjectNameLength)
	}
	description = strings.TrimSpace(description)
	if len(description) > MaxProjectDescriptionLength {
		return nil, fmt.Errorf("%w: project description must be %d characters or less", ErrValidation, MaxProjectDescriptionLength)
	}
	return storage.CreateProject(userID, name, description)
}

// RenameProject updates a project name (owner only) and returns the updated project.
func RenameProject(ctx context.Context, userID, projectID int, name string) (*storage.Project, error) {
	return UpdateProject(ctx, userID, projectID, &name, nil, nil)
}

// UpdateProject patches name, description, and/or workflow_mode (owner only).
func UpdateProject(ctx context.Context, userID, projectID int, name, description, workflowMode *string) (*storage.Project, error) {
	_ = ctx
	var trimmedName string
	if name != nil {
		trimmedName = strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, fmt.Errorf("%w: project name is required", ErrValidation)
		}
		if len(trimmedName) > MaxProjectNameLength {
			return nil, fmt.Errorf("%w: project name must be %d characters or less", ErrValidation, MaxProjectNameLength)
		}
	}

	var trimmedDescription *string
	if description != nil {
		d := strings.TrimSpace(*description)
		if len(d) > MaxProjectDescriptionLength {
			return nil, fmt.Errorf("%w: project description must be %d characters or less", ErrValidation, MaxProjectDescriptionLength)
		}
		trimmedDescription = &d
	}

	proj, err := storage.GetAccessibleProjectByID(projectID, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !storage.RoleCanManage(proj.Role) {
		return nil, ErrForbidden
	}

	var namePtr *string
	if name != nil {
		namePtr = &trimmedName
	}
	if namePtr != nil || trimmedDescription != nil {
		if err := storage.UpdateProject(projectID, proj.OwnerUserID, namePtr, trimmedDescription); err != nil {
			return nil, err
		}
	}

	if workflowMode != nil {
		if _, err := SetProjectWorkflowMode(ctx, userID, projectID, *workflowMode); err != nil {
			return nil, err
		}
	}

	return storage.GetProjectByID(projectID, proj.OwnerUserID)
}

// ReorderProjectsForUser sets the display order of all owned projects.
func ReorderProjectsForUser(ctx context.Context, userID int, orderedIDs []int) error {
	_ = ctx
	if len(orderedIDs) == 0 {
		return fmt.Errorf("%w: project_ids is required", ErrValidation)
	}
	if err := storage.ReorderProjects(userID, orderedIDs); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	return nil
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
