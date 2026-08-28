package domain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
)

func TestRemovedTagIsProtected(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive Protect Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID

	if _, err := CreateTag(ctx, 1, "removed", &pid); !errors.Is(err, ErrValidation) {
		t.Fatalf("create reserved: err=%v want validation", err)
	}

	tags, err := ListTags(ctx, 1, &pid)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var removed *storage.Tag
	for i := range tags {
		if storage.IsRemovedTagName(tags[i].Name) {
			removed = &tags[i]
			break
		}
	}
	if removed == nil || !removed.Protected {
		t.Fatalf("expected protected removed tag, got %+v", tags)
	}
	if _, err := RenameTag(ctx, 1, removed.ID, "gone"); !errors.Is(err, ErrValidation) {
		t.Fatalf("rename protected: err=%v want validation", err)
	}
	color := "#dc3545"
	if _, err := UpdateTag(ctx, 1, removed.ID, nil, &color); !errors.Is(err, ErrValidation) {
		t.Fatalf("recolor protected: err=%v want validation", err)
	}
	if err := DeleteTag(ctx, 1, removed.ID); !errors.Is(err, ErrValidation) {
		t.Fatalf("delete protected: err=%v want validation", err)
	}
}

func TestCannotManuallyAssignRemovedTag(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive Assign Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	tags, err := ListTags(ctx, 1, &pid)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var removedID int
	for _, tg := range tags {
		if storage.IsRemovedTagName(tg.Name) {
			removedID = tg.ID
			break
		}
	}
	if removedID == 0 {
		t.Fatal("missing removed tag")
	}
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Assign archive", ProjectID: &pid, TagIDs: []int{removedID}})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("create with removed tag_ids: err=%v want validation", err)
	}
	_ = taskID
	if _, err := storage.GetOrCreateTagByName(1, &pid, "removed"); err == nil {
		t.Fatal("GetOrCreateTagByName should reject reserved name")
	}
}

func TestArchiveRestoreAndListFilter(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive List Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	work, err := CreateTag(ctx, 1, "work", &pid)
	if err != nil {
		t.Fatalf("work tag: %v", err)
	}
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Active then archive", ProjectID: &pid, TagIDs: []int{work.ID}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	uid := 1
	listed, total, err := tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, "UTC", tasks.ListFilters{ProjectFilter: &pid})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !containsTaskID(listed, taskID) {
		t.Fatalf("active task missing from default list, total=%d", total)
	}

	if err := ArchiveTask(ctx, 1, taskID); err != nil {
		t.Fatalf("archive: %v", err)
	}
	listed, _, err = tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, "UTC", tasks.ListFilters{ProjectFilter: &pid})
	if err != nil {
		t.Fatalf("list after archive: %v", err)
	}
	if containsTaskID(listed, taskID) {
		t.Fatal("archived task should be hidden from default list")
	}

	removed, err := storage.FindTagByName(1, &pid, storage.RemovedTagName)
	if err != nil {
		t.Fatalf("find removed: %v", err)
	}
	listed, _, err = tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, "UTC", tasks.ListFilters{ProjectFilter: &pid, TagFilter: &removed.ID})
	if err != nil {
		t.Fatalf("list by id: %v", err)
	}
	if !containsTaskID(listed, taskID) {
		t.Fatal("filter by removed id should show archived task")
	}
	listed, _, err = tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, "UTC", tasks.ListFilters{ProjectFilter: &pid, TagNameFilter: "removed"})
	if err != nil {
		t.Fatalf("list by name: %v", err)
	}
	if !containsTaskID(listed, taskID) {
		t.Fatal("filter by removed name should show archived task")
	}

	got, err := storage.GetTagsForTask(taskID)
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if !taskHasRemoved(got) {
		t.Fatalf("expected removed tag after archive, got %+v", got)
	}

	ids := []int{work.ID}
	if _, err := UpdateTask(ctx, 1, taskID, UpdateTaskInput{TagIDs: &ids}); err != nil {
		t.Fatalf("update tags: %v", err)
	}
	got, err = storage.GetTagsForTask(taskID)
	if err != nil {
		t.Fatalf("tags after save: %v", err)
	}
	if !taskHasRemoved(got) {
		t.Fatal("saving other tags must keep removed")
	}

	if err := RestoreTask(ctx, 1, taskID); err != nil {
		t.Fatalf("restore: %v", err)
	}
	listed, _, err = tasks.ReturnPaginationForUserWithFilters(1, 50, &uid, "UTC", tasks.ListFilters{ProjectFilter: &pid})
	if err != nil {
		t.Fatalf("list after restore: %v", err)
	}
	if !containsTaskID(listed, taskID) {
		t.Fatal("restored task should reappear")
	}
}

