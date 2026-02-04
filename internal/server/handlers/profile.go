package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
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
	itemsPerPageStr := r.FormValue("items_per_page")
	var itemsPerPage int
	if itemsPerPageStr != "" {
		if v, err := strconv.Atoi(itemsPerPageStr); err == nil {
			itemsPerPage = v
		}
	}
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

	if itemsPerPage > 0 {
		_, err = db.Exec(context.Background(), "UPDATE users SET user_name = $1, timezone = $2, items_per_page = $3 WHERE email = $4", userName, timezone, itemsPerPage, email)
	} else {
		_, err = db.Exec(context.Background(), "UPDATE users SET user_name = $1, timezone = $2 WHERE email = $3", userName, timezone, email)
	}
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
	if itemsPerPage > 0 {
		session.Values["items_per_page"] = itemsPerPage
	}
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

	// fetch user to get items_per_page
	user, _ := storage.GetUserByEmail(email)
	itemsPerPage := 15
	if user != nil && user.ItemsPerPage > 0 {
		itemsPerPage = user.ItemsPerPage
	}

	context := map[string]interface{}{
		"UserEmail":    email,
		"Email":        email,
		"Timezone":     timezone,
		"Status":       statusMsg,
		"Name":         user_name,
		"ItemsPerPage": itemsPerPage,
		"LoggedIn":     loggedIn,
		"Permissions":  permissions,
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

// APIChangePassword allows a user to change their password
func APIChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email, _, _, _, loggedIn, _ := utils.GetSessionUserWithTimezone(r)
	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate all fields are provided
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "All password fields are required")
		return
	}

	// Validate new passwords match
	if newPassword != confirmPassword {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "New passwords do not match")
		return
	}

	// Validate new password length
	if len(newPassword) < 8 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "New password must be at least 8 characters long")
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Internal server error")
		return
	}
	defer db.Close()

	// Get user's current password hash
	var currentHashedPassword string
	err = db.QueryRow(context.Background(), "SELECT password FROM users WHERE email = $1", email).Scan(&currentHashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Internal server error")
		return
	}

	// Verify current password is correct
	err = bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(currentPassword))
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Current password is incorrect")
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Internal server error")
		return
	}

	// Update password in database
	_, err = db.Exec(context.Background(), "UPDATE users SET password = $1 WHERE email = $2", string(hashedPassword), email)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Internal server error")
		return
	}

	// Send password changed email
	siteName := "GoTodo"
	if settings, err := storage.GetSiteSettings(); err == nil && settings != nil && settings.SiteName != "" {
		siteName = settings.SiteName
	}
	subject := fmt.Sprintf("%s - Password Changed", siteName)
	body := fmt.Sprintf(`Hello,

Your password has been changed for %s

If you did not request this, please reach out to support.

This email cannot receive replies. Please do not reply to this email.
`, siteName)

	err = utils.SendEmail(subject, body, email)
	if err != nil {
		fmt.Printf("Warning: Failed to send password changed email to %s: %v\n", email, err)
		// Continue anyway - password was updated successfully
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "success")
}
