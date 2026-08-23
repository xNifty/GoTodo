package tasks_test

import (
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	os.Setenv("SESSION_KEY", "test-session-key-for-unit-tests-32chars!!")
	port := uint32(5438)
	// Pin a Maven-published binary version (DefaultConfig alone can drift and 404).
	// Isolate RuntimePath for parallel go test ./... against other packages.
	runtimePath := filepath.Join(os.TempDir(), "gotodo-embedded-pg-tasks")
	db := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V16).
		Port(port).
		Database("gotodo_test").
		RuntimePath(runtimePath))
	if err := db.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}

	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", fmt.Sprintf("%d", port))
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "gotodo_test")

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://postgres:postgres@localhost:%d/gotodo_test?sslmode=disable", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT, user_name TEXT);
		CREATE TABLE saved_views (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			filter_json JSONB NOT NULL DEFAULT '{}',
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (user_id, name)
		);
		CREATE TABLE projects (
			id SERIAL PRIMARY KEY,
			user_id INT,
			name TEXT,
			workflow_mode VARCHAR(16) NOT NULL DEFAULT 'classic'
		);
		CREATE TABLE project_statuses (
			id SERIAL PRIMARY KEY,
			project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			is_done BOOLEAN NOT NULL DEFAULT FALSE,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE project_members (
			project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL,
			role VARCHAR(16) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (project_id, user_id)
		);
		CREATE TABLE tags (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			project_id INTEGER,
			name TEXT NOT NULL,
			color VARCHAR(7) DEFAULT '#6c757d',
			protected BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE tasks (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE,
			time_stamp TIMESTAMP DEFAULT NOW(),
			is_favorite BOOLEAN DEFAULT FALSE,
			position INTEGER DEFAULT 0,
			priority SMALLINT DEFAULT 0,
			user_id INTEGER,
			project_id INTEGER,
			parent_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
			date_modified TIMESTAMP,
			due_date DATE,
			status_id INTEGER,
			estimate_points INTEGER,
			claimed_by INTEGER
		);
		CREATE TABLE task_tags (
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (task_id, tag_id)
		);
		CREATE TABLE task_time_entries (
			id SERIAL PRIMARY KEY,
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL,
			minutes INTEGER NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE task_events (
			id SERIAL PRIMARY KEY,
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL,
			event_type VARCHAR(32) NOT NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE task_github_issues (
			task_id INTEGER PRIMARY KEY REFERENCES tasks(id) ON DELETE CASCADE,
			issue_id BIGINT NOT NULL DEFAULT 0,
			issue_number INTEGER NOT NULL,
			issue_url TEXT NOT NULL DEFAULT '',
			issue_state VARCHAR(16) NOT NULL DEFAULT 'open',
			issue_title TEXT NOT NULL DEFAULT '',
			last_synced_at TIMESTAMPTZ,
			last_sync_error TEXT NOT NULL DEFAULT '',
			UNIQUE (issue_id),
			CHECK (issue_number > 0)
		);
		INSERT INTO users (id, email) VALUES
			(1, 'user@example.com'),
			(2, 'other@example.com'),
			(3, 'viewer@example.com'),
			(4, 'editor@example.com');
		INSERT INTO projects (id, user_id, name) VALUES
			(1, 1, 'Owned project'),
			(2, 2, 'Shared project');
		INSERT INTO project_members (project_id, user_id, role) VALUES
			(1, 1, 'owner'),
			(2, 2, 'owner'),
			(2, 3, 'viewer'),
			(2, 4, 'editor');
		INSERT INTO tags (id, user_id, name, color) VALUES (1, 1, 'work', '#0d6efd'), (2, 1, 'personal', '#198754');
		INSERT INTO tasks (title, description, user_id, completed, is_favorite, position, priority, project_id, due_date) VALUES
		 ('Favorite task', 'fav desc', 1, false, true, 1, 2, NULL, CURRENT_DATE),
		 ('Open task', 'open desc', 1, false, false, 2, 1, 1, CURRENT_DATE + 1),
		 ('Done task', 'done desc', 1, true, false, 3, 0, 1, CURRENT_DATE - 1),
		 ('Tagged task', 'has work tag', 1, false, false, 4, 0, NULL, NULL),
		 ('Shared owner task', 'owned by project owner', 2, false, false, 5, 0, 2, NULL);
		INSERT INTO task_tags (task_id, tag_id) VALUES (4, 1);
	`)
	pool.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = db.Stop()
	os.Exit(code)
}

func TestReturnPaginationForUserWithFilters(t *testing.T) {
	userID := 1
	timezone := "America/New_York"
	project := 1
	projectZero := 0

	cases := []struct {
		name    string
		filters tasks.ListFilters
	}{
		{"all", tasks.ListFilters{}},
		{"incomplete", tasks.ListFilters{StatusFilter: "incomplete"}},
		{"complete", tasks.ListFilters{StatusFilter: "complete"}},
		{"completed alias", tasks.ListFilters{StatusFilter: "Completed"}},
		{"project incomplete", tasks.ListFilters{ProjectFilter: &project, StatusFilter: "incomplete"}},
		{"no project complete", tasks.ListFilters{ProjectFilter: &projectZero, StatusFilter: "complete"}},
		{"due today", tasks.ListFilters{DueFilter: "today"}},
		{"due overdue", tasks.ListFilters{DueFilter: "overdue"}},
		{"tag filter", tasks.ListFilters{TagFilter: intPtr(1)}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, total, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &userID, timezone, tc.filters)
			if err != nil {
				t.Fatalf("ReturnPaginationForUserWithFilters: %v", err)
			}
			if total < 0 {
				t.Fatalf("expected non-negative total, got %d", total)
			}
		})
	}
}

func TestSearchTasksForUserWithFilters(t *testing.T) {
	userID := 1
	timezone := "America/New_York"

	favoriteResults, favoriteTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "Favorite", &userID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("favorite search: %v", err)
	}
	if favoriteTotal != 1 || len(favoriteResults) != 1 {
		t.Fatalf("expected one favorite search result, got total %d and %d tasks", favoriteTotal, len(favoriteResults))
	}
	if !favoriteResults[0].IsFavorite {
		t.Fatalf("expected search result %q to preserve favorite status", favoriteResults[0].Title)
	}

	_, total, err := tasks.SearchTasksForUserWithFilters(1, 10, "task", &userID, timezone, tasks.ListFilters{StatusFilter: "incomplete"})
	if err != nil {
		t.Fatalf("SearchTasksForUserWithFilters: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected 3 incomplete search matches, got %d", total)
	}

	tagID := 1
	tasksList, tagTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &userID, timezone, tasks.ListFilters{TagFilter: &tagID})
	if err != nil {
		t.Fatalf("tag filter list: %v", err)
	}
	if tagTotal != 1 {
		t.Fatalf("expected 1 tagged task, got total %d", tagTotal)
	}
	if len(tasksList) != 1 || tasksList[0].Title != "Tagged task" {
		t.Fatalf("expected tagged task on page, got %v", tasksList)
	}

	_, nameTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &userID, timezone, tasks.ListFilters{TagNameFilter: "WORK"})
	if err != nil {
		t.Fatalf("tag name filter list: %v", err)
	}
	if nameTotal != 1 {
		t.Fatalf("expected 1 task matching tag name, got total %d", nameTotal)
	}

	_, tagSearchTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "work", &userID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("tag name search: %v", err)
	}
	if tagSearchTotal != 1 {
		t.Fatalf("expected 1 task matching tag name search, got %d", tagSearchTotal)
	}

	idResults, idTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "4", &userID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("numeric id search: %v", err)
	}
	if idTotal != 1 || len(idResults) != 1 || idResults[0].Title != "Tagged task" {
		t.Fatalf("expected Tagged task for id 4, got total %d tasks %v", idTotal, idResults)
	}

	hashResults, hashTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "#4", &userID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("hash id search: %v", err)
	}
	if hashTotal != 1 || len(hashResults) != 1 || hashResults[0].ID != 4 {
		t.Fatalf("expected task 4 for #4, got total %d tasks %v", hashTotal, hashResults)
	}
}

func TestSharedProjectVisibilityByRole(t *testing.T) {
	timezone := "America/New_York"
	sharedProject := 2
	viewerID := 3
	editorID := 4

	viewerHome, viewerHomeTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &viewerID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("viewer home list: %v", err)
	}
	if viewerHomeTotal != 0 || len(viewerHome) != 0 {
		t.Fatalf("viewer should not see shared tasks on home list, got total %d tasks %v", viewerHomeTotal, titles(viewerHome))
	}

	viewerScoped, viewerScopedTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &viewerID, timezone, tasks.ListFilters{ProjectFilter: &sharedProject})
	if err != nil {
		t.Fatalf("viewer project list: %v", err)
	}
	if viewerScopedTotal != 1 || len(viewerScoped) != 1 || viewerScoped[0].Title != "Shared owner task" {
		t.Fatalf("viewer should see shared task in project view, got total %d tasks %v", viewerScopedTotal, titles(viewerScoped))
	}

	matchHome, err := tasks.TaskMatchesFilters(5, viewerID, timezone, tasks.ListFilters{}, "")
	if err != nil {
		t.Fatalf("viewer home TaskMatchesFilters: %v", err)
	}
	if matchHome {
		t.Fatal("viewer TaskMatchesFilters should be false on home list")
	}
	matchScoped, err := tasks.TaskMatchesFilters(5, viewerID, timezone, tasks.ListFilters{ProjectFilter: &sharedProject}, "")
	if err != nil {
		t.Fatalf("viewer scoped TaskMatchesFilters: %v", err)
	}
	if !matchScoped {
		t.Fatal("viewer TaskMatchesFilters should be true for project filter")
	}

	editorHome, editorHomeTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &editorID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("editor home list: %v", err)
	}
	if editorHomeTotal != 1 || len(editorHome) != 1 || editorHome[0].Title != "Shared owner task" {
		t.Fatalf("editor should see shared tasks on home list, got total %d tasks %v", editorHomeTotal, titles(editorHome))
	}

	editorScoped, editorScopedTotal, err := tasks.ReturnPaginationForUserWithFilters(1, 10, &editorID, timezone, tasks.ListFilters{ProjectFilter: &sharedProject})
	if err != nil {
		t.Fatalf("editor project list: %v", err)
	}
	if editorScopedTotal != 1 || len(editorScoped) != 1 || editorScoped[0].Title != "Shared owner task" {
		t.Fatalf("editor should see shared task in project view, got total %d tasks %v", editorScopedTotal, titles(editorScoped))
	}

	_, searchHomeTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "Shared", &viewerID, timezone, tasks.ListFilters{})
	if err != nil {
		t.Fatalf("viewer home search: %v", err)
	}
	if searchHomeTotal != 0 {
		t.Fatalf("viewer home search should hide shared tasks, got %d", searchHomeTotal)
	}
	_, searchScopedTotal, err := tasks.SearchTasksForUserWithFilters(1, 10, "Shared", &viewerID, timezone, tasks.ListFilters{ProjectFilter: &sharedProject})
	if err != nil {
		t.Fatalf("viewer project search: %v", err)
	}
	if searchScopedTotal != 1 {
		t.Fatalf("viewer project search should find shared task, got %d", searchScopedTotal)
	}
}

func TestFetchTaskByIDAttachesChildrenAndParent(t *testing.T) {
	pool, err := storage.OpenDatabase()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	var parentID, childID int
	err = pool.QueryRow(context.Background(),
		`INSERT INTO tasks (title, description, user_id, completed, is_favorite, position, priority, project_id)
		 VALUES ('Parent with kids', 'root', 1, false, false, 50, 0, 1)
		 RETURNING id`).Scan(&parentID)
	if err != nil {
		t.Fatalf("insert parent: %v", err)
	}
	err = pool.QueryRow(context.Background(),
		`INSERT INTO tasks (title, description, user_id, completed, is_favorite, position, priority, project_id, parent_id)
		 VALUES ('Nested child', 'child', 1, false, false, 1, 0, 1, $1)
		 RETURNING id`, parentID).Scan(&childID)
	if err != nil {
		t.Fatalf("insert child: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1", parentID)
	})

	parent, err := tasks.FetchTaskByIDForUser(parentID, 1, "UTC", 1)
	if err != nil {
		t.Fatalf("fetch parent: %v", err)
	}
	if parent.ChildCount != 1 || len(parent.Children) != 1 {
		t.Fatalf("expected parent to include 1 child, got count=%d children=%d", parent.ChildCount, len(parent.Children))
	}
	if parent.Children[0].ID != childID || parent.Children[0].Title != "Nested child" {
		t.Fatalf("unexpected child on parent: %+v", parent.Children[0])
	}

	child, err := tasks.FetchTaskByIDForUser(childID, 1, "UTC", 1)
	if err != nil {
		t.Fatalf("fetch child: %v", err)
	}
	if child.ParentID != parentID {
		t.Fatalf("expected child parent_id %d, got %d", parentID, child.ParentID)
	}
	if child.ChildCount != 0 || len(child.Children) != 0 {
		t.Fatalf("expected no nested children on subtask, got count=%d children=%d", child.ChildCount, len(child.Children))
	}
}

func titles(list []tasks.Task) []string {
	out := make([]string, len(list))
	for i, task := range list {
		out[i] = task.Title
	}
	return out
}

func intPtr(n int) *int {
	return &n
}
