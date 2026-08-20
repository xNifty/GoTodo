package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxTagsPerTask = 5

const tagSelectCols = `id, user_id, project_id, name, COALESCE(color, '#6c757d')`

// Tag is a label in a personal (inbox) or project namespace.
type Tag struct {
	ID        int
	UserID    int
	ProjectID *int
	Name      string
	Color     string
}

var tagPalette = []string{
	"#6c757d", "#0d6efd", "#198754", "#dc3545",
	"#fd7e14", "#6610f2", "#20c997", "#d63384",
}

func tagColorForID(id int) string {
	if len(tagPalette) == 0 {
		return "#6c757d"
	}
	return tagPalette[(id-1)%len(tagPalette)]
}

func scanTag(scanner interface{ Scan(...any) error }) (Tag, error) {
	var t Tag
	var projectID sql.NullInt64
	err := scanner.Scan(&t.ID, &t.UserID, &projectID, &t.Name, &t.Color)
	if err != nil {
		return t, err
	}
	if projectID.Valid && projectID.Int64 > 0 {
		pid := int(projectID.Int64)
		t.ProjectID = &pid
	}
	return t, nil
}

func isNoRows(err error) bool {
	return err != nil && (errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows"))
}

// CreateTagsTables creates tags and task_tags tables.
func CreateTagsTables() error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			project_id INTEGER,
			name TEXT NOT NULL,
			color VARCHAR(7) DEFAULT '#6c757d',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS task_tags (
			task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (task_id, tag_id)
		);
		CREATE INDEX IF NOT EXISTS idx_task_tags_tag_id ON task_tags(tag_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags tables: %v", err)
	}
	return nil
}

// MigrateTagsAddProjectID adds project-scoped tags, splits legacy user-owned
// tags onto the projects (and inbox) that use them, and replaces the old
// UNIQUE(user_id, name) constraint.
func MigrateTagsAddProjectID() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	if _, err := pool.Exec(context.Background(),
		`ALTER TABLE tags ADD COLUMN IF NOT EXISTS project_id INTEGER`); err != nil {
		return fmt.Errorf("failed to add tags.project_id: %v", err)
	}

	// Cloning a personal tag onto a project inserts another row with the same
	// (user_id, name). The legacy UNIQUE(user_id, name) constraint must be
	// gone before that split, or PostgreSQL raises tags_user_id_name_key.
	if err := dropLegacyTagsUserNameUnique(pool); err != nil {
		return err
	}

	if err := migrateSplitLegacyTags(pool); err != nil {
		return err
	}
	if err := migrateMergeDuplicateTagNames(pool); err != nil {
		return err
	}

	if _, err := pool.Exec(context.Background(),
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_personal_lower_name
			ON tags (user_id, lower(name)) WHERE project_id IS NULL`); err != nil {
		return fmt.Errorf("failed to create personal tag unique index: %v", err)
	}
	if _, err := pool.Exec(context.Background(),
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_project_lower_name
			ON tags (project_id, lower(name)) WHERE project_id IS NOT NULL`); err != nil {
		return fmt.Errorf("failed to create project tag unique index: %v", err)
	}

	var fkExists bool
	if err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tags_project')`).Scan(&fkExists); err != nil {
		return err
	}
	if !fkExists && tableExists(pool, "projects") {
		if _, err := pool.Exec(context.Background(),
			`ALTER TABLE tags ADD CONSTRAINT fk_tags_project
			 FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE`); err != nil {
			return fmt.Errorf("failed to add tags.project_id FK: %v", err)
		}
	}
	return nil
}

