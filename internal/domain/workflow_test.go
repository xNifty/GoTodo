package domain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"GoTodo/internal/storage"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	os.Setenv("SESSION_KEY", "test-session-key-for-unit-tests-32chars!!")
	port := uint32(5441)
	// Isolate RuntimePath so go test ./... can run this package in parallel with
	// internal/tasks (which also starts embedded-postgres).
	runtimePath := filepath.Join(os.TempDir(), "gotodo-embedded-pg-domain")
	db := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V16).
		Port(port).
		Database("gotodo_workflow_test").
		RuntimePath(runtimePath))
	if err := db.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}

	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", fmt.Sprintf("%d", port))
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "gotodo_workflow_test")

	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf("postgres://postgres:postgres@localhost:%d/gotodo_workflow_test?sslmode=disable", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}

	if err := storage.CreateUsersTable(); err != nil {
		fmt.Fprintf(os.Stderr, "users: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateUsersAddName(); err != nil {
		fmt.Fprintf(os.Stderr, "user_name: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateProjectsTable(); err != nil {
		fmt.Fprintf(os.Stderr, "projects: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateProjectsAddWorkflowMode(); err != nil {
		fmt.Fprintf(os.Stderr, "workflow mode: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateProjectSharingTables(); err != nil {
		fmt.Fprintf(os.Stderr, "sharing: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateTasksTable(); err != nil {
		fmt.Fprintf(os.Stderr, "tasks: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddProjectID(); err != nil {
		fmt.Fprintf(os.Stderr, "project_id: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddPosition(); err != nil {
		fmt.Fprintf(os.Stderr, "position: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddIsFavorite(); err != nil {
		fmt.Fprintf(os.Stderr, "favorite: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddPriority(); err != nil {
		fmt.Fprintf(os.Stderr, "priority: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddDateModified(); err != nil {
		fmt.Fprintf(os.Stderr, "date_modified: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddDueDate(); err != nil {
		fmt.Fprintf(os.Stderr, "due_date: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddParentID(); err != nil {
		fmt.Fprintf(os.Stderr, "parent_id: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateProjectWorkflowTables(); err != nil {
		fmt.Fprintf(os.Stderr, "workflow tables: %v\n", err)
		os.Exit(1)
	}
	if err := storage.MigrateTasksAddWorkflowFields(); err != nil {
		fmt.Fprintf(os.Stderr, "task workflow: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateTaskEventsTable(); err != nil {
		fmt.Fprintf(os.Stderr, "task events: %v\n", err)
		os.Exit(1)
	}
	if err := storage.CreateTagsTables(); err != nil {
		fmt.Fprintf(os.Stderr, "tags: %v\n", err)
		os.Exit(1)
	}

	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (id, email, password, role_id) VALUES
			(1, 'owner@example.com', 'x', 1),
			(2, 'editor@example.com', 'x', 1)
		ON CONFLICT DO NOTHING;
		SELECT setval(pg_get_serial_sequence('users','id'), (SELECT MAX(id) FROM users));
	`)
	pool.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = db.Stop()
	os.Exit(code)
}

func TestKanbanEnableDisableAndStatusCap(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Board Proj")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	updated, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban)
	if err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	if updated.WorkflowMode != storage.WorkflowKanban {
		t.Fatalf("mode=%q", updated.WorkflowMode)
	}

	statuses, err := ListProjectStatusesForUser(ctx, 1, proj.ID)
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("want 3 default statuses, got %d", len(statuses))
	}

	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{
		Title:     "Kanban task",
		ProjectID: &pid,
		Priority:  1,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, err = SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowClassic)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("disable with tasks: err=%v want conflict", err)
	}

	fields, err := storage.GetWorkflowFieldsForTasks([]int{taskID})
	if err != nil {
		t.Fatalf("workflow fields: %v", err)
	}
	f := fields[taskID]
	if f.StatusID == 0 || f.StatusName == "" {
		t.Fatalf("expected status on new task, got %+v", f)
	}

	// Fill up to max statuses (3 defaults + 5 custom = 8).
	for i := 0; i < 5; i++ {
		_, err := CreateProjectStatusForUser(ctx, 1, proj.ID, CreateProjectStatusInput{
			Name: fmt.Sprintf("Extra %d", i+1),
		})
		if err != nil {
			t.Fatalf("create status %d: %v", i+1, err)
		}
	}
	_, err = CreateProjectStatusForUser(ctx, 1, proj.ID, CreateProjectStatusInput{Name: "Over"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("status cap: err=%v want conflict", err)
	}

	// Completed sync via status.
	var doneID int
	statuses, _ = ListProjectStatusesForUser(ctx, 1, proj.ID)
	for _, s := range statuses {
		if s.IsDone {
			doneID = s.ID
			break
		}
	}
	sid := doneID
	statusPtr := &sid
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{StatusID: &statusPtr}); err != nil {
		t.Fatalf("set done status: %v", err)
	}
	pool, _ := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)
	var completed bool
	if err := pool.QueryRow(ctx, `SELECT completed FROM tasks WHERE id=$1`, taskID).Scan(&completed); err != nil {
		t.Fatal(err)
	}
	if !completed {
		t.Fatal("expected completed=true after done status")
	}

	// Time entry.
	entry, err := AddTimeEntryForUser(ctx, 1, taskID, 30, "pairing")
	if err != nil {
		t.Fatalf("add time: %v", err)
	}
	if entry.Minutes != 30 {
		t.Fatalf("minutes=%d", entry.Minutes)
	}

	// Delete task so disable is allowed.
	if err := DeleteTask(ctx, 1, taskID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowClassic); err != nil {
		t.Fatalf("disable empty kanban: %v", err)
	}
}

func TestKanbanEditorCannotManageStatuses(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Shared Board")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	_, err = CreateProjectStatusForUser(ctx, 2, proj.ID, CreateProjectStatusInput{Name: "Blocked"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor create status: err=%v want forbidden", err)
	}
}
