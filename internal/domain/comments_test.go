package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"GoTodo/internal/storage"
)

func TestCommentsViewerCanPostNonMemberCannot(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Discuss Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 3, storage.RoleViewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Talk about this", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	c, err := AddCommentForUser(ctx, 3, taskID, "Viewer note")
	if err != nil {
		t.Fatalf("viewer post: %v", err)
	}
	if c.Body != "Viewer note" || c.UserID != 3 {
		t.Fatalf("comment %+v", c)
	}

	if _, err := AddCommentForUser(ctx, 2, taskID, "Outsider"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("non-member err=%v, want not found", err)
	}

	personalID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Inbox only"})
	if err != nil {
		t.Fatalf("personal: %v", err)
	}
	if _, err := AddCommentForUser(ctx, 1, personalID, "Nope"); !errors.Is(err, ErrValidation) {
		t.Fatalf("personal task err=%v, want validation", err)
	}
}

func TestCommentsSoftDeleteAuthorAndOwner(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Delete Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 3, storage.RoleViewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Delete thread", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	own, err := AddCommentForUser(ctx, 2, taskID, "Editor says hi")
	if err != nil {
		t.Fatalf("editor post: %v", err)
	}
	if err := DeleteCommentForUser(ctx, 2, taskID, own.ID); err != nil {
		t.Fatalf("author delete: %v", err)
	}
	listed, err := ListCommentsForUser(ctx, 1, taskID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 || listed[0].DeletedAt == nil || listed[0].Body != "" {
		t.Fatalf("author tombstone %+v", listed)
	}
	if listed[0].DeletedByKind != storage.CommentDeletedByUser {
		t.Fatalf("kind=%q want user", listed[0].DeletedByKind)
	}

	other, err := AddCommentForUser(ctx, 3, taskID, "Viewer says hi")
	if err != nil {
		t.Fatalf("viewer post: %v", err)
	}
	if err := DeleteCommentForUser(ctx, 2, taskID, other.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor deleting other err=%v, want forbidden", err)
	}
	if err := DeleteCommentForUser(ctx, 1, taskID, other.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
	listed, err = ListCommentsForUser(ctx, 1, taskID)
	if err != nil {
		t.Fatalf("list2: %v", err)
	}
	var found *storage.TaskComment
	for i := range listed {
		if listed[i].ID == other.ID {
			found = &listed[i]
			break
		}
	}
	if found == nil || found.DeletedAt == nil || found.Body != "" {
		t.Fatalf("owner tombstone %+v", found)
	}
	if found.DeletedByKind != storage.CommentDeletedByOwner {
		t.Fatalf("kind=%q want owner", found.DeletedByKind)
	}

	ownerOwn, err := AddCommentForUser(ctx, 1, taskID, "Owner own")
	if err != nil {
		t.Fatalf("owner post: %v", err)
	}
	if err := DeleteCommentForUser(ctx, 1, taskID, ownerOwn.ID); err != nil {
		t.Fatalf("owner self-delete: %v", err)
	}
	listed, err = ListCommentsForUser(ctx, 1, taskID)
	if err != nil {
		t.Fatalf("list3: %v", err)
	}
	for _, c := range listed {
		if c.ID == ownerOwn.ID && c.DeletedByKind != storage.CommentDeletedByUser {
			t.Fatalf("owner self-delete kind=%q want user", c.DeletedByKind)
		}
	}
}

func TestCommentsNotifyOtherMembers(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Notify Comments", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 3, storage.RoleViewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Need eyes", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := AddCommentForUser(ctx, 1, taskID, "Please review"); err != nil {
		t.Fatalf("post: %v", err)
	}

	forOwner, _, err := storage.ListUserNotifications(1, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range forOwner {
		if n.TaskID == taskID && n.Type == storage.NotificationTaskCommented {
			t.Fatal("actor should not receive own comment notification")
		}
	}

	foundEditor := false
	forEditor, _, err := storage.ListUserNotifications(2, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range forEditor {
		if n.TaskID == taskID && n.Type == storage.NotificationTaskCommented {
			foundEditor = true
			if n.Body != "Please review" {
				t.Fatalf("body=%q", n.Body)
			}
		}
	}
	if !foundEditor {
		t.Fatal("editor should receive task_commented")
	}

	foundViewer := false
	forViewer, _, err := storage.ListUserNotifications(3, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range forViewer {
		if n.TaskID == taskID && n.Type == storage.NotificationTaskCommented {
			foundViewer = true
		}
	}
	if !foundViewer {
		t.Fatal("viewer should receive task_commented")
	}
}

func TestParseCommentMentions(t *testing.T) {
	names := ParseCommentMentions("Hey @Bob_Editor and @alice, also ping @Bob_Editor again")
	if len(names) != 2 || names[0] != "Bob_Editor" || names[1] != "alice" {
		t.Fatalf("names=%v", names)
	}
	if ParseCommentMentions("email me@host.com please") != nil {
		t.Fatal("email should not be a mention")
	}
	if ParseCommentMentions("@@alice") != nil {
		t.Fatal("double-at should not mention")
	}
	got := ParseCommentMentions("(@carol_viewer) and #12")
	if len(got) != 1 || got[0] != "carol_viewer" {
		t.Fatalf("paren mention=%v", got)
	}
	if ParseCommentMentions("hi @ab") != nil {
		t.Fatal("short token is not a username")
	}
}

func TestCommentsMentionNotifiesMemberOnce(t *testing.T) {
	ctx := context.Background()
	if err := storage.SetUsername(1, "alice", false); err != nil {
		t.Fatalf("set alice: %v", err)
	}
	if err := storage.SetUsername(2, "bob_editor", false); err != nil {
		t.Fatalf("set bob: %v", err)
	}
	if err := storage.SetUsername(3, "carol_viewer", false); err != nil {
		t.Fatalf("set carol: %v", err)
	}

	proj, err := CreateProject(ctx, 1, "Mention Proj", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 3, storage.RoleViewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Need Bob", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := AddCommentForUser(ctx, 1, taskID, "Please look @bob_editor"); err != nil {
		t.Fatalf("post: %v", err)
	}

	forEditor, _, err := storage.ListUserNotifications(2, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	foundMention := false
	for _, n := range forEditor {
		if n.TaskID != taskID {
			continue
		}
		if n.Type == storage.NotificationTaskCommented {
			t.Fatal("mentioned member should not also get task_commented")
		}
		if n.Type == storage.NotificationTaskMentioned {
			foundMention = true
			if !strings.Contains(n.Title, "mentioned") {
				t.Fatalf("title=%q", n.Title)
			}
			if n.Body != "Please look @bob_editor" {
				t.Fatalf("body=%q", n.Body)
			}
		}
	}
	if !foundMention {
		t.Fatal("editor should receive task_mentioned")
	}

	foundViewerComment := false
	forViewer, _, err := storage.ListUserNotifications(3, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range forViewer {
		if n.TaskID == taskID && n.Type == storage.NotificationTaskMentioned {
			t.Fatal("unmentioned viewer should not receive task_mentioned")
		}
		if n.TaskID == taskID && n.Type == storage.NotificationTaskCommented {
			foundViewerComment = true
		}
	}
	if !foundViewerComment {
		t.Fatal("unmentioned viewer should still receive task_commented")
	}

	forOwner, _, err := storage.ListUserNotifications(1, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range forOwner {
		if n.TaskID == taskID && (n.Type == storage.NotificationTaskCommented || n.Type == storage.NotificationTaskMentioned) {
			t.Fatal("actor should not receive a mention or comment notification")
		}
	}
}

func TestCommentsMentionIgnoresNonMembers(t *testing.T) {
	ctx := context.Background()
	if err := storage.SetUsername(1, "alice", false); err != nil {
		t.Fatalf("set alice: %v", err)
	}
	if err := storage.SetUsername(2, "bob_editor", false); err != nil {
		t.Fatalf("set bob: %v", err)
	}

	proj, err := CreateProject(ctx, 1, "No Ghost Mentions", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Ghost ping", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, err := AddCommentForUser(ctx, 1, taskID, "Calling @not_a_member"); err != nil {
		t.Fatalf("post: %v", err)
	}

	forEditor, _, err := storage.ListUserNotifications(2, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range forEditor {
		if n.TaskID == taskID && n.Type == storage.NotificationTaskMentioned {
			t.Fatal("non-member mention should not create task_mentioned")
		}
		if n.TaskID == taskID && n.Type == storage.NotificationTaskCommented {
			found = true
		}
	}
	if !found {
		t.Fatal("editor should still get task_commented")
	}
}

func TestSearchUsernamesProjectMembersOnly(t *testing.T) {
	ctx := context.Background()
	if err := storage.SetUsername(1, "alice", false); err != nil {
		t.Fatalf("set alice: %v", err)
	}
	if err := storage.SetUsername(2, "bob_editor", false); err != nil {
		t.Fatalf("set bob: %v", err)
	}
	if err := storage.SetUsername(3, "carol_viewer", false); err != nil {
		t.Fatalf("set carol: %v", err)
	}

	proj, err := CreateProject(ctx, 1, "Search Mentions", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}

	other, err := CreateProject(ctx, 3, "Other Search Proj", "")
	if err != nil {
		t.Fatalf("other project: %v", err)
	}

	hits, err := SearchUsernames(ctx, 1, "bo", proj.ID)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 || hits[0] != "bob_editor" {
		t.Fatalf("hits=%v, want [bob_editor]", hits)
	}

	carolHits, err := SearchUsernames(ctx, 1, "ca", proj.ID)
	if err != nil {
		t.Fatalf("carol search: %v", err)
	}
	if len(carolHits) != 0 {
		t.Fatalf("non-member carol should not appear, got %v", carolHits)
	}

	if _, err := SearchUsernames(ctx, 2, "bo", other.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("outsider search err=%v, want not found", err)
	}

	aliceHits, err := SearchUsernames(ctx, 2, "al", proj.ID)
	if err != nil {
		t.Fatalf("member searching owner: %v", err)
	}
	if len(aliceHits) != 1 || aliceHits[0] != "alice" {
		t.Fatalf("alice hits=%v", aliceHits)
	}
}

func TestParseCommentTaskIDs(t *testing.T) {
	ids := ParseCommentTaskIDs("Check #191 and also [[42]] plus #191 again")
	if len(ids) != 2 || ids[0] != 191 || ids[1] != 42 {
		t.Fatalf("ids=%v", ids)
	}
	if ParseCommentTaskIDs("no refs") != nil {
		t.Fatal("expected nil")
	}
}

func TestCommentTaskLinksRespectAccess(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Link Proj", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	pid := proj.ID
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Discussion home", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	personalID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Secret inbox item"})
	if err != nil {
		t.Fatalf("personal: %v", err)
	}

	body := fmt.Sprintf("Check #%d", personalID)
	if _, err := AddCommentForUser(ctx, 1, taskID, body); err != nil {
		t.Fatalf("post: %v", err)
	}

	forOwner, err := ListCommentsForUser(ctx, 1, taskID)
	if err != nil {
		t.Fatalf("owner list: %v", err)
	}
	if len(forOwner) != 1 || len(forOwner[0].Links) != 1 || forOwner[0].Links[0].ID != personalID {
		t.Fatalf("owner links %+v", forOwner)
	}
	if forOwner[0].Links[0].Title != "Secret inbox item" {
		t.Fatalf("title=%q", forOwner[0].Links[0].Title)
	}

	forEditor, err := ListCommentsForUser(ctx, 2, taskID)
	if err != nil {
		t.Fatalf("editor list: %v", err)
	}
	if len(forEditor) != 1 || len(forEditor[0].Links) != 0 {
		t.Fatalf("editor should not see inaccessible task title, got %+v", forEditor[0].Links)
	}
}
