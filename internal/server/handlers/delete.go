package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
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

	// Set response headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Trigger the next item fetch
	w.Header().Set("HX-Trigger", "taskDeleted")
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

	// Check if we're on the last page
	pageSize := utils.AppConstants.PageSize
	var totalTasks int
	err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&totalTasks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error counting tasks")
		return
	}

	lastPage := (totalTasks + pageSize - 1) / pageSize
	if currentPage >= lastPage {
		// If we're on the last page, don't return anything
		return
	}

	// Get the first item from the next page using a window function
	var task tasks.Task
	err = db.QueryRow(context.Background(),
		`WITH numbered_tasks AS (
			SELECT 
				id, 
				title, 
				description, 
				completed,
				TO_CHAR(time_stamp, 'YYYY/MM/DD HH:MI AM') AS date_added,
				ROW_NUMBER() OVER (ORDER BY id) as row_num
			FROM tasks
		)
		SELECT 
			id, 
			title, 
			description, 
			completed, 
			date_added
		FROM numbered_tasks
		WHERE row_num = $1
		ORDER BY id`,
		currentPage*pageSize).Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.DateAdded)

	if err != nil {
		// No more items
		return
	}

	task.Page = currentPage
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, "todo.html", &task); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
