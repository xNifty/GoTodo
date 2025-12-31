package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"net/http"
	"strconv"
	"strings"
)

// APIReorderTasks updates positions for tasks within a favorite/non-favorite group
func APIReorderTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	order := r.FormValue("order") // comma-separated IDs
	isFavStr := r.FormValue("is_favorite")
	pageStr := r.FormValue("page")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	isFav := false
	if isFavStr == "true" || isFavStr == "1" {
		isFav = true
	}

	if order == "" {
		http.Error(w, "Missing order", http.StatusBadRequest)
		return
	}

	idStrs := strings.Split(order, ",")
	ids := make([]int, 0, len(idStrs))
	for _, s := range idStrs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "Invalid id in order", http.StatusBadRequest)
			return
		}
		ids = append(ids, v)
	}

	email, _, _, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
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

	// Validate that all provided IDs belong to the user and match is_favorite
	for _, id := range ids {
		var exists bool
		err = db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1 AND user_id = $2 AND COALESCE(is_favorite,false) = $3)", id, userID, isFav).Scan(&exists)
		if err != nil {
			http.Error(w, "Error validating tasks", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Task does not belong to user or mismatched favorite group", http.StatusBadRequest)
			return
		}
	}

	// Fetch all task IDs in this user's group ordered by position so we can renumber globally
	rowsAll, err := db.Query(context.Background(), "SELECT id FROM tasks WHERE user_id = $1 AND COALESCE(is_favorite,false) = $2 ORDER BY position ASC, id ASC", userID, isFav)
	if err != nil {
		http.Error(w, "Error fetching task list for reorder", http.StatusInternalServerError)
		return
	}
	defer rowsAll.Close()

	allIDs := make([]int, 0)
	for rowsAll.Next() {
		var tid int
		if err := rowsAll.Scan(&tid); err != nil {
			http.Error(w, "Error reading task ids", http.StatusInternalServerError)
			return
		}
		allIDs = append(allIDs, tid)
	}

	if len(allIDs) == 0 {
		// Nothing to reorder
	} else {
		// Compute page window start index and clamp
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

		start := (page - 1) * pageSize
		if start < 0 {
			start = 0
		}
		// Ensure the replacement window fits in the allIDs slice
		if start > len(allIDs)-len(ids) {
			start = len(allIDs) - len(ids)
			if start < 0 {
				start = 0
			}
		}

		// Replace the slice segment with the new ordering provided by the client
		for i, id := range ids {
			if start+i < len(allIDs) {
				allIDs[start+i] = id
			} else {
				// Append if somehow beyond end
				allIDs = append(allIDs, id)
			}
		}

		// Update positions for allIDs inside a transaction
		tx, err := db.Begin(context.Background())
		if err != nil {
			http.Error(w, "Error starting transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(context.Background())

		for idx, id := range allIDs {
			pos := idx + 1
			_, err := tx.Exec(context.Background(), "UPDATE tasks SET position = $1 WHERE id = $2 AND user_id = $3", pos, id, userID)
			if err != nil {
				http.Error(w, "Error updating positions", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(context.Background()); err != nil {
			http.Error(w, "Error committing position updates", http.StatusInternalServerError)
			return
		}
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

	userPtr := &userID
	taskList, totalTasks, err := tasks.ReturnPaginationForUser(page, pageSize, userPtr, timezone)
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Split into favorites and non-favorites
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
