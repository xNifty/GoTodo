package utils

import (
	"GoTodo/internal/sessionstore"
	"net/http"
)

// IntOrZero returns the value of p or 0 if p is nil.
func IntOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// SetFlash adds a flash message to the session
func SetFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		return
	}
	session.AddFlash(message)
	_ = session.Save(r, w)
}
