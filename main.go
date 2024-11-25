package main

import (
	"GoTodo/internal/server"
	"GoTodo/internal/storage"
	// "GoTodo/internal/tasks"
	// "fmt"
	//"os"
)

func main() {
	//runArgs := os.Args
	storage.CreateDatabase()
	server.StartServer()
	// if runArgs != nil && len(runArgs) > 1 {
	// 	if runArgs[1] == "server" {
	// 		server.StartServer()
	// 	}
	// }

	// for {
	// 	fmt.Println("\nTODO App")
	// 	fmt.Println("1. List Tasks")
	// 	fmt.Println("2. Add Task")
	// 	fmt.Println("3. Complete Task")
	// 	fmt.Println("4. Incomplete Task")
	// 	fmt.Println("5. Delete Task")
	// 	fmt.Println("7. List Next ID")
	// 	fmt.Println("8. Delete All Tasks")
	// 	fmt.Println("9. Exit Program")
	// 	fmt.Print("Enter your choice: ")
	//
	// 	var choice int
	// 	_, err := fmt.Scanln(&choice)
	// 	if err != nil {
	// 		fmt.Print("Invalid choice\n")
	// 		continue
	// 	}
	//
	// 	switch choice {
	// 	case 1:
	// 		tasks.ListTasks()
	// 	case 2:
	// 		tasks.AddTask()
	// 	case 3:
	// 		tasks.MarkTaskComplete()
	// 	case 4:
	// 		tasks.MarkTaskIncomplete()
	// 	case 5:
	// 		tasks.DeleteTask()
	// 	case 7:
	// 		nextId := storage.GetNextID()
	// 		fmt.Printf("\nNext ID: %d\n\n", nextId)
	// 	case 8:
	// 		storage.DeleteAllTasks()
	// 	case 9:
	// 		fmt.Println("\nSee you next time!")
	// 		os.Exit(0)
	// 	default:
	// 		fmt.Println("Invalid choice")
	// 	}
	// }
}
