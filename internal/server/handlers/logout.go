package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/sessionstore"
	"net/http"
	"strings"
)

func APILogout(w http.ResponseWriter, r *http.Request) {
	// Try to get the session. If there's an expired securecookie error, clear it and try again.
	var err error
	if sessionstore.Store != nil {
		// Use the cookie store to get session
		sess, e := sessionstore.Store.Get(r, "session")
		err = e
		if err != nil {
			if strings.Contains(err.Error(), "securecookie: expired timestamp") {
				sessionstore.ClearSessionCookie(w, r)
				// try again
				sess, err = sessionstore.Store.Get(r, "session")
			}
		}
		if err == nil && sess != nil {
			// clear values and expire
			sess.Values = make(map[interface{}]interface{})
			sess.Options.MaxAge = -1
			_ = sess.Save(r, w)
		}
	}

	basePath := utils.GetBasePath()

	// Redirect with logout parameter
	w.Header().Set("HX-Redirect", basePath+"/?logged_out=true")
	w.WriteHeader(http.StatusOK)
}
