package tasks

import (
	"GoTodo/internal/storage"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const RED = "\033[31m"
const GREEN = "\033[32m"
const RESET = "\033[0m"

func ListTasks() {
	db := storage.OpenDatebase()
	defer db.Close()

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	rows, err := db.Query("SELECT id, title, description, completed FROM tasks")

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

		err = rows.Scan(&id, &title, &description, &completed)

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
	db := storage.OpenDatebase()
	defer db.Close()

	var tasks []Task

	rows, err := db.Query("SELECT id, title, description, completed FROM tasks")

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
	fmt.Println(tasks)
	return tasks
}
