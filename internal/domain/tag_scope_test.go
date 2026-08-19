package domain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"GoTodo/internal/storage"
)

func TestProjectTagsAreSharedAndScoped(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Tag Scope Proj", "")
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
	ownerTag, err := CreateTag(ctx, 1, "urgent", &pid)
	if err != nil {
		t.Fatalf("owner create: %v", err)
	}
	if ownerTag.ProjectID == nil || *ownerTag.ProjectID != pid {
		t.Fatalf("expected project tag, got %+v", ownerTag)
	}

	editorTag, err := CreateTag(ctx, 2, "urgent", &pid)
	if err != nil {
		t.Fatalf("editor get-or-create: %v", err)
	}
	if editorTag.ID != ownerTag.ID {
		t.Fatalf("editor should reuse project tag %d, got %d", ownerTag.ID, editorTag.ID)
	}

	_, err = CreateTag(ctx, 3, "blocked", &pid)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer create: err=%v want forbidden", err)
	}

	listed, err := ListTags(ctx, 3, &pid)
	if err != nil {
		t.Fatalf("viewer list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != ownerTag.ID {
		t.Fatalf("viewer should see project tags, got %+v", listed)
	}

	if err := DeleteTag(ctx, 3, ownerTag.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer delete: err=%v want forbidden", err)
	}

	personal, err := CreateTag(ctx, 2, "urgent", nil)
	if err != nil {
		t.Fatalf("editor personal tag: %v", err)
	}
	if personal.ProjectID != nil {
		t.Fatalf("personal tag should have nil project_id")
	}
	if personal.ID == ownerTag.ID {
		t.Fatal("personal and project tags must be distinct")
	}
}

func TestSetTaskTagsRejectsCrossNamespace(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Tag Assign Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	projectTag, err := CreateTag(ctx, 1, "proj-only", &pid)
	if err != nil {
		t.Fatalf("project tag: %v", err)
	}
	personalTag, err := CreateTag(ctx, 1, "inbox-only", nil)
	if err != nil {
		t.Fatalf("personal tag: %v", err)
	}

	inboxID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Inbox tagged", TagIDs: []int{personalTag.ID}})
	if err != nil {
		t.Fatalf("inbox task: %v", err)
	}
	if err := storage.SetTaskTags(inboxID, 1, []int{projectTag.ID}); err == nil {
		t.Fatal("expected invalid tag selection for project tag on inbox task")
	}

	projTaskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Project tagged", ProjectID: &pid, TagIDs: []int{projectTag.ID}})
	if err != nil {
		t.Fatalf("project task: %v", err)
	}
	if err := storage.SetTaskTags(projTaskID, 1, []int{personalTag.ID}); err == nil {
		t.Fatal("expected invalid tag selection for personal tag on project task")
	}
}

func TestMoveTaskRemapsTagsByName(t *testing.T) {
	ctx := context.Background()
	src, err := CreateProject(ctx, 1, "Tag Move Src", "")
	if err != nil {
		t.Fatalf("src project: %v", err)
	}
	dst, err := CreateProject(ctx, 1, "Tag Move Dst", "")
	if err != nil {
		t.Fatalf("dst project: %v", err)
	}
	srcID, dstID := src.ID, dst.ID
	srcTag, err := CreateTag(ctx, 1, "shared-name", &srcID)
	if err != nil {
		t.Fatalf("src tag: %v", err)
	}
	dstTag, err := CreateTag(ctx, 1, "shared-name", &dstID)
	if err != nil {
		t.Fatalf("dst tag: %v", err)
	}
	onlySrc, err := CreateTag(ctx, 1, "src-only", &srcID)
	if err != nil {
		t.Fatalf("src-only tag: %v", err)
	}

	taskID, err := CreateTask(ctx, 1, CreateTaskInput{
		Title:     "Moving",
		ProjectID: &srcID,
		TagIDs:    []int{srcTag.ID, onlySrc.ID},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	dstPtr := &dstID
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{ProjectID: &dstPtr}); err != nil {
		t.Fatalf("move: %v", err)
	}
	got, err := storage.GetTagsForTask(taskID)
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(got) != 1 || got[0].ID != dstTag.ID {
		t.Fatalf("after move want dest shared-name tag %d, got %+v", dstTag.ID, got)
	}
}

func TestListTagsQueryScopes(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Tag List Proj", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	pid := proj.ID
	if _, err := CreateTag(ctx, 1, "personal-a", nil); err != nil {
		t.Fatalf("personal: %v", err)
	}
	if _, err := CreateTag(ctx, 1, "project-a", &pid); err != nil {
		t.Fatalf("project: %v", err)
	}

	zero := 0
	personal, err := ListTags(ctx, 1, &zero)
	if err != nil {
		t.Fatalf("personal list: %v", err)
	}
	for _, tg := range personal {
		if tg.ProjectID != nil {
			t.Fatalf("personal list included project tag %+v", tg)
		}
	}

	projectTags, err := ListTags(ctx, 1, &pid)
	if err != nil {
		t.Fatalf("project list: %v", err)
	}
	if len(projectTags) != 1 || !strings.EqualFold(projectTags[0].Name, "project-a") {
		t.Fatalf("project list: %+v", projectTags)
	}

	all, err := ListTags(ctx, 1, nil)
	if err != nil {
		t.Fatalf("all list: %v", err)
	}
	if len(all) < 2 {
		t.Fatalf("all list should include personal and project tags, got %d", len(all))
	}
}

func TestProjectTagShareLinkOwnerOnly(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Tag Share Proj", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	pid := proj.ID
	tag, err := CreateTag(ctx, 1, "share-me", &pid)
	if err != nil {
		t.Fatalf("tag: %v", err)
	}

	_, err = CreateShareLinkForScope(ctx, 2, storage.ShareScopeTag, tag.ID, nil)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor share: err=%v want forbidden", err)
	}

	link, err := CreateShareLinkForScope(ctx, 1, storage.ShareScopeTag, tag.ID, nil)
	if err != nil {
		t.Fatalf("owner share: %v", err)
	}
	if link == nil || link.ScopeID != tag.ID {
		t.Fatalf("unexpected link %+v", link)
	}
}
