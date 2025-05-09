package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func APIDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("id")
	currentPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if currentPage < 1 {
		currentPage = 1
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error opening database")
		return
	}
	defer db.Close()

	// Delete the task from the database
	_, err = db.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting task")
		return
	}

	// Get total number of tasks after deletion
	var totalTasks int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&totalTasks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error counting tasks")
		return
	}

	// Calculate last page
	pageSize := utils.AppConstants.PageSize
	lastPage := (totalTasks + pageSize - 1) / pageSize

	// If current page is beyond last page, we need to reload the previous page
	if currentPage > lastPage && currentPage > 1 {
		w.Header().Set("HX-Trigger", "reload-previous-page")
	} else {
		// Otherwise, just reload the current page
		w.Header().Set("HX-Trigger", "taskDeleted")
	}

	// Set response headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func APIGetNextItem(w http.ResponseWriter, r *http.Request) {
	currentPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if currentPage < 1 {
		currentPage = 1
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error opening database")
		return
	}
	defer db.Close()

	// Get total number of tasks first
	var totalTasks int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&totalTasks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Calculate last page
	pageSize := utils.AppConstants.PageSize
	lastPage := (totalTasks + pageSize - 1) / pageSize

	// If current page is beyond last page, trigger reload of previous page
	if currentPage > lastPage && currentPage > 1 {
		w.Header().Set("HX-Trigger", "reload-previous-page")
		w.WriteHeader(http.StatusOK)
		return
	}

	offset := (currentPage - 1) * pageSize

	// Check how many items are on the current page
	var itemsOnPage int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks LIMIT $1 OFFSET $2", pageSize, offset).Scan(&itemsOnPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If there are no items on the current page and we're not on page 1, reload previous page
	if itemsOnPage == 0 && currentPage > 1 {
		w.Header().Set("HX-Trigger", "reload-previous-page")
		w.WriteHeader(http.StatusOK)
		return
	}

	// If we're on page 1 and the total tasks is less than or equal to page size,
	// we don't need to pull up any items
	if currentPage == 1 && totalTasks <= pageSize {
		w.WriteHeader(http.StatusOK)
		return
	}

	// If we're on the last page and it's full, no need to pull up
	if currentPage == lastPage && itemsOnPage == pageSize {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only pull up the next item if there is a gap on the page
	nextItemOffset := offset + itemsOnPage
	if nextItemOffset >= totalTasks {
		w.WriteHeader(http.StatusOK)
		return
	}

	row := db.QueryRow(context.Background(),
		`SELECT id, title, description, completed, TO_CHAR(time_stamp, 'YYYY/MM/DD HH:MI AM') AS date_added
		 FROM tasks ORDER BY id LIMIT 1 OFFSET $1`, nextItemOffset)

	var task tasks.Task
	err = row.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.DateAdded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	task.Page = currentPage
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, "todo.html", &task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
