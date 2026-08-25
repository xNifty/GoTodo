package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
)

func TestProjectSprintsCRUDAndTaskAssignment(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Sprint Board", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}

	listed, err := ListProjectSprintsForUser(ctx, 1, proj.ID)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected no sprints, got %d", len(listed))
	}

	sprint, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint 1",
		StartDate: "2026-08-24",
		EndDate:   "2026-09-06",
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	if sprint.Name != "Sprint 1" {
		t.Fatalf("name=%q", sprint.Name)
	}
	if storage.FormatSprintDate(sprint.StartDate) != "2026-08-24" {
		t.Fatalf("start=%q", storage.FormatSprintDate(sprint.StartDate))
	}
	if storage.FormatSprintDate(sprint.EndDate) != "2026-09-06" {
		t.Fatalf("end=%q", storage.FormatSprintDate(sprint.EndDate))
	}

	if _, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint 1",
		StartDate: "2026-09-07",
		EndDate:   "2026-09-20",
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate name err=%v", err)
	}

	if _, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Bad dates",
		StartDate: "2026-09-20",
		EndDate:   "2026-09-07",
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("inverted dates err=%v", err)
	}

	if _, err := CreateProjectSprintForUser(ctx, 2, proj.ID, CreateProjectSprintInput{
		Name:      "Editor sprint",
		StartDate: "2026-09-07",
		EndDate:   "2026-09-20",
	}); !errors.Is(err, ErrForbidden) && !errors.Is(err, ErrNotFound) {
		// editor is not a member of this project
		t.Fatalf("editor create err=%v", err)
	}

	updated, err := UpdateProjectSprintForUser(ctx, 1, proj.ID, sprint.ID, UpdateProjectSprintInput{
		Name: strPtr("Sprint One"),
	})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if updated.Name != "Sprint One" {
		t.Fatalf("renamed=%q", updated.Name)
	}

	pid := proj.ID
	sid := sprint.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{
		Title:     "In sprint",
		ProjectID: &pid,
		SprintID:  &sid,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	fields, err := storage.GetWorkflowFieldsForTasks([]int{taskID})
	if err != nil {
		t.Fatalf("workflow fields: %v", err)
	}
	if fields[taskID].SprintID != sid {
		t.Fatalf("sprint_id=%d want %d", fields[taskID].SprintID, sid)
	}
	if fields[taskID].SprintName != "Sprint One" {
		t.Fatalf("sprint_name=%q", fields[taskID].SprintName)
	}

	clear := (*int)(nil)
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{SprintID: &clear}); err != nil {
		t.Fatalf("clear sprint: %v", err)
	}
	fields, _ = storage.GetWorkflowFieldsForTasks([]int{taskID})
	if fields[taskID].SprintID != 0 {
		t.Fatalf("expected backlog, got %d", fields[taskID].SprintID)
	}

	set := &sid
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{SprintID: &set}); err != nil {
		t.Fatalf("set sprint: %v", err)
	}

	changed := eventsOfType(t, taskID, 1, "sprint_changed")
	if len(changed) < 2 {
		t.Fatalf("sprint_changed count=%d want >= 2", len(changed))
	}

	backlogID, err := CreateTask(ctx, 1, CreateTaskInput{
		Title:     "Backlog item",
		ProjectID: &pid,
	})
	if err != nil {
		t.Fatalf("create backlog: %v", err)
	}

	tz := "UTC"
	uid := 1
	sprintFilter := sid
	listedTasks, _, err := tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, tz, tasks.ListFilters{
		ProjectFilter:      &pid,
		WorkflowClaimScope: "all",
		SprintFilter:       &sprintFilter,
	})
	if err != nil {
		t.Fatalf("filter sprint: %v", err)
	}
	if !sprintListHasID(listedTasks, taskID) || sprintListHasID(listedTasks, backlogID) {
		t.Fatalf("sprint filter mismatch: %+v", sprintTaskIDs(listedTasks))
	}

	zero := 0
	backlogTasks, _, err := tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, tz, tasks.ListFilters{
		ProjectFilter:      &pid,
		WorkflowClaimScope: "all",
		SprintFilter:       &zero,
	})
	if err != nil {
		t.Fatalf("filter backlog: %v", err)
	}
	if !sprintListHasID(backlogTasks, backlogID) || sprintListHasID(backlogTasks, taskID) {
		t.Fatalf("backlog filter mismatch: %+v", sprintTaskIDs(backlogTasks))
	}

	sprint2, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint Two",
		StartDate: "2026-09-07",
		EndDate:   "2026-09-20",
	})
	if err != nil {
		t.Fatalf("create sprint 2: %v", err)
	}
	sid2 := sprint2.ID
	set2 := &sid2
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{SprintID: &set2}); err != nil {
		t.Fatalf("move to sprint 2: %v", err)
	}
	fields, _ = storage.GetWorkflowFieldsForTasks([]int{taskID})
	if fields[taskID].SprintID != sid2 {
		t.Fatalf("moved sprint_id=%d want %d", fields[taskID].SprintID, sid2)
	}
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{SprintID: &set}); err != nil {
		t.Fatalf("move back to sprint 1: %v", err)
	}

	if err := DeleteProjectSprintForUser(ctx, 1, proj.ID, sprint.ID, nil); err != nil {
		t.Fatalf("delete sprint: %v", err)
	}
	fields, _ = storage.GetWorkflowFieldsForTasks([]int{taskID})
	if fields[taskID].SprintID != 0 {
		t.Fatalf("task should return to backlog after sprint delete, got %d", fields[taskID].SprintID)
	}
}

