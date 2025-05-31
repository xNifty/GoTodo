package handlers

import (
	"fmt"
	"net/http"
)

func ValidateDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if len(description) > MaxDescriptionLength {
		w.Header().Set("HX-Trigger", "description-error")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Description must be %d characters or less", MaxDescriptionLength)
		return
	}

	// If validation passes, return empty response
	w.WriteHeader(http.StatusOK)
}
