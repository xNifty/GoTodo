package tasks

import (
	"GoTodo/internal/storage"
	"fmt"
)

func MarkTaskComplete() {
	db := storage.OpenDatebase()
	defer db.Close()

	fmt.Print("\nEnter task ID to mark as complete: ")
	var id int
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	stmt, err := db.Prepare("UPDATE tasks SET completed = 1 WHERE id = ?")

	if err != nil {
		fmt.Println("Error in MarkTaskComplete (prepare):", err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		fmt.Println("Error in MarkTaskComplete (exec):", err)
	} else {
		fmt.Print("Task marked as complete!")
		return
	}

	fmt.Println("Task not found")
}

func MarkTaskIncomplete() {
	db := storage.OpenDatebase()
	defer db.Close()

	fmt.Print("\nEnter task ID to mark as incomplete: ")
	var id int
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	stmt, err := db.Prepare("UPDATE tasks SET completed = 0 WHERE id = ?")

	if err != nil {
		fmt.Println("Error in MarkTaskIncomplete (prepare):", err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		fmt.Println("Error in MarkTaskIncomplete (exec):", err)
	} else {
		fmt.Println("Task marked as incomplete!")
		return
	}

	fmt.Println("Task not found")
}