func dropLegacyTagsUserNameUnique(pool *pgxpool.Pool) error {
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_user_id_name_key`); err != nil {
		return fmt.Errorf("failed to drop tags_user_id_name_key: %v", err)
	}
	// Some installs created this as a unique index rather than a table constraint.
	if _, err := pool.Exec(ctx, `DROP INDEX IF EXISTS tags_user_id_name_key`); err != nil {
		return fmt.Errorf("failed to drop tags_user_id_name_key index: %v", err)
	}
	return nil
}

func migrateSplitLegacyTags(pool *pgxpool.Pool) error {
	if !tableExists(pool, "tasks") || !tableExists(pool, "tags") {
		return nil
	}
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `
		DELETE FROM task_tags a
		WHERE EXISTS (
			SELECT 1
			FROM task_tags b
			JOIN tags ta ON ta.id = a.tag_id
			JOIN tags tb ON tb.id = b.tag_id
			WHERE a.task_id = b.task_id
			  AND lower(ta.name) = lower(tb.name)
			  AND a.tag_id > b.tag_id
		)`); err != nil {
		return fmt.Errorf("failed to collapse duplicate task tags: %v", err)
	}

	rows, err := pool.Query(ctx, `SELECT `+tagSelectCols+` FROM tags WHERE project_id IS NULL`)
	if err != nil {
		return fmt.Errorf("failed to list personal tags for migration: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		t, err := scanTag(rows)
		if err != nil {
			return err
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range tags {
		projRows, err := pool.Query(ctx, `
			SELECT DISTINCT t.project_id
			FROM task_tags tt
			JOIN tasks t ON t.id = tt.task_id
			WHERE tt.tag_id = $1 AND t.project_id IS NOT NULL
			ORDER BY t.project_id`, t.ID)
		if err != nil {
			return fmt.Errorf("failed to list tag projects: %v", err)
		}
		var projectIDs []int
		for projRows.Next() {
			var pid int
			if err := projRows.Scan(&pid); err != nil {
				projRows.Close()
				return err
			}
			projectIDs = append(projectIDs, pid)
		}
		projRows.Close()
		if err := projRows.Err(); err != nil {
			return err
		}

		var inboxUsed bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM task_tags tt
				JOIN tasks t ON t.id = tt.task_id
				WHERE tt.tag_id = $1 AND t.project_id IS NULL
			)`, t.ID).Scan(&inboxUsed); err != nil {
			return err
		}

		if len(projectIDs) == 0 {
			continue
		}
		if len(projectIDs) == 1 && !inboxUsed {
			if _, err := pool.Exec(ctx, `UPDATE tags SET project_id = $1 WHERE id = $2 AND project_id IS NULL`,
				projectIDs[0], t.ID); err != nil {
				return fmt.Errorf("failed to convert tag to project scope: %v", err)
			}
			continue
		}

		for _, pid := range projectIDs {
			destID, err := findOrCloneProjectTag(pool, t, pid)
			if err != nil {
				return err
			}
			if destID == t.ID {
				continue
			}
			if _, err := pool.Exec(ctx, `
				UPDATE task_tags tt
				SET tag_id = $1
				FROM tasks tsk
				WHERE tt.task_id = tsk.id
				  AND tt.tag_id = $2
				  AND tsk.project_id = $3
				  AND NOT EXISTS (
					SELECT 1 FROM task_tags existing
					WHERE existing.task_id = tt.task_id AND existing.tag_id = $1
				  )`, destID, t.ID, pid); err != nil {
				return fmt.Errorf("failed to re-point task tags: %v", err)
			}
			if _, err := pool.Exec(ctx, `
				DELETE FROM task_tags tt
				USING tasks tsk
				WHERE tt.task_id = tsk.id AND tt.tag_id = $1 AND tsk.project_id = $2`,
				t.ID, pid); err != nil {
				return err
			}
		}
	}
	return nil
}

