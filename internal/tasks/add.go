package tasks

import (
        "fmt"
)

// AddTask handles adding a new task to the list and returns the updated next ID.
func AddTask(taskList *[]Task, nextID int) int {
        var title string
        var description string

	fmt.Print("Enter task title: ")
        fmt.Scanf("%s\n", &title)

	fmt.Print("Enter task description (optional): ")
        fmt.Scanf("%s\n", &description)

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
