package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "internal/server/templates/about.html")
	utils.RenderTemplate(w, "about.html", nil)
}