func TestArchiveCascadesToChildren(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive Cascade Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	parentID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Parent archive", ProjectID: &pid})
	if err != nil {
		t.Fatalf("parent: %v", err)
	}
	childID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Child archive", ProjectID: &pid, ParentID: &parentID})
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	if err := ArchiveTask(ctx, 1, parentID); err != nil {
		t.Fatalf("archive: %v", err)
	}
	for _, id := range []int{parentID, childID} {
		got, err := storage.GetTagsForTask(id)
		if err != nil {
			t.Fatalf("tags %d: %v", id, err)
		}
		if !taskHasRemoved(got) {
			t.Fatalf("task %d should be archived", id)
		}
	}
	if err := RestoreTask(ctx, 1, parentID); err != nil {
		t.Fatalf("restore: %v", err)
	}
	for _, id := range []int{parentID, childID} {
		got, err := storage.GetTagsForTask(id)
		if err != nil {
			t.Fatalf("tags %d: %v", id, err)
		}
		if taskHasRemoved(got) {
			t.Fatalf("task %d should be restored", id)
		}
	}
}

func TestArchiveForbiddenForViewer(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive Viewer Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	if err := storage.UpsertProjectMember(pid, 3, storage.RoleViewer); err != nil {
		t.Fatalf("add viewer: %v", err)
	}
	if err := storage.UpsertProjectMember(pid, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Viewer archive", ProjectID: &pid})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := ArchiveTask(ctx, 3, taskID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer archive: err=%v want forbidden", err)
	}
	if err := ArchiveTask(ctx, 2, taskID); err != nil {
		t.Fatalf("editor archive: %v", err)
	}
	if err := RestoreTask(ctx, 3, taskID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer restore: err=%v want forbidden", err)
	}
	if err := RestoreTask(ctx, 2, taskID); err != nil {
		t.Fatalf("editor restore: %v", err)
	}
}

func TestArchiveDoesNotCountAgainstTagLimit(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Archive Max Tags Proj", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	var ids []int
	for _, name := range []string{"one", "two", "three", "four", "five"} {
		tg, err := CreateTag(ctx, 1, name, &pid)
		if err != nil {
			t.Fatalf("tag %s: %v", name, err)
		}
		ids = append(ids, tg.ID)
	}
	taskID, err := CreateTask(ctx, 1, CreateTaskInput{Title: "Full tags", ProjectID: &pid, TagIDs: ids})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := ArchiveTask(ctx, 1, taskID); err != nil {
		t.Fatalf("archive with 5 user tags: %v", err)
	}
	got, err := storage.GetTagsForTask(taskID)
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("expected 5 user tags + removed, got %d", len(got))
	}
}

func containsTaskID(list []tasks.Task, id int) bool {
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

func taskHasRemoved(tags []storage.Tag) bool {
	for _, tg := range tags {
		if storage.IsRemovedTagName(tg.Name) {
			return true
		}
	}
	return false
}

func TestRemovedTagNameHelper(t *testing.T) {
	if !storage.IsRemovedTagName("Removed") || storage.IsRemovedTagName("work") {
		t.Fatal("IsRemovedTagName mismatch")
	}
	if !strings.EqualFold(storage.RemovedTagName, "removed") {
		t.Fatalf("canonical name %q", storage.RemovedTagName)
	}
}
