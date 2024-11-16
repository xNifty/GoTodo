package tasks

import "fmt"

func ListTasks(taskList []Task) {
	if len(taskList) == 0 {
		fmt.Println("No tasks")
		return
	}

	fmt.Println("\nTasks:")
	for _, task := range taskList {
		status := "Incomplete"
		if task.Completed { 
			status = "Complete"
		}
		fmt.Printf("%d. %s: %s (%s)\n", task.ID, task.Title, task.Description, status)
	}
}
