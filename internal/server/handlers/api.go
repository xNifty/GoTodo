package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

var taskTemplate = template.Must(template.New("task").Parse(`
<tr>
    <td>{{ .ID }}</td>
    <td>{{ .Title }}</td>
    <td>{{ .Description }}</td>
	<td>{{ if .Completed }}<font color="green">Complete</font>{{ else }}<font color="red">Incomplete</font>{{ end }}</td>
</tr>
`))

var taskPartialTemplate *template.Template

var pageSize int = 15

func APIReturnTasks(w http.ResponseWriter, r *http.Request) {
	//pageSize := 15

	var page int

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

	//fmt.Println("\nPage, early: ", page)

	// Fetch tasks for the current page
	tasks, totalTasks, err := tasks.ReturnPagination(page, pageSize)
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

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request method: ", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	pageStr := r.FormValue("currentPage")
	fmt.Println("currentPageStr: ", pageStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Default to page 1 if no valid page is provided
	}

	fmt.Println("Page: ", page)

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new task into the database
	_, err = db.Exec(context.Background(), "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3)", title, description, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pageSize := 15
	tasks, totalTasks, err := tasks.ReturnPagination(page, pageSize)
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalTasks {
		nextDisabled = "disabled"
	}

	// Prepare the context for pagination and task list
	context := map[string]interface{}{
		"Tasks":        tasks,
		"PreviousPage": page - 1,
		"NextPage":     page + 1,
		"CurrentPage":  pageStr,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// If we're on the last page, return only the updated task list (with new task at the end)
	if page*pageSize >= totalTasks {
		// Append the new task to the task list
		task := tasks[len(tasks)-1] // Get the newly added task
		taskPartialTemplate, err := template.ParseFiles("internal/server/templates/partials/todo.html")
		if err != nil {
			http.Error(w, "Error rendering task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Render the task as a partial (this will update the list immediately in HTMX)
		err = taskPartialTemplate.Execute(w, task)
		if err != nil {
			http.Error(w, "Error executing task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// For pages that aren't the last, just return the full task list and pagination
		err := utils.RenderTemplate(w, "pagination.html", context)
		if err != nil {
			http.Error(w, "Error rendering pagination template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func APIDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("id")
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

	// Return an empty HTML response for the deleted task row
	fmt.Fprintf(w, "<tr id=\"task-%s\"></tr>", taskID)
}

func APIConfirmDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	modalTemplate, err := template.ParseFiles("internal/server/templates/partials/confirm.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		ID string
	}{
		ID: id,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = modalTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func APIUpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var completed bool

	err = db.QueryRow(context.Background(), "SELECT completed FROM tasks WHERE id = $1", id).Scan(&completed)

	if err != nil {
		http.Error(w, "Task not found.", http.StatusInternalServerError)
		return
	}

	updatedStatus := !completed

	_, err = db.Exec(context.Background(), "UPDATE tasks SET completed = $1 WHERE id = $2", updatedStatus, id)

	if err != nil {
		http.Error(w, "Failed to update task status.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; character=utf-8")
	fmt.Fprintf(w, `<span 
        class="badge %s"
        hx-get="/api/update-task-status?id=%s" 
        hx-target="#task-%s .badge" 
        hx-swap="outerHTML"
        style="cursor: pointer;">
        %s
    </span>`,
		map[bool]string{true: "bg-success", false: "bg-secondary"}[updatedStatus],
		id,
		id,
		map[bool]string{true: "Complete", false: "Incomplete"}[updatedStatus],
	)
}
