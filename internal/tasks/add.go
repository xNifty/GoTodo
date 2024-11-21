package tasks

import (
	"GoTodo/internal/storage"
	"bufio"
	"fmt"
	"os"
)

func AddTask() {
	db := storage.OpenDatebase()
	defer db.Close()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter task title: ")
	scanner.Scan()
	title := scanner.Text()

	fmt.Print("Enter task description (optional): ")
	scanner.Scan()
	description := scanner.Text()

	if title == "" {
		fmt.Println("Title cannot be empty")
		return
	}

	stmt, err := db.Prepare("INSERT INTO tasks (title, description, completed) VALUES (?, ?, 0)")
	if err != nil {
		fmt.Println("Error in AddTask (prepare):", err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(title, description)

	if err != nil {
		fmt.Println("Error in AddTask (exec):", err)
	}

	fmt.Println("\nTask added successfully!")
	db.Close()
}
