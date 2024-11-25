package handlers

import (
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"fmt"
	"html/template"
	"net/http"
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

func APIReturnTasks(w http.ResponseWriter, r *http.Request) {
	tasks := tasks.ReturnTaskList()
	w.Header().Set("Content-Type", "text/html; character=utf-8")
	//fmt.Println(tasks)
	for _, task := range tasks {
		if err := taskTemplate.Execute(w, task); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `<div id="status" style="background-color: #f8d7da; color: #721c24; padding: 10px; margin-bottom: 10px; border: 1px solid #f5c6cb;">Title is required.</div>`)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error opening database")
		return
	}

	defer db.Close()

	_, err = db.Exec(context.Background(), "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3)", title, description, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error inserting task")
		return
	}

	tasks := tasks.ReturnTaskList()
	task := tasks[len(tasks)-1]
	w.Header().Set("Content-Type", "text/html; character=utf-8")

	taskPartialTemplate, err = template.ParseFiles("internal/server/templates/partials/todo.html")

	if err = taskPartialTemplate.Execute(w, task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, `<div id="status" style="background-color: #d4edda; color: #155724; padding: 10px; margin-bottom: 10px; border: 1px solid #c3e6cb;">Task added successfully.</div>`)
	//
	// APIReturnTasks(w, r)
}

func APIDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	res, err := tasks.DeleteTask(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		if res {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
	}
}

func APIConfirmDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Simulate fetching the task details (if necessary)
	task := struct {
		ID string
	}{
		ID: id,
	}

	// Render the modal content
	tmpl := template.Must(template.ParseFiles("internal/server/templates/partials/confirm.html"))
	if err := tmpl.Execute(w, task); err != nil {
		http.Error(w, "Failed to load confirmation modal", http.StatusInternalServerError)
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
