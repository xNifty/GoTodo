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
	pool, err := storage.OpenDatabase()
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

	fmt.Fprintf(writer, strings.Join(underlines, "\t")+"\n")

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
	pool, err := storage.OpenDatabase()
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
	pool, err := storage.OpenDatabase()
	if err != nil {
		return nil, 0, err
	}
	defer storage.CloseDatabase(pool)

	var tasks []Task
	offset := (page - 1) * pageSize
	// Fetch paginated tasks
	rows, err := pool.Query(context.Background(),
		"SELECT id, title, description, completed, time_stamp FROM tasks ORDER BY id LIMIT $1 OFFSET $2",
		pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
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
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&totalTasks)
	if err != nil {
		return nil, 0, err
	}
	// fmt.Println("\nTotal tasks:", totalTasks)
	//fmt.Println("\nTasks:", tasks)
	return tasks, totalTasks, nil
}

func SearchTasks(page, pageSize int, searchQuery string) ([]Task, int, error) {
	pool, err := storage.OpenDatabase()
	if err != nil {
		return nil, 0, err
	}

	defer storage.CloseDatabase(pool)

	var tasks []Task
	offset := (page - 1) * pageSize
	searchQuery = "%" + searchQuery + "%"

	rows, err := pool.Query(context.Background(),
		"SELECT id, title, description, completed, time_stamp FROM tasks WHERE title ILIKE $1 OR description ILIKE $1 ORDER BY id LIMIT $2 OFFSET $3",
		searchQuery, pageSize, offset)

	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.DateAdded); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	var totalTasks int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks WHERE title ILIKE $1 OR description ILIKE $1", searchQuery).Scan(&totalTasks)

	if err != nil {
		return nil, 0, err
	}

	return tasks, totalTasks, nil

}
