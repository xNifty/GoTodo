package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/tasks"
	"net/http"
	"strconv"
)

func APIReturnTasks(w http.ResponseWriter, r *http.Request) {
	pageSize := utils.AppConstants.PageSize

	searchQuery := r.URL.Query().Get("search")

	// Parse "page" query parameter
	var currentPage int
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		var err error
		currentPage, err = strconv.Atoi(pageParam)
		if err != nil || currentPage < 1 {
			currentPage = 1
		}
	} else {
		currentPage = 1
	}
	page := currentPage

	// Get user ID if logged in
	email, _, _, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	var userID *int
	if loggedIn {
		userID = getUserIDFromEmail(email)
	}

	// Fetch tasks for the current page
	var taskList []tasks.Task
	var totalTasks int
	var err error

	if searchQuery != "" {
		taskList, totalTasks, err = tasks.SearchTasksForUser(page, pageSize, searchQuery, userID, timezone)
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
		taskList, totalTasks, err = tasks.ReturnPaginationForUser(page, pageSize, userID, timezone)
		if err != nil {
			http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Validate and clamp page number to valid range
	lastPage := (totalTasks + pageSize - 1) / pageSize
	if lastPage < 1 {
		lastPage = 1
	}
	if page > lastPage {
		page = lastPage
	}
	if page < 1 {
		page = 1
	}

	// If page was adjusted, we need to refetch with the correct page
	if page != currentPage {
		if searchQuery != "" {
			taskList, totalTasks, err = tasks.SearchTasksForUser(page, pageSize, searchQuery, userID, timezone)
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
			taskList, totalTasks, err = tasks.ReturnPaginationForUser(page, pageSize, userID, timezone)
			if err != nil {
				http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
				return
			}
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
		"TotalTasks":   totalTasks,
		"LoggedIn":     loggedIn,
		"Timezone":     timezone,
		"TotalPages":   pagination.TotalPages,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
