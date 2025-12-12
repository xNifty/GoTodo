package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"context"
	"net/http"
	"strings"
)

// APIUpdateProfile updates the user's name and timezone
func APIUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, _, _, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userName := r.FormValue("user_name")
	timezone := r.FormValue("timezone")
	if userName == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if timezone == "" {
		http.Error(w, "Timezone is required", http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), "UPDATE users SET user_name = $1, timezone = $2 WHERE email = $3", userName, timezone, email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		if strings.Contains(err.Error(), "securecookie: expired timestamp") {
			sessionstore.ClearSessionCookie(w, r)
			// Require re-login when session cookie was expired
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	session.Values["user_name"] = userName
	session.Values["timezone"] = timezone
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, permissions, timezone, loggedIn, user_name := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	status := r.URL.Query().Get("status")
	var statusMsg string
	if status == "success" {
		statusMsg = "Timezone updated successfully!"
	}

	context := map[string]interface{}{
		"UserEmail":   email,
		"Email":       email,
		"Timezone":    timezone,
		"Status":      statusMsg,
		"Name":        user_name,
		"LoggedIn":    loggedIn,
		"Permissions": permissions,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, r, "profile.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func APIUpdateTimezone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, _, _, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	timezone := r.FormValue("timezone")
	if timezone == "" {
		http.Error(w, "Timezone is required", http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), "UPDATE users SET timezone = $1 WHERE email = $2", timezone, email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		if strings.Contains(err.Error(), "securecookie: expired timestamp") {
			sessionstore.ClearSessionCookie(w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.Values["timezone"] = timezone
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	basePath := utils.GetBasePath()

	http.Redirect(w, r, basePath+"/profile?status=success", http.StatusSeeOther)
}
