package tasks

import (
	"GoTodo/internal/storage"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const (
	RED   = "\033[31m"
	GREEN = "\033[32m"
	RESET = "\033[0m"
)

func ListTasks() {
	pool, _ := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	rows, err := pool.Query(context.Background(), "SELECT id, title, description, completed, time_stamp FROM tasks ORDER BY id")
	if err != nil {
		fmt.Println("Error in ListTasks (query):", err)
	}
	defer rows.Close()

	headers := []string{"ID", "Title", "Description", "Status"}

	fmt.Fprintln(writer, "\n"+strings.Join(headers, "\t"))

	underlines := make([]string, len(headers))
	for i, header := range headers {
		underlines[i] = strings.Repeat("-", len(header))
	}

	fmt.Fprintf(writer, "%s", strings.Join(underlines, "\t")+"\n")

	for rows.Next() {
		var id int
		var title string
		var description string
		var completed bool

		err := rows.Scan(&id, &title, &description, &completed)

		if err != nil {
			fmt.Println("Error in ListTasks (scan):", err)
		}

		status := RED + "Incomplete" + RESET
		if completed {
			status = GREEN + "Complete" + RESET
		}
		fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", id, title, description, status)
	}
	writer.Flush()
	fmt.Println()
}

func ReturnTaskList() []Task {
	pool, _ := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)

	var tasks []Task

	rows, err := pool.Query(context.Background(), "SELECT id, title, description, completed, time_stamp FROM tasks ORDER BY id")

	if err != nil {
		fmt.Println("Error in ListTasks (query):", err)
		return tasks
	}

	defer rows.Close()

	for rows.Next() {
		var task Task

		err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed)

		if err != nil {
			fmt.Println("Error in ListTasks (scan):", err)
			return tasks
		}
		tasks = append(tasks, task)

	}
	return tasks
}

func ReturnPagination(page, pageSize int) ([]Task, int, error) {
	return ReturnPaginationForUser(page, pageSize, nil)
}

func ReturnPaginationForUser(page, pageSize int, userID *int) ([]Task, int, error) {
	pool, err := storage.OpenDatabase()
	if err != nil {
		return nil, 0, err
	}
	defer storage.CloseDatabase(pool)

	var tasks []Task
	offset := (page - 1) * pageSize

	// Build query based on whether user is logged in
	query := `SELECT id, title, description, completed, 
		TO_CHAR(time_stamp, 'YYYY/MM/DD HH:MI AM') AS date_added 
		FROM tasks `

	var countQuery string
	var rows interface {
		Next() bool
		Scan(...interface{}) error
		Close()
	}

	if userID == nil {
		// Not logged in - don't show any tasks
		return tasks, 0, nil
	}

	// Logged in - filter by user_id
	query += `WHERE user_id = $3 ORDER BY id LIMIT $1 OFFSET $2`
	rows, err = pool.Query(context.Background(), query, pageSize, offset, *userID)
	if err != nil {
		return nil, 0, err
	}
	countQuery = "SELECT COUNT(*) FROM tasks WHERE user_id = $1"
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.DateAdded); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	// Fetch total task count for pagination controls
	var totalTasks int
	err = pool.QueryRow(context.Background(), countQuery, *userID).Scan(&totalTasks)
	if err != nil {
		return nil, 0, err
	}
	return tasks, totalTasks, nil
}

func SearchTasks(page, pageSize int, searchQuery string) ([]Task, int, error) {
	return SearchTasksForUser(page, pageSize, searchQuery, nil)
}

func SearchTasksForUser(page, pageSize int, searchQuery string, userID *int) ([]Task, int, error) {
	pool, err := storage.OpenDatabase()
	if err != nil {
		return nil, 0, err
	}

	defer storage.CloseDatabase(pool)

	var tasks []Task
	offset := (page - 1) * pageSize
	searchPattern := "%" + searchQuery + "%"

	// If not logged in, return empty results
	if userID == nil {
		return tasks, 0, nil
	}

	rows, err := pool.Query(context.Background(),
		`SELECT id,
			title, 
			description,
			completed, 
			TO_CHAR(time_stamp, 'YYYY/MM/DD HH:MM AM')  as date_added
		 FROM tasks 
		 WHERE (title ILIKE $1 OR description ILIKE $1) AND user_id = $4
		 ORDER BY id 
		 LIMIT $2 OFFSET $3`,
		searchPattern, pageSize, offset, *userID)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var totalTasks int = 0

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.DateAdded); err != nil {
			return nil, 0, err
		}
		totalTasks++
		tasks = append(tasks, task)
	}

	return tasks, totalTasks, nil

}
