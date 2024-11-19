package main

/*
	Create function to grab next ID from database
*/

import (
	"GoTodo/internal/server"
	"GoTodo/internal/storage"
	"GoTodo/internal/tasks"
	"fmt"
	"os"
)

func main() {
	//manager := tasks.NewTaskManager()

	runArgs := os.Args
	storage.CreateDatabase()

	if runArgs != nil && len(runArgs) > 1 {
		if runArgs[1] == "server" {
			server.StartServer()
		}
	}

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
			fmt.Print("Invalid choice\n")
			continue
		}

		switch choice {
		case 1:
			tasks.ListTasks()
		case 2:
			tasks.AddTask()
		case 3:
			tasks.MarkTaskComplete()
		case 4:
			tasks.MarkTaskIncomplete()
		case 5:
			tasks.DeleteTask()
		case 6:
			fmt.Println("\nSee you next time!\n")
			os.Exit(0)
		case 7:
			nextId := storage.GetNextID()
			fmt.Printf("\nNext ID: %d\n\n", nextId)
		default:
			fmt.Println("Invalid choice\n")
		}
	}
}
