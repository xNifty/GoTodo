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
	searchQuery := r.URL.Query().Get("search")

	// Parse "page" query parameter
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			page = 1
		}
	} else {
		page = 1
	}

	// Fetch tasks for the current page
	var taskList []tasks.Task
	var totalTasks int
	var err error

	if searchQuery != "" {
		taskList, totalTasks, err = tasks.SearchTasks(page, pageSize, searchQuery)
		if err != nil {
			http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Highlight search matches
		for i, task := range taskList {
			taskList[i].Title = highlightMatches(task.Title, searchQuery)
			taskList[i].Description = highlightMatches(task.Description, searchQuery)
		}
	} else {
		taskList, totalTasks, err = tasks.ReturnPagination(page, pageSize)
		if err != nil {
			http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Set the Page field for each task
	for i := range taskList {
		taskList[i].Page = page
	}

	pagination := utils.GetPaginationData(page, pageSize, totalTasks)

	// Set response header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create a context for the tasks and pagination
	context := map[string]interface{}{
		"Tasks":        taskList,
		"PreviousPage": pagination.PreviousPage,
		"NextPage":     pagination.NextPage,
		"CurrentPage":  pagination.CurrentPage,
		"PrevDisabled": pagination.PrevDisabled,
		"NextDisabled": pagination.NextDisabled,
		"SearchQuery":  searchQuery,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
