package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "index.html", map[string]string{
		"Title": "Home",
	})
}
