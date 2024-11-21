package handlers

import (
	"GoTodo/internal/tasks"
	"encoding/json"
	"net/http"
)

func APIHandler(w http.ResponseWriter, r *http.Request) {
	tasks := tasks.ReturnTaskList()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
