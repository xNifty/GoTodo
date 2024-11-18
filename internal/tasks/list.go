package tasks

import (
	"GoTodo/internal/storage"
	"fmt"
)

func ListTasks() {
	db := storage.OpenDatebase()
	defer db.Close()

	rows, err := db.Query("SELECT id, title, description, completed FROM tasks")

	if err != nil {
		fmt.Println("Error in ListTasks (query):", err)
	}

	defer rows.Close()

	fmt.Println("\nTasks:")
	for rows.Next() {
		var id int
		var title string
		var description string
		var completed bool

		err = rows.Scan(&id, &title, &description, &completed)

		if err != nil {
			fmt.Println("Error in ListTasks (scan):", err)
		}

		status := "Incomplete"
		if completed {
			status = "Complete"
		}
		fmt.Printf("%d. %s: %s (%s)\n", id, title, description, status)
	}
	fmt.Println()
}
