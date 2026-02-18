package handlers

import (
	"GoTodo/internal/server/utils"
	"net/http"
)

func APIDismissAnnouncement(w http.ResponseWriter, r *http.Request) {
	session, err := utils.GetSession(r)
	if err != nil {
		http.Error(w, "Session Error", http.StatusInternalServerError)
		return
	}

	session.Values["announcement_dismissed"] = true

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session Save Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
