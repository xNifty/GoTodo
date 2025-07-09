package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "about.html", nil)
}
