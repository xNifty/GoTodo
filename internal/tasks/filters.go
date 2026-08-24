package tasks

import (
	"fmt"
	"strings"
)

// ListFilters holds query filters for task list and search endpoints.
type ListFilters struct {
	ProjectFilter      *int
	StatusFilter       string
	DueFilter          string
	CompletedFilter    string
	PriorityFilter     *int
	TagFilter          *int
	TagNameFilter      string
	Sort               string
	WorkflowClaimScope string // "mine" | "all" | ""
	// SprintFilter: nil = no filter; &0 = backlog (no sprint); &n = sprint n.
	SprintFilter *int
}

func (f ListFilters) projectCondition(tablePrefix string) string {
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	if f.ProjectFilter == nil {
		return ""
	}
	if *f.ProjectFilter == 0 {
		return fmt.Sprintf(" AND (%sproject_id IS NULL)", prefix)
	}
	return fmt.Sprintf(" AND (%sproject_id = %d)", prefix, *f.ProjectFilter)
}

func (f ListFilters) statusCondition(tablePrefix string) string {
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	switch normalizeListStatusFilter(f.StatusFilter) {
	case "complete":
		return fmt.Sprintf(" AND %scompleted = true", prefix)
	case "incomplete":
		return fmt.Sprintf(" AND (%scompleted IS NULL OR %scompleted = false)", prefix, prefix)
	default:
		return ""
	}
}

func normalizeListStatusFilter(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "complete", "completed":
		return "complete"
	case "incomplete":
		return "incomplete"
	default:
		return ""
	}
}

func (f ListFilters) priorityCondition(tablePrefix string) string {
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	if f.PriorityFilter == nil {
		return ""
	}
	return fmt.Sprintf(" AND (%spriority = %d)", prefix, *f.PriorityFilter)
}

func (f ListFilters) sprintCondition(tablePrefix string) string {
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	if f.SprintFilter == nil {
		return ""
	}
	if *f.SprintFilter == 0 {
		return fmt.Sprintf(" AND (%ssprint_id IS NULL)", prefix)
	}
	return fmt.Sprintf(" AND (%ssprint_id = %d)", prefix, *f.SprintFilter)
}

func (f ListFilters) orderByClause(tablePrefix string) string {
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	if f.Sort == "priority" {
		return fmt.Sprintf(" ORDER BY %spriority DESC, %sposition", prefix, prefix)
	}
	return fmt.Sprintf(" ORDER BY %sposition", prefix)
}

// workflowClaimCondition hides unclaimed kanban tasks when scope is "mine".
// The personal list is an active work queue: completed kanban work is also
// hidden unless the caller explicitly asks for completed tasks (status=complete
// or completed=week), so done cards stay on the board but leave the list.
func (f ListFilters) workflowClaimCondition(tablePrefix string, argIdx int) string {
	if strings.ToLower(strings.TrimSpace(f.WorkflowClaimScope)) != "mine" {
		return ""
	}
	prefix := ""
	if tablePrefix != "" {
		prefix = tablePrefix + "."
	}
	hideDoneKanban := normalizeListStatusFilter(f.StatusFilter) != "complete" &&
		strings.TrimSpace(f.CompletedFilter) == ""
	doneClause := ""
	if hideDoneKanban {
		doneClause = fmt.Sprintf(" AND (%scompleted IS NULL OR %scompleted = false)", prefix, prefix)
	}
	return fmt.Sprintf(` AND (
		%sproject_id IS NULL
		OR NOT EXISTS (
			SELECT 1 FROM projects p
			WHERE p.id = %sproject_id AND COALESCE(p.workflow_mode, 'classic') = 'kanban'
		)
		OR (%sclaimed_by = $%d%s)
	)`, prefix, prefix, prefix, argIdx, doneClause)
}

func (f ListFilters) appendConditions(baseWhere string, timezone string, tablePrefix string, args []interface{}, userID int) (string, []interface{}) {
	return appendFilterSQL(baseWhere, args, f, timezone, tablePrefix, userID)
}