func findOrCloneProjectTag(pool *pgxpool.Pool, src Tag, projectID int) (int, error) {
	ctx := context.Background()
	var existingID int
	err := pool.QueryRow(ctx, `
		SELECT id FROM tags
		WHERE project_id = $1 AND lower(name) = lower($2)
		ORDER BY id ASC LIMIT 1`, projectID, src.Name).Scan(&existingID)
	if err == nil {
		return existingID, nil
	}
	if !isNoRows(err) {
		return 0, err
	}

	var newID int
	err = pool.QueryRow(ctx, `
		INSERT INTO tags (user_id, project_id, name, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, src.UserID, projectID, src.Name, src.Color).Scan(&newID)
	if err != nil {
		if lookupErr := pool.QueryRow(ctx, `
			SELECT id FROM tags
			WHERE project_id = $1 AND lower(name) = lower($2)
			ORDER BY id ASC LIMIT 1`, projectID, src.Name).Scan(&existingID); lookupErr == nil {
			return existingID, nil
		}
		return 0, fmt.Errorf("failed to clone project tag: %v", err)
	}
	return newID, nil
}

func migrateMergeDuplicateTagNames(pool *pgxpool.Pool) error {
	ctx := context.Background()

	personalRows, err := pool.Query(ctx, `
		SELECT user_id, lower(name)
		FROM tags
		WHERE project_id IS NULL
		GROUP BY user_id, lower(name)
		HAVING COUNT(*) > 1`)
	if err != nil {
		return fmt.Errorf("failed to list personal tag duplicates: %v", err)
	}
	type pair struct {
		scopeID int
		name    string
	}
	var personal []pair
	for personalRows.Next() {
		var p pair
		if err := personalRows.Scan(&p.scopeID, &p.name); err != nil {
			personalRows.Close()
			return err
		}
		personal = append(personal, p)
	}
	personalRows.Close()
	if err := personalRows.Err(); err != nil {
		return err
	}
	for _, p := range personal {
		if err := mergeTagsMatching(pool, `
			SELECT id FROM tags
			WHERE project_id IS NULL AND user_id = $1 AND lower(name) = $2
			ORDER BY id ASC`, p.scopeID, p.name); err != nil {
			return err
		}
	}

	projectRows, err := pool.Query(ctx, `
		SELECT project_id, lower(name)
		FROM tags
		WHERE project_id IS NOT NULL
		GROUP BY project_id, lower(name)
		HAVING COUNT(*) > 1`)
	if err != nil {
		return fmt.Errorf("failed to list project tag duplicates: %v", err)
	}
	var projects []pair
	for projectRows.Next() {
		var p pair
		if err := projectRows.Scan(&p.scopeID, &p.name); err != nil {
			projectRows.Close()
			return err
		}
		projects = append(projects, p)
	}
	projectRows.Close()
	if err := projectRows.Err(); err != nil {
		return err
	}
	for _, p := range projects {
		if err := mergeTagsMatching(pool, `
			SELECT id FROM tags
			WHERE project_id = $1 AND lower(name) = $2
			ORDER BY id ASC`, p.scopeID, p.name); err != nil {
			return err
		}
	}
	return nil
}

func tableExists(pool *pgxpool.Pool, name string) bool {
	var exists bool
	_ = pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
		name).Scan(&exists)
	return exists
}

func mergeTagsMatching(pool *pgxpool.Pool, idSQL string, scopeID int, name string) error {
	ctx := context.Background()
	rows, err := pool.Query(ctx, idSQL, scopeID, name)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(ids) < 2 {
		return nil
	}
	canonical := ids[0]
	for _, dupe := range ids[1:] {
		if _, err := pool.Exec(ctx, `
			UPDATE task_tags tt SET tag_id = $1
			WHERE tt.tag_id = $2
			  AND NOT EXISTS (
				SELECT 1 FROM task_tags existing
				WHERE existing.task_id = tt.task_id AND existing.tag_id = $1
			  )`, canonical, dupe); err != nil {
			return fmt.Errorf("failed to merge task_tags: %v", err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM task_tags WHERE tag_id = $1`, dupe); err != nil {
			return err
		}
		if tableExists(pool, "saved_views") {
			if _, err := pool.Exec(ctx, `
				UPDATE saved_views
				SET filter_json = jsonb_set(filter_json, '{tag}', to_jsonb($1::text), true)
				WHERE filter_json->>'tag' = $2`, fmt.Sprintf("%d", canonical), fmt.Sprintf("%d", dupe)); err != nil {
				return err
			}
		}
		if _, err := pool.Exec(ctx, `DELETE FROM tags WHERE id = $1`, dupe); err != nil {
			return err
		}
	}
	return nil
}

