package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"context"
	"net/http"
)

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, _, timezone, loggedIn := utils.GetSessionUserWithTimezone(r)
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
		"Email":    email,
		"Timezone": timezone,
		"Status":   statusMsg,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := utils.RenderTemplate(w, "profile.html", context); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func APIUpdateTimezone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, _, _, loggedIn := utils.GetSessionUserWithTimezone(r)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.Values["timezone"] = timezone
	err = session.Save(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?status=success", http.StatusSeeOther)
}
