package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

// APIGetLoginPartial renders the login partial with basePath available
func APIGetLoginPartial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := utils.Templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
