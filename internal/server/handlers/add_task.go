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
	"strings"
)

const MaxDescriptionLength = 100

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Request method: ", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	pageStr := strings.TrimSpace(r.FormValue("currentPage"))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Default to page 1 if no valid page is provided
	}

	// Validate description length
	if len(description) > MaxDescriptionLength {
		// On validation failure, return a 200 status with the error message
		// and use HX-Retarget and HX-Reswap to update the error div specifically
		// Tell the client this was a validation error so JS won't close the sidebar
		w.Header().Set("X-Validation-Error", "true")
		w.Header().Set("HX-Trigger", "description-error")   // Keep trigger for potential JS handling
		w.Header().Set("HX-Retarget", "#description-error") // Target the specific error div
		w.Header().Set("HX-Reswap", "innerHTML")            // Swap the content inside the error div
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Description must be %d characters or less", MaxDescriptionLength) // The content to swap
		return
	}

	if title == "" {
		// Title missing — return validation error and appropriate message
		w.Header().Set("X-Validation-Error", "true")
		w.Header().Set("HX-Trigger", "description-error")   // Keep trigger for potential JS handling
		w.Header().Set("HX-Retarget", "#description-error") // Target the specific error div
		w.Header().Set("HX-Reswap", "innerHTML")            // Swap the content inside the error div
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Title is required")
		return
		// http.Error(w, "Title is required", http.StatusBadRequest)
		// return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		fmt.Println("We failed to open the database.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get user ID from session (fallback to querying by email if not present)
	email, _, _, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Please log in to add tasks.")
		return
	}

	var userID int
	if uid := utils.GetSessionUserID(r); uid != nil {
		userID = *uid
	} else {
		// fallback to DB lookup if session doesn't contain user_id
		err = db.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
		if err != nil {
			fmt.Printf("Error getting user ID: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Determine next position within non-favorite group for this user
	var nextPos int
	err = db.QueryRow(context.Background(), "SELECT COALESCE(MAX(position),0) + 1 FROM tasks WHERE user_id = $1 AND (is_favorite IS NULL OR is_favorite = false)", userID).Scan(&nextPos)
	if err != nil {
		fmt.Printf("Error determining next position: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert the new task into the database with user_id and position
	_, err = db.Exec(context.Background(), "INSERT INTO tasks (title, description, completed, user_id, time_stamp, position) VALUES ($1, $2, $3, $4, NOW() AT TIME ZONE 'UTC', $5)", title, description, false, userID, nextPos)
	if err != nil {
		fmt.Println("We failed to insert into the database.")
		fmt.Println("Failed values:", title, description, false)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// After successful insertion, determine the correct page to display
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

	// Open a new DB connection to count total tasks (or reuse db if possible)
	var totalTasks int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE user_id = $1", userID).Scan(&totalTasks)
	if err != nil {
		http.Error(w, "Error counting tasks after add: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the new last page
	lastPage := (totalTasks + pageSize - 1) / pageSize
	if lastPage < 1 {
		lastPage = 1
	}

	// If the new task caused a new page, go to the last page
	if page < lastPage {
		page = lastPage
	}

	taskList, totalTasks, err := tasks.ReturnPaginationForUser(page, pageSize, &userID, timezone)
	if err != nil {
		http.Error(w, "Error fetching tasks after add: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination button states based on new totalTasks
	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalTasks {
		nextDisabled = "disabled"
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1

	// Split into favorites and non-favorites for rendering and allow separate sortable containers
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

	// Create a context for rendering pagination.html
	context := map[string]interface{}{
		"FavoriteTasks":    favs,
		"Tasks":            nonFavs,
		"PreviousPage":     prevPage,
		"NextPage":         nextPage,
		"CurrentPage":      page,
		"PrevDisabled":     prevDisabled,
		"NextDisabled":     nextDisabled,
		"TotalTasks":       totalTasks,
		"LoggedIn":         true,
		"TotalPages":       (totalTasks + pageSize - 1) / pageSize,
		"Pages":            utils.GetPaginationData(page, pageSize, totalTasks, userID).Pages,
		"HasRightEllipsis": utils.GetPaginationData(page, pageSize, totalTasks, userID).HasRightEllipsis,
		"CompletedTasks":   utils.GetCompletedTasksCount(&userID),
		"IncompleteTasks":  utils.GetIncompleteTasksCount(&userID),
		"PerPage":          pageSize,
	}

	// Set headers for successful addition
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "task-added") // Signal JS to close sidebar and clear form

	// Render the updated task list into the main task-container
	if err := utils.RenderTemplate(w, r, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering tasks after add: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response (HTMX will handle the swap due to hx-target and hx-swap on the form)
}
