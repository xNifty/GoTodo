package utils

import "GoTodo/internal/tasks"

type PaginationData struct {
	PreviousPage         int
	NextPage             int
	CurrentPage          int
	PrevDisabled         string
	NextDisabled         string
	TotalPages           int
	Pages                []int
	HasRightEllipsis     bool
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

	// Build a small sliding window of page numbers to display.
	// Behavior: show pages 1..windowSize at the start; once current >= windowSize,
	// start the window at the current page. Always offer an ellipsis + last page
	// when there are pages beyond the window.
	windowSize := 4
	var start int
	if page >= windowSize {
		start = page
	} else {
		start = 1
	}
	end := start + windowSize - 1
	if end > totalPages {
		end = totalPages
	}

	pages := make([]int, 0)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	hasRightEllipsis := false
	if end < totalPages {
		hasRightEllipsis = true
	}

	return PaginationData{
		PreviousPage:         prevPage,
		NextPage:             nextPage,
		CurrentPage:          page,
		PrevDisabled:         prevDisabled,
		NextDisabled:         nextDisabled,
		TotalPages:           totalPages,
		Pages:                pages,
		HasRightEllipsis:     hasRightEllipsis,
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
