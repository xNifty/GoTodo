package tasks

import (
	"GoTodo/internal/storage"
	"context"
	"fmt"
)

func DeleteTask() {
	pool, err := storage.OpenDatabase()
	defer pool.Close()

	fmt.Print("\nEnter task ID to delete: ")
	var id int
	_, err = fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	_, err = pool.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		fmt.Println("Error in DeleteTask (prepare):", err)
	}

	if err != nil {
		fmt.Println("Error in DeleteTask (exec):", err)
	} else {
		fmt.Println("Task deleted")
		return
	}

	fmt.Println("Task not found")
}
