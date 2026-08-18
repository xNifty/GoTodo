package domain

import (
	"context"
	"testing"

	"GoTodo/internal/storage"
)

func TestListShareLinksForUserProjectMembersSeeOwnerLinks(t *testing.T) {
	ctx := context.Background()
	proj, err := CreateProject(ctx, 1, "Share Link Visibility", "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := storage.UpsertProjectMember(proj.ID, 2, storage.RoleEditor); err != nil {
		t.Fatalf("add editor: %v", err)
	}

	link, err := CreateShareLinkForScope(ctx, 1, storage.ShareScopeProject, proj.ID, nil)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	ownerLinks, err := ListShareLinksForUser(ctx, 1, storage.ShareScopeProject, proj.ID)
	if err != nil {
		t.Fatalf("owner list: %v", err)
	}
	if len(ownerLinks) == 0 {
		t.Fatal("owner should see share links")
	}

	editorLinks, err := ListShareLinksForUser(ctx, 2, storage.ShareScopeProject, proj.ID)
	if err != nil {
		t.Fatalf("editor list: %v", err)
	}
	found := false
	for _, l := range editorLinks {
		if l.ID == link.ID && l.Token == link.Token {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("editor should see owner's project share link, got %+v", editorLinks)
	}
}
