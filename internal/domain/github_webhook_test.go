package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"GoTodo/internal/storage"
)

type ghWebhookFixture struct {
	projectID   int
	taskID      int
	owner       string
	repo        string
	issueID     int64
	issueNumber int
	secret      string
	todoID      int
	progressID  int
	doneID      int
}

func setupGitHubWebhookFixture(t *testing.T, suffix string) ghWebhookFixture {
	t.Helper()
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "GH Sync "+suffix, "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := SetProjectWorkflowMode(ctx, 1, proj.ID, storage.WorkflowKanban); err != nil {
		t.Fatalf("enable kanban: %v", err)
	}
	statuses, err := ListProjectStatusesForUser(ctx, 1, proj.ID)
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	var todoID, progressID, doneID int
	for _, s := range statuses {
		switch {
		case s.IsDefault:
			todoID = s.ID
		case s.IsDone:
			doneID = s.ID
		case s.Name == "In Progress":
			progressID = s.ID
		}
	}
	if todoID == 0 || progressID == 0 || doneID == 0 {
		t.Fatalf("missing default statuses: %+v", statuses)
	}

	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{
		Title:     "Linked issue",
		ProjectID: &pid,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	owner := "acme-" + suffix
	repo := "board-" + suffix
	secret := "webhook-secret-" + suffix
	if _, err := storage.UpsertProjectGitHubRepo(proj.ID, 1, owner, repo, 1000, secret); err != nil {
		t.Fatalf("link repo: %v", err)
	}

	issueID := time.Now().UnixNano()
	issueNumber := int(issueID % 100000)
	if issueNumber <= 0 {
		issueNumber = 1
	}
	if _, err := storage.UpsertTaskGitHubIssue(
		taskID, issueID, issueNumber,
		fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber),
		"open", "Linked issue",
	); err != nil {
		t.Fatalf("link issue: %v", err)
	}

	return ghWebhookFixture{
		projectID:   proj.ID,
		taskID:      taskID,
		owner:       owner,
		repo:        repo,
		issueID:     issueID,
		issueNumber: issueNumber,
		secret:      secret,
		todoID:      todoID,
		progressID:  progressID,
		doneID:      doneID,
	}
}

func setTaskStatus(t *testing.T, taskID, statusID int) {
	t.Helper()
	sid := statusID
	statusPtr := &sid
	if _, err := UpdateTask(context.Background(), 1, taskID, UpdateTaskInput{StatusID: &statusPtr}); err != nil {
		t.Fatalf("set status: %v", err)
	}
}

func applyIssueWebhook(t *testing.T, f ghWebhookFixture, state, title string) {
	t.Helper()
	err := ApplyGitHubIssueWebhookState(
		context.Background(),
		f.owner, f.repo, f.issueID, f.issueNumber,
		state, title,
		fmt.Sprintf("https://github.com/%s/%s/issues/%d", f.owner, f.repo, f.issueNumber),
		f.secret,
	)
	if err != nil {
		t.Fatalf("webhook %s: %v", state, err)
	}
}

func TestGitHubWebhookSkipsStatusWhenNotDefaultOrDone(t *testing.T) {
	f := setupGitHubWebhookFixture(t, "in-progress")
	setTaskStatus(t, f.taskID, f.progressID)

	applyIssueWebhook(t, f, "open", "Still in progress")
	st, err := storage.GetTaskProjectStatus(f.taskID)
	if err != nil {
		t.Fatalf("status after open: %v", err)
	}
	if st == nil || st.ID != f.progressID {
		t.Fatalf("status after open webhook: %+v want in-progress %d", st, f.progressID)
	}
	issue, err := storage.GetTaskGitHubIssue(f.taskID)
	if err != nil || issue == nil {
		t.Fatalf("issue after open: %v %+v", err, issue)
	}
	if issue.IssueState != "open" {
		t.Fatalf("issue_state=%q want open", issue.IssueState)
	}
	if issue.IssueTitle != "Still in progress" {
		t.Fatalf("issue_title=%q", issue.IssueTitle)
	}

	applyIssueWebhook(t, f, "closed", "Closed on GitHub")
	st, err = storage.GetTaskProjectStatus(f.taskID)
	if err != nil {
		t.Fatalf("status after closed: %v", err)
	}
	if st == nil || st.ID != f.progressID {
		t.Fatalf("status after closed webhook: %+v want in-progress %d", st, f.progressID)
	}
	issue, err = storage.GetTaskGitHubIssue(f.taskID)
	if err != nil || issue == nil {
		t.Fatalf("issue after closed: %v %+v", err, issue)
	}
	if issue.IssueState != "closed" {
		t.Fatalf("issue_state=%q want closed", issue.IssueState)
	}
	if issue.IssueTitle != "Closed on GitHub" {
		t.Fatalf("issue_title=%q", issue.IssueTitle)
	}
}

func TestGitHubWebhookSyncsStatusFromDefaultAndDone(t *testing.T) {
	t.Run("default closed moves to done", func(t *testing.T) {
		f := setupGitHubWebhookFixture(t, "from-default")
		st, err := storage.GetTaskProjectStatus(f.taskID)
		if err != nil || st == nil || st.ID != f.todoID {
			t.Fatalf("expected default status, got %+v err=%v", st, err)
		}

		applyIssueWebhook(t, f, "closed", "Finished")
		st, err = storage.GetTaskProjectStatus(f.taskID)
		if err != nil {
			t.Fatalf("status after closed: %v", err)
		}
		if st == nil || st.ID != f.doneID {
			t.Fatalf("status after closed webhook: %+v want done %d", st, f.doneID)
		}
		issue, err := storage.GetTaskGitHubIssue(f.taskID)
		if err != nil || issue == nil {
			t.Fatalf("issue: %v %+v", err, issue)
		}
		if issue.IssueState != "closed" {
			t.Fatalf("issue_state=%q want closed", issue.IssueState)
		}
	})

	t.Run("done open moves to default", func(t *testing.T) {
		f := setupGitHubWebhookFixture(t, "from-done")
		setTaskStatus(t, f.taskID, f.doneID)

		applyIssueWebhook(t, f, "open", "Reopened")
		st, err := storage.GetTaskProjectStatus(f.taskID)
		if err != nil {
			t.Fatalf("status after open: %v", err)
		}
		if st == nil || st.ID != f.todoID {
			t.Fatalf("status after open webhook: %+v want default %d", st, f.todoID)
		}
		issue, err := storage.GetTaskGitHubIssue(f.taskID)
		if err != nil || issue == nil {
			t.Fatalf("issue: %v %+v", err, issue)
		}
		if issue.IssueState != "open" {
			t.Fatalf("issue_state=%q want open", issue.IssueState)
		}
	})
}
