package server

import (
	//"GoTodo/internal/storage"
	"GoTodo/internal/server/handlers"
	"GoTodo/internal/server/utils"
	"fmt"
	"html/template"
	"net/http"
)

func doNothing(w http.ResponseWriter, r *http.Request) {}

func StartServer() {

	utils.Templates = template.Must(template.ParseGlob("internal/server/templates/*.html"))
	utils.Templates = template.Must(utils.Templates.ParseGlob("internal/server/templates/partials/*.html"))

	fs := http.FileServer(http.Dir("internal/server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/favicon.ico", doNothing)
	http.HandleFunc("/api/fetch-tasks", handlers.APIReturnTasks)
	http.HandleFunc("/api/add-task", handlers.APIAddTask)
	http.HandleFunc("/api/confirm", handlers.APIConfirmDelete)
	http.HandleFunc("/api/delete-task", handlers.APIDeleteTask)
	http.HandleFunc("/api/update-status", handlers.APIUpdateTaskStatus)
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
