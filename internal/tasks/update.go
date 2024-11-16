package tasks

import "fmt"

func MarkTaskComplete(taskList *[]Task) {
	fmt.Print("Enter task ID to mark as complete: ")
	var id int
	_, err := fmt.Scan(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
 
	for i, task := range *taskList {
		if task.ID == id {
			(*taskList)[i].Completed = true
			fmt.Println("Task marked as complete")
			return
		}
	}
 
	fmt.Println("Task not found")
}

func MarkTaskIncomplete(taskList *[]Task) {
	fmt.Print("Enter task ID to mark as incomplete: ")
	var id int
	_, err := fmt.Scan(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
 
	for i, task := range *taskList {
		if task.ID == id {
			(*taskList)[i].Completed = false
			fmt.Println("Task marked as incomplete")
			return
		}
	}
 
	fmt.Println("Task not found")
}