func TestSprintRequiresKanbanAndValidDates(t *testing.T) {
	ctx := context.Background()
	classic, err := CreateProject(ctx, 1, "Classic No Sprint", "")
	if err != nil {
		t.Fatalf("create classic: %v", err)
	}
	if _, err := CreateProjectSprintForUser(ctx, 1, classic.ID, CreateProjectSprintInput{
		Name:      "Nope",
		StartDate: "2026-08-24",
		EndDate:   "2026-09-06",
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("classic sprint err=%v", err)
	}

	kanban, err := CreateProject(ctx, 1, "Kanban Sprint Dates", "")
	if err != nil {
		t.Fatalf("create kanban: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, kanban.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	if _, err := CreateProjectSprintForUser(ctx, 1, kanban.ID, CreateProjectSprintInput{
		Name:      "Missing dates",
		StartDate: "",
		EndDate:   "2026-09-06",
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("missing start err=%v", err)
	}

	pid := kanban.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "No sprint on classic move", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	bad := 999999
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{SprintID: ptrToIntPtr(bad)}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid sprint_id err=%v", err)
	}
}

func TestSprintIsActiveWindow(t *testing.T) {
	start, _ := time.Parse("2006-01-02", "2026-08-24")
	end, _ := time.Parse("2006-01-02", "2026-09-06")
	now, _ := time.Parse(time.RFC3339, "2026-08-24T18:00:00Z")
	if !storage.SprintIsActive(start, end, now) {
		t.Fatal("expected active on start date")
	}
	before, _ := time.Parse(time.RFC3339, "2026-08-23T23:00:00Z")
	if storage.SprintIsActive(start, end, before) {
		t.Fatal("expected inactive before start")
	}
	after, _ := time.Parse(time.RFC3339, "2026-09-07T00:00:00Z")
	if storage.SprintIsActive(start, end, after) {
		t.Fatal("expected inactive after end")
	}
}

func TestSprintDatesOverlap(t *testing.T) {
	parse := func(s string) time.Time {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	if !storage.SprintDatesOverlap(parse("2026-08-01"), parse("2026-08-15"), parse("2026-08-15"), parse("2026-08-31")) {
		t.Fatal("sharing an endpoint day should overlap")
	}
	if storage.SprintDatesOverlap(parse("2026-08-01"), parse("2026-08-14"), parse("2026-08-15"), parse("2026-08-31")) {
		t.Fatal("adjacent ranges should not overlap")
	}
	if !storage.SprintDatesOverlap(parse("2026-08-01"), parse("2026-08-31"), parse("2026-08-10"), parse("2026-08-12")) {
		t.Fatal("contained range should overlap")
	}
}

func TestSprintRejectsOverlappingDates(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "No Overlap Board", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	first, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint A",
		StartDate: "2026-08-01",
		EndDate:   "2026-08-14",
	})
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	if _, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint B",
		StartDate: "2026-08-15",
		EndDate:   "2026-08-31",
	}); err != nil {
		t.Fatalf("adjacent sprint should be allowed: %v", err)
	}
	_, err = CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Overlap",
		StartDate: "2026-08-10",
		EndDate:   "2026-08-20",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("overlapping create err=%v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "Sprint A") {
		t.Fatalf("overlap error should name the other sprint, got %v", err)
	}
	_, err = UpdateProjectSprintForUser(ctx, 1, proj.ID, first.ID, UpdateProjectSprintInput{
		EndDate: strPtr("2026-08-20"),
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("overlapping update err=%v", err)
	}
	if _, err := UpdateProjectSprintForUser(ctx, 1, proj.ID, first.ID, UpdateProjectSprintInput{
		EndDate: strPtr("2026-08-14"),
	}); err != nil {
		t.Fatalf("updating a sprint to its own range should be allowed: %v", err)
	}
}

