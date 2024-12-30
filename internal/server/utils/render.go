package utils

import (
	"fmt"
	"html/template"
	"net/http"
)

var Templates *template.Template

func InitializeTemplates() error {
	var err error
	Templates, err = template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseGlob("internal/server/templates/*.html")
	if err != nil {
		return err
	}
	_, err = Templates.ParseGlob("internal/server/templates/partials/*.html")
	return err
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	err := Templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		fmt.Println("Error parsing template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
