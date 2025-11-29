package handlers

import (
	"GoTodo/internal/sessionstore"
	"GoTodo/internal/storage"
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func APILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		w.Header().Set("HX-Retarget", "#login-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Email and password are required.")
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error.")
		return
	}
	defer db.Close()

	var hashedPassword string
	var roleID int
	row := db.QueryRow(context.Background(), "SELECT password, role_id FROM users WHERE email = $1", email)
	err = row.Scan(&hashedPassword, &roleID)
	if err != nil {
		w.Header().Set("HX-Retarget", "#login-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Invalid username or password.")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		w.Header().Set("HX-Retarget", "#login-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Invalid username or password.")
		return
	}

	permissions, err := storage.GetPermissionsByRoleID(roleID)
	if err != nil {
		fmt.Printf("Error fetching permissions: %v\n", err)
		permissions = []string{}
	}

	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		fmt.Printf("Error getting session: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Session error.")
		return
	}

	session.Values["email"] = email
	session.Values["role_id"] = roleID
	session.Values["permissions"] = permissions

	err = session.Save(r, w)
	if err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to save session.")
		return
	}

	w.Header().Set("HX-Trigger", "login-success")
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, " ")
}
