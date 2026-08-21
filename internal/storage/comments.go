package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	MaxTaskCommentBody = 2000

	CommentDeletedByUser  = "user"
	CommentDeletedByOwner = "owner"
)

// TaskComment is a discussion post on a project task.
type TaskComment struct {
	ID              int
	TaskID          int
	UserID          int
	UserName        string
	Body            string
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID int
	DeletedByKind   string
	Links           []TaskCommentLink
}

// TaskCommentLink is a task mentioned in a comment that the viewer can open.
type TaskCommentLink struct {
	ID    int
	Title string
}

// CreateTaskCommentsTable creates the task discussion store.
func CreateTaskCommentsTable() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS task_comments (
			id SERIAL PRIMARY KEY,
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			body TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			deleted_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			deleted_by_kind TEXT CHECK (deleted_by_kind IS NULL OR deleted_by_kind IN ('user', 'owner'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_task_comments_task_created
			ON task_comments (task_id, created_at ASC, id ASC)`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(context.Background(), s); err != nil {
			return fmt.Errorf("failed to create task_comments: %v", err)
		}
	}
	return nil
}

func scanTaskComment(row interface {
	Scan(dest ...any) error
}, c *TaskComment) error {
	var deletedAt sql.NullTime
	var deletedBy sql.NullInt64
	var deletedKind sql.NullString
	err := row.Scan(
		&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt,
		&deletedAt, &deletedBy, &deletedKind, &c.UserName,
	)
	if err != nil {
		return err
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
	if deletedBy.Valid {
		c.DeletedByUserID = int(deletedBy.Int64)
	}
	if deletedKind.Valid {
		c.DeletedByKind = deletedKind.String
	}
	return nil
}

const taskCommentSelect = `SELECT c.id, c.task_id, COALESCE(c.user_id, 0), COALESCE(c.body, ''), c.created_at,
		c.deleted_at, c.deleted_by_user_id, c.deleted_by_kind,
		COALESCE(u.user_name, u.email, '')
	 FROM task_comments c
	 LEFT JOIN users u ON u.id = c.user_id`

// ListTaskComments returns discussion posts for a task, oldest first.
func ListTaskComments(taskID int) ([]TaskComment, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(),
		taskCommentSelect+` WHERE c.task_id = $1 ORDER BY c.created_at ASC, c.id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TaskComment
	for rows.Next() {
		var c TaskComment
		if err := scanTaskComment(rows, &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetTaskComment loads one comment by id.
func GetTaskComment(commentID int) (*TaskComment, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var c TaskComment
	err = scanTaskComment(pool.QueryRow(context.Background(),
		taskCommentSelect+` WHERE c.id = $1`, commentID), &c)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, err
	}
	return &c, nil
}

// CreateTaskComment inserts a discussion post and returns it with author name.
func CreateTaskComment(taskID, userID int, body string) (*TaskComment, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	body = strings.TrimSpace(body)
	var id int
	err = pool.QueryRow(context.Background(),
		`INSERT INTO task_comments (task_id, user_id, body)
		 VALUES ($1, $2, $3) RETURNING id`,
		taskID, userID, body).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetTaskComment(id)
}

// SoftDeleteTaskComment clears the body and records who deleted it.
func SoftDeleteTaskComment(commentID, deletedByUserID int, kind string) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if kind != CommentDeletedByUser && kind != CommentDeletedByOwner {
		return fmt.Errorf("invalid delete kind")
	}

	tag, err := pool.Exec(context.Background(),
		`UPDATE task_comments
		 SET body = '', deleted_at = NOW(), deleted_by_user_id = $1, deleted_by_kind = $2
		 WHERE id = $3 AND deleted_at IS NULL`,
		deletedByUserID, kind, commentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}
	return nil
}
