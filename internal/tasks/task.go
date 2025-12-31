package tasks

import (
	"fmt"
	"sync"
)

type Task struct {
	ID          int
	Title       string
	Description string
	Completed   bool
	DateAdded   string
	Page        int
	IsFavorite  bool
	Position    int
}

type TaskManager struct {
	tasks map[int]Task
	mutex sync.Mutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[int]Task),
	}
}

func (tm *TaskManager) GetTasks() []Task {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tasks := make([]Task, 0, len(tm.tasks))

	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func (t *Task) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	return nil
}
