package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	email, _, permissions, loggedIn := utils.GetSessionUser(r)

	context := map[string]interface{}{
		"LoggedIn":   loggedIn,
		"UserEmail":  email,
		"Permissions": permissions,
	}

	utils.RenderTemplate(w, "about.html", context)
}
