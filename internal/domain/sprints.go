package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"GoTodo/internal/live"
	"GoTodo/internal/storage"
)

// CreateProjectSprintInput is the create payload for a sprint.
type CreateProjectSprintInput struct {
	Name        string
	Description string
	StartDate   string
	EndDate     string
}

// UpdateProjectSprintInput is a partial sprint update.
type UpdateProjectSprintInput struct {
	Name        *string
	Description *string
	StartDate   *string
	EndDate     *string
}

func normalizeSprintName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", fmt.Errorf("%w: sprint name is required", ErrValidation)
	}
	if len(name) > storage.MaxSprintNameLen {
		return "", fmt.Errorf("%w: sprint name must be %d characters or less", ErrValidation, storage.MaxSprintNameLen)
	}
	return name, nil
}

func normalizeSprintDescription(raw string) (string, error) {
	desc := strings.TrimSpace(raw)
	if len(desc) > storage.MaxSprintDescriptionLen {
		return "", fmt.Errorf("%w: sprint description must be %d characters or less", ErrValidation, storage.MaxSprintDescriptionLen)
	}
	return desc, nil
}

func parseSprintDateRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	start, err := storage.ParseSprintDate(startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: start_date %s", ErrValidation, err.Error())
	}
	end, err := storage.ParseSprintDate(endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: end_date %s", ErrValidation, err.Error())
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: end_date must be on or after start_date", ErrValidation)
	}
	return start, end, nil
}

func rejectOverlappingSprint(projectID, excludeID int, start, end time.Time) error {
	hit, err := storage.FindOverlappingProjectSprint(projectID, start, end, excludeID)
	if err != nil {
		return err
	}
	if hit == nil {
		return nil
	}
	return fmt.Errorf("%w: dates overlap %s (%s – %s)", ErrValidation, hit.Name,
		storage.FormatSprintDate(hit.StartDate), storage.FormatSprintDate(hit.EndDate))
}

func sprintConflictError(err error) error {
	if err == nil {
		return nil
	}
	low := strings.ToLower(err.Error())
	if strings.Contains(low, "unique") || strings.Contains(low, "duplicate") {
		return fmt.Errorf("%w: a sprint with this name already exists", ErrConflict)
	}
	if strings.Contains(low, "check") {
		return fmt.Errorf("%w: end_date must be on or after start_date", ErrValidation)
	}
	return err
}

