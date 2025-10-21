package handlers

import (
	"GoTodo/internal/storage"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render login form (assume template exists)
		http.ServeFile(w, r, "internal/server/templates/login.html")
		return
	}

	// POST: process login
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required.", http.StatusBadRequest)
		return
	}

	user, err := storage.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Invalid email or password.", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, "Invalid email or password.", http.StatusUnauthorized)
		return
	}

	session, _ := store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// You will need to implement storage.GetUserByEmail and create a login.html template.