func TestSubtaskInheritsParentSprint(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Inherit Sprint", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	sprint, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Inherit Me",
		StartDate: "2026-08-01",
		EndDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	pid := proj.ID
	sid := sprint.ID
	parentID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Parent", ProjectID: &pid, SprintID: &sid})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	childID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Child", ParentID: &parentID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	fields, err := storage.GetWorkflowFieldsForTasks([]int{childID})
	if err != nil {
		t.Fatalf("fields: %v", err)
	}
	if fields[childID].SprintID != sid {
		t.Fatalf("child sprint_id=%d want %d", fields[childID].SprintID, sid)
	}
}

func TestSprintBoardSeparatesParentAndChildAssignments(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Mixed Sprint Board", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	sprint, err := CreateProjectSprintForUser(ctx, 1, proj.ID, CreateProjectSprintInput{
		Name:      "Sprint 1",
		StartDate: "2026-08-01",
		EndDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	pid := proj.ID
	sid := sprint.ID
	parentID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Kanban Project 2", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	childID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Sub task", ProjectID: &pid, ParentID: &parentID, SprintID: &sid})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	tz := "UTC"
	uid := 1
	zero := 0
	backlog, _, err := tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, tz, tasks.ListFilters{
		ProjectFilter:      &pid,
		WorkflowClaimScope: "all",
		SprintFilter:       &zero,
	})
	if err != nil {
		t.Fatalf("backlog list: %v", err)
	}
	parent := sprintTopLevel(backlog, parentID)
	if parent == nil {
		t.Fatalf("backlog missing parent: %v", sprintTaskIDs(backlog))
	}
	if sprintNestedHasID(*parent, childID) {
		t.Fatal("sprint-1 subtask should not nest under a backlog parent")
	}
	if sprintTopLevel(backlog, childID) != nil {
		t.Fatal("sprint-1 subtask should not appear as a backlog card")
	}

	sprintTasks, _, err := tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, tz, tasks.ListFilters{
		ProjectFilter:      &pid,
		WorkflowClaimScope: "all",
		SprintFilter:       &sid,
	})
	if err != nil {
		t.Fatalf("sprint list: %v", err)
	}
	if sprintTopLevel(sprintTasks, parentID) != nil {
		t.Fatal("backlog parent should not appear on sprint 1")
	}
	child := sprintTopLevel(sprintTasks, childID)
	if child == nil {
		t.Fatalf("sprint-1 missing subtask card: %v", sprintTaskIDs(sprintTasks))
	}
	if child.ParentID != parentID {
		t.Fatalf("subtask parent_id=%d want %d", child.ParentID, parentID)
	}
	if child.ParentTitle != "Kanban Project 2" {
		t.Fatalf("subtask parent_title=%q", child.ParentTitle)
	}
}

func strPtr(s string) *string { return &s }

func ptrToIntPtr(v int) **int {
	p := &v
	return &p
}

func sprintListHasID(list []tasks.Task, id int) bool {
	for _, t := range list {
		if t.ID == id {
			return true
		}
		for _, c := range t.Children {
			if c.ID == id {
				return true
			}
		}
	}
	return false
}

func sprintTopLevel(list []tasks.Task, id int) *tasks.Task {
	for i := range list {
		if list[i].ID == id {
			return &list[i]
		}
	}
	return nil
}

func sprintNestedHasID(parent tasks.Task, id int) bool {
	for _, c := range parent.Children {
		if c.ID == id {
			return true
		}
	}
	return false
}

func sprintTaskIDs(list []tasks.Task) []int {
	ids := make([]int, 0, len(list))
	for _, t := range list {
		ids = append(ids, t.ID)
	}
	return ids
}
