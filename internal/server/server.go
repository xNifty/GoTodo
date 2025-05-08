package server

import (
	//"GoTodo/internal/storage"
	"GoTodo/internal/server/handlers"
	"GoTodo/internal/server/utils"
	"fmt"

	// "html/template"
	"net/http"
)

// Literally just used to prevent favicon.ico from being requestedi
// TODO:: Add a favicon
func doNothing(w http.ResponseWriter, r *http.Request) {}

func StartServer() error {
	err := utils.InitializeTemplates()
	if err != nil {
		return fmt.Errorf("failed to initialize templates: %v", err)
	}

	fs := http.FileServer(http.Dir("internal/server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/favicon.ico", doNothing)
	http.HandleFunc("/api/fetch-tasks", handlers.APIReturnTasks)
	http.HandleFunc("/api/add-task", handlers.APIAddTask)
	http.HandleFunc("/api/confirm", handlers.APIConfirmDelete)
	http.HandleFunc("/api/delete-task", handlers.APIDeleteTask)
	http.HandleFunc("/api/get-next-item", handlers.APIGetNextItem)
	http.HandleFunc("/api/update-status", handlers.APIUpdateTaskStatus)
	http.HandleFunc("/about", handlers.AboutHandler)
	http.HandleFunc("/search", handlers.SearchHandler)

	fmt.Println("Starting server on :8080")
	return http.ListenAndServe(":8080", nil)
}
