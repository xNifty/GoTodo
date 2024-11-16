package tasks

import (
        "fmt"
        "os"
        "bufio"
)

// AddTask handles adding a new task to the list and returns the updated next ID.
func AddTask(taskList *[]Task, nextID int) int {
        scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter task title: ")
	scanner.Scan()
        title := scanner.Text()

	fmt.Print("Enter task description (optional): ")
	scanner.Scan()
        description := scanner.Text()

	newTask := Task{
		ID:          nextID,
		Title:       title,
		Description: description,
		Completed:   false,
	}

	if err := newTask.Validate(); err != nil {
		fmt.Println("Error:", err)
		return nextID
	}

	*taskList = append(*taskList, newTask)
	fmt.Println("Task added successfully!")

	return nextID + 1
}
