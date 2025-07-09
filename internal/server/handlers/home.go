package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/tasks"
	"net/http"
	"regexp"
	"strconv"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := utils.AppConstants.PageSize
	searchQuery := r.URL.Query().Get("search")

	var taskList []tasks.Task
	var totalTasks int
	var err error

	if searchQuery != "" {
		taskList, totalTasks, err = tasks.SearchTasks(page, pageSize, searchQuery)
	} else {
		taskList, totalTasks, err = tasks.ReturnPagination(page, pageSize)
	}

	if err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		for i, task := range taskList {
			taskList[i].Title = highlightMatches(task.Title, searchQuery)
			taskList[i].Description = highlightMatches(task.Description, searchQuery)
		}
	}

	// Set the page number for each task
	for i := range taskList {
		taskList[i].Page = page
	}

	pagination := utils.GetPaginationData(page, pageSize, totalTasks)

	// Create a context for the tasks and pagination
	context := map[string]interface{}{
		"Tasks":        taskList,
		"CurrentPage":  page,
		"PreviousPage": pagination.PreviousPage,
		"NextPage":     pagination.NextPage,
		"PrevDisabled": pagination.PrevDisabled,
		"NextDisabled": pagination.NextDisabled,
	}

	// Render the tasks and pagination controls
	if err := utils.RenderTemplate(w, "index.html", context); err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	pageSize := utils.AppConstants.PageSize
	var page int

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			page = 1
		}
	} else {
		page = 1
	}

	searchQuery := r.FormValue("search")

	taskList, totalTasks, err := tasks.SearchTasks(page, pageSize, searchQuery)
	if err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		for i, task := range taskList {
			taskList[i].Title = highlightMatches(task.Title, searchQuery)
			taskList[i].Description = highlightMatches(task.Description, searchQuery)
		}
	}

	// Set the page number for each task
	for i := range taskList {
		taskList[i].Page = page
	}

	pagination := utils.GetPaginationData(page, pageSize, totalTasks)

	context := map[string]interface{}{
		"Tasks":        taskList,
		"TotalResults": totalTasks,
		"SearchQuery":  searchQuery,
		"CurrentPage":  page,
		"PreviousPage": pagination.PreviousPage,
		"NextPage":     pagination.NextPage,
		"PrevDisabled": pagination.PrevDisabled,
		"NextDisabled": pagination.NextDisabled,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func highlightMatches(text, searchQuery string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(searchQuery))
	text = re.ReplaceAllString(text, "<mark>$0</mark>")
	return text
}
