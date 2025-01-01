package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/tasks"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func SortHandler(w http.ResponseWriter, r *http.Request) {
	column := r.URL.Query().Get("column")
	direction := r.URL.Query().Get("direction")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize := utils.AppConstants.PageSize

	taskList, totalTasks, err := tasks.ReturnPagination(page, pageSize)

	if err != nil {
		http.Error(w, "Error fetching tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch column {
	case "id":
		sort.Slice(taskList, func(i, j int) bool {
			if direction == "asc" {
				return taskList[i].ID < taskList[j].ID
			}
			return taskList[i].ID > taskList[j].ID
		})
	case "date":
		sort.Slice(taskList, func(i, j int) bool {
			dateA, errA := time.Parse(time.RFC3339, taskList[i].DateAdded)
			dateB, errB := time.Parse(time.RFC3339, taskList[j].DateAdded)
			if errA != nil || errB != nil {
				return false
			}
			if direction == "asc" {
				return dateA.Before(dateB)
			}
			return dateA.After(dateB)
		})
	}

	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalTasks {
		nextDisabled = "disabled"
	}

	context := map[string]interface{}{
		"Tasks":        taskList,
		"PreviousPage": page - 1,
		"NextPage":     page + 1,
		"CurrentPage":  page,
		"PrevDisabled": prevDisabled,
		"NextDisabled": nextDisabled,
	}

	if err := utils.RenderTemplate(w, "pagination.html", context); err != nil {
		http.Error(w, "Error executing task partial: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
