package tasks

import (
	"GoTodo/internal/storage"
	"fmt"
)

func DeleteTask() {
	db := storage.OpenDatebase()
	defer db.Close()

	fmt.Print("\nEnter task ID to delete: ")
	var id int
	_, err := fmt.Scan(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	stmt, err := db.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		fmt.Println("Error in DeleteTask (prepare):", err)
	}

	defer stmt.Close()
	_, err = stmt.Exec(id)

	if err != nil {
		fmt.Println("Error in DeleteTask (exec):", err)
	} else {
		fmt.Println("Task deleted\n")
		return
	}

	fmt.Println("Task not found")
}
