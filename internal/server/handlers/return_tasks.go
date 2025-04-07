package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/tasks"
	"net/http"
	"strconv"
)

func APIReturnTasks(w http.ResponseWriter, r *http.Request) {
	pageSize := utils.AppConstants.PageSize

	var page int

	// Parse "page" query parameter
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		page, err := strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			page = 1
		}
	} else {
		page = 1
	}

	//fmt.Println("\nPage, early: ", page)

	// Fetch tasks for the current page
	tasks, totalTasks, err := tasks.ReturnPagination(page, utils.AppConstants.PageSize, "")
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination button states
	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled" // Disable on the first page
	}

	nextDisabled := ""
	if page*pageSize >= totalTasks {
		nextDisabled = "disabled" // Disable if next page is unavailable
	}

	// Set response header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1

	if page*pageSize >= totalTasks {
		nextPage = page
	}

	// Create a context for the tasks and pagination
	context := map[string]interface{}{
		"Tasks":        tasks,
		"PreviousPage": prevPage,
		"NextPage":     nextPage,
		"CurrentPage":  page,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}

}
