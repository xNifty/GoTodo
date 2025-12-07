package utils

import "GoTodo/internal/tasks"

type PaginationData struct {
	PreviousPage         int
	NextPage             int
	CurrentPage          int
	PrevDisabled         string
	NextDisabled         string
	TotalPages           int
	TotalCompletedTasks  int
	TotalIncompleteTasks int
}

func GetPaginationData(page, pageSize, totalItems, userID int) PaginationData {
	prevDisabled := ""

	totalPages := (totalItems + pageSize - 1) / pageSize
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalItems {
		nextDisabled = "disabled"
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1
	if page*pageSize >= totalItems {
		nextPage = page
	}

	return PaginationData{
		PreviousPage:         prevPage,
		NextPage:             nextPage,
		CurrentPage:          page,
		PrevDisabled:         prevDisabled,
		NextDisabled:         nextDisabled,
		TotalPages:           totalPages,
		TotalCompletedTasks:  GetCompletedTasksCount(&userID),
		TotalIncompleteTasks: GetIncompleteTasksCount(&userID),
	}
}

func GetCompletedTasksCount(userID *int) int {
	count := 0
	var ts = tasks.ReturnTaskListForUser(userID)
	for _, task := range ts {
		if task.Completed {
			count++
		}
	}
	return count
}

func GetIncompleteTasksCount(userID *int) int {
	count := 0
	var ts = tasks.ReturnTaskListForUser(userID)
	for _, task := range ts {
		if !task.Completed {
			count++
		}
	}
	return count
}
