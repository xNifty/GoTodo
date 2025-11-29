package utils

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

var Templates *template.Template
var BasePath string

func InitializeTemplates() error {
	var err error
	BasePath = os.Getenv("BASE_PATH")
	if BasePath == "" {
		BasePath = "/"
	}

	BasePath = strings.TrimSuffix(BasePath, "/")
	if BasePath == "" {
		BasePath = "/"
	}

	Templates, err = template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"hasPermission": func(permissions []string, permission string) bool {
			for _, p := range permissions {
				if p == permission {
					return true
				}
			}
			return false
		},
		"basePath": func() string {
			return GetBasePath()
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
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// GetBasePath returns the base path for use in templates
func GetBasePath() string {
	return BasePath
}
