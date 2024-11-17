package tasks

import "fmt"

func ListTasks(taskList []Task) {
	if len(taskList) == 0 {
		fmt.Println("\nNo tasks\n")
		return
	}

	fmt.Println("\nTasks:")
	for i, task := range taskList {
		status := "Incomplete"
		if task.Completed { 
			status = "Complete"
		}
		if i == len(taskList)-1 {
			fmt.Printf("%d. %s: %s (%s)\n\n", task.ID, task.Title, task.Description, status)
		} else {
			fmt.Printf("%d. %s: %s (%s)\n", task.ID, task.Title, task.Description, status)
		}
	}
}
