package tasks

import (
	"GoTodo/internal/storage"
	"context"
	"fmt"
)

func DeleteTask(id string) (bool, error) {
	pool, err := storage.OpenDatabase()
	defer pool.Close()

	_, err = pool.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		fmt.Println("Error in DeleteTask (prepare):", err)
		return false, err
	}

	if err != nil {
		fmt.Println("Error in DeleteTask (exec):", err)
		return false, err
	} else {
		fmt.Println("Task deleted")
		return true, nil
	}
}
