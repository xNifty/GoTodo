package tasks

import (
	"GoTodo/internal/storage"
	"context"
	"fmt"
)

func MarkTaskComplete() {
	pool := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)

	fmt.Print("\nEnter task ID to mark as complete: ")
	var id int
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	_, err = pool.Exec(context.Background(), "UPDATE tasks SET completed = true WHERE id = $1", id)

	if err != nil {
		fmt.Println("Error in MarkTaskComplete (prepare):", err)
	} else {
		fmt.Println("Task marked as complete!")
		return
	}

	fmt.Println("Task not found")
}

func MarkTaskIncomplete() {
	pool := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)

	fmt.Print("\nEnter task ID to mark as incomplete: ")
	var id int
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	_, err = pool.Exec(context.Background(), "UPDATE tasks SET completed = false WHERE id = $1", id)

	if err != nil {
		fmt.Println("Error in MarkTaskIncomplete (prepare):", err)
	} else {
		fmt.Println("Task marked as incomplete!")
		return
	}

	fmt.Println("Task not found")
}
