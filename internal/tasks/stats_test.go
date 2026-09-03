package tasks_test

import (
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestGetDashboardStats(t *testing.T) {
	stats, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats == nil {
		t.Fatal("expected stats")
	}
	if stats.ByProject == nil {
		t.Error("ByProject must be non-nil")
	}
	if stats.ByPriority == nil {
		t.Error("ByPriority must be non-nil")
	}
	if stats.CompletionsLast7Days == nil {
		t.Error("CompletionsLast7Days must be non-nil")
	}
	if len(stats.ByProject) == 0 {
		t.Error("expected open tasks grouped by project")
	}
	if len(stats.CompletionsLast7Days) != 7 {
		t.Fatalf("expected 7 chart days, got %d", len(stats.CompletionsLast7Days))
	}
	if stats.DueThisWeekCount < stats.DueTodayCount {
		t.Fatalf("due_this_week_count (%d) should include due today (%d)", stats.DueThisWeekCount, stats.DueTodayCount)
	}
}

func TestGetDashboardStatsEmptyBreakdownsEncodeAsArrays(t *testing.T) {
	// Seeded user 3 has no tasks, so project/priority GROUP BYs return zero rows.
	stats, err := tasks.GetDashboardStats(3, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.ByProject == nil {
		t.Fatal("ByProject must be non-nil empty slice, not nil")
	}
	if stats.ByPriority == nil {
		t.Fatal("ByPriority must be non-nil empty slice, not nil")
	}
	if len(stats.ByProject) != 0 || len(stats.ByPriority) != 0 {
		t.Fatalf("expected empty breakdowns, got by_project=%v by_priority=%v", stats.ByProject, stats.ByPriority)
	}

	raw, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, `"by_project":[]`) {
		t.Errorf("expected by_project:[], got %s", body)
	}
	if !strings.Contains(body, `"by_priority":[]`) {
		t.Errorf("expected by_priority:[], got %s", body)
	}
	if strings.Contains(body, `"by_project":null`) || strings.Contains(body, `"by_priority":null`) {
		t.Errorf("breakdown fields must not encode as null: %s", body)
	}
}

func TestGetDashboardStatsCompletionNotDoubleCounted(t *testing.T) {
	before, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats baseline: %v", err)
	}
	weekBefore := before.CompletedThisWeek
	todayBefore := before.CompletionsLast7Days[len(before.CompletionsLast7Days)-1].Count

	pool, err := storage.OpenDatabase()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer storage.CloseDatabase(pool)

	ctx := context.Background()
	var taskID int
	err = pool.QueryRow(ctx, `
		INSERT INTO tasks (title, user_id, completed, time_stamp, date_modified)
		VALUES ('count-once', 1, true, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
		RETURNING id`).Scan(&taskID)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM task_events WHERE task_id = $1", taskID)
		_, _ = pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)
	})

	_, err = pool.Exec(ctx,
		`INSERT INTO task_events (task_id, user_id, event_type, created_at) VALUES ($1, 1, 'completed', NOW() AT TIME ZONE 'UTC')`,
		taskID,
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}

	after, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats after completion: %v", err)
	}
	todayAfter := after.CompletionsLast7Days[len(after.CompletionsLast7Days)-1].Count

	if after.CompletedThisWeek != weekBefore+1 {
		t.Fatalf("week completions: want %d, got %d", weekBefore+1, after.CompletedThisWeek)
	}
	if todayAfter != todayBefore+1 {
		t.Fatalf("today chart completions: want %d, got %d", todayBefore+1, todayAfter)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO task_events (task_id, user_id, event_type, created_at) VALUES ($1, 1, 'completed', NOW() AT TIME ZONE 'UTC')`,
		taskID,
	)
	if err != nil {
		t.Fatalf("insert duplicate completion event: %v", err)
	}

	afterToggle, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats after duplicate event: %v", err)
	}
	if afterToggle.CompletedThisWeek != after.CompletedThisWeek {
		t.Fatalf("week completions after toggle spam: want %d, got %d", after.CompletedThisWeek, afterToggle.CompletedThisWeek)
	}
	todayAfterToggle := afterToggle.CompletionsLast7Days[len(afterToggle.CompletionsLast7Days)-1].Count
	if todayAfterToggle != todayAfter {
		t.Fatalf("today chart after toggle spam: want %d, got %d", todayAfter, todayAfterToggle)
	}
}

func TestGetDashboardStatsCountsOverdueChildOfCompletedParent(t *testing.T) {
	before, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats baseline: %v", err)
	}

	pool, err := storage.OpenDatabase()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer storage.CloseDatabase(pool)

	ctx := context.Background()
	var parentID, childID int
	err = pool.QueryRow(ctx,
		`INSERT INTO tasks (title, user_id, completed, due_date)
		 VALUES ('Stats completed parent', 1, true, ((NOW() AT TIME ZONE 'America/New_York')::date - 4))
		 RETURNING id`).Scan(&parentID)
	if err != nil {
		t.Fatalf("insert parent: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO tasks (title, user_id, completed, parent_id, due_date)
		 VALUES ('Stats overdue child', 1, false, $1, ((NOW() AT TIME ZONE 'America/New_York')::date - 2))
		 RETURNING id`, parentID).Scan(&childID)
	if err != nil {
		t.Fatalf("insert child: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM tasks WHERE id IN ($1, $2)", childID, parentID)
	})

	after, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats after child: %v", err)
	}
	if after.OverdueCount != before.OverdueCount+1 {
		t.Fatalf("overdue_count: want %d, got %d", before.OverdueCount+1, after.OverdueCount)
	}
}

func TestGetDashboardStatsExcludesUnclaimedKanbanOverdue(t *testing.T) {
	before, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats baseline: %v", err)
	}
	digestBefore, err := tasks.GetOverdueCount(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetOverdueCount baseline: %v", err)
	}

	pool, err := storage.OpenDatabase()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer storage.CloseDatabase(pool)

	ctx := context.Background()
	var projectID, taskID int
	err = pool.QueryRow(ctx,
		`INSERT INTO projects (id, user_id, name, workflow_mode)
		 VALUES (90, 1, 'Kanban overdue board', 'kanban')
		 RETURNING id`).Scan(&projectID)
	if err != nil {
		t.Fatalf("insert kanban project: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO tasks (title, user_id, completed, project_id, due_date, claimed_by)
		 VALUES ('Unclaimed kanban overdue', 1, false, $1, ((NOW() AT TIME ZONE 'America/New_York')::date - 1), NULL)
		 RETURNING id`, projectID).Scan(&taskID)
	if err != nil {
		t.Fatalf("insert unclaimed kanban task: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)
		_, _ = pool.Exec(ctx, "DELETE FROM projects WHERE id = $1", projectID)
	})

	after, err := tasks.GetDashboardStats(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetDashboardStats after unclaimed kanban: %v", err)
	}
	if after.OverdueCount != before.OverdueCount {
		t.Fatalf("dashboard overdue_count should ignore unclaimed kanban, want %d got %d", before.OverdueCount, after.OverdueCount)
	}

	digestAfter, err := tasks.GetOverdueCount(1, "America/New_York")
	if err != nil {
		t.Fatalf("GetOverdueCount after unclaimed kanban: %v", err)
	}
	if digestAfter != digestBefore+1 {
		t.Fatalf("digest overdue count should still include unclaimed kanban, want %d got %d", digestBefore+1, digestAfter)
	}
}
