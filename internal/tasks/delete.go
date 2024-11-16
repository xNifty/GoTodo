package tasks

import "fmt"

func DeleteTask(taskList *[]Task) {
	fmt.Print("Enter task ID to delete: ")
	var id int
	_, err := fmt.Scan(&id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	for i, task := range *taskList {
		if task.ID == id {
			*taskList = append((*taskList)[:i], (*taskList)[i+1:]...)
			fmt.Println("Task deleted")
			return
		}
	}

	fmt.Println("Task not found")
}
