package utils

import (
	"net/http"
)

func RequireHTMX(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") != "true" {
			basePath := GetBasePath()
			http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
