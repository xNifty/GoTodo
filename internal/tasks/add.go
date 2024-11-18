package tasks

import (
	"GoTodo/internal/storage"
	"bufio"
	"fmt"
	"os"
)

// AddTask handles adding a new task to the list and returns the updated next ID.
func AddTask(nextID int) int {
	db := storage.OpenDatebase()
	defer db.Close()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter task title: ")
	scanner.Scan()
	title := scanner.Text()

	fmt.Print("Enter task description (optional): ")
	scanner.Scan()
	description := scanner.Text()

	if title == "" || description == "" {
		fmt.Println("Title cannot be empty")
		return nextID
	}

	stmt, err := db.Prepare("INSERT INTO tasks (title, description, completed) VALUES (?, ?, 0)")
	if err != nil {
		fmt.Println("Error in AddTask (prepare):", err)
	}

	defer stmt.Close()

	//fmt.Printf("title %s, description: %s", title, description)
	_, err = stmt.Exec(title, description)

	//_, err = stmt.Exec(title, description, false)
	//fmt.Printf("call: %v", call)
	if err != nil {
		fmt.Println("Error in AddTask (exec):", err)
	}
	ListTasks()
	//*taskList = append(*taskList, newTask)
	fmt.Println("\nTask added successfully!\n")
	db.Close()
	return nextID + 1
}
