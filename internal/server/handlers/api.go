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

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div id="status" style="background-color: #d4edda; color: #155724; padding: 10px; margin-bottom: 10px; border: 1px solid #c3e6cb;">Task added successfully.</div>`)
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
