package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"net/http"
)

func APILogout(w http.ResponseWriter, r *http.Request) {
	session, err := sessionstore.Store.Get(r, "session")
	if err == nil {
		session.Values = make(map[interface{}]interface{})
		session.Options.MaxAge = -1
		session.Save(r, w)
	}

	basePath := utils.GetBasePath()

	// Redirect with logout parameter
	w.Header().Set("HX-Redirect", basePath+"/?logged_out=true")
	w.WriteHeader(http.StatusOK)
}
