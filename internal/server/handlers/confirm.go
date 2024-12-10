package handlers

import (
	"html/template"
	"net/http"
)

func APIConfirmDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	modalTemplate, err := template.ParseFiles("internal/server/templates/partials/confirm.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		ID string
	}{
		ID: id,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = modalTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
