package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	email, _, loggedIn := utils.GetSessionUser(r)

	context := map[string]interface{}{
		"LoggedIn":  loggedIn,
		"UserEmail": email,
	}

	utils.RenderTemplate(w, "about.html", context)
}
