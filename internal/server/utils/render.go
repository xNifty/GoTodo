package utils

import (
	"fmt"
	"html/template"
	"net/http"
)

var Templates *template.Template

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	err := Templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		fmt.Println("Error parsing template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
