package main

import (
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"fmt"
	"os"
)

func main() {
	//manager := tasks.NewTaskManager()

	nextId := 1

	var taskList []tasks.Task

	storage.CreateDatabase()

	for {
		fmt.Println("TODO App")
		fmt.Println("1. List Tasks")
		fmt.Println("2. Add Task")
		fmt.Println("3. Complete Task")
		fmt.Println("4. Incomplete Task")
		fmt.Println("5. Delete Task")
		fmt.Println("6. Exit")
		fmt.Println("7. List Current Id")
		fmt.Print("Enter your choice: ")

		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("Invalid choice\n")
			continue
		}

		switch choice {
		case 1:
			tasks.ListTasks()
		case 2:
			nextId = tasks.AddTask(nextId)
		case 3:
			tasks.MarkTaskComplete()
		case 4:
			tasks.MarkTaskIncomplete()
		case 5:
			tasks.DeleteTask(&taskList)
		case 6:
			fmt.Println("Exiting...\n")
			os.Exit(0)
		case 7:
			fmt.Println("Current Id: \n", nextId)
		default:
			fmt.Println("Invalid choice\n")
		}
	}
}
