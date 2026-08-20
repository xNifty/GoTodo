package domain

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"GoTodo/internal/live"
	"GoTodo/internal/storage"
)

const maxCommentPreview = 120

var commentTaskRefPattern = regexp.MustCompile(`\[\[(\d+)\]\]|#(\d+)\b`)

func requireProjectTaskAccess(taskID, userID int) (projectID int, err error) {
	canRead, _, projectID, err := storage.CanUserAccessTask(taskID, userID)
	if err != nil {
		return 0, err
	}
	if !canRead {
		return 0, ErrNotFound
	}
	if projectID <= 0 {
		return 0, fmt.Errorf("%w: discussion is only available on project tasks", ErrValidation)
	}
	return projectID, nil
}

// ParseCommentTaskIDs returns unique task IDs referenced as [[123]] or #123.
func ParseCommentTaskIDs(body string) []int {
	matches := commentTaskRefPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(matches))
	ids := make([]int, 0, len(matches))
	for _, m := range matches {
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// ResolveCommentTaskLinks returns titles for referenced tasks the user can access.
func ResolveCommentTaskLinks(userID int, body string) []storage.TaskCommentLink {
	ids := ParseCommentTaskIDs(body)
	if len(ids) == 0 {
		return nil
	}
	out := make([]storage.TaskCommentLink, 0, len(ids))
	for _, id := range ids {
		canRead, _, _, err := storage.CanUserAccessTask(id, userID)
		if err != nil || !canRead {
			continue
		}
		td, err := storage.GetTaskTitleDescription(id)
		if err != nil || td == nil {
			continue
		}
		title := strings.TrimSpace(td.Title)
		if title == "" {
			title = fmt.Sprintf("Task #%d", id)
		}
		out = append(out, storage.TaskCommentLink{ID: id, Title: title})
	}
	return out
}

func attachCommentLinks(userID int, comments []storage.TaskComment) {
	for i := range comments {
		if comments[i].DeletedAt != nil || comments[i].Body == "" {
			continue
		}
		comments[i].Links = ResolveCommentTaskLinks(userID, comments[i].Body)
	}
}

// ListCommentsForUser lists discussion posts on a project task the user can read.
func ListCommentsForUser(ctx context.Context, userID, taskID int) ([]storage.TaskComment, error) {
	_ = ctx
	if _, err := requireProjectTaskAccess(taskID, userID); err != nil {
		return nil, err
	}
	comments, err := storage.ListTaskComments(taskID)
	if err != nil {
		return nil, err
	}
	if comments == nil {
		comments = []storage.TaskComment{}
	}
	attachCommentLinks(userID, comments)
	return comments, nil
}

// AddCommentForUser posts a discussion message. Any project member (including viewers) may post.
func AddCommentForUser(ctx context.Context, userID, taskID int, body string) (*storage.TaskComment, error) {
	_ = ctx
	projectID, err := requireProjectTaskAccess(taskID, userID)
	if err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("%w: comment cannot be empty", ErrValidation)
	}
	if len(body) > storage.MaxTaskCommentBody {
		return nil, fmt.Errorf("%w: comment must be %d characters or less", ErrValidation, storage.MaxTaskCommentBody)
	}
	comment, err := storage.CreateTaskComment(taskID, userID, body)
	if err != nil {
		return nil, err
	}
	comment.Links = ResolveCommentTaskLinks(userID, comment.Body)
	NotifyProjectMembersTaskCommented(taskID, userID, projectID, body)
	live.AfterTaskChange(userID, taskID, live.TypeTaskCommented)
	return comment, nil
}

// DeleteCommentForUser soft-deletes a comment. Authors delete as "user";
// project owners deleting someone else's comment delete as "owner".
func DeleteCommentForUser(ctx context.Context, userID, taskID, commentID int) error {
	_ = ctx
	projectID, err := requireProjectTaskAccess(taskID, userID)
	if err != nil {
		return err
	}
	comment, err := storage.GetTaskComment(commentID)
	if err != nil || comment.TaskID != taskID {
		return ErrNotFound
	}
	if comment.DeletedAt != nil {
		return fmt.Errorf("%w: comment already deleted", ErrConflict)
	}

	isAuthor := comment.UserID == userID
	role, _ := storage.GetProjectRole(projectID, userID)
	isOwner := storage.RoleCanManage(role)

	if !isAuthor && !isOwner {
		return ErrForbidden
	}

	kind := storage.CommentDeletedByUser
	if !isAuthor && isOwner {
		kind = storage.CommentDeletedByOwner
	}
	if err := storage.SoftDeleteTaskComment(commentID, userID, kind); err != nil {
		return err
	}
	live.AfterTaskChange(userID, taskID, live.TypeTaskCommented)
	return nil
}

func commentPreview(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return "New comment"
	}
	if len(body) <= maxCommentPreview {
		return body
	}
	return body[:maxCommentPreview-3] + "..."
}
