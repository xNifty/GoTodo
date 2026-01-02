package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// APIToggleFavorite toggles the is_favorite flag for a task and reloads the task list
func APIToggleFavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}

	email, _, _, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Prevent banned users from performing actions
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userID int
	if uid := utils.GetSessionUserID(r); uid != nil {
		userID = *uid
	} else {
		err = db.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Get current favorite status
	var isFav bool
	err = db.QueryRow(context.Background(), "SELECT COALESCE(is_favorite,false) FROM tasks WHERE id = $1 AND user_id = $2", id, userID).Scan(&isFav)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// If setting to favorite, ensure user has fewer than 5 favorites
	if !isFav {
		var favCount int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND is_favorite = true", userID).Scan(&favCount)
		if err != nil {
			http.Error(w, "Error checking favorites", http.StatusInternalServerError)
			return
		}
		if favCount >= 5 {
			// Trigger client-side event via HTMX without returning an error status
			// Provide a small JSON payload so the client can show a custom message
			w.Header().Set("HX-Trigger", `{"favorite-limit-reached":{"message":"You can only favorite up to 5 tasks"}}`)
			w.WriteHeader(http.StatusNoContent) // 204 prevents HTMX from swapping content
			return
		}
	}

	// Toggle favorite
	_, err = db.Exec(context.Background(), "UPDATE tasks SET is_favorite = NOT COALESCE(is_favorite,false) WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "Error updating favorite", http.StatusInternalServerError)
		return
	}

	// Determine page size from session
	pageSize := utils.AppConstants.PageSize
	if sess, err := sessionstore.Store.Get(r, "session"); err == nil && sess != nil {
		if val, ok := sess.Values["items_per_page"]; ok {
			switch tv := val.(type) {
			case int:
				if tv > 0 {
					pageSize = tv
				}
			case int64:
				if int(tv) > 0 {
					pageSize = int(tv)
				}
			case float64:
				if int(tv) > 0 {
					pageSize = int(tv)
				}
			case string:
				if v, err := strconv.Atoi(tv); err == nil && v > 0 {
					pageSize = v
				}
			}
		}
	}

	// Fetch updated tasks and render the pagination partial
	userPtr := &userID
	taskList, totalTasks, err := tasks.ReturnPaginationForUser(page, pageSize, userPtr, timezone)
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Split into favorites and non-favorites for rendering and set page
	favs := make([]tasks.Task, 0)
	nonFavs := make([]tasks.Task, 0)
	for i := range taskList {
		taskList[i].Page = page
		if taskList[i].IsFavorite {
			favs = append(favs, taskList[i])
		} else {
			nonFavs = append(nonFavs, taskList[i])
		}
	}

	uid := userID
	pagination := utils.GetPaginationData(page, pageSize, totalTasks, uid)

	context := map[string]interface{}{
		"FavoriteTasks":    favs,
		"Tasks":            nonFavs,
		"PreviousPage":     pagination.PreviousPage,
		"NextPage":         pagination.NextPage,
		"CurrentPage":      pagination.CurrentPage,
		"PrevDisabled":     pagination.PrevDisabled,
		"NextDisabled":     pagination.NextDisabled,
		"SearchQuery":      "",
		"TotalTasks":       totalTasks,
		"LoggedIn":         true,
		"Timezone":         timezone,
		"TotalPages":       pagination.TotalPages,
		"Pages":            pagination.Pages,
		"HasRightEllipsis": pagination.HasRightEllipsis,
		"PerPage":          pageSize,
		"CompletedTasks":   pagination.TotalCompletedTasks,
		"IncompleteTasks":  pagination.TotalIncompleteTasks,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, r, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
