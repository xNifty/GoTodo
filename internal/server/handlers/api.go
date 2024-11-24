package handlers

import (
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func APIReturnTasks(w http.ResponseWriter, r *http.Request) {
	tasks := tasks.ReturnTaskList()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `<div id="feedback" style="background-color: #f8d7da; color: #721c24; padding: 10px; margin-bottom: 10px; border: 1px solid #f5c6cb;">Title is required.</div>`)
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
	fmt.Fprintf(w, `<div id="feedback" style="background-color: #d4edda; color: #155724; padding: 10px; margin-bottom: 10px; border: 1px solid #c3e6cb;">Task added successfully.</div>`)
}
