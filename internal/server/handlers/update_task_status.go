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
	fmt.Fprintf(w, `<button 
		class="badge %s status-column"
		hx-get="`+basePath+`/api/update-status?id=%s" 
		hx-target="#task-%s .badge" 
		hx-swap="outerHTML"
		style="cursor: pointer;">
		%s
	</button>`,
		map[bool]string{true: "bg-success", false: "bg-danger"}[updatedStatus],
		id,
		id,
		map[bool]string{true: "Complete", false: "Incomplete"}[updatedStatus],
	)
}
