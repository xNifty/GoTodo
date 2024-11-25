package server

import (
	//"GoTodo/internal/storage"
	"GoTodo/internal/server/handlers"
	"GoTodo/internal/server/utils"
	"fmt"
	"html/template"
	"net/http"
)

func StartServer() {

	utils.Templates = template.Must(template.ParseGlob("internal/server/templates/*.html"))
	utils.Templates = template.Must(utils.Templates.ParseGlob("internal/server/templates/partials/*.html"))

	fs := http.FileServer(http.Dir("internal/server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/api/fetch-tasks", handlers.APIReturnTasks)
	http.HandleFunc("/api/add-task", handlers.APIAddTask)
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
