package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCreateProjectValidation(t *testing.T) {
	ctx := context.Background()
	_, err := CreateProject(ctx, 1, "", "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: err=%v", err)
	}
	_, err = CreateProject(ctx, 1, strings.Repeat("x", MaxProjectNameLength+1), "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("long name: err=%v", err)
	}
	_, err = CreateProject(ctx, 1, "ok", strings.Repeat("d", MaxProjectDescriptionLength+1))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("long description: err=%v", err)
	}
}

func TestRenameProjectValidation(t *testing.T) {
	ctx := context.Background()
	_, err := RenameProject(ctx, 1, 1, "  ")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("blank name: err=%v", err)
	}
}

func TestUpdateProjectDescriptionValidation(t *testing.T) {
	ctx := context.Background()
	long := strings.Repeat("d", MaxProjectDescriptionLength+1)
	_, err := UpdateProject(ctx, 1, 1, nil, &long, nil)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("long description: err=%v", err)
	}
}

func TestReorderProjectsForUserValidation(t *testing.T) {
	ctx := context.Background()
	err := ReorderProjectsForUser(ctx, 1, nil)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("empty ids: err=%v", err)
	}
}

func TestCreateTagValidation(t *testing.T) {
	ctx := context.Background()
	_, err := CreateTag(ctx, 1, "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: err=%v", err)
	}
	_, err = CreateTag(ctx, 1, strings.Repeat("t", MaxTagNameLength+1))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("long name: err=%v", err)
	}
}

func TestRenameTagValidation(t *testing.T) {
	ctx := context.Background()
	_, err := RenameTag(ctx, 1, 1, "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: err=%v", err)
	}
}
