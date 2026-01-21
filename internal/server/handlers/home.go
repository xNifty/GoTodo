package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"net/http"
	"regexp"
	"strconv"
)

func getUserIDFromEmail(email string) *int {
	// First try to read user_id from the session (avoid extra DB lookup)
	// Note: we don't have *http.Request here, so callers may prefer using
	// utils.GetSessionUserID directly. This function remains for backward
	// compatibility and will perform a DB lookup by email if needed.
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
	// determine page size from session if present
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
	searchQuery := r.URL.Query().Get("search")

	loggedOut := r.URL.Query().Get("logged_out") == "true"
	accountCreated := r.URL.Query().Get("account_created") == "true"

	email, _, permissions, timezone, loggedIn, _ := utils.GetSessionUserWithTimezone(r)

	var taskList []tasks.Task
	var totalTasks int
	var err error
	var userID *int

	isSearching := false

	// Get user ID if logged in (prefer session-stored ID)
	if loggedIn {
		if uid := utils.GetSessionUserID(r); uid != nil {
			userID = uid
		} else {
			userID = getUserIDFromEmail(email)
		}
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

	// Split into favorite and non-favorite lists and set page number
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

	// Avoid dereferencing nil userID; use 0 for anonymous users
	uid := 0
	if userID != nil {
		uid = *userID
	}
	pagination := utils.GetPaginationData(page, pageSize, totalTasks, uid)

	// Create a context for the tasks and pagination
	context := map[string]interface{}{
		"FavoriteTasks":    favs,
		"Tasks":            nonFavs,
		"CurrentPage":      page,
		"PreviousPage":     pagination.PreviousPage,
		"NextPage":         pagination.NextPage,
		"PrevDisabled":     pagination.PrevDisabled,
		"NextDisabled":     pagination.NextDisabled,
		"Pages":            pagination.Pages,
		"HasRightEllipsis": pagination.HasRightEllipsis,
		"PerPage":          pageSize,
		"LoggedIn":         loggedIn,
		"UserEmail":        email,
		"Permissions":      permissions,
		"LoggedOut":        loggedOut,
		"AccountCreated":   accountCreated,
		"TotalTasks":       totalTasks,
		"TotalPages":       pagination.TotalPages,
		"IsSearching":      isSearching,
		"Title":            "GoTodo - Home",
		"CompletedTasks":   utils.GetCompletedTasksCount(userID),
		"IncompleteTasks":  utils.GetIncompleteTasksCount(userID),
	}

	// Include user's projects for the sidebar project select
	if loggedIn && userID != nil {
		if projs, err := storage.GetProjectsForUser(*userID); err == nil {
			projList := make([]map[string]interface{}, 0)
			for _, p := range projs {
				projList = append(projList, map[string]interface{}{"ID": p.ID, "Name": p.Name, "Selected": false})
			}
			context["Projects"] = projList
		}
	}

	// Render the tasks and pagination controls
	if err := utils.RenderTemplate(w, r, "index.html", context); err != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// determine page size from session if present
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
		if uid := utils.GetSessionUserID(r); uid != nil {
			userID = uid
		} else {
			userID = getUserIDFromEmail(email)
		}
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

	// Set the page number for each task and split into favorites/non-favorites
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

	// Avoid dereferencing nil userID; use 0 for anonymous users
	uid := 0
	if userID != nil {
		uid = *userID
	}
	pagination := utils.GetPaginationData(page, pageSize, totalTasks, uid)

	context := map[string]interface{}{
		"FavoriteTasks":    favs,
		"Tasks":            nonFavs,
		"TotalResults":     totalTasks,
		"SearchQuery":      searchQuery,
		"CurrentPage":      page,
		"PreviousPage":     pagination.PreviousPage,
		"NextPage":         pagination.NextPage,
		"PrevDisabled":     pagination.PrevDisabled,
		"NextDisabled":     pagination.NextDisabled,
		"TotalPages":       pagination.TotalPages,
		"Pages":            pagination.Pages,
		"HasRightEllipsis": pagination.HasRightEllipsis,
		"LoggedIn":         loggedIn,
		"UserEmail":        email,
		"Permissions":      permissions,
		"LoggedOut":        loggedOut,
		"IsSearching":      isSearching,
		"TotalTasks":       totalTasks,
		"CompletedTasks":   utils.GetCompletedTasksCount(userID),
		"IncompleteTasks":  utils.GetIncompleteTasksCount(userID),
	}

	if err := utils.RenderTemplate(w, r, "pagination.html", context); err != nil {
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
