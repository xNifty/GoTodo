package handlers

import (
	"GoTodo/internal/storage"
	"context"
	"fmt"
	"net/http"
)

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
	fmt.Fprintf(w, `<button 
        class="badge %s"
        hx-get="/api/update-status?id=%s" 
        hx-target="#task-%s .badge" 
        hx-swap="outerHTML"
        style="cursor: pointer;">
        %s
    </button>`,
		map[bool]string{true: "bg-success", false: "bg-danger"}[updatedStatus],
		id,
		id,
		map[bool]string{true: "Complete", false: "Incomplete"}[updatedStatus],
	)
}
