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
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		for i, task := range taskList {
			taskList[i].Title = highlightMatches(task.Title, searchQuery)
			taskList[i].Description = highlightMatches(task.Description, searchQuery)
		}
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

	// Create a context for the tasks and pagination
	context := map[string]interface{}{
		"Tasks":        taskList,
		"PreviousPage": prevPage,
		"NextPage":     nextPage,
		"CurrentPage":  page,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
	}

	// Render the tasks and pagination controls
	if err := utils.RenderTemplate(w, "index.html", context); err != nil {
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

	taskList, totalResults, err := tasks.SearchTasks(page, pageSize, searchQuery)
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		for i, task := range taskList {
			taskList[i].Title = highlightMatches(task.Title, searchQuery)
			taskList[i].Description = highlightMatches(task.Description, searchQuery)
		}
	}

	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalResults {
		nextDisabled = "disabled"
	}

	context := map[string]interface{}{
		"Tasks":        taskList,
		"TotalResults": totalResults,
		"SearchQuery":  searchQuery,
		"PreviousPage": page - 1,
		"NextPage":     page + 1,
		"CurrentPage":  page,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func highlightMatches(text, searchQuery string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(searchQuery))
	text = re.ReplaceAllString(text, "<mark>$0</mark>")
	return text
}
