package tasks

import (
	"GoTodo/internal/storage"
	"bufio"
	"context"
	"fmt"
	"os"
)

func AddTask() {
	pool := storage.OpenDatabase()
	defer storage.CloseDatabase(pool)
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

	_, err := pool.Exec(context.Background(), "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, false)", title, description)
	if err != nil {
		fmt.Println("Error in AddTask (prepare):", err)
		return
	}

	fmt.Println("\nTask added successfully!")
}