// ListProjectSprintsForUser returns sprints if the user can access the project.
func ListProjectSprintsForUser(ctx context.Context, userID, projectID int) ([]storage.ProjectSprint, error) {
	_ = ctx
	proj, err := storage.GetAccessibleProjectByID(projectID, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if proj.WorkflowMode != storage.WorkflowKanban {
		return []storage.ProjectSprint{}, nil
	}
	return storage.ListProjectSprints(projectID)
}

// CreateProjectSprintForUser adds a sprint (owner only).
func CreateProjectSprintForUser(ctx context.Context, userID, projectID int, in CreateProjectSprintInput) (*storage.ProjectSprint, error) {
	_ = ctx
	if _, err := requireKanbanOwner(projectID, userID); err != nil {
		return nil, err
	}
	name, err := normalizeSprintName(in.Name)
	if err != nil {
		return nil, err
	}
	desc, err := normalizeSprintDescription(in.Description)
	if err != nil {
		return nil, err
	}
	start, end, err := parseSprintDateRange(in.StartDate, in.EndDate)
	if err != nil {
		return nil, err
	}
	if err := rejectOverlappingSprint(projectID, 0, start, end); err != nil {
		return nil, err
	}
	n, err := storage.CountProjectSprints(projectID)
	if err != nil {
		return nil, err
	}
	if n >= storage.MaxProjectSprints {
		return nil, fmt.Errorf("%w: a maximum of %d sprints is allowed", ErrConflict, storage.MaxProjectSprints)
	}
	s, err := storage.CreateProjectSprint(projectID, name, desc, start, end)
	if err != nil {
		return nil, sprintConflictError(err)
	}
	_ = storage.LogProjectEvent(projectID, userID, "sprint_added", map[string]interface{}{
		"sprint_id": s.ID, "name": s.Name,
	})
	live.AfterProjectChange(userID, projectID, live.TypeProjectUpdated)
	return s, nil
}

// UpdateProjectSprintForUser updates a sprint (owner only).
func UpdateProjectSprintForUser(ctx context.Context, userID, projectID, sprintID int, in UpdateProjectSprintInput) (*storage.ProjectSprint, error) {
	_ = ctx
	if _, err := requireKanbanOwner(projectID, userID); err != nil {
		return nil, err
	}
	cur, err := storage.GetProjectSprint(projectID, sprintID)
	if err != nil {
		return nil, ErrNotFound
	}
	var name *string
	if in.Name != nil {
		n, err := normalizeSprintName(*in.Name)
		if err != nil {
			return nil, err
		}
		name = &n
	}
	var description *string
	if in.Description != nil {
		d, err := normalizeSprintDescription(*in.Description)
		if err != nil {
			return nil, err
		}
		description = &d
	}
	startRaw := storage.FormatSprintDate(cur.StartDate)
	endRaw := storage.FormatSprintDate(cur.EndDate)
	if in.StartDate != nil {
		startRaw = *in.StartDate
	}
	if in.EndDate != nil {
		endRaw = *in.EndDate
	}
	start, end, err := parseSprintDateRange(startRaw, endRaw)
	if err != nil {
		return nil, err
	}
	if err := rejectOverlappingSprint(projectID, sprintID, start, end); err != nil {
		return nil, err
	}
	s, err := storage.UpdateProjectSprint(projectID, sprintID, name, description, &start, &end)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "sprint not found") {
			return nil, ErrNotFound
		}
		return nil, sprintConflictError(err)
	}
	_ = storage.LogProjectEvent(projectID, userID, "sprint_updated", map[string]interface{}{
		"sprint_id": s.ID, "name": s.Name,
	})
	live.AfterProjectChange(userID, projectID, live.TypeProjectUpdated)
	return s, nil
}

// DeleteProjectSprintForUser deletes a sprint, optionally moving tasks first.
func DeleteProjectSprintForUser(ctx context.Context, userID, projectID, sprintID int, moveToSprintID *int) error {
	_ = ctx
	if _, err := requireKanbanOwner(projectID, userID); err != nil {
		return err
	}
	cur, err := storage.GetProjectSprint(projectID, sprintID)
	if err != nil {
		return ErrNotFound
	}

	n, err := storage.CountTasksWithSprint(sprintID)
	if err != nil {
		return err
	}
	if n > 0 && moveToSprintID != nil {
		if *moveToSprintID == sprintID {
			return fmt.Errorf("%w: move_to_sprint_id must be a different sprint", ErrValidation)
		}
		if *moveToSprintID > 0 {
			if _, err := storage.GetProjectSprint(projectID, *moveToSprintID); err != nil {
				return fmt.Errorf("%w: invalid move_to_sprint_id", ErrValidation)
			}
		}
		if err := storage.MoveTasksFromSprint(sprintID, moveToSprintID); err != nil {
			return err
		}
	}

	if err := storage.DeleteProjectSprint(projectID, sprintID); err != nil {
		return err
	}
	_ = storage.LogProjectEvent(projectID, userID, "sprint_deleted", map[string]interface{}{
		"name": cur.Name,
	})
	live.AfterProjectChange(userID, projectID, live.TypeProjectUpdated)
	return nil
}

func applyTaskSprint(taskID, projectID int, sprintID *int) error {
	if sprintID == nil || *sprintID <= 0 {
		return storage.SetTaskSprintID(taskID, nil)
	}
	if projectID <= 0 {
		return fmt.Errorf("%w: sprint requires a kanban project", ErrValidation)
	}
	mode, err := storage.GetProjectWorkflowMode(projectID)
	if err != nil {
		return err
	}
	if mode != storage.WorkflowKanban {
		return fmt.Errorf("%w: sprint requires a kanban project", ErrValidation)
	}
	if _, err := storage.GetProjectSprint(projectID, *sprintID); err != nil {
		return fmt.Errorf("%w: invalid sprint_id", ErrValidation)
	}
	return storage.SetTaskSprintID(taskID, sprintID)
}
