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
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/api/", handlers.APIHandler)
	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
