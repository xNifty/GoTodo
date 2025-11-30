package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"net/http"
	"regexp"
	"strconv"
)

func getUserIDFromEmail(email string) *int {
	pool, err := storage.OpenDatabase()
	if err != nil {
		return nil
	}
	defer storage.CloseDatabase(pool)

	var userID int
	err = pool.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		return nil
	}

	return &userID
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := utils.AppConstants.PageSize
	searchQuery := r.URL.Query().Get("search")

	loggedOut := r.URL.Query().Get("logged_out") == "true"

	email, _, permissions, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)

	var taskList []tasks.Task
	var totalTasks int
	var err error
	var userID *int

	isSearching := false

	// Get user ID if logged in
	if loggedIn {
		userID = getUserIDFromEmail(email)
	}

	if searchQuery != "" {
		taskList, totalTasks, err = tasks.SearchTasksForUser(page, pageSize, searchQuery, userID, timezone)
	} else {
		taskList, totalTasks, err = tasks.ReturnPaginationForUser(page, pageSize, userID, timezone)
	}

	if err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if searchQuery != "" {
		isSearching = true
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
		"LoggedIn":     loggedIn,
		"UserEmail":    email,
		"Permissions":  permissions,
		"LoggedOut":    loggedOut,
		"TotalTasks":   totalTasks,
		"TotalPages":   pagination.TotalPages,
		"IsSearching":  isSearching,
		"Title":        "GoTodo - Home",
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
	var userID *int
	var taskList []tasks.Task
	var totalTasks int
	var err error

	isSearching := false

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		var err error
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			page = 1
		}
	} else {
		page = 1
	}

	loggedOut := r.URL.Query().Get("logged_out") == "true"

	email, _, permissions, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)

	searchQuery := r.FormValue("search")

	if loggedIn {
		userID = getUserIDFromEmail(email)
	}

	if searchQuery != "" {
		isSearching = true
		taskList, totalTasks, err = tasks.SearchTasksForUser(page, pageSize, searchQuery, userID, timezone)
	} else {
		taskList, totalTasks, err = tasks.ReturnPaginationForUser(page, pageSize, userID, timezone)
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

	context := map[string]interface{}{
		"Tasks":        taskList,
		"TotalResults": totalTasks,
		"SearchQuery":  searchQuery,
		"CurrentPage":  page,
		"PreviousPage": pagination.PreviousPage,
		"NextPage":     pagination.NextPage,
		"PrevDisabled": pagination.PrevDisabled,
		"NextDisabled": pagination.NextDisabled,
		"TotalPages":   pagination.TotalPages,
		"LoggedIn":     loggedIn,
		"UserEmail":    email,
		"Permissions":  permissions,
		"LoggedOut":    loggedOut,
		"IsSearching":  isSearching,
		"TotalTasks":   totalTasks,
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
