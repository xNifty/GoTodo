package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/tasks"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	taskList := tasks.ReturnTaskList()
	utils.RenderTemplate(w, "index.html", map[string]interface{}{
		"Title": "TODO App",
		"Tasks": taskList,
	})
}