const accessibleTagCondition = `
	(
		(tg.project_id IS NULL AND tg.user_id = $%[1]d)
		OR EXISTS (
			SELECT 1 FROM projects p
			LEFT JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $%[1]d
			WHERE p.id = tg.project_id AND (p.user_id = $%[1]d OR pm.user_id IS NOT NULL)
		)
	)`

// GetTag returns a tag by id.
func GetTag(id int) (*Tag, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	t, err := scanTag(pool.QueryRow(context.Background(),
		`SELECT `+tagSelectCols+` FROM tags WHERE id = $1`, id))
	if err != nil {
		if isNoRows(err) {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("tag not found")
	}
	return &t, nil
}

// GetTagByID returns a tag by id if the user can access it.
func GetTagByID(id, userID int) (*Tag, error) {
	t, err := GetTag(id)
	if err != nil {
		return nil, err
	}
	ok, err := UserCanAccessTag(userID, *t)
	if err != nil || !ok {
		return nil, fmt.Errorf("tag not found")
	}
	return t, nil
}

// UserCanAccessTag reports whether the user can see the tag.
func UserCanAccessTag(userID int, tag Tag) (bool, error) {
	if tag.ProjectID == nil {
		return tag.UserID == userID, nil
	}
	_, err := GetAccessibleProjectByID(*tag.ProjectID, userID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// UserCanManageTag reports whether the user can create/rename/delete the tag.
func UserCanManageTag(userID int, tag Tag) (bool, error) {
	if tag.ProjectID == nil {
		return tag.UserID == userID, nil
	}
	proj, err := GetAccessibleProjectByID(*tag.ProjectID, userID)
	if err != nil {
		return false, nil
	}
	return RoleCanWrite(proj.Role), nil
}

// UserCanManageTagNamespace reports whether the user can add/delete tags in a namespace.
// projectID nil means personal tags.
func UserCanManageTagNamespace(userID int, projectID *int) (bool, error) {
	if projectID == nil || *projectID <= 0 {
		return true, nil
	}
	proj, err := GetAccessibleProjectByID(*projectID, userID)
	if err != nil {
		return false, fmt.Errorf("project not found")
	}
	return RoleCanWrite(proj.Role), nil
}

// GetAccessibleTags returns personal tags plus tags on projects the user can access.
func GetAccessibleTags(userID int) ([]Tag, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	cond := fmt.Sprintf(accessibleTagCondition, 1)
	rows, err := pool.Query(context.Background(),
		`SELECT `+tagSelectCols+` FROM tags tg WHERE `+cond+` ORDER BY LOWER(tg.name), tg.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	return collectTags(rows)
}

// GetTagsForUser returns all tags accessible to the user.
func GetTagsForUser(userID int) ([]Tag, error) {
	return GetAccessibleTags(userID)
}

// GetPersonalTags returns inbox tags owned by the user.
func GetPersonalTags(userID int) ([]Tag, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(),
		`SELECT `+tagSelectCols+` FROM tags WHERE project_id IS NULL AND user_id = $1 ORDER BY LOWER(name)`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	return collectTags(rows)
}

// GetProjectTags returns tags belonging to a project.
func GetProjectTags(projectID int) ([]Tag, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(),
		`SELECT `+tagSelectCols+` FROM tags WHERE project_id = $1 ORDER BY LOWER(name)`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	return collectTags(rows)
}

func collectTags(rows pgx.Rows) ([]Tag, error) {
	defer rows.Close()
	var out []Tag
	for rows.Next() {
		t, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []Tag{}
	}
	return out, nil
}

// FindTagByName returns a tag in the given namespace (case-insensitive).
// projectID nil means personal tags for userID.
func FindTagByName(userID int, projectID *int, name string) (*Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var t Tag
	if projectID == nil || *projectID <= 0 {
		t, err = scanTag(pool.QueryRow(context.Background(),
			`SELECT `+tagSelectCols+` FROM tags
			 WHERE project_id IS NULL AND user_id = $1 AND LOWER(name) = LOWER($2)`,
			userID, name))
	} else {
		t, err = scanTag(pool.QueryRow(context.Background(),
			`SELECT `+tagSelectCols+` FROM tags
			 WHERE project_id = $1 AND LOWER(name) = LOWER($2)`,
			*projectID, name))
	}
	if err != nil {
		if isNoRows(err) {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, err
	}
	return &t, nil
}

// GetOrCreateTagByName returns an existing tag or creates one in the namespace.
func GetOrCreateTagByName(userID int, projectID *int, name string) (*Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("tag name must be 50 characters or less")
	}
	if existing, err := FindTagByName(userID, projectID, name); err == nil {
		return existing, nil
	}

	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var t Tag
	if projectID == nil || *projectID <= 0 {
		t, err = scanTag(pool.QueryRow(context.Background(),
			`INSERT INTO tags (user_id, name) VALUES ($1, $2)
			 RETURNING `+tagSelectCols, userID, name))
	} else {
		t, err = scanTag(pool.QueryRow(context.Background(),
			`INSERT INTO tags (user_id, project_id, name) VALUES ($1, $2, $3)
			 RETURNING `+tagSelectCols, userID, *projectID, name))
	}
	if err != nil {
		if existing, findErr := FindTagByName(userID, projectID, name); findErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create tag: %v", err)
	}
	t.Color = tagColorForID(t.ID)
	_, _ = pool.Exec(context.Background(), "UPDATE tags SET color = $1 WHERE id = $2", t.Color, t.ID)
	return &t, nil
}

// DeleteTag removes a tag by id.
func DeleteTag(id int) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	tag, err := pool.Exec(context.Background(), "DELETE FROM tags WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

// UpdateTag renames a tag.
func UpdateTag(id int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("tag name is required")
	}
	if len(name) > 50 {
		return fmt.Errorf("tag name must be 50 characters or less")
	}

	existing, err := GetTag(id)
	if err != nil {
		return err
	}

	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	var exists bool
	if existing.ProjectID == nil {
		if err := pool.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM tags
			 WHERE project_id IS NULL AND user_id = $1 AND LOWER(name) = LOWER($2) AND id != $3)`,
			existing.UserID, name, id).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check tag name: %v", err)
		}
	} else {
		if err := pool.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM tags
			 WHERE project_id = $1 AND LOWER(name) = LOWER($2) AND id != $3)`,
			*existing.ProjectID, name, id).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check tag name: %v", err)
		}
	}
	if exists {
		return fmt.Errorf("a tag with that name already exists")
	}

	tag, err := pool.Exec(context.Background(), "UPDATE tags SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		return fmt.Errorf("failed to update tag: %v", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func taskTagNamespace(taskID int) (*int, error) {
	pid, err := GetTaskProjectID(taskID)
	if err != nil {
		return nil, err
	}
	if pid <= 0 {
		return nil, nil
	}
	return &pid, nil
}

func tagMatchesTaskNamespace(tag Tag, taskProjectID *int, userID int) bool {
	if taskProjectID == nil {
		return tag.ProjectID == nil && tag.UserID == userID
	}
	return tag.ProjectID != nil && *tag.ProjectID == *taskProjectID
}

// SetTaskTags replaces all tags on a task (max MaxTagsPerTask).
func SetTaskTags(taskID, userID int, tagIDs []int) error {
	if len(tagIDs) > MaxTagsPerTask {
		return fmt.Errorf("maximum %d tags per task", MaxTagsPerTask)
	}

	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	canRead, writeRole, _, accessErr := CanUserAccessTask(taskID, userID)
	if accessErr != nil {
		return accessErr
	}
	if !canRead || !RoleCanWrite(writeRole) {
		return fmt.Errorf("not authorized")
	}

	taskProjectID, err := taskTagNamespace(taskID)
	if err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		tag, err := GetTag(tagID)
		if err != nil || !tagMatchesTaskNamespace(*tag, taskProjectID, userID) {
			return fmt.Errorf("invalid tag selection")
		}
	}

	_, err = pool.Exec(context.Background(), "DELETE FROM task_tags WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to clear task tags: %v", err)
	}

	for _, tagID := range tagIDs {
		_, err = pool.Exec(context.Background(), "INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)", taskID, tagID)
		if err != nil {
			return fmt.Errorf("failed to assign tag: %v", err)
		}
	}
	return nil
}

// RemapTaskTagsForProjectChange keeps tags whose names exist in the destination namespace and drops the rest.
func RemapTaskTagsForProjectChange(taskID, userID int, newProjectID *int) error {
	existing, err := GetTagsForTask(taskID)
	if err != nil {
		return err
	}
	if len(existing) == 0 {
		return nil
	}
	var dest []Tag
	if newProjectID == nil || *newProjectID <= 0 {
		dest, err = GetPersonalTags(userID)
	} else {
		dest, err = GetProjectTags(*newProjectID)
	}
	if err != nil {
		return err
	}
	byName := make(map[string]int, len(dest))
	for _, t := range dest {
		byName[strings.ToLower(t.Name)] = t.ID
	}
	var ids []int
	seen := make(map[int]bool)
	for _, t := range existing {
		id, ok := byName[strings.ToLower(t.Name)]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return SetTaskTags(taskID, userID, ids)
}

// GetTagsForTask returns tags assigned to a single task.
func GetTagsForTask(taskID int) ([]Tag, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(), `
		SELECT `+tagSelectCols+`
		FROM tags tg
		JOIN task_tags tt ON tt.tag_id = tg.id
		WHERE tt.task_id = $1
		ORDER BY LOWER(tg.name)`, taskID)
	if err != nil {
		return nil, err
	}
	return collectTags(rows)
}

// GetTagsForTasks batch-loads tags for multiple tasks.
func GetTagsForTasks(taskIDs []int) (map[int][]Tag, error) {
	result := make(map[int][]Tag)
	if len(taskIDs) == 0 {
		return result, nil
	}

	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(), `
		SELECT tt.task_id, `+qualifyTagSelect("tg")+`
		FROM task_tags tt
		JOIN tags tg ON tg.id = tt.tag_id
		WHERE tt.task_id = ANY($1)
		ORDER BY tt.task_id, LOWER(tg.name)`, taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID int
		var projectID sql.NullInt64
		var t Tag
		if err := rows.Scan(&taskID, &t.ID, &t.UserID, &projectID, &t.Name, &t.Color); err != nil {
			return nil, err
		}
		if projectID.Valid && projectID.Int64 > 0 {
			pid := int(projectID.Int64)
			t.ProjectID = &pid
		}
		result[taskID] = append(result[taskID], t)
	}
	return result, nil
}

func qualifyTagSelect(alias string) string {
	return fmt.Sprintf(`%s.id, %s.user_id, %s.project_id, %s.name, COALESCE(%s.color, '#6c757d')`,
		alias, alias, alias, alias, alias)
}

// ResolveTaskTagIDs parses tag_ids from form and creates new tags from comma-separated names.
func ResolveTaskTagIDs(userID int, projectID *int, tagIDStrs []string, newTagsCSV string) ([]int, error) {
	seen := make(map[int]bool)
	var ids []int

	for _, s := range tagIDStrs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := parseInt(s)
		if err != nil {
			continue
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	for _, part := range strings.Split(newTagsCSV, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		t, err := GetOrCreateTagByName(userID, projectID, name)
		if err != nil {
			return nil, err
		}
		if !seen[t.ID] {
			seen[t.ID] = true
			ids = append(ids, t.ID)
		}
	}

	if len(ids) > MaxTagsPerTask {
		return nil, fmt.Errorf("maximum %d tags per task", MaxTagsPerTask)
	}
	return ids, nil
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
