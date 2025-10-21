package utils

import (
	"GoTodo/internal/sessionstore"
	"fmt"
	"net/http"
)

func GetSessionUser(r *http.Request) (email string, roleID int, loggedIn bool) {
	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		fmt.Printf("GetSessionUser error getting session: %v\n", err)
		return "", 0, false
	}

	emailVal, ok := session.Values["email"]
	if !ok {
		return "", 0, false
	}

	email, ok = emailVal.(string)
	if !ok {
		return "", 0, false
	}

	roleIDVal, ok := session.Values["role_id"]
	if !ok {
		return email, 0, true
	}

	roleID, ok = roleIDVal.(int)
	if !ok {
		return email, 0, true
	}
	return email, roleID, true
}

// RequireAuth is a middleware that checks if a user is logged in
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, loggedIn := GetSessionUser(r)
		if !loggedIn {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
