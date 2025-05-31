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

const MaxDescriptionLength = 100

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Request method: ", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	pageStr := r.FormValue("currentPage")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Default to page 1 if no valid page is provided
	}

	// Validate description length
	if len(description) > MaxDescriptionLength {
		// On validation failure, return a 200 status with the error message
		// and use HX-Retarget and HX-Reswap to update the error div specifically
		w.Header().Set("HX-Trigger", "description-error")   // Keep trigger for potential JS handling
		w.Header().Set("HX-Retarget", "#description-error") // Target the specific error div
		w.Header().Set("HX-Reswap", "innerHTML")            // Swap the content inside the error div
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Description must be %d characters or less", MaxDescriptionLength) // The content to swap
		return
	}

	if title == "" {
		// Handle empty title error - maybe similar HX-Retarget for title error div?
		// For now, just return bad request
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		fmt.Println("We failed to open the database.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new task into the database
	_, err = db.Exec(context.Background(), "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3)", title, description, false)
	if err != nil {
		fmt.Println("We failed to insert into the database.")
		fmt.Println("Failed values:", title, description, false)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// After successful insertion, render the updated task list (pagination.html)
	pageSize := utils.AppConstants.PageSize
	// We might need the total tasks and tasks for the current page here to render pagination correctly

	// Fetch tasks for the current page again to get the updated list
	taskList, totalTasks, err := tasks.ReturnPagination(page, pageSize)
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

	// Create a context for rendering pagination.html
	context := map[string]interface{}{
		"Tasks":        taskList,
		"PreviousPage": prevPage,
		"NextPage":     nextPage,
		"CurrentPage":  page,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
		// Assuming SearchQuery might need to be preserved, pass it if available
		// "SearchQuery":  r.FormValue("search"), // Need to get search query from form if needed
	}

	// Set headers for successful addition
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "task-added") // Signal JS to close sidebar and clear form

	// Render the updated task list into the main task-container
	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering tasks after add: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response (HTMX will handle the swap due to hx-target and hx-swap on the form)
}
