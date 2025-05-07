package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Request method: ", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	pageStr := r.FormValue("currentPage")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Default to page 1 if no valid page is provided
	}

	// fmt.Println("Page: ", page)

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := storage.OpenDatabase()
	if err != nil {
		fmt.Println("We failed to open the database.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new task into the database
	_, err = db.Exec(context.Background(), "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3)", title, description, false)
	if err != nil {
		fmt.Println("We failed to insert into the database.")
		fmt.Println("Failed values:", title, description, false)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//utils.AppConstants.PageSize := 15
	_, totalTasks, err := tasks.ReturnPagination(page, utils.AppConstants.PageSize)
	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "task-added")

	if page*utils.AppConstants.PageSize >= totalTasks {
		// Last page, so add the new task to the response
		tasks, _, err := tasks.ReturnPagination(page, utils.AppConstants.PageSize)

		if err != nil {
			http.Error(w, "Error rendering task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}
		prevDisabled := ""
		if page == 1 {
			prevDisabled = "disabled" // Disable on the first page
		}
		prevPage := page - 1
		if prevPage < 1 {
			prevPage = 1
		}
		// Render just the new task
		context := map[string]interface{}{
			"Tasks":        tasks,
			"PreviousPage": prevPage,
			"NextPage":     page,
			"CurrentPage":  page,
			"PrevDisabled": prevDisabled,
			"NextDisabled": "disabled",
		}

		err = utils.RenderTemplate(w, "pagination.html", context)
		if err != nil {
			fmt.Println("Error executing task partial: ", err)
			http.Error(w, "Error executing task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		// Not on the last page; no update needed
		// Last page, so add the new task to the response
		tasks, _, err := tasks.ReturnPagination(page, utils.AppConstants.PageSize)

		if err != nil {
			http.Error(w, "Error rendering task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}
		prevDisabled := ""
		if page == 1 {
			prevDisabled = "disabled" // Disable on the first page
		}

		prevPage := page - 1
		if prevPage < 1 {
			prevPage = 1
		}
		// Render just the new task
		context := map[string]interface{}{
			"Tasks":        tasks,
			"PreviousPage": prevPage,
			"NextPage":     page + 1,
			"CurrentPage":  page,
			"PrevDisabled": prevDisabled,
			"NextDisabled": "",
		}

		err = utils.RenderTemplate(w, "pagination.html", context)
		if err != nil {
			fmt.Println("Error executing task partial: ", err)
			http.Error(w, "Error executing task partial: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("New page is now added")
	}
}
