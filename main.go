package main

import (
	"fmt"
	"os"
	"GoTodo/internal/tasks"
)

func main() {
	//manager := tasks.NewTaskManager()

	nextId := 1

	var taskList []tasks.Task

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
			fmt.Println("Invalid choice")
			continue
		}

		switch choice {
		case 1:
			fmt.Println(taskList)
		case 2:
			nextId = tasks.AddTask(&taskList, nextId)
		//case 3:
		//	markTaskComplete(&taskList)
		//case 4:
		//	deleteTask(&taskList)
		//case 5:
		//	markTaskIncomplete(&taskList)
		case 6:
			fmt.Println("Exiting...")
			os.Exit(0)
		case 7:
			fmt.Println("Current Id: ", nextId)
		default:
			fmt.Println("Invalid choice")
		}
	}
}
