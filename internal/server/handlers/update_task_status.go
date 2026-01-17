package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

func APIUpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	page := r.URL.Query().Get("page")

	// Require logged-in user and enforce ban check + ownership
	email, _, _, loggedIn := utils.GetSessionUser(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if isBanned, err := storage.IsUserBanned(email); err == nil && isBanned {
		sessionstore.ClearSessionCookie(w, r)
		basePath := utils.GetBasePath()
		w.Header().Set("HX-Redirect", basePath)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, " ")
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var completed bool

	// Ensure task exists and belongs to the current user
	var ownerID int
	err = db.QueryRow(context.Background(), "SELECT completed, user_id FROM tasks WHERE id = $1", id).Scan(&completed, &ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Task not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Task not found.", http.StatusInternalServerError)
		return
	}

	// Verify ownership
	var userID int
	if uid := utils.GetSessionUserID(r); uid != nil {
		userID = *uid
	} else {
		err = db.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
		if err != nil {
			http.Error(w, "Error getting user ID", http.StatusInternalServerError)
			return
		}
	}
	if ownerID != userID {
		http.Error(w, "Not authorized to update this task.", http.StatusForbidden)
		return
	}

	updatedStatus := !completed

	_, err = db.Exec(context.Background(), "UPDATE tasks SET completed = $1 WHERE id = $2", updatedStatus, id)

	if err != nil {
		http.Error(w, "Failed to update task status.", http.StatusInternalServerError)
		return
	}

	if err := db.QueryRow(context.Background(), "SELECT user_id FROM tasks WHERE id = $1", id).Scan(&ownerID); err == nil {
		var completedCount int
		var incompleteCount int
		_ = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND completed = true", ownerID).Scan(&completedCount)
		_ = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND completed = false", ownerID).Scan(&incompleteCount)
		// Emit HTMX trigger with counts payload so client can update badges
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"taskCountsChanged":{"completed":%d,"incomplete":%d}}`, completedCount, incompleteCount))
	}

	basePath := utils.GetBasePath()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Include the status-column class so mobile styles remain consistent after HTMX swaps
	icons := map[bool]string{
		true:  `<i class="bi bi-toggle-on"></i>`,
		false: `<i class="bi bi-toggle-off"></i>`,
	}
	labels := map[bool]string{true: "Complete", false: "Incomplete"}

	// Build the actions column (status toggle + optional edit + delete)
	// Include an off-screen label for accessibility
	editBtn := ""
	if !updatedStatus {
		// only show edit when task is incomplete
		editBtn = fmt.Sprintf(`<button class="btn btn-link p-0 mx-2 edit-btn" style="text-decoration:none;" hx-get="%s/api/edit?id=%s&page=%s" hx-target="#sidebar .sidebar-body" hx-swap="innerHTML" aria-label="Edit task"><i class="bi bi-pencil"></i></button>`, basePath, id, page)
	}

	deleteBtn := fmt.Sprintf(`<button hx-get="%s/api/confirm?id=%s&page=%s" hx-target="#modal .modal-content" hx-trigger="click" data-bs-toggle="modal" data-bs-target="#modal" class="btn btn-link p-0 delete-column" aria-label="Delete task" style="text-decoration:none;"><i class="bi bi-trash text-danger"></i></button>`, basePath, id, page)

	// status button (icon + visible label)
	statusBtn := fmt.Sprintf(`<button class="badge %s status-column" hx-get="%s/api/update-status?id=%s&page=%s" hx-target="#task-%s .actions-column" hx-swap="outerHTML" aria-label="Toggle complete" style="cursor: pointer; display: inline-flex; align-items: center; justify-content: center; padding:.35rem; gap:.4rem;">%s %s</button>`,
		map[bool]string{true: "bg-success", false: "bg-danger text-white"}[updatedStatus],
		basePath,
		id,
		page,
		id,
		icons[updatedStatus],
		labels[updatedStatus],
	)

	// Wrap into a td matching the template so HTMX can replace it cleanly
	actionsHTML := fmt.Sprintf(`<td class="actions-column" data-label="Actions"><div class="d-flex align-items-center gap-2 justify-content-start">%s%s%s</div></td>`, statusBtn, editBtn, deleteBtn)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, actionsHTML)
}
